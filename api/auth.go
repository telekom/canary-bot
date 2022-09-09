package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"

	connect "github.com/bufbuild/connect-go"
	"google.golang.org/grpc/credentials"
)

func (a *Api) NewAuthInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(
			func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				//a.log.Debugf("API tokens: %+v", a.set.Tokens)
				authToken := req.Header().Get("Authorization")
				// check if token is set
				if authToken == "" {
					a.log.Warnw("Request", "host", req.Header().Get("X-Forwarded-Host"), "auth", "failed")
					return nil, connect.NewError(
						connect.CodeUnauthenticated,
						errors.New("no token provided"),
					)
				}

				// check if token is correct
				for _, t := range a.set.Tokens {
					if authToken[7:] == t {
						a.log.Infow("Request", "host", req.Header().Get("X-Forwarded-Host"), "auth", "succeded")
						return next(ctx, req)
					}
				}
				a.log.Warnw("Request", "host", req.Header().Get("X-Forwarded-Host"), "auth", "failed")
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("auth failed"),
				)
			})
	}
	return connect.UnaryInterceptorFunc(interceptor)
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
