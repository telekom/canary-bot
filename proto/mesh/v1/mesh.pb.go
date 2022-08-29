// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: v1/mesh.proto

package meshv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type JoinMeshRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IAmNode *Node `protobuf:"bytes,1,opt,name=i_am_node,json=iAmNode,proto3" json:"i_am_node,omitempty"`
}

func (x *JoinMeshRequest) Reset() {
	*x = JoinMeshRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_mesh_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinMeshRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinMeshRequest) ProtoMessage() {}

func (x *JoinMeshRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_mesh_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinMeshRequest.ProtoReflect.Descriptor instead.
func (*JoinMeshRequest) Descriptor() ([]byte, []int) {
	return file_v1_mesh_proto_rawDescGZIP(), []int{0}
}

func (x *JoinMeshRequest) GetIAmNode() *Node {
	if x != nil {
		return x.IAmNode
	}
	return nil
}

type JoinMeshResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NameUnique bool    `protobuf:"varint,1,opt,name=name_unique,json=nameUnique,proto3" json:"name_unique,omitempty"`
	MyName     string  `protobuf:"bytes,2,opt,name=my_name,json=myName,proto3" json:"my_name,omitempty"`
	Nodes      []*Node `protobuf:"bytes,3,rep,name=nodes,proto3" json:"nodes,omitempty"`
}

func (x *JoinMeshResponse) Reset() {
	*x = JoinMeshResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_mesh_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinMeshResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinMeshResponse) ProtoMessage() {}

func (x *JoinMeshResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_mesh_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinMeshResponse.ProtoReflect.Descriptor instead.
func (*JoinMeshResponse) Descriptor() ([]byte, []int) {
	return file_v1_mesh_proto_rawDescGZIP(), []int{1}
}

func (x *JoinMeshResponse) GetNameUnique() bool {
	if x != nil {
		return x.NameUnique
	}
	return false
}

func (x *JoinMeshResponse) GetMyName() string {
	if x != nil {
		return x.MyName
	}
	return ""
}

func (x *JoinMeshResponse) GetNodes() []*Node {
	if x != nil {
		return x.Nodes
	}
	return nil
}

type NodeDiscoveryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NewNode *Node `protobuf:"bytes,1,opt,name=new_node,json=newNode,proto3" json:"new_node,omitempty"`
	IAmNode *Node `protobuf:"bytes,2,opt,name=i_am_node,json=iAmNode,proto3" json:"i_am_node,omitempty"`
}

func (x *NodeDiscoveryRequest) Reset() {
	*x = NodeDiscoveryRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_mesh_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeDiscoveryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeDiscoveryRequest) ProtoMessage() {}

func (x *NodeDiscoveryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_mesh_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeDiscoveryRequest.ProtoReflect.Descriptor instead.
func (*NodeDiscoveryRequest) Descriptor() ([]byte, []int) {
	return file_v1_mesh_proto_rawDescGZIP(), []int{2}
}

func (x *NodeDiscoveryRequest) GetNewNode() *Node {
	if x != nil {
		return x.NewNode
	}
	return nil
}

func (x *NodeDiscoveryRequest) GetIAmNode() *Node {
	if x != nil {
		return x.IAmNode
	}
	return nil
}

type Node struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Target string `protobuf:"bytes,2,opt,name=target,proto3" json:"target,omitempty"`
}

func (x *Node) Reset() {
	*x = Node{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_mesh_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Node) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Node) ProtoMessage() {}

func (x *Node) ProtoReflect() protoreflect.Message {
	mi := &file_v1_mesh_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Node.ProtoReflect.Descriptor instead.
func (*Node) Descriptor() ([]byte, []int) {
	return file_v1_mesh_proto_rawDescGZIP(), []int{3}
}

func (x *Node) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Node) GetTarget() string {
	if x != nil {
		return x.Target
	}
	return ""
}

type Probes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Probes []*Probe `protobuf:"bytes,1,rep,name=probes,proto3" json:"probes,omitempty"`
}

func (x *Probes) Reset() {
	*x = Probes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_mesh_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Probes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Probes) ProtoMessage() {}

