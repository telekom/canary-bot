package api

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"sync"

	third_party "canary-bot/proto/api/third_party"
	apiv1 "canary-bot/proto/api/v1"
	apiv1connect "canary-bot/proto/api/v1/apiv1connect"

	connect "github.com/bufbuild/connect-go"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/encoding/protojson"
)

// Backend implements the protobuf interface
type Backend struct {
	mu      *sync.RWMutex
	samples []*apiv1.Sample
}

// New initializes a new Backend struct.
func New() *Backend {
	return &Backend{
		mu:      &sync.RWMutex{},
		samples: []*apiv1.Sample{{Id: "1234"}},
	}
}

// ListUsers lists all users in the store.
func (b *Backend) ListSamples(ctx context.Context, req *connect.Request[apiv1.ListSampleRequest], srv *connect.ServerStream[apiv1.Sample]) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sample := range b.samples {
		err := srv.Send(sample)
		if err != nil {
			return err
		}
	}

	return nil
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

func Run() error {
	log := grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)

	addr := "localhost:8080"
	// Note: this will succeed asynchronously, once we've started the server below.
	conn, err := grpc.DialContext(
		context.Background(),
		"dns:///"+addr,
		grpc.WithInsecure(),
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

	mux := http.NewServeMux()
	mux.Handle("/", getOpenAPIHandler())
	mux.Handle(apiv1connect.NewSampleServiceHandler(New()))
	mux.Handle("/api/v1/", gwmux)
	server := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	log.Info("Serving Connect, gRPC-Gateway and OpenAPI Documentation on http://", addr)
	return server.ListenAndServe()
}
