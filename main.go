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

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/telekom/canary-bot/mesh"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	envPrefix = "MESH"
)

var cmd = &cobra.Command{
	Use:   "cbot",
	Short: "Canary Bot collecting & distributing sample data in a mesh",
	Long: `Canary Bot collecting & distributing sample data in a mesh.
	
Initialize multiple Canary Bots to create a Canary Mesh.
All bots will gather data and spread it in the mesh.

Features:
 - auto re-join
 - http based communication (gRPC)
 - partly based on the gossip protocol standard
 - mutal TLS, edge-terminating TLS, no TLS

All flags can be set by environment variables. Please use the Prefix ` + envPrefix + `.
Multiple targets can be set comma-separated.

Example 1
- eade-terminating TLS - 2 targets - different join & listen address for Kubernetes szenario -
cbot --name owl --join-address bird-owl.com:443 --listen-adress localhost --listen-port 8080 --api-port 8081 -t bird-goose.com:443 -t bird-eagle.net:8080 --ca-cert-path path/to/cert.cer

Example 2
- mutal TLS - 2 targets - join & listen-address is external IP from network interface
cbot --name swan -t bird-goose.com:443 -t bird-eagle.net:8080 --ca-cert-path path/to/cert.cer --server-cert-path path/to/cert.cer --server-key ZWFzdGVyZWdn 
`,
	PersistentPreRun: initSettings,
	Run:              run,
}

var (
	defaults mesh.SetupConfiguration
	set      mesh.SetupConfiguration
)

// Will create the
func run(cmd *cobra.Command, args []string) {
	mesh.CreateCanaryMesh(mesh.StandardProductionRoutineConfig(), &set)
}

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// The init function is the very first function that ist running,
// even before the main is executed.\n
// The default mesh settings will me set.
// All cmd flags will be defined.
func init() {
	defaults = mesh.SetupConfiguration{
		Targets:        []string{},
		Name:           "",
		JoinAddress:    "",
		ListenAddress:  "",
		ListenPort:     8081,
		ApiPort:        8080,
		ServerCertPath: "",
		ServerKeyPath:  "",
		ServerCert:     nil,
		ServerKey:      nil,
		CaCertPath:     []string{},
		CaCert:         nil,
		Tokens:         []string{},
		CleanupNodes:   false,
		CleanupSamples: false,
		Debug:          false,
		DebugGrpc:      false,
	}

	// Targets for joining
	cmd.Flags().StringSliceVarP(&set.Targets, "target", "t", defaults.Targets, "Comma-seperated or multi-flag list of targets for joining the mesh.\nFormat: [IP|ADDRESS]:PORT")

	// ssttings for this node
	cmd.Flags().StringVarP(&set.Name, "name", "n", defaults.Name, "Name of the node, has to be unique in mesh (mandatory)")
	cmd.Flags().StringVar(&set.ListenAddress, "listen-address", defaults.ListenAddress, "Address or IP the server of the node will bind to; eg. 0.0.0.0, localhost (default outbound IP of the network interface)")
	cmd.Flags().Int64Var(&set.ListenPort, "listen-port", defaults.ListenPort, "Listening port of this node")
	cmd.Flags().StringVar(&set.JoinAddress, "join-address", defaults.JoinAddress, "Address of this node; nodes in the mesh will use the domain to connect; eg. test.de, localhost (default outbound IP of the network interface)")

	// API
	cmd.Flags().Int64VarP(&set.ApiPort, "api-port", "p", defaults.ApiPort, "API port of this node")

	// TLS server side
	cmd.Flags().StringVar(&set.ServerCertPath, "server-cert-path", defaults.ServerCertPath, "Path to the server cert file e.g. cert/server-cert.pem - use with server-key-path to enable TLS")
	cmd.Flags().StringVar(&set.ServerKeyPath, "server-key-path", defaults.ServerKeyPath, "Path to the server key file e.g. cert/server-key.pem - use with server-cert-path to enable TLS")
	cmd.Flags().BytesBase64Var(&set.ServerCert, "server-cert", defaults.ServerCert, "Base64 encoded server cert, use with server-key to enable TLS")
	cmd.Flags().BytesBase64Var(&set.ServerKey, "server-key", defaults.ServerKey, "Base64 encoded server key, use with server-cert to enable TLS")

	// TLS client side
	cmd.Flags().StringSliceVar(&set.CaCertPath, "ca-cert-path", defaults.CaCertPath, "Path to ca cert file/s to enable TLS")
	cmd.Flags().BytesBase64Var(&set.CaCert, "ca-cert", defaults.CaCert, "Base64 encoded ca cert to enable TLS, support for multiple ca certs by ca-cert-path flag")

	// Auth API
	cmd.Flags().StringSliceVar(&set.Tokens, "token", defaults.Targets, "Comma-seperated or multi-flag list of tokens to protect the sample data API. (optional)")

	// Cleanup database mode
	cmd.Flags().BoolVar(&set.CleanupNodes, "cleanup-nodes", defaults.CleanupNodes, "Enable cleanup mode for nodes (default disabled)")
	cmd.Flags().BoolVar(&set.CleanupSamples, "cleanup-samples", defaults.CleanupSamples, "Enable cleanup mode for measurment samples (default disabled)")

	// Logging mode
	cmd.Flags().BoolVar(&set.Debug, "debug", defaults.Debug, "Set logging to debug mode")
	cmd.Flags().BoolVar(&set.DebugGrpc, "debug-grpc", defaults.DebugGrpc, "Enable more logging for grpc")
}

// Before the run function gets executed
// all env variables need to be loaded (bindEnvToFlags).
// Viper gets initialised.
func initSettings(cmd *cobra.Command, args []string) {
	v := viper.New()
	// Set environment variable prefix
	v.SetEnvPrefix(envPrefix)
	// Bind to environment variables
	v.AutomaticEnv()
	// Bind the current command's flags to viper
	bindEnvToFlags(cmd, v)
}

// Environment variables will be bind to cmd flags,
// if the flag is not set.
// Env vars have a defined prefix.
func bindEnvToFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Mapping Flag with "-" to uppercase env with "_" --listen-port to <PREFIX>_LISTEN_PORT
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			err := v.BindEnv(f.Name, envPrefix+"_"+envVarSuffix)
			if err != nil {
				log.Printf("Could not bind env varibale %v", envPrefix+"_"+envVarSuffix)
			}
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.GetString(f.Name)
			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				log.Printf("Could not apply viper config to flag: %v", v.GetString(f.Name))
			}
		}
	})
}
