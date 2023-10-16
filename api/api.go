/*
 * MIT License
 *
 * Copyright (c) 2022 Johan Brandhorst-Satzkorn
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

// https://github.com/johanbrandhorst/connect-gateway-example

package api

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"strconv"
	"time"

	connect "github.com/bufbuild/connect-go"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/telekom/canary-bot/data"
	h "github.com/telekom/canary-bot/helper"
	"github.com/telekom/canary-bot/metric"
	"github.com/telekom/canary-bot/proto/api/third_party"
	apiv1 "github.com/telekom/canary-bot/proto/api/v1"
	"github.com/telekom/canary-bot/proto/api/v1/apiv1connect"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

// StartApi starts the API server of the canary
func StartApi(data data.Database, metrics metric.Metrics, config *Configuration, log *zap.SugaredLogger) error {
	a := &Api{
		data:    data,
		metrics: metrics,
		config:  config,
		log:     log,
	}

	if config.DebugGrpc {
		grpc_zap.ReplaceGrpcLoggerV2(log.Named("grpc").Desugar())
	}

	var opts []grpc.DialOption

	// TLS for http proxy server
	tlsCredentials, err := h.LoadServerTLSCredentials(
		config.ServerCertPath,
		config.ServerKeyPath,
		config.ServerCert,
		config.ServerKey,
	)
	if err != nil {
		log.Warnw("Cannot load TLS server credentials - using insecure connection for incoming requests")
		log.Debugw("Cannot load TLS credentials", "error", err.Error())
	}

	// TLS for client connect from http proxy server to grpc server
	// just load it if TLS is activated, not considered for edge-terminated TLS
	var tlsClientCredentials credentials.TransportCredentials
	if tlsCredentials != nil {
		tlsClientCredentials, err = h.LoadClientTLSCredentials(config.CaCertPath, config.CaCert)
	}

	if err != nil {
		log.Debugw("Cannot load TLS client credentials - starting insecure connection to grpc server")
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(tlsClientCredentials))
	}

	addr := config.Address + ":" + strconv.FormatInt(config.Port, 10)
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

	err = apiv1.RegisterApiServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	// Auth
	interceptors := connect.WithInterceptors(a.NewAuthInterceptor())

	mux := http.NewServeMux()

	// Open API Handler + Endpoint
	openApiHandler, err := getOpenAPIHandler()
	if err != nil {
		log.Warn("Could not start the OpenAPI Endpoint ", err)
	} else {
		mux.Handle("/", openApiHandler)
	}

	mux.Handle(apiv1connect.NewApiServiceHandler(a, interceptors))
	mux.Handle("/api/v1/", gwmux)
	mux.Handle("/metrics",
		a.NewAuthHandler(
			metrics.Handler(a.data,
				promhttp.HandlerFor(
					metrics.GetRegistry(),
					promhttp.HandlerOpts{
						EnableOpenMetrics: true,
					},
				),
			),
		),
	)
	server := &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: time.Minute,
	}
	log.Info("Serving Connect, gRPC-Gateway and OpenAPI Documentation on ", addr)

	// TLS ready
	if tlsCredentials != nil {
		server.TLSConfig = tlsCredentials
		return server.ListenAndServeTLS("", "")
	}

	return server.ListenAndServe()
}

func getOpenAPIHandler() (http.Handler, error) {
	err := mime.AddExtensionType(".svg", "image/svg+xml")
	if err != nil {
		return nil, errors.New("Couldn't add extension type: " + err.Error())
	}
	// Use subdirectory in embedded files
	subFS, err := fs.Sub(third_party.OpenAPI, "OpenAPI")
	if err != nil {
		return nil, errors.New("Couldn't create sub filesystem: " + err.Error())
	}
	return http.FileServer(http.FS(subFS)), nil
}
