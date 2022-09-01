package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"mime"
	"net/http"
	"strconv"
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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

// Api implements the protobuf interface
type Api struct {
	data data.Database
	set  mesh.Settings
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

func NewApi(data data.Database, set mesh.Settings, log *zap.SugaredLogger) error {

	a := &Api{
		data: data,
		set:  set,
		log:  log,
	}

	// TLS client
	var opts []grpc.DialOption
	tlsClientCredentials, err := loadClientTLSCredentials(set.CaCertPath, set.CaCert)
	if err != nil {
		log.Debugw("Cannot load TLS credentials - starting insecure connection", "error", err.Error())
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(tlsClientCredentials))
	}

	addr := set.ListenAddress + ":" + strconv.FormatInt(set.ApiPort, 10)
	// Note: this will succeed asynchronously, once we've started the server below.
	conn, err := grpc.DialContext(
		context.Background(),
		"dns:///"+addr,
		//grpc.WithTransportCredentials(credentials.NewTLS(tlsCredentials)),
		//grpc.WithInsecure(),
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

	mux := http.NewServeMux()
	mux.Handle("/", getOpenAPIHandler())
	mux.Handle(apiv1connect.NewSampleServiceHandler(a))
	mux.Handle("/api/v1/", gwmux)
	server := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	log.Info("Serving Connect, gRPC-Gateway and OpenAPI Documentation on ", addr)

	// TLS
	tlsCredentials, err := loadTLSCredentials(
		set.ServerCertPath,
		set.ServerKeyPath,
		set.ServerCert,
		set.ServerKey,
	)

	if err != nil {
		log.Warnw("Cannot load TLS credentials - using insecure connection")
		log.Debugw("Cannot load TLS credentials", "error", err.Error())
	}
	if tlsCredentials != nil {
		server.TLSConfig = tlsCredentials
		return server.ListenAndServeTLS("", "")
	}

	return server.ListenAndServe()
}

func loadTLSCredentials(serverCert_path string, serverKey_path string, serverCert_b64 []byte, serverKey_b64 []byte) (*tls.Config, error) {
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

	return config, nil
}

func loadClientTLSCredentials(caCert_Path string, caCert_b64 []byte) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server certificate

	var pemServerCA []byte
	var err error

	if caCert_Path != "" {
		pemServerCA, err = ioutil.ReadFile(caCert_Path)
	} else if caCert_b64 != nil {
		_, err = base64.StdEncoding.Decode(pemServerCA, caCert_b64)
	} else {
		return nil, errors.New("Neither ca cert path nor base64 encoded ca cert set")
	}

	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("Failed to add server ca certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}