func (x *Probes) ProtoReflect() protoreflect.Message {
	mi := &file_v1_mesh_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Probes.ProtoReflect.Descriptor instead.
func (*Probes) Descriptor() ([]byte, []int) {
	return file_v1_mesh_proto_rawDescGZIP(), []int{4}
}

func (x *Probes) GetProbes() []*Probe {
	if x != nil {
		return x.Probes
	}
	return nil
}

type Probe struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	From  string `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	To    string `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
	Key   int64  `protobuf:"varint,3,opt,name=key,proto3" json:"key,omitempty"`
	Value string `protobuf:"bytes,4,opt,name=value,proto3" json:"value,omitempty"`
	Ts    int64  `protobuf:"varint,5,opt,name=ts,proto3" json:"ts,omitempty"`
}

func (x *Probe) Reset() {
	*x = Probe{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_mesh_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Probe) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Probe) ProtoMessage() {}

func (x *Probe) ProtoReflect() protoreflect.Message {
	mi := &file_v1_mesh_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Probe.ProtoReflect.Descriptor instead.
func (*Probe) Descriptor() ([]byte, []int) {
	return file_v1_mesh_proto_rawDescGZIP(), []int{5}
}

func (x *Probe) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *Probe) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *Probe) GetKey() int64 {
	if x != nil {
		return x.Key
	}
	return 0
}

func (x *Probe) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Probe) GetTs() int64 {
	if x != nil {
		return x.Ts
	}
	return 0
}

var File_v1_mesh_proto protoreflect.FileDescriptor

var file_v1_mesh_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x76, 0x31, 0x2f, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3c, 0x0a, 0x0f, 0x4a, 0x6f, 0x69, 0x6e, 0x4d, 0x65, 0x73,
	0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x29, 0x0a, 0x09, 0x69, 0x5f, 0x61, 0x6d,
	0x5f, 0x6e, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6d, 0x65,
	0x73, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x07, 0x69, 0x41, 0x6d, 0x4e,
	0x6f, 0x64, 0x65, 0x22, 0x71, 0x0a, 0x10, 0x4a, 0x6f, 0x69, 0x6e, 0x4d, 0x65, 0x73, 0x68, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x6e, 0x61, 0x6d, 0x65, 0x5f,
	0x75, 0x6e, 0x69, 0x71, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x6e, 0x61,
	0x6d, 0x65, 0x55, 0x6e, 0x69, 0x71, 0x75, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x6d, 0x79, 0x5f, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x79, 0x4e, 0x61, 0x6d,
	0x65, 0x12, 0x23, 0x0a, 0x05, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x0d, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52,
	0x05, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x22, 0x6b, 0x0a, 0x14, 0x4e, 0x6f, 0x64, 0x65, 0x44, 0x69,
	0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x28,
	0x0a, 0x08, 0x6e, 0x65, 0x77, 0x5f, 0x6e, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0d, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52,
	0x07, 0x6e, 0x65, 0x77, 0x4e, 0x6f, 0x64, 0x65, 0x12, 0x29, 0x0a, 0x09, 0x69, 0x5f, 0x61, 0x6d,
	0x5f, 0x6e, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6d, 0x65,
	0x73, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x07, 0x69, 0x41, 0x6d, 0x4e,
	0x6f, 0x64, 0x65, 0x22, 0x32, 0x0a, 0x04, 0x4e, 0x6f, 0x64, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x22, 0x30, 0x0a, 0x06, 0x50, 0x72, 0x6f, 0x62, 0x65,
	0x73, 0x12, 0x26, 0x0a, 0x06, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x0e, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x62,
	0x65, 0x52, 0x06, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x73, 0x22, 0x63, 0x0a, 0x05, 0x50, 0x72, 0x6f,
	0x62, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x74, 0x6f, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x0e,
	0x0a, 0x02, 0x74, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x74, 0x73, 0x32, 0x8d,
	0x02, 0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x68, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x41,
	0x0a, 0x08, 0x4a, 0x6f, 0x69, 0x6e, 0x4d, 0x65, 0x73, 0x68, 0x12, 0x18, 0x2e, 0x6d, 0x65, 0x73,
	0x68, 0x2e, 0x76, 0x31, 0x2e, 0x4a, 0x6f, 0x69, 0x6e, 0x4d, 0x65, 0x73, 0x68, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x4a,
	0x6f, 0x69, 0x6e, 0x4d, 0x65, 0x73, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x38, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x48, 0x0a, 0x0d, 0x4e,
	0x6f, 0x64, 0x65, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x12, 0x1d, 0x2e, 0x6d,
	0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x44, 0x69, 0x73, 0x63, 0x6f,
	0x76, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x37, 0x0a, 0x0a, 0x50, 0x75, 0x73, 0x68, 0x50, 0x72, 0x6f,
	0x62, 0x65, 0x73, 0x12, 0x0f, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72,
	0x6f, 0x62, 0x65, 0x73, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x42, 0x25,
	0x5a, 0x23, 0x63, 0x61, 0x6e, 0x61, 0x72, 0x79, 0x2d, 0x62, 0x6f, 0x74, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x3b, 0x6d,
	0x65, 0x73, 0x68, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_v1_mesh_proto_rawDescOnce sync.Once
	file_v1_mesh_proto_rawDescData = file_v1_mesh_proto_rawDesc
)

