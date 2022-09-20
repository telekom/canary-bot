/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"strconv"

	"gitlab.devops.telekom.de/caas/canary-bot/api"
	"gitlab.devops.telekom.de/caas/canary-bot/data"
	mesh "gitlab.devops.telekom.de/caas/canary-bot/mesh"

	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	h "gitlab.devops.telekom.de/caas/canary-bot/helper"

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

	logger.Debugf("CLI settings: %+v", set)

	// validate if node name for this node is set
	if set.Name == defaults.Name {
		logger.Fatalln("Please set a name for the creating node. It has to be unique in the mesh.")
	}

	// validate if target(s) is/are set
	if len(set.Targets) == 0 {
		logger.Fatal("No target(s) set, please set to join a (future) mesh")
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
	if set.JoinAddress == "" {
		logger.Info("JoinAddress flag not set - getting external IP automatically")
		if err != nil {
			logger.Fatalln("Could not get external IP, please use join-address flag")
		} else {
			set.JoinAddress = externalIP + strconv.FormatInt(set.ListenPort, 10)
		}
	}

	// enable pprof for profiling
	if set.DebugProfile {
		go func() {
			logger.Info("Starting go debugging profiler pprof on port 6060")
			logger.Debug(http.ListenAndServe(set.ListenAddress+":6060", nil))
		}()
	}

	// get tokens; generate one if none is set
	if len(set.Tokens) == 0 {
		newToken := h.GenerateRandomToken(64)
		set.Tokens = append(set.Tokens, newToken)
		logger.Infow("No API tokens set - generated new token", "token", newToken)
		logger.Info("Please use and set new token as environment variable")

	}

	// prepare in-memory database
	mData, err := data.NewMemDB(logger.Named("database"))
	if err != nil {
		logger.Fatalf("Could not create Memory Database (MemDB) - Error: %+v", err)
	}

	// prepare mesh
	_, err = mesh.NewMesh(mData, &mesh.Config{
		JoinInterval:          time.Second * 3,
		PingInterval:          time.Second * 10,
		PingRetryAmount:       3,
		PingRetryDelay:        time.Second * 5,
		TimeoutRetryPause:     time.Minute,
		TimeoutRetryAmount:    3,
		TimeoutRetryDelay:     time.Second * 30,
		BroadcastToAmount:     2,
		PushSampleInterval:    time.Second * 5,
		PushSampleToAmount:    2,
		PushSampleRetryAmount: 2,
		PushSampleRetryDelay:  time.Second * 10,
		CleanSampleInterval:   time.Minute,
		SampleMaxAge:          time.Hour * 24,

		RttInterval: time.Second * 3,

		StartupSettings: set,
	}, logger)

	if err != nil {
		logger.Fatalf("Could not create Mesh - Error: %+v", err)
	}

	// check TLS mode
	if set.CaCert != nil || len(set.CaCertPath) > 0 {
		if (set.ServerCert != nil || set.ServerCertPath != "") && (set.ServerKey != nil || set.ServerKeyPath != "") {
			logger.Info("Mesh is set to mutal TLS mode")
		} else {
			logger.Info("Mesh is set to edge-terminated TLS mode")
		}
	} else {
		logger.Info("Mesh is set to unsecure mode - no TLS used")
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
		Debug:          false,
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

	// Logging mode
	cmd.Flags().BoolVar(&set.Debug, "debug", defaults.Debug, "Set logging to debug mode")
	cmd.Flags().BoolVar(&set.DebugGrpc, "debug-grpc", defaults.DebugGrpc, "Enable more logging for grpc")
	cmd.Flags().BoolVar(&set.DebugProfile, "debug-pprof", defaults.DebugProfile, "Enable profile debugging on port 6060")
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
