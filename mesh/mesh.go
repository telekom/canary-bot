package mesh

import (
	"canary-bot/api"
	"canary-bot/data"
	h "canary-bot/helper"
	meshv1 "canary-bot/proto/mesh/v1"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Main struct for the mesh
type Mesh struct {
	// Mesh in-memory datastore
	database data.Database
	// Global zap logger
	logger *zap.SugaredLogger
	// Configuration for the timer- and channelRoutines
	routineConfig *RoutineConfiguration
	// Configuration how the bot can connect to the mesh etc.
	setupConfig *SetupConfiguration

	// GRPC client store
	clients map[uint32]*MeshClient
	mu      sync.Mutex

	// Channel if a new node is discovered in the mesh
	newNodeDiscovered chan NodeDiscovered

	// timerRoutine main functionality timers
	pingTicker        *time.Ticker
	pushSampleTicker  *time.Ticker
	cleanSampleTicker *time.Ticker

	// timerRoutine sample measurement timers
	rttTicker *time.Ticker

	// Channels to quit and re-enter mesh joinRoutine
	quitJoinRoutine    chan bool
	restartJoinRoutine chan bool
	joinRoutineDone    bool
}

// A newly discoverd node in the mesh
type NodeDiscovered struct {
	NewNode *meshv1.Node
	From    uint32 // TODO change to name
}

// Create a canary bot & mesh with the desired configuration
// Use a pre-defined routineConfig with e.g. StandardProductionRoutineConfig()
// and define your mesh setip configuration
//
// - Logger will be initialised
// - setupConfig will be checked
// - DB will be created
// - Mesh will be initialised
// - Routines will be started
// - API will be created
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

	// start mesh server
	go func() {
		logger.Info("Starting server, listening vor joining nodes")
		err := m.StartServer()
		if err != nil {
			logger.Debugf("Mesh server error: %+v", err)
			logger.Fatal("Could not start Mesh Server")
		}
	}()

	// start main mesh functionality
	logger.Infow("Starting mesh routines")
	go m.channelRoutines()
	go m.timerRoutines()

	// start API
	apiConfig := &api.Configuration{
		NodeName:       setupConfig.Name,
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

	// start the mesh API
	if err = api.StartApi(database, apiConfig, logger.Named("api")); err != nil {
		logger.Fatal("Could not start API - Error: %+v", err)
	}
}

// Routines that will be executed by timer interrupts.
// In the startup phase, just the joinRoutine timer will run
// After joining a mesh or a node is joining all routines
// will be started and the join Routine will stop.
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

	// Sample measurement: RTT
	m.rttTicker = time.NewTicker(m.routineConfig.RttInterval)
	m.rttTicker.Stop()

	for {
		select {
		case <-joinTicker.C:
			joinTicker.Stop()
			log := m.logger.Named("join-routine")
			// join (future) mesh
			log.Infow("Waiting for a node to join a mesh...")
			connected, isNameUniqueInMesh := m.Join(m.setupConfig.Targets)
			if !isNameUniqueInMesh {
				log.Fatal("The name is not unique in the mesh, please choose another one.")
			}
			if connected {
				log.Infow("Connected to a mesh")
				m.quitJoinRoutine <- true
			} else {
				joinTicker.Reset(m.routineConfig.JoinInterval)
			}

		case <-m.pingTicker.C:
			log := m.logger.Named("ping-routine")
			log.Debugw("Starting")

			// get a random healthy node
			nodes := m.database.GetRandomNodeListByState(NODE_OK, 1)
			if len(nodes) == 0 {
				log.Debugw("No Node connected or all nodes in timeout")
				break
			}

			// ping choosen node
			go m.retryPing(nodes[0].Convert())

		case <-m.pushSampleTicker.C:
			log := m.logger.Named("sample-routine")
			log.Debugw("Starting push sample routine to random nodes", "amount", m.routineConfig.PushSampleToAmount)

			// get random, configured amount of healty nodes
			nodes := m.database.GetRandomNodeListByState(NODE_OK, m.routineConfig.PushSampleToAmount)
			if len(nodes) == 0 {
				log.Debugw("No node connected or all nodes in timeout")
				break
			}

			// push own measurement samples to choosen nodes
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
			// measure round-trip-time samples
			go m.Rtt()

		case <-m.restartJoinRoutine:
			// stop ticker and re-enter joinRoutine
			joinTicker.Reset(m.routineConfig.JoinInterval)
			m.joinRoutineDone = false

			m.pingTicker.Stop()
			m.pushSampleTicker.Stop()
			m.cleanSampleTicker.Stop()
			m.rttTicker.Stop()
			m.logger.Debug("Start joinRoutine again, stopping all timer routines")
		case <-m.quitJoinRoutine:
			joinTicker.Stop()
			m.joinRoutineDone = true
			// starting ticker after joinRoutine
			m.pingTicker.Reset(m.routineConfig.PingInterval)
			m.pushSampleTicker.Reset(m.routineConfig.PushSampleInterval)
			m.cleanSampleTicker.Reset(m.routineConfig.CleanSampleInterval)
			m.rttTicker.Reset(m.routineConfig.RttInterval)
			m.logger.Info("Starting pings")
			m.logger.Debug("Stop joinRoutine, starting all timer routines")
		}
	}
}

// Routines that will be executed by event/channel interrupts.
// Events:
// - nodeDiscovered: A new node is discovered in the mesh
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

// Will call the ping method with set retry configuration.
// Database nodes and samples will be updated.
func (m *Mesh) retryPing(node *meshv1.Node) {
	log := m.logger.Named("ping-routine")
	log.Debugw("Retry routine started", "node", node.Name)

	// start retry ping logic
	for r := 1; r <= m.routineConfig.PingRetryAmount; r++ {
		// Ping the node
		err := m.ping(node)

		// Ping ok; return
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

// Will call the pushSample method with set retry configuration.
// Database nodes and samples will be updated.
func (m *Mesh) retryPushSample(node *meshv1.Node) {
	log := m.logger.Named("sample-routine")
	log.Debugw("Push sample retry routine started", "node", node.Name)

	// start retry pushSample logic
	for r := 1; r <= m.routineConfig.PushSampleRetryAmount; r++ {
		// Push all samples to node
		err := m.pushSamples(node)

		// Push ok; return
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

// Get the ID of a node
// Hash integer value of the target field (name of node)
func GetId(n *meshv1.Node) uint32 {
	return h.Hash(n.Target)
}

// Get the ID of a sample
// Hash integer value of the concatenated From, To and Key field
func GetSampleId(p *meshv1.Sample) uint32 {
	return h.Hash(p.From + p.To + strconv.FormatInt(p.Key, 10))
}

// Setup the Logger
func getLogger(debug bool, pprofAddress string) *zap.SugaredLogger {
	if debug {
		// starting pprof for memory and cpu analysis
		go func() {
			log.Println("Starting go debugging profiler pprof on port 6060")
			http.ListenAndServe(pprofAddress+":6060", nil)
		}()

		// using debug logger
		logger, err := zap.NewDevelopment()
		if err != nil {
			log.Fatalf("Could not start debug logging - error: %+v", err)
		}

		return logger.Sugar()
	}

	// using prod logger in non-debug mode
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Could not start production logging - error: %+v", err)
	}

	return logger.Sugar()
}
