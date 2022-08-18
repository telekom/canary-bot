package mesh

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"log"
	"net"
	"strconv"

	"canary-bot/data"
	meshv1 "canary-bot/proto/gen/mesh/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MeshServer struct {
	meshv1.UnimplementedMeshServiceServer
	log  *zap.SugaredLogger
	data *data.Database
	name *string

	newNodeDiscovered chan NodeDiscovered
}

func (s *MeshServer) JoinMesh(ctx context.Context, req *meshv1.JoinMeshRequest) (*meshv1.JoinMeshResponse, error) {
	// Check if name of joining node is unique in mesh
	if s.data.GetNodeByName(req.IAmNode.Name).Id != 0 || *s.name == req.IAmNode.Name {
		return &meshv1.JoinMeshResponse{NameUnique: false, MyName: *s.name, Nodes: []*meshv1.Node{}}, nil
	}
	s.newNodeDiscovered <- NodeDiscovered{req.IAmNode, GetId(req.IAmNode)}

	var nodes []*meshv1.Node
	for _, datanode := range s.data.GetNodeList() {
		nodes = append(nodes, &meshv1.Node{Name: datanode.Name, Target: datanode.Target})
	}
	res := meshv1.JoinMeshResponse{NameUnique: true, MyName: *s.name, Nodes: nodes}
	return &res, nil
}

func (s *MeshServer) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *MeshServer) NodeDiscovery(ctx context.Context, req *meshv1.NodeDiscoveryRequest) (*emptypb.Empty, error) {
	s.newNodeDiscovered <- NodeDiscovered{req.NewNode, GetId(req.IAmNode)}
	return &emptypb.Empty{}, nil
}

func (s *MeshServer) PushProbes(ctx context.Context, req *meshv1.Probes) (*emptypb.Empty, error) {
	for _, probe := range req.Probes {
		if probe.Ts > s.data.GetProbeTs(GetProbeId(probe)) {
			s.data.SetProbe(&data.Probe{
				From:  probe.From,
				To:    probe.To,
				Key:   probe.Key,
				Value: probe.Value,
				Ts:    probe.Ts,
			})
		}
	}
	log.Println("SERVER GOT REQUEST")
	log.Printf("Safe probes: %+v", req.Probes)
	log.Printf("All my now probes: %+v", s.data.GetProbeList())
	return &emptypb.Empty{}, nil
}

func (m *Mesh) StartServer() error {
	meshServer := &MeshServer{
		log:               m.log,
		data:              &m.database,
		name:              &m.config.StartupSettings.Name,
		newNodeDiscovered: m.newNodeDiscovered,
	}

	listenAdd := m.config.StartupSettings.ListenAddress + ":" + strconv.FormatInt(m.config.StartupSettings.ListenPort, 10)

	m.log.Infof("Start listening to %+v\n", listenAdd)
	lis, err := net.Listen("tcp", listenAdd)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}

	// TLS
	tlsCredentials, err := loadTLSCredentials(
		m.config.StartupSettings.ServerCertPath,
		m.config.StartupSettings.ServerKeyPath,
		m.config.StartupSettings.ServerCert,
		m.config.StartupSettings.ServerKey,
	)
	if err != nil {
		m.log.Warnf("Cannot load TLS credentials - using insecure connection")
		m.log.Debugf("Cannot load TLS credentials - err: %v", err)
	}

	if tlsCredentials != nil {
		opts = append(opts, grpc.Creds(tlsCredentials))
	}

	grpcServer := grpc.NewServer(opts...)
	meshv1.RegisterMeshServiceServer(grpcServer, meshServer)
	reflection.Register(grpcServer)
	grpcServer.Serve(lis)
	return nil
}

func loadTLSCredentials(serverCert_path string, serverKey_path string, serverCert_b64 []byte, serverKey_b64 []byte) (credentials.TransportCredentials, error) {
	// Load server certificate and key
	var serverCert tls.Certificate
	var err error

	if serverCert_path != "" && serverKey_path != "" {
		serverCert, err = tls.LoadX509KeyPair(serverCert_path, serverKey_path)
	} else if serverCert_b64 != nil && serverKey_b64 != nil {
		var cert []byte
		var key []byte
		_, err = base64.StdEncoding.Decode(cert, serverCert_b64)
		if err != nil {
			return nil, err
		}
		_, err = base64.StdEncoding.Decode(key, serverCert_b64)
		serverCert, err = tls.X509KeyPair(cert, key)
	} else {
		return nil, errors.New("Neither server cert and key path nor base64 encoded cert and key set")
	}

	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}
