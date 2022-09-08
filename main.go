/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"canary-bot/api"
	"canary-bot/data"
	mesh "canary-bot/mesh"
	meshv1 "canary-bot/proto/mesh/v1"
	"strconv"

	h "canary-bot/helper"
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
	// environment variables prefix
	set.EnvPrefix = envPrefix
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

	// get IP of this node if no bind-address and/or domain set
	externalIP, err := h.ExternalIP()
	if set.ListenAddress == "" {
		logger.Info("Bind-address flag not set - getting external IP automatically")
		if err != nil {
			logger.Fatalln("Could not get external IP, please use bind-address flag")
		} else {
			set.ListenAddress = externalIP
		}
	}
	if set.Domain == "" {
		logger.Info("Domain flag not set - getting external IP automatically")
		if err != nil {
			logger.Fatalln("Could not get external IP, please use domain flag")
		} else {
			set.Domain = externalIP + strconv.FormatInt(set.ListenPort, 10)
		}
	}

	// get tokens; generate one if none is set
	if len(set.Tokens) == 0 {
		newToken := h.GenerateRandomToken(64)
		set.Tokens = append(set.Tokens, newToken)
		logger.Infow("No API tokens set - generated new token", "token", newToken)
		logger.Info("Please use and set new token as environment variable")

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
		PingInterval:          time.Second * 10,
		PingRetryAmount:       3,
		PingRetryDelay:        time.Second * 5,
		TimeoutRetryPause:     time.Second * 5,
		TimeoutRetryAmount:    3,
		TimeoutRetryDelay:     time.Second * 5,
		BroadcastToAmount:     2,
		PushSampleInterval:    time.Second * 5,
		PushSampleToAmount:    1,
		PushSampleRetryAmount: 2,
		PushSampleRetryDelay:  time.Second * 1,

		StartupSettings: set,
	}, logger)

	if err != nil {
		logger.Fatalf("Could not create Mesh - Error: %+v", err)
	}

	// check TLS mode
	if set.CaCert != nil || set.CaCertPath != "" {
		if (set.ServerCert != nil || set.ServerCertPath != "") && (set.ServerKey != nil || set.ServerKeyPath != "") {
			logger.Info("Mesh is set to mutal TLS mode")
		} else {
			logger.Info("Mesh is set to edge-terminated TLS mode")
		}
	} else {
		logger.Info("Mesh is set to unsecure mode - no TLS used")
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
	if err = api.NewApi(mData, &set, logger.Named("api")); err != nil {
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
		Domain:         "",
		ListenAddress:  "",
		ListenPort:     8081,
		ApiPort:        8080,
		ServerCertPath: "",
		ServerKeyPath:  "",
		ServerCert:     nil,
		ServerKey:      nil,
		CaCertPath:     "",
		CaCert:         nil,
		Tokens:         []string{},
		Debug:          false,
	}

	// Targets for joining
	cmd.Flags().StringSliceVarP(&set.Targets, "target", "t", defaults.Targets, "Comma-seperated or multi-flag list of targets for joining the mesh.\nFormat: [IP|ADDRESS]:PORT")

	// ssttings for this node
	cmd.Flags().StringVarP(&set.Name, "name", "n", defaults.Name, "Name of the node, has to be unique in mesh (mandatory)")
	cmd.Flags().StringVarP(&set.ListenAddress, "bind-address", "a", defaults.ListenAddress, "Address or IP the server of the node will bind to; eg. 0.0.0.0, localhost (default outbound IP of the network interface)")
	cmd.Flags().Int64VarP(&set.ListenPort, "bind-port", "b", defaults.ListenPort, "Listening port of this node")
	cmd.Flags().StringVar(&set.Domain, "domain", defaults.Domain, "Domain of this node; nodes in the mesh will use the domain to connect; eg. test.de, localhost (default outbound IP of the network interface)")

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

	// Auth API
	cmd.Flags().StringSliceVar(&set.Tokens, "token", defaults.Targets, "Comma-seperated or multi-flag list of tokens to protect the sample data API. (optional)")

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
