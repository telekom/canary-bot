package api

import (
	"context"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"strconv"
	"sync"
	"time"

	"canary-bot/data"
	"canary-bot/mesh"
	third_party "canary-bot/proto/api/third_party"
	apiv1 "canary-bot/proto/api/v1"
	apiv1connect "canary-bot/proto/api/v1/apiv1connect"

	connect "github.com/bufbuild/connect-go"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

// Api implements the protobuf interface
type Api struct {
	mu   *sync.RWMutex
	data data.Database
	set  mesh.Settings
	log  *zap.SugaredLogger
}

// ListUsers lists all users in the store.
func (b *Api) ListSamples(ctx context.Context, req *connect.Request[apiv1.ListSampleRequest], srv *connect.ServerStream[apiv1.Sample]) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	b.data.GetProbeList()
	for _, sample := range b.data.GetProbeList() {
		err := srv.Send(
			&apiv1.Sample{
				From:   sample.From,
				To:     sample.To,
				Sample: data.SampleName[sample.Key],
				Value:  sample.Value,
				Ts:     time.Unix(sample.Ts, 0).String(),
			})
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

func NewApi(data data.Database, set mesh.Settings, log *zap.SugaredLogger) error {

	a := &Api{
		mu:   &sync.RWMutex{},
		data: data,
		set:  set,
		log:  log,
	}

	addr := set.ListenAddress + ":" + strconv.FormatInt(set.ApiPort, 10)
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
	mux.Handle(apiv1connect.NewSampleServiceHandler(a))
	mux.Handle("/api/v1/", gwmux)
	server := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	log.Info("Serving Connect, gRPC-Gateway and OpenAPI Documentation on http://", addr)
	return server.ListenAndServe()
}
