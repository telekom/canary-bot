package mesh

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"gitlab.devops.telekom.de/caas/canary-bot/data"
	h "gitlab.devops.telekom.de/caas/canary-bot/helper"
	meshv1 "gitlab.devops.telekom.de/caas/canary-bot/proto/mesh/v1"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MeshClient struct {
	conn   *grpc.ClientConn
	client meshv1.MeshServiceClient
}

//bool NameUnique
func (m *Mesh) Join(targets []string) (bool, bool) {
	log := m.log.Named("join-routine")
	var res *meshv1.JoinMeshResponse
	log.Debugw("Starting")

	for index, target := range targets {
		log.Debugf("Index %+v Targets: %+v", index, targets)
		node := &meshv1.Node{Name: "", Target: target}

		err := m.initClient(node, false, false, false)
		if err != nil {
			m.log.Debug("Could not connect to client, joinMesh request failed")
			if index != len(targets)-1 {
				log.Debugw("Trying next node", "error", err)
				continue
			}
			return false, true
		}

		res, err = m.clients[GetId(node)].client.JoinMesh(
			context.Background(),
			&meshv1.Node{
				Name:   m.config.StartupSettings.Name,
				Target: m.config.StartupSettings.JoinAddress,
			})

		if err != nil {
			m.log.Debug("Client connected, but joinMesh request failed")
			if index != len(targets)-1 {
				log.Debugw("Trying next node", "error", err)
				continue
			}
			return false, true
		}

		// check if name of node is unique in mesh response
		if !res.NameUnique {
			log.Debugw("Node name is not unique in mesh")
			return true, false
		}

		// save join-requested node as node in mesh
		node.Name = res.MyName
		m.database.SetNode(data.Convert(node, NODE_OK))
		//m.clients[GetId(node)].conn.Close()

		log.Infow("Joined mesh", "name", node.Name, "target", node.Target)
		break
	}
	for _, node := range res.Nodes {
		if GetId(node) != GetId(&meshv1.Node{
			Name:   m.config.StartupSettings.Name,
			Target: m.config.StartupSettings.JoinAddress,
		}) {
			m.database.SetNode(data.Convert(node, NODE_OK))
		}
	}
	return true, true
}

func (m *Mesh) Ping(node *meshv1.Node) error {
	log := m.log.Named("ping-routine")
	err := m.initClient(node, false, false, false)
	if err != nil {
		log.Debugw("Could not connect to client")
		return err
	}
	_, err = m.clients[GetId(node)].client.Ping(
		context.Background(),
		&meshv1.Node{
			Name:   m.config.StartupSettings.Name,
			Target: m.config.StartupSettings.JoinAddress,
		})
	if err != nil {
		log.Debugw("Ping failed")
		return err
	}

	return nil
}

func (m *Mesh) NodeDiscovery(toNode *meshv1.Node, newNode *meshv1.Node) {
	log := m.log.Named("discovery-routine")
	err := m.initClient(toNode, false, false, false)
	if err != nil {
		log.Warnw("Could not connect to client - skip Node Discover Request", "node", toNode.Name)
		return
	}
	_, err = m.clients[GetId(toNode)].client.NodeDiscovery(
		context.Background(),
		&meshv1.NodeDiscoveryRequest{
			NewNode: newNode,
			IAmNode: &meshv1.Node{
				Name:   m.config.StartupSettings.Name,
				Target: m.config.StartupSettings.ListenAddress + ":" + strconv.FormatInt(m.config.StartupSettings.ListenPort, 10),
			}})
	if err != nil {
		log.Warnf("Could not start request to client - skip Node Discover Request", "node", toNode.Name, "error", err)
	}
	return
}

func (m *Mesh) PushSamples(node *meshv1.Node) error {
	log := m.log.Named("sample-routine")
	err := m.initClient(node, false, false, false)
	if err != nil {
		log.Debugw("Could not connect to client")
		return err
	}

	var samples []*meshv1.Sample
	databaseSamples := m.database.GetSampleList()
	if len(databaseSamples) == 0 {
		log.Debugw("No samples found for push - will not push")
		return nil
	}
	for _, sample := range m.database.GetSampleList() {
		samples = append(samples, &meshv1.Sample{From: sample.From, To: sample.To, Key: sample.Key, Value: sample.Value, Ts: sample.Ts})
	}

	_, err = m.clients[GetId(node)].client.PushSamples(context.Background(), &meshv1.Samples{Samples: samples})
	if err != nil {
		log.Debugw("Could not send samples", "error", err)
		return err
	}
	return nil
}

