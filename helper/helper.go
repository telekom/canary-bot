/*
 * canary-bot
 *
 * (C) 2022, Maximilian Schubert, Deutsche Telekom IT GmbH
 *
 * Deutsche Telekom IT GmbH and all other contributors /
 * copyright owners license this file to you under the Apache
 * License, Version 2.0 (the "License"); you may not use this
 * file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package helper

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"google.golang.org/grpc/credentials"
	"hash/fnv"
	"log"
	"math/big"
	"net"
	"os"
)

// ExternalIP returns the external IP of the host
func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Fatal(err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			log.Printf("Lookup Outbound IP res: %+v\n", ip)
			return ip.String(), nil
		}
	}
	return "", errors.New("Could not get outbound IP. Are you connected to the network?")
}

func Hash(s string) (uint32, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0, errors.New("Generating a hash value failed: " + err.Error())
	}
	return h.Sum32(), nil
}

// charset is the alphabet for the random string generation
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"0123456789"

func stringWithCharset(n int64, chars string) (string, error) {
	ret := make([]byte, n)
	for i := int64(0); i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		ret[i] = chars[num.Int64()]
	}

	return string(ret), nil
}

func GenerateRandomToken(length int64) string {
	token, err := stringWithCharset(length, charset)
	if err != nil {
		panic("Could not generate a random token, please check func GenerateRandomToken")
	}
	return token
}

// LoadClientTLSCredentials loads a certificate from disk and creates
// a transport credentials object for gRPC usage.
func LoadClientTLSCredentials(cacertPaths []string, cacertB64 []byte) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server certificate

	certPool := x509.NewCertPool()

	if len(cacertPaths) > 0 {
		for _, path := range cacertPaths {
			/* #nosec G304*/
			pemServerCA, err := os.ReadFile(path)
			if err != nil {
				panic("Failed to add server ca certificate, path not found (security issue): " + path)
			}
			if !certPool.AppendCertsFromPEM(pemServerCA) {
				return nil, fmt.Errorf("Failed to add server ca certificate")
			}
		}
	} else if cacertB64 != nil {
		var pemServerCA []byte
		_, err := base64.StdEncoding.Decode(pemServerCA, cacertB64)
		if err != nil || !certPool.AppendCertsFromPEM(pemServerCA) {
			return nil, fmt.Errorf("Failed to add server ca certificate")
		}
	} else {
		return nil, errors.New("Neither ca cert path nor base64 encoded ca cert set")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}

	return credentials.NewTLS(config), nil
}

// LoadServerTLSCredentials loads a certificate from disk and creates
// a transport credentials object for gRPC usage.
func LoadServerTLSCredentials(servercertPath string, serverkeyPath string, servercertB64 []byte, serverkeyB64 []byte) (*tls.Config, error) {
	// Load server certificate and key //credentials.NewTLS(config)
	var serverCert tls.Certificate
	var err error

	if servercertPath != "" && serverkeyPath != "" {
		serverCert, err = tls.LoadX509KeyPair(servercertPath, serverkeyPath)
	} else if servercertB64 != nil && serverkeyB64 != nil {
		var cert []byte
		var key []byte
		_, err = base64.StdEncoding.Decode(cert, servercertB64)
		if err != nil {
			return nil, err
		}
		_, err = base64.StdEncoding.Decode(key, servercertB64)
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
		MinVersion:   tls.VersionTLS12,
	}

	return config, nil
}
