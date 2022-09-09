// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: v1/mesh.proto

package meshv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MeshServiceClient is the client API for MeshService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MeshServiceClient interface {
	JoinMesh(ctx context.Context, in *Node, opts ...grpc.CallOption) (*JoinMeshResponse, error)
	Ping(ctx context.Context, in *Node, opts ...grpc.CallOption) (*emptypb.Empty, error)
	NodeDiscovery(ctx context.Context, in *NodeDiscoveryRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	PushSamples(ctx context.Context, in *Samples, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type meshServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMeshServiceClient(cc grpc.ClientConnInterface) MeshServiceClient {
	return &meshServiceClient{cc}
}

func (c *meshServiceClient) JoinMesh(ctx context.Context, in *Node, opts ...grpc.CallOption) (*JoinMeshResponse, error) {
	out := new(JoinMeshResponse)
	err := c.cc.Invoke(ctx, "/mesh.v1.MeshService/JoinMesh", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *meshServiceClient) Ping(ctx context.Context, in *Node, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/mesh.v1.MeshService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *meshServiceClient) NodeDiscovery(ctx context.Context, in *NodeDiscoveryRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/mesh.v1.MeshService/NodeDiscovery", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *meshServiceClient) PushSamples(ctx context.Context, in *Samples, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/mesh.v1.MeshService/PushSamples", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MeshServiceServer is the server API for MeshService service.
// All implementations must embed UnimplementedMeshServiceServer
// for forward compatibility
type MeshServiceServer interface {
	JoinMesh(context.Context, *Node) (*JoinMeshResponse, error)
	Ping(context.Context, *Node) (*emptypb.Empty, error)
	NodeDiscovery(context.Context, *NodeDiscoveryRequest) (*emptypb.Empty, error)
	PushSamples(context.Context, *Samples) (*emptypb.Empty, error)
	mustEmbedUnimplementedMeshServiceServer()
}

// UnimplementedMeshServiceServer must be embedded to have forward compatible implementations.
type UnimplementedMeshServiceServer struct {
}

func (UnimplementedMeshServiceServer) JoinMesh(context.Context, *Node) (*JoinMeshResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method JoinMesh not implemented")
}
func (UnimplementedMeshServiceServer) Ping(context.Context, *Node) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedMeshServiceServer) NodeDiscovery(context.Context, *NodeDiscoveryRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NodeDiscovery not implemented")
}
func (UnimplementedMeshServiceServer) PushSamples(context.Context, *Samples) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PushSamples not implemented")
}
func (UnimplementedMeshServiceServer) mustEmbedUnimplementedMeshServiceServer() {}

// UnsafeMeshServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MeshServiceServer will
// result in compilation errors.
type UnsafeMeshServiceServer interface {
	mustEmbedUnimplementedMeshServiceServer()
}

func RegisterMeshServiceServer(s grpc.ServiceRegistrar, srv MeshServiceServer) {
	s.RegisterService(&MeshService_ServiceDesc, srv)
}

func _MeshService_JoinMesh_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Node)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MeshServiceServer).JoinMesh(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mesh.v1.MeshService/JoinMesh",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MeshServiceServer).JoinMesh(ctx, req.(*Node))
	}
	return interceptor(ctx, in, info, handler)
}

func _MeshService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Node)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MeshServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mesh.v1.MeshService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MeshServiceServer).Ping(ctx, req.(*Node))
	}
	return interceptor(ctx, in, info, handler)
}

func _MeshService_NodeDiscovery_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeDiscoveryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MeshServiceServer).NodeDiscovery(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mesh.v1.MeshService/NodeDiscovery",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MeshServiceServer).NodeDiscovery(ctx, req.(*NodeDiscoveryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MeshService_PushSamples_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Samples)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MeshServiceServer).PushSamples(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mesh.v1.MeshService/PushSamples",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MeshServiceServer).PushSamples(ctx, req.(*Samples))
	}
	return interceptor(ctx, in, info, handler)
}

// MeshService_ServiceDesc is the grpc.ServiceDesc for MeshService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MeshService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mesh.v1.MeshService",
	HandlerType: (*MeshServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "JoinMesh",
			Handler:    _MeshService_JoinMesh_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _MeshService_Ping_Handler,
		},
		{
			MethodName: "NodeDiscovery",
			Handler:    _MeshService_NodeDiscovery_Handler,
		},
		{
			MethodName: "PushSamples",
			Handler:    _MeshService_PushSamples_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "v1/mesh.proto",
}