func (m *Mesh) initClient(to *meshv1.Node, blocking bool, wait bool, forceReconnect bool) error {
	nodeId := GetId(to)
	log := m.log.Named("client")
	log.Debugw("Init client")

	if m.config.StartupSettings.DebugGrpc {
		grpc_zap.ReplaceGrpcLoggerV2(log.Named("grpc").Desugar())
	}

	if _, exists := m.clients[nodeId]; !exists {
		var opts []grpc.DialOption

		// TLS
		tlsCredentials, err := h.LoadClientTLSCredentials(m.config.StartupSettings.CaCertPath, m.config.StartupSettings.CaCert)
		if err != nil {
			log.Debugw("Cannot load TLS credentials - starting insecure connection", "error", err.Error())
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(tlsCredentials))
		}

		// blocking
		if blocking {
			opts = append(opts, grpc.WithBlock())
		}

		// wait for connection
		if wait {
			opts = append(opts, grpc.WithDefaultCallOptions(grpc.WaitForReady(true)))
		}

		// dial
		conn, err := grpc.Dial(to.Target, opts...)
		if err != nil {
			log.Debugw("Dial error", "error", err)
			return err
		}

		client := meshv1.NewMeshServiceClient(conn)

		m.mu.Lock()
		m.clients[nodeId] = &MeshClient{
			client: client,
			conn:   conn,
		}
		m.mu.Unlock()
	} else {
		log.Debugw("Client already existed")
	}
	return nil
}

func (m *Mesh) closeClient(to *meshv1.Node) {
	m.mu.Lock()
	m.clients[GetId(to)].conn.Close()
	// remove client
	delete(m.clients, GetId(to))
	m.mu.Unlock()
}

func (m *Mesh) Rtt() {
	log := m.log.Named("rtt")
	log.Debugw("Starting RTT measurement")
	var opts []grpc.DialOption
	var rttStartH, rttStart, rttEnd time.Time

	nodes := m.database.GetNodeListByState(NODE_OK)
	if nodes == nil {
		log.Debugw("No Node suitable for RTT measurement")
		return
	}
	// select random node for RTT measurment
	node := nodes[rand.Intn(len(nodes))]
	log.Debugw("Node selected", "node", node.Name)
	// grpc logging
	if m.config.StartupSettings.DebugGrpc {
		grpc_zap.ReplaceGrpcLoggerV2(log.Named("grpc").Desugar())
	}

	// TLS
	tlsCredentials, err := h.LoadClientTLSCredentials(m.config.StartupSettings.CaCertPath, m.config.StartupSettings.CaCert)
	if err != nil {
		log.Debugw("Cannot load TLS credentials - starting insecure connection", "error", err.Error())
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(tlsCredentials))
	}

	// blocking
	opts = append(opts, grpc.WithBlock())

	// start RTT with TCP handshake
	rttStartH = time.Now()
	// dial
	conn, err := grpc.Dial(node.Target, opts...)
	defer conn.Close()
	if err != nil {
		log.Debugw("Dial error", "error", err)
		return
	}

	client := meshv1.NewMeshServiceClient(conn)
	if err != nil {
		log.Debugw("Could not connect to client")
		return
	}

	// start RTT without TCP handshake
	rttStart = time.Now()

	// send request
	_, err = client.Rtt(context.Background(), &emptypb.Empty{})
	// end RTT
	rttEnd = time.Now()

	if err != nil {
		log.Debugw("RTT failed")
		return
	}
	log.Debugw("RTT succeded")
	// RTT with handshake
	rttH := rttEnd.Sub(rttStartH)
	// RTT without handshale
	rtt := rttEnd.Sub(rttStart)

	// safe samples
	m.database.SetSample(
		&data.Sample{
			From:  m.config.StartupSettings.Name,
			To:    node.Name,
			Key:   data.RTT_TOTAL,
			Value: strconv.FormatInt(rttH.Nanoseconds(), 10),
			Ts:    time.Now().Unix(),
		},
	)

	m.database.SetSample(
		&data.Sample{
			From:  m.config.StartupSettings.Name,
			To:    node.Name,
			Key:   data.RTT_REQUEST,
			Value: strconv.FormatInt(rtt.Nanoseconds(), 10),
			Ts:    time.Now().Unix(),
		},
	)

	return
}
