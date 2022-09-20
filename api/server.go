package api

import (
	"context"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"strconv"
	"time"

	"gitlab.devops.telekom.de/caas/canary-bot/data"
	h "gitlab.devops.telekom.de/caas/canary-bot/helper"
	"gitlab.devops.telekom.de/caas/canary-bot/mesh"

	third_party "gitlab.devops.telekom.de/caas/canary-bot/proto/api/third_party"
	apiv1 "gitlab.devops.telekom.de/caas/canary-bot/proto/api/v1"
	apiv1connect "gitlab.devops.telekom.de/caas/canary-bot/proto/api/v1/apiv1connect"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

	connect "github.com/bufbuild/connect-go"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

// Api implements the protobuf interface
type Api struct {
	data data.Database
	set  *mesh.Settings
	log  *zap.SugaredLogger
}

// ListUsers lists all users in the store.
func (b *Api) ListSamples(ctx context.Context, req *connect.Request[apiv1.ListSampleRequest]) (*connect.Response[apiv1.ListSampleResponse], error) {
	samples := []*apiv1.Sample{}

	for _, sample := range b.data.GetSampleList() {
		samples = append(samples, &apiv1.Sample{
			From:   sample.From,
			To:     sample.To,
			Sample: data.SampleName[sample.Key],
			Value:  sample.Value,
			Ts:     time.Unix(sample.Ts, 0).String(),
		})
	}

	return connect.NewResponse(&apiv1.ListSampleResponse{
		Samples: samples,
	}), nil
}

func getOpenAPIHandler() http.Handler {
	mime.AddExtensionType(".svg", "image/svg+xml")
	// Use subdirectory in embedded files
	subFS, err := fs.Sub(third_party.OpenAPI, "OpenAPI")
	if err != nil {
		panic("couldn't create sub filesystem: " + err.Error())
	}
	return http.FileServer(http.FS(subFS))
}

func NewApi(data data.Database, set *mesh.Settings, log *zap.SugaredLogger) error {
	a := &Api{
		data: data,
		set:  set,
		log:  log,
	}

	if set.DebugGrpc {
		grpc_zap.ReplaceGrpcLoggerV2(log.Named("grpc").Desugar())
	}

	var opts []grpc.DialOption

	// TLS for http proxy server
	tlsCredentials, err := h.LoadServerTLSCredentials(
		set.ServerCertPath,
		set.ServerKeyPath,
		set.ServerCert,
		set.ServerKey,
	)

	if err != nil {
		log.Warnw("Cannot load TLS server credentials - using insecure connection for incoming requests")
		log.Debugw("Cannot load TLS credentials", "error", err.Error())
	}

	// TLS for client connect from http proxy server to grpc server
	// just load it if TLS is activated, not considered for edge-terminated TLS
	var tlsClientCredentials credentials.TransportCredentials
	if tlsCredentials != nil {
		tlsClientCredentials, err = h.LoadClientTLSCredentials(set.CaCertPath, set.CaCert)

	}

	if err != nil {
		log.Debugw("Cannot load TLS client credentials - starting insecure connection to grpc server")
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(tlsClientCredentials))
	}

	addr := set.ListenAddress + ":" + strconv.FormatInt(set.ApiPort, 10)
	// Note: this will succeed asynchronously, once we've started the server below.
	conn, err := grpc.DialContext(
		context.Background(),
		"dns:///"+addr,
		opts...,
	)
	if err != nil {
		return fmt.Errorf("failed to dial server: %w", err)
	}

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption("*", &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{UseProtoNames: true},
			},
		}),
	)

	err = apiv1.RegisterSampleServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	// Auth
	interceptors := connect.WithInterceptors(a.NewAuthInterceptor())

	mux := http.NewServeMux()
	mux.Handle("/", getOpenAPIHandler())
	mux.Handle(apiv1connect.NewSampleServiceHandler(a, interceptors))
	mux.Handle("/api/v1/", gwmux)
	server := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	log.Info("Serving Connect, gRPC-Gateway and OpenAPI Documentation on ", addr)

	// TLS ready
	if tlsCredentials != nil {
		server.TLSConfig = tlsCredentials
		return server.ListenAndServeTLS("", "")
	}

	return server.ListenAndServe()
}
