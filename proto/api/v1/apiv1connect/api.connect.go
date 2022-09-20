// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: v1/api.proto

package apiv1connect

import (
	context "context"
	errors "errors"
	connect_go "github.com/bufbuild/connect-go"
	v1 "gitlab.devops.telekom.de/caas/canary-bot/proto/api/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect_go.IsAtLeastVersion0_1_0

const (
	// SampleServiceName is the fully-qualified name of the SampleService service.
	SampleServiceName = "api.v1.SampleService"
)

// SampleServiceClient is a client for the api.v1.SampleService service.
type SampleServiceClient interface {
	ListSamples(context.Context, *connect_go.Request[v1.ListSampleRequest]) (*connect_go.Response[v1.ListSampleResponse], error)
}

// NewSampleServiceClient constructs a client for the api.v1.SampleService service. By default, it
// uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewSampleServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) SampleServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &sampleServiceClient{
		listSamples: connect_go.NewClient[v1.ListSampleRequest, v1.ListSampleResponse](
			httpClient,
			baseURL+"/api.v1.SampleService/ListSamples",
			opts...,
		),
	}
}

// sampleServiceClient implements SampleServiceClient.
type sampleServiceClient struct {
	listSamples *connect_go.Client[v1.ListSampleRequest, v1.ListSampleResponse]
}

// ListSamples calls api.v1.SampleService.ListSamples.
func (c *sampleServiceClient) ListSamples(ctx context.Context, req *connect_go.Request[v1.ListSampleRequest]) (*connect_go.Response[v1.ListSampleResponse], error) {
	return c.listSamples.CallUnary(ctx, req)
}

// SampleServiceHandler is an implementation of the api.v1.SampleService service.
type SampleServiceHandler interface {
	ListSamples(context.Context, *connect_go.Request[v1.ListSampleRequest]) (*connect_go.Response[v1.ListSampleResponse], error)
}

// NewSampleServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewSampleServiceHandler(svc SampleServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle("/api.v1.SampleService/ListSamples", connect_go.NewUnaryHandler(
		"/api.v1.SampleService/ListSamples",
		svc.ListSamples,
		opts...,
	))
	return "/api.v1.SampleService/", mux
}

// UnimplementedSampleServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedSampleServiceHandler struct{}

func (UnimplementedSampleServiceHandler) ListSamples(context.Context, *connect_go.Request[v1.ListSampleRequest]) (*connect_go.Response[v1.ListSampleResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("api.v1.SampleService.ListSamples is not implemented"))
}