func file_v1_mesh_proto_rawDescGZIP() []byte {
	file_v1_mesh_proto_rawDescOnce.Do(func() {
		file_v1_mesh_proto_rawDescData = protoimpl.X.CompressGZIP(file_v1_mesh_proto_rawDescData)
	})
	return file_v1_mesh_proto_rawDescData
}

var file_v1_mesh_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_v1_mesh_proto_goTypes = []interface{}{
	(*JoinMeshRequest)(nil),      // 0: mesh.v1.JoinMeshRequest
	(*JoinMeshResponse)(nil),     // 1: mesh.v1.JoinMeshResponse
	(*NodeDiscoveryRequest)(nil), // 2: mesh.v1.NodeDiscoveryRequest
	(*Node)(nil),                 // 3: mesh.v1.Node
	(*Probes)(nil),               // 4: mesh.v1.Probes
	(*Probe)(nil),                // 5: mesh.v1.Probe
	(*emptypb.Empty)(nil),        // 6: google.protobuf.Empty
}
var file_v1_mesh_proto_depIdxs = []int32{
	3, // 0: mesh.v1.JoinMeshRequest.i_am_node:type_name -> mesh.v1.Node
	3, // 1: mesh.v1.JoinMeshResponse.nodes:type_name -> mesh.v1.Node
	3, // 2: mesh.v1.NodeDiscoveryRequest.new_node:type_name -> mesh.v1.Node
	3, // 3: mesh.v1.NodeDiscoveryRequest.i_am_node:type_name -> mesh.v1.Node
	5, // 4: mesh.v1.Probes.probes:type_name -> mesh.v1.Probe
	0, // 5: mesh.v1.MeshService.JoinMesh:input_type -> mesh.v1.JoinMeshRequest
	6, // 6: mesh.v1.MeshService.Ping:input_type -> google.protobuf.Empty
	2, // 7: mesh.v1.MeshService.NodeDiscovery:input_type -> mesh.v1.NodeDiscoveryRequest
	4, // 8: mesh.v1.MeshService.PushProbes:input_type -> mesh.v1.Probes
	1, // 9: mesh.v1.MeshService.JoinMesh:output_type -> mesh.v1.JoinMeshResponse
	6, // 10: mesh.v1.MeshService.Ping:output_type -> google.protobuf.Empty
	6, // 11: mesh.v1.MeshService.NodeDiscovery:output_type -> google.protobuf.Empty
	6, // 12: mesh.v1.MeshService.PushProbes:output_type -> google.protobuf.Empty
	9, // [9:13] is the sub-list for method output_type
	5, // [5:9] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_v1_mesh_proto_init() }
func file_v1_mesh_proto_init() {
	if File_v1_mesh_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_v1_mesh_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinMeshRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_v1_mesh_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinMeshResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_v1_mesh_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeDiscoveryRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_v1_mesh_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Node); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_v1_mesh_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Probes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_v1_mesh_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Probe); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_v1_mesh_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_v1_mesh_proto_goTypes,
		DependencyIndexes: file_v1_mesh_proto_depIdxs,
		MessageInfos:      file_v1_mesh_proto_msgTypes,
	}.Build()
	File_v1_mesh_proto = out.File
	file_v1_mesh_proto_rawDesc = nil
	file_v1_mesh_proto_goTypes = nil
	file_v1_mesh_proto_depIdxs = nil
}
