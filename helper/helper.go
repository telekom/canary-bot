package helper

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
	"time"

	"google.golang.org/grpc/credentials"
)

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
	return "", errors.New("could not get outbound IP. Are you connected to the network")
}

func LookupIP(url string) (string, error) {
	ips, err := net.LookupIP(url)
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", nil
	}
	return ips[0].String(), nil
}

func LookupAddress(ip string) (string, error) {
	addr, err := net.LookupAddr(ip)
	if err != nil {
		return "", err
	}
	if len(addr) == 0 {
		return "", nil
	}
	return addr[0], nil
}

func ValidateAddress(domain string) bool {
	RegExp := regexp.MustCompile(`^(([a-zA-Z]{1})|([a-zA-Z]{1}[a-zA-Z]{1})|([a-zA-Z]{1}[0-9]{1})|([0-9]{1}[a-zA-Z]{1})|([a-zA-Z0-9][a-zA-Z0-9-_]{1,61}[a-zA-Z0-9]))\.([a-zA-Z]{2,6}|[a-zA-Z0-9-]{2,30}\.[a-zA-Z]{2,3})$`)
	return RegExp.MatchString(domain)
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// ------------------
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"0123456789"

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GenerateRandomToken(length int) string {
	return stringWithCharset(length, charset)
}

// ------------------
func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// LoadClientTLSCredentials loads a certificate from disk and returns the credentials
func LoadClientTLSCredentials(cacertPaths []string, cacertB64 []byte) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server certificate

	certPool := x509.NewCertPool()

	if len(cacertPaths) > 0 {
		for _, path := range cacertPaths {
			pemServerCA, err := os.ReadFile(path)
			if err != nil || !certPool.AppendCertsFromPEM(pemServerCA) {
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
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}

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
	}

	return config, nil
}
