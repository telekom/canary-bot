/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"canary-bot/api"
	"canary-bot/data"
	mesh "canary-bot/mesh"
	meshv1 "canary-bot/proto/mesh/v1"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	envPrefix = "MESH"
)

var cmd = &cobra.Command{
	Use:   "cbot",
	Short: "Canary Bot collecting & distributing data in a mesh",
	Long: `Initialize multiple Canary Bots to create a Canary Mesh.
All bots will gather data and spread it in the mesh.

Features:
 - http based communication (gRPC)
 - partly based on the gossip protocol standard
 - mutal TLS enabled

All flags can be set by environment variables. Please use the Prefix ` + envPrefix + `
Multiple targets can be set comma-separated.
`,
	PersistentPreRun: initSettings,
	Run:              run,
}

var defaults mesh.Settings
var set mesh.Settings

func run(cmd *cobra.Command, args []string) {
	// prepare logging
	var zapLogger *zap.Logger
	var err error
	if set.Debug {
		zapLogger, err = zap.NewDevelopment()
	} else {
		zapLogger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf("Could not start Logging - error: %+v", err)
	}
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	// validate if node name for this node is set
	if set.Name == defaults.Name {
		logger.Fatalln("Please set a name for the creating node. It has to be unique in the mesh.")
	}

	// prepare targets
	var joinNodes []*meshv1.Node
	for _, target := range set.Targets {
		joinNodes = append(joinNodes, &meshv1.Node{
			Name:   "",
			Target: target,
		},
		)
	}

	// prepare in-memory database
	mData, err := data.NewMemDB(logger.Named("database"))
	if err != nil {
		logger.Fatalf("Could not create Memory Database (MemDB) - Error: %+v", err)
	}

	// prepare mesh
	m, err := mesh.NewMesh(mData, &mesh.Config{
		PingInterval:         time.Second * 10,
		PingRetryAmount:      3,
		PingRetryDelay:       time.Second * 5,
		TimeoutRetryPause:    time.Second * 5,
		TimeoutRetryAmount:   3,
		TimeoutRetryDelay:    time.Second * 5,
		BroadcastToAmount:    2,
		PushProbeInterval:    time.Second * 5,
		PushProbeToAmount:    1,
		PushProbeRetryAmount: 2,
		PushProbeRetryDelay:  time.Second * 1,

		StartupSettings: set,
	}, logger)

	if err != nil {
		logger.Fatalf("Could not create Mesh - Error: %+v", err)
	}

	// check if this node ist first in mesh otherwise join
	if len(set.Targets) == 0 {
		logger.Info("No target set - first node in mesh")
	} else {
		isNameUniqueInMesh, err := m.Join(joinNodes)
		if err != nil {
			logger.Fatalf("Could not join Mesh - Error: %+v", err)
		}
		if !isNameUniqueInMesh {
			logger.Fatal("The name is not unique in the mesh, please choose another one.")
			// TODO generate random node name
		}
	}

	// start API
	if err = api.NewApi(mData, set, logger.Named("api")); err != nil {
		logger.Fatal("Could not start API - Error: %+v", err)
	}

	incomingSigs := make(chan os.Signal, 1)
	signal.Notify(incomingSigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)
	select {
	case <-incomingSigs:
	}
}

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	defaults = mesh.Settings{
		Targets:        []string{},
		Name:           "",
		ListenAddress:  "",
		ListenPort:     8081,
		ApiPort:        8080,
		ServerCertPath: "",
		ServerKeyPath:  "",
		ServerCert:     nil,
		ServerKey:      nil,
		CaCertPath:     "",
		CaCert:         nil,
		Debug:          false,
	}

	// Targets for joining
	cmd.Flags().StringSliceVarP(&set.Targets, "target", "t", defaults.Targets, "Comma-seperated or multi-flag list of targets for joining the mesh.\nFormat: [IP|ADDRESS]:PORT")

	// ssttings for this node
	cmd.Flags().StringVarP(&set.Name, "name", "n", defaults.Name, "Name of the node, has to be unique in mesh")
	cmd.Flags().StringVarP(&set.ListenAddress, "address", "a", defaults.ListenAddress, "Domain (or IP) of this node")
	cmd.Flags().Int64VarP(&set.ListenPort, "listen-port", "l", defaults.ListenPort, "Listening port of this node")

	// API
	cmd.Flags().Int64VarP(&set.ApiPort, "api-port", "p", defaults.ApiPort, "API port of this node")

	// TLS server side
	cmd.Flags().StringVar(&set.ServerCertPath, "server-cert-path", defaults.ServerCertPath, "Path to the server cert file e.g. cert/server-cert.pem - use with server-key-path to enable TLS")
	cmd.Flags().StringVar(&set.ServerKeyPath, "server-key-path", defaults.ServerKeyPath, "Path to the server key file e.g. cert/server-key.pem - use with server-cert-path to enable TLS")
	cmd.Flags().BytesBase64Var(&set.ServerCert, "server-cert", defaults.ServerCert, "Base64 encoded server cert, use with server-key to enable TLS")
	cmd.Flags().BytesBase64Var(&set.ServerKey, "server-key", defaults.ServerKey, "Base64 encoded server key, use with server-cert to enable TLS")

	// TLS client side
	cmd.Flags().StringVar(&set.CaCertPath, "ca-cert-path", defaults.CaCertPath, "Path to ca cert file to enable TLS")
	cmd.Flags().BytesBase64Var(&set.CaCert, "ca-cert", defaults.CaCert, "Base64 encoded ca cert to enable TLS")

	// Logging mode
	cmd.Flags().BoolVar(&set.Debug, "debug", defaults.Debug, "Set logging to debug mode")

}

func initSettings(cmd *cobra.Command, args []string) {
	v := viper.New()
	// Set environment variable prefix
	v.SetEnvPrefix(envPrefix)
	// Bind to environment variables
	v.AutomaticEnv()
	// Bind the current command's flags to viper
	bindEnvToFlags(cmd, v)
}

func bindEnvToFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Mapping Flag with "-" to uppercase env with "_" --listen-port to <PREFIX>_LISTEN_PORT
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			v.BindEnv(f.Name, envPrefix+"_"+envVarSuffix)
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.GetString(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
