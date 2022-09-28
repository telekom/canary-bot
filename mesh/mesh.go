package mesh

import (
	"canary-bot/data"
	h "canary-bot/helper"
	meshv1 "canary-bot/proto/mesh/v1"
	"context"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Mesh struct {
	database data.Database
	log      *zap.SugaredLogger
	config   *Config

	mu      sync.Mutex
	clients map[uint32]*MeshClient

	newNodeDiscovered chan NodeDiscovered
	timeoutNode       chan *meshv1.Node

	pingTicker        *time.Ticker
	pushSampleTicker  *time.Ticker
	cleanSampleTicker *time.Ticker

	rttTicker *time.Ticker

	quitJoinRoutine    chan bool
	restartJoinRoutine chan bool
	joinRoutineDone    bool
}

type Config struct {
	// Timeout for every grpc request
	RequestTimeout time.Duration

	// Join config
	JoinInterval time.Duration

	// Ping config
	PingInterval       time.Duration
	PingRetryAmount    int
	PingRetryDelay     time.Duration
	TimeoutRetryPause  time.Duration
	TimeoutRetryAmount int
	TimeoutRetryDelay  time.Duration

	// Node discovery
	BroadcastToAmount int

	// Push samples
	PushSampleInterval    time.Duration
	PushSampleToAmount    int
	PushSampleRetryAmount int
	PushSampleRetryDelay  time.Duration

	// Clean samples
	CleanSampleInterval time.Duration
	SampleMaxAge        time.Duration

	// Sample: RTT
	RttInterval time.Duration

	// User settings from flags and env vars
	StartupSettings Settings
}

type Settings struct {
	// remote target
	Targets []string

	// local config
	Name          string
	JoinAddress   string
	ListenAddress string
	ListenPort    int64

	// API
	ApiPort int64

	// TLS server side
	ServerCertPath string
	ServerKeyPath  string
	ServerCert     []byte
	ServerKey      []byte
	// TLS client side
	CaCertPath []string
	CaCert     []byte

	//Auth API
	Tokens []string

	//Logging
	Debug     bool
	DebugGrpc bool
}

type NodeDiscovered struct {
	NewNode *meshv1.Node
	From    uint32 // TODO change to name
}

func NewMesh(db data.Database, conf *Config, logger *zap.SugaredLogger) (*Mesh, error) {
	defer logger.Sync()

	m := &Mesh{
		database:           db,
		log:                logger,
		config:             conf,
		clients:            map[uint32]*MeshClient{},
		newNodeDiscovered:  make(chan NodeDiscovered),
		timeoutNode:        make(chan *meshv1.Node),
		quitJoinRoutine:    make(chan bool, 1),
		restartJoinRoutine: make(chan bool, 1),
		joinRoutineDone:    false,
	}
	m.log.Info("Starting mesh")

	go func() {
		m.log.Info("Starting server, listening vor joining nodes")
		err := m.StartServer()
		if err != nil {
			m.log.Debugf("Mesh server error: %+v", err)
			m.log.Fatal("Could not start Mesh Server")
		}
	}()

	m.log.Infow("Starting mesh routines")
	go m.channelRoutines()
	go m.timerRoutines()

	return m, nil
}

func (m *Mesh) timerRoutines() {

	// Timer to send ping to node
	joinTicker := time.NewTicker(m.config.JoinInterval)
	// Timer to send ping to node
	m.pingTicker = time.NewTicker(m.config.PingInterval)
	m.pingTicker.Stop()
	// Timer to send samples to node
	m.pushSampleTicker = time.NewTicker(m.config.PushSampleInterval)
	m.pushSampleTicker.Stop()
	// Timer to clean sampels from removed nodes
	m.cleanSampleTicker = time.NewTicker(m.config.CleanSampleInterval)
	m.cleanSampleTicker.Stop()

	// Sample: RTT
	m.rttTicker = time.NewTicker(m.config.RttInterval)
	m.rttTicker.Stop()

	for {
		select {
		case <-joinTicker.C:
			log := m.log.Named("join-routine")
			// join (future) mesh
			log.Infow("Waiting for a node to join a mesh...")
			connected, isNameUniqueInMesh := m.Join(m.config.StartupSettings.Targets)
			if !isNameUniqueInMesh {
				log.Fatal("The name is not unique in the mesh, please choose another one.")
				// TODO generate random node name?
			}
			if connected {
				log.Infow("Connected to a mesh")
				m.quitJoinRoutine <- true
			}

		case <-m.pingTicker.C:
			log := m.log.Named("ping-routine")
			log.Debugw("Starting")
			nodes := m.database.GetNodeListByState(NODE_OK)
			if nodes == nil {
				log.Debugw("No Node connected or all nodes in timeout")
				break
			}

			go m.RetryPing(context.Background(), nodes[rand.Intn(len(nodes))].Convert(), m.config.PingRetryAmount, m.config.PingRetryDelay, false)

		case <-m.pushSampleTicker.C:
			log := m.log.Named("sample-routine")
			log.Debugw("Starting push sample routine to random nodes", "amount", m.config.PushSampleToAmount)
			nodes := m.database.GetNodeListByState(NODE_OK)
			if nodes == nil {
				log.Debugw("No Node connected or all nodes in timeout")
				break
			}

			for broadcastCount := 0; broadcastCount < m.config.PushSampleToAmount; broadcastCount++ {
				if len(nodes) <= 0 {
					log.Debug("Stopping routine prematurely - no more known nodes")
					break
				}
				randomIndex := rand.Intn(len(nodes))
				randomNode := nodes[randomIndex]

				log.Debugw("Pushing samples", "node", randomNode.Name)
				go m.RetryPushSample(context.Background(), randomNode.Convert(), m.config.PushSampleRetryAmount, m.config.PushSampleRetryDelay)

				// Remove node already started broadcast to from list
				nodes[randomIndex] = nodes[len(nodes)-1]
				nodes = nodes[:len(nodes)-1]

			}
		case <-m.cleanSampleTicker.C:
			// check if sample is too old and delete
			for _, sample := range m.database.GetSampleList() {
				if time.Unix(sample.Ts, 0).Before(time.Now().Add(-1 * m.config.SampleMaxAge)) {
					m.log.Infow("Delete old sample", "from", sample.From, "to", sample.To, "key", data.SampleName[sample.Key], "maxAge", m.config.SampleMaxAge.String())
					m.database.DeleteSample(sample.Id)
				}
			}

		case <-m.rttTicker.C:
			go m.Rtt()

		case <-m.restartJoinRoutine:
			joinTicker.Reset(m.config.JoinInterval)
			m.joinRoutineDone = false
			// starting ticker after join routine
			m.pingTicker.Stop()
			m.pushSampleTicker.Stop()
			m.cleanSampleTicker.Stop()
			m.rttTicker.Stop()
			m.log.Debug("Start joinRoutine again, stopping all timer routines")
		case <-m.quitJoinRoutine:
			joinTicker.Stop()
			m.joinRoutineDone = true
			// starting ticker after join routine
			m.pingTicker.Reset(m.config.PingInterval)
			m.pushSampleTicker.Reset(m.config.PushSampleInterval)
			m.cleanSampleTicker.Reset(m.config.CleanSampleInterval)
			m.rttTicker.Reset(m.config.RttInterval)
			m.log.Info("Starting pings")
			m.log.Debug("Stop joinRoutine, starting all timer routines")
		}
	}
}

func (m *Mesh) channelRoutines() {
	for {
		select {
		case nodeDiscovered := <-m.newNodeDiscovered:
			log := m.log.Named("discovery-routine")
			// quit joinMesh routine if discovery is received before
			if !m.joinRoutineDone {
				m.quitJoinRoutine <- true
			}
			if m.database.GetNodeByName(nodeDiscovered.NewNode.Name).Id != 0 {
				log.Info("Node is rejoining node")
				m.database.SetNode(data.Convert(nodeDiscovered.NewNode, NODE_OK))
				break
			}
			log.Info("Node joined - new node")

			log.Debugw("Starting discovery broadcast routine to random nodes", "amount", m.config.BroadcastToAmount)
			nodes := m.database.GetNodeListByState(NODE_OK)

			// Remove node from list that sent the discovery request
			for i, node := range nodes {
				if node.Id == nodeDiscovered.From {
					nodes[i] = nodes[len(nodes)-1]
					nodes = nodes[:len(nodes)-1]
				}
			}

			for broadcastCount := 0; broadcastCount < m.config.BroadcastToAmount; broadcastCount++ {
				if len(nodes) <= 0 {
					log.Debug("Stopping routine prematurely - no more known nodes")
					break
				}
				randomIndex := rand.Intn(len(nodes))
				randomNode := nodes[randomIndex]

				log.Infow("Sending Discovery Broadcast", "node", randomNode.Name)
				go m.NodeDiscovery(randomNode.Convert(), nodeDiscovered.NewNode)

				// Remove node already started broadcast to from list
				nodes[randomIndex] = nodes[len(nodes)-1]
				nodes = nodes[:len(nodes)-1]

			}

			m.database.SetNode(data.Convert(nodeDiscovered.NewNode, NODE_OK))
		case node := <-m.timeoutNode:
			log := m.log.Named("ping-routine")
			m.database.SetNode(data.Convert(node, NODE_TIMEOUT_RETRY))
			log.Debugw("Start Timeout Retry Routine", "delay", m.config.TimeoutRetryPause.String())
			go func() {
				time.Sleep(m.config.TimeoutRetryPause)
				m.RetryPing(context.Background(), node, m.config.TimeoutRetryAmount, m.config.TimeoutRetryDelay, true)
			}()
		}
	}
}

func (m *Mesh) RetryPing(ctx context.Context, node *meshv1.Node, retries int, delay time.Duration, timedOut bool) {
	// TODO refactor retries & delay -> directly from class
	// TODO DB Get node, anstelle timeout bool
	log := m.log.Named("ping-routine")
	log.Debugw("Retry routine started", "node", node.Name)
	for r := 0; ; r++ {
		err := m.Ping(node)

		if err == nil || r >= retries {
			if err != nil {
				log.Warnw("Retry timeout - limit reached", "node", node.Name, "limit", retries)
				if !timedOut {
					m.database.SetNode(data.Convert(node, NODE_TIMEOUT))
					m.timeoutNode <- node
				} else {
					m.database.DeleteNode(GetId(node))
					if len(m.database.GetNodeList()) == 0 {
						m.restartJoinRoutine <- true
					}
					log.Warnw("Removed node from mesh", "node", node.Name)
				}
			} else {
				m.database.SetNode(data.Convert(node, NODE_OK))
				log.Infow("Ping ok", "node", node.Name)
			}
			break
		}

		if !timedOut {
			log.Infow("Ping failed", "node", node.Name, "timeout", m.config.RequestTimeout)
			m.database.SetNode(data.Convert(node, NODE_RETRY))
		}
		log.Debugw("Retrying", "node", m.database.GetNode(GetId(node)).Name, "delay", delay)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			log.Warnw("Context error", "error", ctx.Err().Error())
		}
	}
}

func (m *Mesh) RetryPushSample(ctx context.Context, node *meshv1.Node, retries int, delay time.Duration) {
	log := m.log.Named("sample-routine")
	log.Debugw("Push sample retry routine started", "node", node.Name)
	for r := 0; ; r++ {
		err := m.PushSamples(node)

		if err == nil || r >= retries {
			if err != nil {
				log.Debugw("Push sample retry timeout - limit reached", "node", node.Name, "limit", retries)
			} else {
				log.Debug("Push samples ok")
			}
			m.database.SetNodeTsNow(GetId(node))
			break
		}

		log.Debugw("Retrying", "node", node.Name, "delay", delay)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			log.Warnw("Push sample retry context error", "error", ctx.Err())
		}
	}
}

func GetId(n *meshv1.Node) uint32 {
	return h.Hash(n.Target)
}

func GetSampleId(p *meshv1.Sample) uint32 {
	return h.Hash(p.From + p.To + strconv.FormatInt(p.Key, 10))
}
