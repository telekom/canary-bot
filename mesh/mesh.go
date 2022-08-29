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

	m.log.Infow("Starting routines")

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
			log := m.log.Named("ping-routine")
			log.Debugw("Starting")
			nodes := m.database.GetNodeListByState(NODE_OK)
			if nodes == nil {
				log.Debugw("No Node connected or all nodes in timeout")
				break
			}

			go m.RetryPing(context.Background(), nodes[rand.Intn(len(nodes))].Convert(), m.config.PingRetryAmount, m.config.PingRetryDelay, false)

		case <-pushProbeTicker.C:
			log := m.log.Named("probe-routine")
			log.Debugw("Starting push probe routine to random nodes", "amount", m.config.PushProbeToAmount)
			nodes := m.database.GetNodeListByState(NODE_OK)
			if nodes == nil {
				log.Debugw("No Node connected or all nodes in timeout")
				break
			}

			for broadcastCount := 0; broadcastCount < m.config.PushProbeToAmount; broadcastCount++ {
				if len(nodes) <= 0 {
					log.Debug("Stopping routine prematurely - no more known nodes")
					break
				}
				randomIndex := rand.Intn(len(nodes))
				randomNode := nodes[randomIndex]

				log.Debugw("Pushing probes", "node", randomNode.Name)
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
			log := m.log.Named("discovery-routine")
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
				log.Warnw("Retry timeout - limit reached", "node", node.Name, "limit", retries)
				if !timedOut {
					m.database.SetNode(data.Convert(node, NODE_TIMEOUT))
					m.timeoutNode <- node
				} else {
					m.database.DeleteNode(GetId(node))
					log.Warnw("Removed node from mesh", "node", node.Name)
				}
			} else {
				log.Infow("Ping ok", "node", node.Name)
			}
			break
		}

		if !timedOut {
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

func (m *Mesh) RetryPushProbe(ctx context.Context, node *meshv1.Node, retries int, delay time.Duration) {
	log := m.log.Named("probe-routine")
	log.Debugw("Push probe retry routine started", "node", node.Name)
	for r := 0; ; r++ {
		err := m.PushProbes(node)

		if err == nil || r >= retries {
			if err != nil {
				log.Debugw("Push probe retry timeout - limit reached", "node", node.Name, "limit", retries)
			} else {
				log.Debug("Push probes ok")
			}
			m.database.SetNodeTsNow(GetId(node))
			break
		}

		log.Debugw("Retrying", "node", node.Name, "delay", delay)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			log.Warnw("Push probe retry context error", "error", ctx.Err())
		}
	}
}

func GetId(n *meshv1.Node) uint32 {
	return h.Hash(n.Target)
}

func GetProbeId(p *meshv1.Probe) uint32 {
	return h.Hash(p.From + p.To + strconv.FormatInt(p.Key, 10))
}
