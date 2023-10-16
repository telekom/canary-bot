/*
 * canary-bot
 *
 * (C) 2022, Maximilian Schubert, Deutsche Telekom IT GmbH
 *
 * Deutsche Telekom IT GmbH and all other contributors /
 * copyright owners license this file to you under the Apache
 * License, Version 2.0 (the "License"); you may not use this
 * file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package mesh

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/telekom/canary-bot/api"
	"github.com/telekom/canary-bot/data"
	h "github.com/telekom/canary-bot/helper"
	"github.com/telekom/canary-bot/metric"
	meshv1 "github.com/telekom/canary-bot/proto/mesh/v1"

	"go.uber.org/zap"
)

// Mesh is the internal mesh representation
type Mesh struct {
	// Mesh in-memory datastore
	database data.Database
	// Metrics for bot
	metrics metric.Metrics
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
	pingTicker       *time.Ticker
	pushSampleTicker *time.Ticker
	cleanupTicker    *time.Ticker

	// timerRoutine sample measurement timers
	rttTicker *time.Ticker

	// Channels to quit and re-enter mesh joinRoutine
	quitJoinRoutine    chan bool
	restartJoinRoutine chan bool
	joinRoutineDone    bool
}

// NodeDiscovered represents a newly discovered node in the mesh
type NodeDiscovered struct {
	NewNode *meshv1.Node
	From    uint32 // TODO change to name
}

// CreateCanaryMesh creates a canary bot & mesh with the desired configuration
// Use a pre-defined routineConfig with e.g. StandardProductionRoutineConfig()
// and define your mesh setup configuration
//
// - Logger will be initialized
//
// - setupConfig will be checked
//
// - DB will be created
//
// - Mesh will be initialized
//
// - Routines will be started
//
// - API will be created
func CreateCanaryMesh(routineConfig *RoutineConfiguration, setupConfig *SetupConfiguration) {
	// prepare logging
	logger := setupLogger(setupConfig.Debug, setupConfig.ListenAddress)
	defer logger.Sync()

	// SetupConfiguration
	logger.Debugf("CLI settings: %+v", setupConfig)

	// Try to set missing setupConfigurations
	setupConfig.setDefaults(logger)
	// Get info from configuration combination
	setupConfig.checkDefaults(logger)

	// prepare the in-memory database
	database, err := data.NewMemDB(logger.Named("database"))
	if err != nil {
		logger.Fatalf("Could not create Memory Database (MemDB) - Error: %+v", err)
	}

	// init metrics
	metrics := metric.InitMetrics()

	m := &Mesh{
		database:           database,
		metrics:            metrics,
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

	// start the main mesh functionality
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
	if err = api.StartApi(database, metrics, apiConfig, logger.Named("api")); err != nil {
		logger.Fatal("Could not start API - Error: %+v", err)
	}
}

// timeRoutines initializes the go routines that will be executed by timer interrupts.
// In the startup phase, just the joinRoutine timer will run
// After joining a mesh or a node is joining the mesh all routines
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
	// Timer to clean samples from removed nodes
	m.cleanupTicker = time.NewTicker(m.routineConfig.CleanupInterval)
	m.cleanupTicker.Stop()

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
			nodes := m.database.GetRandomNodeListByState(NodeOk, 1)
			if len(nodes) == 0 {
				log.Debugw("No Node connected or all nodes in timeout")
				break
			}

			// ping chosen node
			go m.retryPing(nodes[0].Convert())

		case <-m.pushSampleTicker.C:
			log := m.logger.Named("sample-routine")
			log.Debugw("Starting push sample routine to random nodes", "amount", m.routineConfig.PushSampleToAmount)

			// get random, configured number of health nodes
			nodes := m.database.GetRandomNodeListByState(NodeOk, m.routineConfig.PushSampleToAmount)
			if len(nodes) == 0 {
				log.Debugw("No node connected or all nodes in timeout")
				break
			}

			// push own measurement samples to chosen nodes
			for _, node := range nodes {
				log.Debugw("Pushing samples", "node", node.Name)
				go m.retryPushSample(node.Convert())
			}

		case <-m.cleanupTicker.C:
			// check if the node is timed-out and over maxAge
			if m.setupConfig.CleanupNodes {
				for _, node := range m.database.GetNodeListByState(NodeDead) {
					if time.Unix(node.StateChangeTs, 0).Before(time.Now().Add(-1 * m.routineConfig.CleanupMaxAge)) {
						m.logger.Infow("Delete old node", "node", node.Name, "maxAge", m.routineConfig.CleanupMaxAge.String())
					}
				}
			}

			// check if the sample is over maxAge
			if m.setupConfig.CleanupSamples {
				for _, sample := range m.database.GetSampleList() {
					if time.Unix(sample.Ts, 0).Before(time.Now().Add(-1 * m.routineConfig.CleanupMaxAge)) {
						m.logger.Infow("Delete old sample", "from", sample.From, "to", sample.To, "key", data.SampleName[sample.Key], "maxAge", m.routineConfig.CleanupMaxAge.String())
						m.database.DeleteSample(sample.Id)
					}
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
			m.cleanupTicker.Stop()
			m.rttTicker.Stop()
			m.logger.Debug("Start joinRoutine again, stopping all timer routines")
		case <-m.quitJoinRoutine:
			joinTicker.Stop()
			m.joinRoutineDone = true
			// starting ticker after joinRoutine
			m.pingTicker.Reset(m.routineConfig.PingInterval)
			m.pushSampleTicker.Reset(m.routineConfig.PushSampleInterval)
			m.cleanupTicker.Reset(m.routineConfig.CleanupInterval)
			m.rttTicker.Reset(m.routineConfig.RttInterval)
			m.logger.Info("Starting pings")
			m.logger.Debug("Stop joinRoutine, starting all timer routines")
		}
	}
}

// channelRoutines initializes the go routines that will be executed by event/channel interrupts.
// Events:
//
// - nodeDiscovered: A new node is discovered in the mesh
func (m *Mesh) channelRoutines() {
	for {
		select {
		case nodeDiscovered := <-m.newNodeDiscovered:
			logger := m.logger.Named("discovery-routine")
			// quit joinMesh routine if discovery is received before
			if !m.joinRoutineDone {
				m.quitJoinRoutine <- true
			}
			if m.database.GetNodeByName(nodeDiscovered.NewNode.Name).Id != 0 {
				logger.Info("Node is rejoining node")
				m.database.SetNode(data.Convert(nodeDiscovered.NewNode, NodeOk))
				break
			}

			logger.Info("Node joined - new node")
			logger.Debugw("Starting discovery broadcast routine to random nodes", "amount", m.routineConfig.BroadcastToAmount)

			nodes := m.database.GetRandomNodeListByState(NodeOk, m.routineConfig.BroadcastToAmount, nodeDiscovered.From)

			if len(nodes) == 0 {
				logger.Debug("Stopping routine prematurely - no more known nodes")
				break
			}

			for _, node := range nodes {
				logger.Infow("Sending Discovery Broadcast", "node", node.Name)
				go m.NodeDiscovery(node.Convert(), nodeDiscovered.NewNode)
			}

			m.database.SetNode(data.Convert(nodeDiscovered.NewNode, NodeOk))
		}
	}
}

// retryPing Will call the ping method with set retry configuration.
// Database nodes and samples will be updated.
func (m *Mesh) retryPing(node *meshv1.Node) {
	logger := m.logger.Named("ping-routine")
	logger.Debugw("Retry routine started", "node", node.Name)

	// start retry ping logic
	for r := 1; r <= m.routineConfig.PingRetryAmount; r++ {
		// Ping the node
		err := m.ping(node)

		// Ping ok; return
		if err == nil {
			m.database.SetNode(data.Convert(node, NodeOk))
			logger.Infow("Ping ok", "node", node.Name, "attempt", r)
			return
		}

		// Ping failed
		logger.Infow("Ping failed", "node", node.Name, "timeout", m.routineConfig.RequestTimeout.String(), "retry in", m.routineConfig.PingRetryDelay.String(), "attempt", r)
		m.database.SetNode(data.Convert(node, NodeTimeout))
		m.database.SetSampleNaN(GetSampleId(&meshv1.Sample{From: m.setupConfig.Name, To: node.Name, Key: data.RttRequest}))
		m.database.SetSampleNaN(GetSampleId(&meshv1.Sample{From: m.setupConfig.Name, To: node.Name, Key: data.RttTotal}))

		if r != m.routineConfig.PingRetryAmount {
			// Retry delay
			time.Sleep(m.routineConfig.PingRetryDelay)
		} else {
			m.database.SetNode(data.Convert(node, NodeDead))
		}
	}

	// Retry limit reached
	logger.Infow("Retry limit reached", "node", node.Name, "limit", m.routineConfig.PingRetryAmount)
	logger.Warnw("Removing node from mesh", "node", node.Name)
	m.database.DeleteNode(GetId(node))

	// Check if node was the last node in mesh
	if len(m.database.GetNodeList()) == 0 {
		m.restartJoinRoutine <- true
	}
}

// retryPushSample will call the pushSample method with set retry configuration.
// Database nodes and samples will be updated.
func (m *Mesh) retryPushSample(node *meshv1.Node) {
	logger := m.logger.Named("sample-routine")
	logger.Debugw("Push sample retry routine started", "node", node.Name)

	// start retry pushSample logic
	for r := 1; r <= m.routineConfig.PushSampleRetryAmount; r++ {
		// Push all samples to node
		err := m.pushSamples(node)

		// Push ok; return
		if err == nil {
			logger.Debug("Push samples ok")
			m.database.SetNodeTsNow(GetId(node))
			return
		}

		// Push failed
		logger.Debugw("Push failed", "node", node.Name, "retry in", m.routineConfig.PushSampleRetryDelay.String(), "attempt", r)

		if r != m.routineConfig.PushSampleRetryAmount {
			// Retry delay
			time.Sleep(m.routineConfig.PushSampleRetryDelay)
		}
	}
}

// GetId returns the hashed ID of a node
func GetId(n *meshv1.Node) uint32 {
	id, err := h.Hash(n.Target)
	if err != nil {
		log.Printf("Could not get the hash value of the sample, please check the hash function")
	}
	return id
}

// GetSampleId returns the ID of a sample
func GetSampleId(p *meshv1.Sample) uint32 {
	id, err := h.Hash(p.From + p.To + strconv.FormatInt(p.Key, 10))
	if err != nil {
		log.Printf("Could not get the hash value of the sample, please check the hash function")
	}
	return id
}

// setupLogger setups the Logger
func setupLogger(debug bool, pprofAddress string) *zap.SugaredLogger {
	if debug {
		// starting pprof for memory and cpu analysis
		// go func() {
		// 	log.Println("Starting go debugging profiler pprof on port 6060")
		// 	http.ListenAndServe(pprofAddress+":6060", nil)
		// }()

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
