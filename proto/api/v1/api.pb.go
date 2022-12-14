// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: v1/api.proto

package apiv1

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/fieldmaskpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// empty sample request
type ListSampleRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListSampleRequest) Reset() {
	*x = ListSampleRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListSampleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListSampleRequest) ProtoMessage() {}

func (x *ListSampleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListSampleRequest.ProtoReflect.Descriptor instead.
func (*ListSampleRequest) Descriptor() ([]byte, []int) {
	return file_v1_api_proto_rawDescGZIP(), []int{0}
}

// response providing a list of measurement samples
type ListSampleResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// list of messured samples
	Samples []*Sample `protobuf:"bytes,1,rep,name=samples,proto3" json:"samples,omitempty"`
}

func (x *ListSampleResponse) Reset() {
	*x = ListSampleResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListSampleResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListSampleResponse) ProtoMessage() {}

func (x *ListSampleResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListSampleResponse.ProtoReflect.Descriptor instead.
func (*ListSampleResponse) Descriptor() ([]byte, []int) {
	return file_v1_api_proto_rawDescGZIP(), []int{1}
}

func (x *ListSampleResponse) GetSamples() []*Sample {
	if x != nil {
		return x.Samples
	}
	return nil
}

// empty node request
type ListNodesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListNodesRequest) Reset() {
	*x = ListNodesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListNodesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListNodesRequest) ProtoMessage() {}

func (x *ListNodesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListNodesRequest.ProtoReflect.Descriptor instead.
func (*ListNodesRequest) Descriptor() ([]byte, []int) {
	return file_v1_api_proto_rawDescGZIP(), []int{2}
}

// response providing a list of known nodes in the mesh
type ListNodesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// list of node names
	Nodes []string `protobuf:"bytes,1,rep,name=nodes,proto3" json:"nodes,omitempty"`
}

func (x *ListNodesResponse) Reset() {
	*x = ListNodesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListNodesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListNodesResponse) ProtoMessage() {}

func (x *ListNodesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListNodesResponse.ProtoReflect.Descriptor instead.
func (*ListNodesResponse) Descriptor() ([]byte, []int) {
	return file_v1_api_proto_rawDescGZIP(), []int{3}
}

func (x *ListNodesResponse) GetNodes() []string {
	if x != nil {
		return x.Nodes
	}
	return nil
}

// a measurement sample
type Sample struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// by whom the sample was messured
	From string `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	// to whom the sample was messured
	To string `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
	// the sample name
	Type string `protobuf:"bytes,3,opt,name=type,proto3" json:"type,omitempty"`
	// the sample value
	Value string `protobuf:"bytes,4,opt,name=value,proto3" json:"value,omitempty"`
	// when the sample was messured
	Ts string `protobuf:"bytes,5,opt,name=ts,proto3" json:"ts,omitempty"`
}

func (x *Sample) Reset() {
	*x = Sample{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_api_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Sample) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Sample) ProtoMessage() {}

func (x *Sample) ProtoReflect() protoreflect.Message {
	mi := &file_v1_api_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Sample.ProtoReflect.Descriptor instead.
func (*Sample) Descriptor() ([]byte, []int) {
	return file_v1_api_proto_rawDescGZIP(), []int{4}
}

func (x *Sample) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *Sample) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *Sample) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Sample) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Sample) GetTs() string {
	if x != nil {
		return x.Ts
	}
	return ""
}

var File_v1_api_proto protoreflect.FileDescriptor

