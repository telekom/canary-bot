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

	pingTicker       *time.Ticker
	pushSampleTicker *time.Ticker

	quitJoinRoutine chan bool
	joinRoutineDone bool
}

type Config struct {
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

	// User settings from flags and env vars
	StartupSettings Settings
}

type Settings struct {
	EnvPrefix string
	// remote target
	Targets []string

	// local config
	Name          string
	Domain        string
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
	CaCertPath string
	CaCert     []byte

	//Auth API
	Tokens []string

	//Logging
	Debug bool
}

type NodeDiscovered struct {
	NewNode *meshv1.Node
	From    uint32 // TODO change to name
}

func NewMesh(db data.Database, conf *Config, logger *zap.SugaredLogger) (*Mesh, error) {
	defer logger.Sync()

	m := &Mesh{
		database:          db,
		log:               logger,
		config:            conf,
		clients:           map[uint32]*MeshClient{},
		newNodeDiscovered: make(chan NodeDiscovered),
		timeoutNode:       make(chan *meshv1.Node),
		quitJoinRoutine:   make(chan bool, 1),
		joinRoutineDone:   false,
	}
	m.log.Info("Starting Mesh")

	go func() {
		err := m.StartServer()
		if err != nil {
			m.log.Debugf("Mesh server error: %+v", err)
			m.log.Fatal("Could not start Mesh Server")
		}
	}()

	m.log.Infow("Starting channel routines")
	go m.channelRoutines()

	m.log.Infow("Starting timer routines")

	go m.timerRoutines()

	return m, nil
}

func (m *Mesh) timerRoutines() {

	// Timer to send ping to node
	joinTicker := time.NewTicker(time.Second * 5)
	// Timer to send ping to node
	m.pingTicker = time.NewTicker(m.config.PingInterval)
	m.pingTicker.Stop()
	// Timer to send samples to node
	m.pushSampleTicker = time.NewTicker(m.config.PushSampleInterval)
	m.pushSampleTicker.Stop()
	// Not used
	quit := make(chan bool)

	for {
		select {
		case <-joinTicker.C:
			log := m.log.Named("join-routine")
			// join (future) mesh
			log.Infow("Waiting for a node for joining ...")
			connected, isNameUniqueInMesh := m.Join(m.config.StartupSettings.Targets)
			if !isNameUniqueInMesh {
				log.Fatal("The name is not unique in the mesh, please choose another one.")
				// TODO generate random node name?
			}
			if connected {
				log.Debug("Join DONE - quit join routine")
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

		case <-quit:
			// Not used
			m.pingTicker.Stop()
			m.pushSampleTicker.Stop()
			return
		case <-m.quitJoinRoutine:
			joinTicker.Stop()
			m.joinRoutineDone = true
			m.pingTicker = time.NewTicker(m.config.PingInterval)
			m.pushSampleTicker = time.NewTicker(m.config.PushSampleInterval)
			m.log.Debug("Quit joinRoutine - ok ... TICKER started")
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
				log.Debug("Quit joinRoutine")
				m.quitJoinRoutine <- true
			}
			if m.database.GetNodeByName(nodeDiscovered.NewNode.Name).Id != 0 {
				log.Info("Node is rejoining node")
				m.database.SetNode(data.Convert(nodeDiscovered.NewNode, NODE_OK))
				break
			}
			log.Debug("Node joined - new node")

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
	log := m.log.Named("ping-routine")
	log.Debugw("Retry routine started", "node", node.Name)
	for r := 0; ; r++ {
		rtt, err := m.Ping(node)

		m.database.SetSample(
			&data.Sample{
				From:  m.config.StartupSettings.Name,
				To:    node.Name,
				Key:   data.RTT,
				Value: strconv.FormatInt(rtt.Nanoseconds(), 10),
				Ts:    time.Now().Unix(),
			},
		)

		if err == nil || r >= retries {
			if err != nil {
				log.Warnw("Retry timeout - limit reached", "node", node.Name, "limit", retries)
				if !timedOut {
					m.database.SetNode(data.Convert(node, NODE_TIMEOUT))
					m.timeoutNode <- node
				} else {
					m.database.DeleteNode(GetId(node))
					log.Warnw("Removed node from mesh", "node", node.Name)
				}
			} else {
				m.database.SetNode(data.Convert(node, NODE_OK))
				log.Infow("Ping ok", "node", node.Name)
			}
			break
		}

		if !timedOut {
			log.Infow("Ping failed", "node", node.Name)
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
