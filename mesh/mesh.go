package mesh

import (
	"canary-bot/api"
	"canary-bot/data"
	h "canary-bot/helper"
	meshv1 "canary-bot/proto/mesh/v1"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Mesh struct {
	database      data.Database
	logger        *zap.SugaredLogger
	routineConfig *RoutineConfiguration
	setupConfig   *SetupConfiguration

	mu      sync.Mutex
	clients map[uint32]*MeshClient

	newNodeDiscovered chan NodeDiscovered

	pingTicker        *time.Ticker
	pushSampleTicker  *time.Ticker
	cleanSampleTicker *time.Ticker

	rttTicker *time.Ticker

	quitJoinRoutine    chan bool
	restartJoinRoutine chan bool
	joinRoutineDone    bool
}

type NodeDiscovered struct {
	NewNode *meshv1.Node
	From    uint32 // TODO change to name
}

func CreateCanaryMesh(routineConfig *RoutineConfiguration, setupConfig *SetupConfiguration) {
	// prepare logging
	logger := getLogger(setupConfig.Debug, setupConfig.ListenAddress)
	defer logger.Sync()

	// SetupConfiguration
	logger.Debugf("CLI settings: %+v", setupConfig)

	// Try to set missing setupConfigurations
	setupConfig.setDefaults(logger)
	// Get info from configuration combination
	setupConfig.checkDefaults(logger)

	// prepare in-memory database
	database, err := data.NewMemDB(logger.Named("database"))
	if err != nil {
		logger.Fatalf("Could not create Memory Database (MemDB) - Error: %+v", err)
	}

	m := &Mesh{
		database:           database,
		logger:             logger,
		routineConfig:      routineConfig,
		setupConfig:        setupConfig,
		clients:            map[uint32]*MeshClient{},
		newNodeDiscovered:  make(chan NodeDiscovered),
		quitJoinRoutine:    make(chan bool, 1),
		restartJoinRoutine: make(chan bool, 1),
		joinRoutineDone:    false,
	}
	logger.Info("Starting mesh")

	go func() {
		logger.Info("Starting server, listening vor joining nodes")
		err := m.StartServer()
		if err != nil {
			logger.Debugf("Mesh server error: %+v", err)
			logger.Fatal("Could not start Mesh Server")
		}
	}()

	logger.Infow("Starting mesh routines")
	go m.channelRoutines()
	go m.timerRoutines()

	// start API
	apiConfig := &api.Configuration{
		Address:        setupConfig.ListenAddress,
		Port:           setupConfig.ApiPort,
		Tokens:         setupConfig.Tokens,
		DebugGrpc:      setupConfig.DebugGrpc,
		ServerCertPath: setupConfig.ServerCertPath,
		ServerKeyPath:  setupConfig.ServerKeyPath,
		ServerCert:     setupConfig.ServerCert,
		ServerKey:      setupConfig.ServerKey,
		CaCertPath:     setupConfig.CaCertPath,
		CaCert:         setupConfig.CaCert,
	}
	if err = api.StartApi(database, apiConfig, logger.Named("api")); err != nil {
		logger.Fatal("Could not start API - Error: %+v", err)
	}
}

func (m *Mesh) timerRoutines() {

	// Timer to send ping to node
	joinTicker := time.NewTicker(m.routineConfig.JoinInterval)
	// Timer to send ping to node
	m.pingTicker = time.NewTicker(m.routineConfig.PingInterval)
	m.pingTicker.Stop()
	// Timer to send samples to node
	m.pushSampleTicker = time.NewTicker(m.routineConfig.PushSampleInterval)
	m.pushSampleTicker.Stop()
	// Timer to clean sampels from removed nodes
	m.cleanSampleTicker = time.NewTicker(m.routineConfig.CleanSampleInterval)
	m.cleanSampleTicker.Stop()

	// Sample: RTT
	m.rttTicker = time.NewTicker(m.routineConfig.RttInterval)
	m.rttTicker.Stop()

	for {
		select {
		case <-joinTicker.C:
			log := m.logger.Named("join-routine")
			// join (future) mesh
			log.Infow("Waiting for a node to join a mesh...")
			connected, isNameUniqueInMesh := m.Join(m.setupConfig.Targets)
			if !isNameUniqueInMesh {
				log.Fatal("The name is not unique in the mesh, please choose another one.")
				// TODO generate random node name?
			}
			if connected {
				log.Infow("Connected to a mesh")
				m.quitJoinRoutine <- true
			}

		case <-m.pingTicker.C:
			log := m.logger.Named("ping-routine")
			log.Debugw("Starting")
			nodes := m.database.GetNodeListByState(NODE_OK)
			if nodes == nil {
				log.Debugw("No Node connected or all nodes in timeout")
				break
			}

			// TODO GetRandomNodeListByState(NODE_OK, amountOfNodes) -> wie bei m.pushSampleTicker?
			go m.retryPing(nodes[rand.Intn(len(nodes))].Convert())

		case <-m.pushSampleTicker.C:
			log := m.logger.Named("sample-routine")
			log.Debugw("Starting push sample routine to random nodes", "amount", m.routineConfig.PushSampleToAmount)

			nodes := m.database.GetRandomNodeListByState(NODE_OK, m.routineConfig.PushSampleToAmount)
			if len(nodes) == 0 {
				log.Debugw("No node connected or all nodes in timeout")
				break
			}

			for _, node := range nodes {
				log.Debugw("Pushing samples", "node", node.Name)
				go m.retryPushSample(node.Convert())
			}

		case <-m.cleanSampleTicker.C:
			// check if sample is too old and delete
			for _, sample := range m.database.GetSampleList() {
				if time.Unix(sample.Ts, 0).Before(time.Now().Add(-1 * m.routineConfig.SampleMaxAge)) {
					m.logger.Infow("Delete old sample", "from", sample.From, "to", sample.To, "key", data.SampleName[sample.Key], "maxAge", m.routineConfig.SampleMaxAge.String())
					m.database.DeleteSample(sample.Id)
				}
			}

		case <-m.rttTicker.C:
			go m.Rtt()

		case <-m.restartJoinRoutine:
			joinTicker.Reset(m.routineConfig.JoinInterval)
			m.joinRoutineDone = false
			// starting ticker after join routine
			m.pingTicker.Stop()
			m.pushSampleTicker.Stop()
			m.cleanSampleTicker.Stop()
			m.rttTicker.Stop()
			m.logger.Debug("Start joinRoutine again, stopping all timer routines")
		case <-m.quitJoinRoutine:
			joinTicker.Stop()
			m.joinRoutineDone = true
			// starting ticker after join routine
			m.pingTicker.Reset(m.routineConfig.PingInterval)
			m.pushSampleTicker.Reset(m.routineConfig.PushSampleInterval)
			m.cleanSampleTicker.Reset(m.routineConfig.CleanSampleInterval)
			m.rttTicker.Reset(m.routineConfig.RttInterval)
			m.logger.Info("Starting pings")
			m.logger.Debug("Stop joinRoutine, starting all timer routines")
		}
	}
}

func (m *Mesh) channelRoutines() {
	for {
		select {
		case nodeDiscovered := <-m.newNodeDiscovered:
			log := m.logger.Named("discovery-routine")
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
			log.Debugw("Starting discovery broadcast routine to random nodes", "amount", m.routineConfig.BroadcastToAmount)

			nodes := m.database.GetRandomNodeListByState(NODE_OK, m.routineConfig.BroadcastToAmount, nodeDiscovered.From)

			if len(nodes) == 0 {
				log.Debug("Stopping routine prematurely - no more known nodes")
				break
			}

			for _, node := range nodes {
				log.Infow("Sending Discovery Broadcast", "node", node.Name)
				go m.NodeDiscovery(node.Convert(), nodeDiscovered.NewNode)
			}

			m.database.SetNode(data.Convert(nodeDiscovered.NewNode, NODE_OK))
		}
	}
}

func (m *Mesh) retryPing(node *meshv1.Node) {
	log := m.logger.Named("ping-routine")
	log.Debugw("Retry routine started", "node", node.Name)

	for r := 1; r <= m.routineConfig.PingRetryAmount; r++ {
		err := m.ping(node)

		// Ping ok
		if err == nil {
			m.database.SetNode(data.Convert(node, NODE_OK))
			log.Infow("Ping ok", "node", node.Name, "attempt", r)
			return
		}

		// Ping failed
		log.Infow("Ping failed", "node", node.Name, "timeout", m.routineConfig.RequestTimeout.String(), "retry in", m.routineConfig.PingRetryDelay.String(), "attempt", r)
		m.database.SetNode(data.Convert(node, NODE_TIMEOUT))
		m.database.SetSampleNaN(GetSampleId(&meshv1.Sample{From: m.setupConfig.Name, To: node.Name, Key: data.RTT_REQUEST}))
		m.database.SetSampleNaN(GetSampleId(&meshv1.Sample{From: m.setupConfig.Name, To: node.Name, Key: data.RTT_TOTAL}))

		if r != m.routineConfig.PingRetryAmount {
			// Retry delay
			time.Sleep(m.routineConfig.PingRetryDelay)
		}
	}

	// Retry limit reached
	log.Infow("Retry limit reached", "node", node.Name, "limit", m.routineConfig.PingRetryAmount)
	log.Warnw("Removing node from mesh", "node", node.Name)
	m.database.DeleteNode(GetId(node))

	// Check if node was last node in mesh
	if len(m.database.GetNodeList()) == 0 {
		m.restartJoinRoutine <- true
	}
}

func (m *Mesh) retryPushSample(node *meshv1.Node) {
	log := m.logger.Named("sample-routine")
	log.Debugw("Push sample retry routine started", "node", node.Name)

	for r := 1; r <= m.routineConfig.PushSampleRetryAmount; r++ {
		err := m.pushSamples(node)

		// Push ok
		if err == nil {
			log.Debug("Push samples ok")
			m.database.SetNodeTsNow(GetId(node))
			return
		}

		// Push failed
		log.Debugw("Push failed", "node", node.Name, "retry in", m.routineConfig.PushSampleRetryDelay.String(), "atempt", r)

		if r != m.routineConfig.PushSampleRetryAmount {
			// Retry delay
			time.Sleep(m.routineConfig.PushSampleRetryDelay)
		}
	}
}

func GetId(n *meshv1.Node) uint32 {
	return h.Hash(n.Target)
}

func GetSampleId(p *meshv1.Sample) uint32 {
	return h.Hash(p.From + p.To + strconv.FormatInt(p.Key, 10))
}

func getLogger(debug bool, pprofAddress string) *zap.SugaredLogger {
	if debug {
		go func() {
			log.Println("Starting go debugging profiler pprof on port 6060")
			http.ListenAndServe(pprofAddress+":6060", nil)
		}()

		logger, err := zap.NewDevelopment()
		if err != nil {
			log.Fatalf("Could not start debug logging - error: %+v", err)
		}

		return logger.Sugar()
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Could not start production logging - error: %+v", err)
	}

	return logger.Sugar()
}