var file_v1_api_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x76, 0x31, 0x2f, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x6d,
	0x61, 0x73, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d,
	0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x13, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x53,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x3e, 0x0a, 0x12,
	0x4c, 0x69, 0x73, 0x74, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x28, 0x0a, 0x07, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x52, 0x07, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x22, 0x12, 0x0a, 0x10,
	0x4c, 0x69, 0x73, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x22, 0x29, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x22, 0x66, 0x0a, 0x06, 0x53,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x74, 0x73, 0x32, 0xc4, 0x01, 0x0a, 0x0a, 0x41, 0x70, 0x69, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x5d, 0x0a, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x73, 0x12, 0x19, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x53,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x61,
	0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x17, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x11,
	0x12, 0x0f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x73, 0x12, 0x57, 0x0a, 0x09, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x12, 0x18,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x6f, 0x64, 0x65,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76,
	0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x15, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0f, 0x12, 0x0d, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x42, 0xd7, 0x02, 0x5a, 0x30, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x6b, 0x6f,
	0x6d, 0x2f, 0x63, 0x61, 0x6e, 0x61, 0x72, 0x79, 0x2d, 0x62, 0x6f, 0x74, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x3b, 0x61, 0x70, 0x69, 0x76, 0x31, 0x92,
	0x41, 0xa1, 0x02, 0x12, 0xf7, 0x01, 0x0a, 0x0a, 0x43, 0x61, 0x6e, 0x61, 0x72, 0x79, 0x20, 0x41,
	0x50, 0x49, 0x12, 0x36, 0x47, 0x65, 0x74, 0x20, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x20, 0x61, 0x6e,
	0x64, 0x20, 0x6d, 0x65, 0x61, 0x73, 0x75, 0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x20, 0x73, 0x61,
	0x6d, 0x70, 0x6c, 0x65, 0x73, 0x20, 0x66, 0x72, 0x6f, 0x6d, 0x20, 0x74, 0x68, 0x65, 0x20, 0x63,
	0x61, 0x6e, 0x61, 0x72, 0x79, 0x2d, 0x6d, 0x65, 0x73, 0x68, 0x22, 0x5d, 0x0a, 0x14, 0x53, 0x63,
	0x68, 0x75, 0x62, 0x65, 0x72, 0x74, 0x2c, 0x20, 0x4d, 0x61, 0x78, 0x69, 0x6d, 0x69, 0x6c, 0x69,
	0x61, 0x6e, 0x12, 0x25, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x6b, 0x6f, 0x6d, 0x2f, 0x63,
	0x61, 0x6e, 0x61, 0x72, 0x79, 0x2d, 0x62, 0x6f, 0x74, 0x1a, 0x1e, 0x6d, 0x61, 0x78, 0x69, 0x6d,
	0x69, 0x6c, 0x69, 0x61, 0x6e, 0x2e, 0x73, 0x63, 0x68, 0x75, 0x62, 0x65, 0x72, 0x74, 0x40, 0x74,
	0x65, 0x6c, 0x65, 0x6b, 0x6f, 0x6d, 0x2e, 0x64, 0x65, 0x2a, 0x4d, 0x0a, 0x12, 0x41, 0x70, 0x61,
	0x63, 0x68, 0x65, 0x20, 0x32, 0x2e, 0x30, 0x20, 0x4c, 0x69, 0x63, 0x65, 0x6e, 0x73, 0x65, 0x12,
	0x37, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x6b, 0x6f, 0x6d, 0x2f, 0x63, 0x61, 0x6e, 0x61,
	0x72, 0x79, 0x2d, 0x62, 0x6f, 0x74, 0x2f, 0x62, 0x6c, 0x6f, 0x62, 0x2f, 0x6d, 0x61, 0x69, 0x6e,
	0x2f, 0x4c, 0x49, 0x43, 0x45, 0x4e, 0x53, 0x45, 0x32, 0x03, 0x31, 0x2e, 0x30, 0x2a, 0x01, 0x02,
	0x32, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73,
	0x6f, 0x6e, 0x3a, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f,
	0x6a, 0x73, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_v1_api_proto_rawDescOnce sync.Once
	file_v1_api_proto_rawDescData = file_v1_api_proto_rawDesc
)

func file_v1_api_proto_rawDescGZIP() []byte {
	file_v1_api_proto_rawDescOnce.Do(func() {
		file_v1_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_v1_api_proto_rawDescData)
	})
	return file_v1_api_proto_rawDescData
}

var file_v1_api_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_v1_api_proto_goTypes = []interface{}{
	(*ListSampleRequest)(nil),  // 0: api.v1.ListSampleRequest
	(*ListSampleResponse)(nil), // 1: api.v1.ListSampleResponse
	(*ListNodesRequest)(nil),   // 2: api.v1.ListNodesRequest
	(*ListNodesResponse)(nil),  // 3: api.v1.ListNodesResponse
	(*Sample)(nil),             // 4: api.v1.Sample
}
var file_v1_api_proto_depIdxs = []int32{
	4, // 0: api.v1.ListSampleResponse.samples:type_name -> api.v1.Sample
	0, // 1: api.v1.ApiService.ListSamples:input_type -> api.v1.ListSampleRequest
	2, // 2: api.v1.ApiService.ListNodes:input_type -> api.v1.ListNodesRequest
	1, // 3: api.v1.ApiService.ListSamples:output_type -> api.v1.ListSampleResponse
	3, // 4: api.v1.ApiService.ListNodes:output_type -> api.v1.ListNodesResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_v1_api_proto_init() }
func file_v1_api_proto_init() {
	if File_v1_api_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_v1_api_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListSampleRequest); i {
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
		file_v1_api_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListSampleResponse); i {
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
		file_v1_api_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListNodesRequest); i {
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
		file_v1_api_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListNodesResponse); i {
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
		file_v1_api_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Sample); i {
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
			RawDescriptor: file_v1_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_v1_api_proto_goTypes,
		DependencyIndexes: file_v1_api_proto_depIdxs,
		MessageInfos:      file_v1_api_proto_msgTypes,
	}.Build()
	File_v1_api_proto = out.File
	file_v1_api_proto_rawDesc = nil
	file_v1_api_proto_goTypes = nil
	file_v1_api_proto_depIdxs = nil
}
