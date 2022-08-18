package mesh

import (
	"canary-bot/data"
	meshv1 "canary-bot/proto/gen/mesh/v1"
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
	"google.golang.org/protobuf/types/known/emptypb"
)

type MeshClient struct {
	conn   *grpc.ClientConn
	client meshv1.MeshServiceClient
}

//bool NameUnique
func (m *Mesh) Join(to []*meshv1.Node) (bool, error) {
	var res *meshv1.JoinMeshResponse
	m.log.Debugf("Starting Join routine")

	for index, node := range to {
		err := m.initClient(node)
		if err != nil {
			if index != len(to)-1 {
				m.log.Debugf("Trying next node to join - err: %+v", err)
				continue
			}
			m.log.Debugf("Error: %+v", err)
			return true, err
		}

		res, err = m.clients[GetId(node)].client.JoinMesh(
			context.Background(),
			&meshv1.JoinMeshRequest{IAmNode: &meshv1.Node{
				Name:   m.config.StartupSettings.Name,
				Target: m.config.StartupSettings.ListenAddress + ":" + strconv.FormatInt(m.config.StartupSettings.ListenPort, 10),
			}})

		if err != nil {
			if index != len(to)-1 {
				m.log.Debugf("Trying next node to join - err: %+v", err)
				continue
			}
			m.log.Debugf("Error: %+v", err)
			return true, err
		}

		// check if name of node is unique in mesh response
		if !res.NameUnique {
			m.log.Debugf("Join response - node name is not unique in mesh")
			return false, nil
		}

		// save join-requested node as node in mesh
		node.Name = res.MyName
		m.database.SetNode(data.Convert(node, NODE_OK))
	}
	for _, node := range res.Nodes {
		if GetId(node) != GetId(&meshv1.Node{
			Name:   m.config.StartupSettings.Name,
			Target: m.config.StartupSettings.ListenAddress + ":" + strconv.FormatInt(m.config.StartupSettings.ListenPort, 10),
		}) {
			m.database.SetNode(data.Convert(node, NODE_OK))
		}
	}

	return true, nil
}

func (m *Mesh) Ping(node *meshv1.Node) (time.Duration, error) {
	err := m.initClient(node)
	if err != nil {
		m.log.Debugf("Could not connect to client!")
		return -1, err
	}
	timeStart := time.Now()
	_, err = m.clients[GetId(node)].client.Ping(context.Background(), &emptypb.Empty{})
	timeEnd := time.Now()
	if err != nil {
		m.log.Debugf("Ping failed.")
		return -1, err
	}
	rtt := timeEnd.Sub(timeStart)

	return rtt, nil
}

func (m *Mesh) NodeDiscovery(toNode *meshv1.Node, newNode *meshv1.Node) {
	err := m.initClient(toNode)
	if err != nil {
		m.log.Warnf("Could not connect to client - skip Node Discover Request to %+v", toNode.Name)
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
		m.log.Warnf("Could not start request to client - skip Node Discover Request to %+v - err: %+v", toNode.Name, err)
	}
	return
}

func (m *Mesh) PushProbes(node *meshv1.Node) error {
	err := m.initClient(node)
	if err != nil {
		m.log.Debugf("Could not connect to client!")
		return err
	}

	var probes []*meshv1.Probe
	databaseProbes := m.database.GetProbeList()
	if len(databaseProbes) == 0 {
		m.log.Debugf("No probes found for push - will not push")
		return nil
	}
	for _, probe := range m.database.GetProbeList() {
		probes = append(probes, &meshv1.Probe{From: probe.From, To: probe.To, Key: probe.Key, Value: probe.Value, Ts: probe.Ts})
	}

	_, err = m.clients[GetId(node)].client.PushProbes(context.Background(), &meshv1.Probes{Probes: probes})
	if err != nil {
		m.log.Debugf("Could not send probes - err: %+v", err)
		return err
	}
	return nil
}

func (m *Mesh) initClient(to *meshv1.Node) error {
	nodeId := GetId(to)

	m.log.Debug("Init Client routine")

	if _, exists := m.clients[nodeId]; !exists {
		var opts []grpc.DialOption

		// TLS
		tlsCredentials, err := loadClientTLSCredentials(m.config.StartupSettings.CaCertPath, m.config.StartupSettings.CaCert)
		if err != nil {
			m.log.Debug("Cannot load TLS credentials - starting insecure connection: ", err)
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(tlsCredentials))
		}

		conn, err := grpc.Dial(to.Target, opts...)
		if err != nil {
			m.log.Debugf("Error: %+v", err)
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
		m.log.Debug("Client already existed")
	}
	return nil
}

func (c *MeshClient) closeClient() {
	c.conn.Close()
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
