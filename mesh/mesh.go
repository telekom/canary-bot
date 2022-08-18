package mesh

import (
	"canary-bot/data"
	h "canary-bot/helper"
	meshv1 "canary-bot/proto/gen/mesh/v1"
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

	// Push probes
	PushProbeInterval    time.Duration
	PushProbeToAmount    int
	PushProbeRetryAmount int
	PushProbeRetryDelay  time.Duration

	// User settings from flags and env vars
	StartupSettings Settings
}

type Settings struct {
	// remote target
	Targets []string

	// local config
	Name          string
	ListenAddress string
	ListenPort    int64

	// TLS server side
	ServerCertPath string
	ServerKeyPath  string
	ServerCert     []byte
	ServerKey      []byte
	// TLS client side
	CaCertPath string
	CaCert     []byte

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
	}
	m.log.Info("Starting Mesh")

	// get IP of this node if no listen-address set
	if conf.StartupSettings.ListenAddress == "" {
		m.log.Info("Address flag not set - getting external IP automatically")
		var err error
		conf.StartupSettings.ListenAddress, err = h.ExternalIP()
		if err != nil {
			m.log.Fatalln("Could not get external IP, please use address flag")
		}

	}

	go func() {
		err := m.StartServer()
		if err != nil {
			m.log.Debugf("Mesh server error: %+v", err)
			m.log.Fatal("Could not start Mesh Server")
		}
	}()

	go m.channelRoutines()

	go m.timerRoutines()

	return m, nil
}

func (m *Mesh) timerRoutines() {
	// Timer to send ping to node
	pingTicker := time.NewTicker(m.config.PingInterval)
	// Timer to send probes to node
	pushProbeTicker := time.NewTicker(m.config.PushProbeInterval)

	// Not used
	quit := make(chan struct{})

	for {
		select {
		case <-pingTicker.C:
			m.log.Debug("Starting Ping Routine")
			nodes := m.database.GetNodeListByState(NODE_OK)
			if nodes == nil {
				m.log.Debug("No Node connected or all nodes in timeout")
				break
			}

			go m.RetryPing(context.Background(), nodes[rand.Intn(len(nodes))].Convert(), m.config.PingRetryAmount, m.config.PingRetryDelay, false)

		case <-pushProbeTicker.C:
			m.log.Debugf("Starting push probe routine to %v random nodes", m.config.PushProbeToAmount)
			nodes := m.database.GetNodeListByState(NODE_OK)
			if nodes == nil {
				m.log.Debug("No Node connected or all nodes in timeout")
				break
			}

			for broadcastCount := 0; broadcastCount < m.config.PushProbeToAmount; broadcastCount++ {
				if len(nodes) <= 0 {
					m.log.Debug("Stopping routine prematurely - no more known nodes")
					break
				}
				randomIndex := rand.Intn(len(nodes))
				randomNode := nodes[randomIndex]

				m.log.Infof("Pushing probes to %+v", randomNode.Name)
				go m.RetryPushProbe(context.Background(), randomNode.Convert(), m.config.PushProbeRetryAmount, m.config.PushProbeRetryDelay)

				// Remove node already started broadcast to from list
				nodes[randomIndex] = nodes[len(nodes)-1]
				nodes = nodes[:len(nodes)-1]

			}

		case <-quit:
			// Not used
			pingTicker.Stop()
			pushProbeTicker.Stop()
			return
		}
	}
}

func (m *Mesh) channelRoutines() {
	for {
		select {
		case nodeDiscovered := <-m.newNodeDiscovered:
			if m.database.GetNode(GetId(nodeDiscovered.NewNode)).Id != 0 {
				m.log.Debug("Node joined - already known")
				m.database.SetNode(data.Convert(nodeDiscovered.NewNode, NODE_OK))
				break
			}
			m.log.Debug("Node joined - new node")

			m.log.Debugf("Starting discovery broadcast routine to %v random nodes", m.config.BroadcastToAmount)
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
					m.log.Debug("Stopping routine prematurely - no more known nodes")
					break
				}
				randomIndex := rand.Intn(len(nodes))
				randomNode := nodes[randomIndex]

				m.log.Infof("Sending Discovery Broadcast to %+v", randomNode)
				go m.NodeDiscovery(randomNode.Convert(), nodeDiscovered.NewNode)

				// Remove node already started broadcast to from list
				nodes[randomIndex] = nodes[len(nodes)-1]
				nodes = nodes[:len(nodes)-1]

			}

			m.database.SetNode(data.Convert(nodeDiscovered.NewNode, NODE_OK))
		case node := <-m.timeoutNode:
			m.database.SetNode(data.Convert(node, NODE_TIMEOUT_RETRY))
			m.log.Debugf("Start Timeout Retry Routine in %v", m.config.TimeoutRetryPause.String())
			go func() {
				time.Sleep(m.config.TimeoutRetryPause)
				m.RetryPing(context.Background(), node, m.config.TimeoutRetryAmount, m.config.TimeoutRetryDelay, true)
			}()
		}
	}
}

func (m *Mesh) RetryPing(ctx context.Context, node *meshv1.Node, retries int, delay time.Duration, timedOut bool) {
	m.log.Infof("Ping retry routine started for: %+v", node)
	for r := 0; ; r++ {
		rtt, err := m.Ping(node)

		m.database.SetProbe(
			&data.Probe{
				From:  m.config.StartupSettings.Name,
				To:    node.Name,
				Key:   data.RTT,
				Value: strconv.FormatInt(rtt.Nanoseconds(), 10),
				Ts:    time.Now().Unix(),
			},
		)

		if err == nil || r >= retries {
			if err != nil {
				m.log.Warnf("Ping retry timeout - limit (%+v) reached for %+v", retries, node)
				if !timedOut {
					m.database.SetNode(data.Convert(node, NODE_TIMEOUT))
					m.timeoutNode <- node
				} else {
					m.database.DeleteNode(GetId(node))
					m.log.Warnf("Removed node from mesh")
				}
			} else {
				m.log.Info("Ping ok")
			}
			break
		}

		if !timedOut {
			m.database.SetNode(data.Convert(node, NODE_RETRY))
		}
		m.log.Debugf("... retrying in %v for %+v", delay, m.database.GetNode(GetId(node)))
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			m.log.Warnf("Ping retry context error: %+v", ctx.Err())
		}
	}
}

func (m *Mesh) RetryPushProbe(ctx context.Context, node *meshv1.Node, retries int, delay time.Duration) {
	m.log.Infof("Push probe retry routine started to %+v", node.Name)
	for r := 0; ; r++ {
		err := m.PushProbes(node)

		if err == nil || r >= retries {
			if err != nil {
				m.log.Warnf("Push probe retry timeout - limit (%+v) reached for %+v", retries, node)
			} else {
				m.log.Debug("Push probe routine ok")
			}
			m.database.SetNodeTsNow(GetId(node))
			break
		}

		m.log.Debugf("... retrying in %v to %+v", delay, node.Name)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			m.log.Warnf("Push probe retry context error: %+v", ctx.Err())
		}
	}
}

func GetId(n *meshv1.Node) uint32 {
	return h.Hash(n.Target)
}

func GetProbeId(p *meshv1.Probe) uint32 {
	return h.Hash(p.From + p.To + strconv.FormatInt(p.Key, 10))
}
