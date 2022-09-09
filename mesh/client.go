package mesh

import (
	"canary-bot/data"
	meshv1 "canary-bot/proto/mesh/v1"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
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

	log.Debugw("Conn loop")
	for index, target := range targets {
		node := &meshv1.Node{Name: "", Target: target}

		log.Debugw("target loop")
		err := m.initClient(node, false, false, false)
		log.Debugw("init client ok")
		if err != nil {
			if index != len(targets)-1 {
				log.Debugw("Trying next node", "error", err)
				continue
			}
			m.log.Debug("Could not connect to client, joinMesh request failed")
			return false, true
		}

		res, err = m.clients[GetId(node)].client.JoinMesh(
			context.Background(),
			&meshv1.Node{
				Name:   m.config.StartupSettings.Name,
				Target: m.config.StartupSettings.Domain,
			})

		if err != nil {
			if index != len(targets)-1 {
				log.Debugw("Trying next node", "error", err)
				continue
			}
			m.log.Debug("Client connected, but joinMesh request failed")
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
		log.Debug("JOIN OK")
		//m.clients[GetId(node)].conn.Close()
		break
	}
	log.Debug("JOIN OK 2")
	for _, node := range res.Nodes {
		if GetId(node) != GetId(&meshv1.Node{
			Name:   m.config.StartupSettings.Name,
			Target: m.config.StartupSettings.ListenAddress + ":" + strconv.FormatInt(m.config.StartupSettings.ListenPort, 10),
		}) {
			m.database.SetNode(data.Convert(node, NODE_OK))
		}
	}

	log.Debug("JOIN OK 3")
	return true, true
}

func (m *Mesh) Ping(node *meshv1.Node) (time.Duration, error) {
	log := m.log.Named("ping-routine")
	err := m.initClient(node, false, false, false)
	if err != nil {
		log.Debugw("Could not connect to client")
		return -1, err
	}
	timeStart := time.Now()
	_, err = m.clients[GetId(node)].client.Ping(
		context.Background(),
		&meshv1.Node{
			Name:   m.config.StartupSettings.Name,
			Target: m.config.StartupSettings.Domain,
		})
	timeEnd := time.Now()
	if err != nil {
		log.Debugw("Ping failed")
		return -1, err
	}
	rtt := timeEnd.Sub(timeStart)

	return rtt, nil
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

	if _, exists := m.clients[nodeId]; !exists {
		var opts []grpc.DialOption

		// TLS
		tlsCredentials, err := loadClientTLSCredentials(m.config.StartupSettings.CaCertPath, m.config.StartupSettings.CaCert)
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
		log.Debugw("DIAL")
		conn, err := grpc.Dial(to.Target, opts...)
		if err != nil {
			log.Debugw("Dial error", "error", err)
			return err
		}
		log.Debugw("DIAL OK")

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

func loadClientTLSCredentials(caCert_Path string, caCert_b64 []byte) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server certificate

	var pemServerCA []byte
	var err error

	if caCert_Path != "" {
		pemServerCA, err = ioutil.ReadFile(caCert_Path)
	} else if caCert_b64 != nil {
		_, err = base64.StdEncoding.Decode(pemServerCA, caCert_b64)
	} else {
		return nil, errors.New("Neither ca cert path nor base64 encoded ca cert set")
	}

	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("Failed to add server ca certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}
