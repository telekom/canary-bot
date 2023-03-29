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
	"context"
	"net"
	"strconv"

	"github.com/telekom/canary-bot/data"
	h "github.com/telekom/canary-bot/helper"
	"github.com/telekom/canary-bot/metric"
	meshv1 "github.com/telekom/canary-bot/proto/mesh/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
)

// Mesh server for incomming requests
type MeshServer struct {
	meshv1.UnimplementedMeshServiceServer
	metrics metric.Metrics
	log     *zap.SugaredLogger
	data    *data.Database
	name    *string

	newNodeDiscovered chan NodeDiscovered
}

// RPC if node wants to join the mesh
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

// PC if node pings
func (s *MeshServer) Ping(ctx context.Context, req *meshv1.Node) (*emptypb.Empty, error) {
	if req != nil {
		s.data.SetNode(data.Convert(req, NODE_OK))
	}
	return &emptypb.Empty{}, nil
}

// RPC if new node is discovered in the mesh
func (s *MeshServer) NodeDiscovery(ctx context.Context, req *meshv1.NodeDiscoveryRequest) (*emptypb.Empty, error) {
	s.newNodeDiscovered <- NodeDiscovered{req.NewNode, GetId(req.IAmNode)}
	return &emptypb.Empty{}, nil
}

// RPC if samples will be sent by node in mesh
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

// PRC if node measures rount-trip-time
// Do not add any functionality that will effect the RTT
func (s *MeshServer) Rtt(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// Start the mesh server.
// Setup gRPC and TLS.
func (m *Mesh) StartServer() error {
	meshServer := &MeshServer{
		log:               m.logger.Named("server"),
		metrics:           m.metrics,
		data:              &m.database,
		name:              &m.setupConfig.Name,
		newNodeDiscovered: m.newNodeDiscovered,
	}

	// gRPC debug mode for more logs
	if m.setupConfig.DebugGrpc {
		grpc_zap.ReplaceGrpcLoggerV2(meshServer.log.Named("grpc").Desugar())
	}

	// address the server will be bound to
	listenAdd := m.setupConfig.ListenAddress + ":" + strconv.FormatInt(m.setupConfig.ListenPort, 10)

	// start TCP listener
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

	// register gRPC listener
	grpcServer := grpc.NewServer(opts...)
	meshv1.RegisterMeshServiceServer(grpcServer, meshServer)
	reflection.Register(grpcServer)
	err = grpcServer.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}
