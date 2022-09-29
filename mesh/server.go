package mesh

import (
	"context"
	"net"
	"strconv"

	"canary-bot/data"
	h "canary-bot/helper"
	meshv1 "canary-bot/proto/mesh/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
)

type MeshServer struct {
	meshv1.UnimplementedMeshServiceServer
	log  *zap.SugaredLogger
	data *data.Database
	name *string

	newNodeDiscovered chan NodeDiscovered
}

func (s *MeshServer) JoinMesh(ctx context.Context, req *meshv1.Node) (*meshv1.JoinMeshResponse, error) {
	s.log.Infow("New join mesh request", "node", req.Name)
	// Check if name of joining node is unique in mesh, let join if state is not ok, let join if target is same
	dbnode := s.data.GetNodeByName(req.Name)
	if (dbnode.Id != 0 && dbnode.State == NODE_OK && dbnode.Target != req.Target) || *s.name == req.Name {
		return &meshv1.JoinMeshResponse{NameUnique: false, MyName: *s.name, Nodes: []*meshv1.Node{}}, nil
	}
	s.newNodeDiscovered <- NodeDiscovered{req, GetId(req)}

	var nodes []*meshv1.Node
	for _, datanode := range s.data.GetNodeList() {
		nodes = append(nodes, &meshv1.Node{Name: datanode.Name, Target: datanode.Target})
	}
	res := meshv1.JoinMeshResponse{NameUnique: true, MyName: *s.name, Nodes: nodes}
	return &res, nil
}

func (s *MeshServer) Ping(ctx context.Context, req *meshv1.Node) (*emptypb.Empty, error) {
	if req != nil {
		s.data.SetNode(data.Convert(req, NODE_OK))
	}
	return &emptypb.Empty{}, nil
}

func (s *MeshServer) NodeDiscovery(ctx context.Context, req *meshv1.NodeDiscoveryRequest) (*emptypb.Empty, error) {
	s.newNodeDiscovered <- NodeDiscovered{req.NewNode, GetId(req.IAmNode)}
	return &emptypb.Empty{}, nil
}

func (s *MeshServer) PushSamples(ctx context.Context, req *meshv1.Samples) (*emptypb.Empty, error) {
	for _, sample := range req.Samples {
		if sample.Ts > s.data.GetSampleTs(GetSampleId(sample)) {
			s.data.SetSample(&data.Sample{
				From:  sample.From,
				To:    sample.To,
				Key:   sample.Key,
				Value: sample.Value,
				Ts:    sample.Ts,
			})
		}
	}
	s.log.Debugw("Safe samples", "count", len(s.data.GetSampleList()))
	return &emptypb.Empty{}, nil
}

func (s *MeshServer) Rtt(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *Mesh) StartServer() error {
	meshServer := &MeshServer{
		log:               m.logger.Named("server"),
		data:              &m.database,
		name:              &m.setupConfig.Name,
		newNodeDiscovered: m.newNodeDiscovered,
	}

	if m.setupConfig.DebugGrpc {
		grpc_zap.ReplaceGrpcLoggerV2(meshServer.log.Named("grpc").Desugar())
	}

	listenAdd := m.setupConfig.ListenAddress + ":" + strconv.FormatInt(m.setupConfig.ListenPort, 10)

	meshServer.log.Infow("Start listening", "address", listenAdd)
	lis, err := net.Listen("tcp", listenAdd)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}

	// TLS
	tlsCredentials, err := h.LoadServerTLSCredentials(
		m.setupConfig.ServerCertPath,
		m.setupConfig.ServerKeyPath,
		m.setupConfig.ServerCert,
		m.setupConfig.ServerKey,
	)
	if err != nil {
		meshServer.log.Warnw("Cannot load TLS credentials - using insecure connection")
		meshServer.log.Debugw("Cannot load TLS credentials", "error", err.Error())
	}

	if tlsCredentials != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsCredentials)))
	}

	grpcServer := grpc.NewServer(opts...)
	meshv1.RegisterMeshServiceServer(grpcServer, meshServer)
	reflection.Register(grpcServer)
	grpcServer.Serve(lis)
	return nil
}
