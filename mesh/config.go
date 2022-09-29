package mesh

import (
	h "canary-bot/helper"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type RoutineConfiguration struct {
	// Timeout for every grpc request
	RequestTimeout time.Duration

	// Join config
	JoinInterval time.Duration

	// Ping config
	PingInterval       time.Duration
	PingRetryAmount    int
	PingRetryDelay     time.Duration
	TimeoutRetryPause  time.Duration
	TimeoutRetryAmount int
	TimeoutRetryDelay  time.Duration

	// Node discovery
	BroadcastToAmount int

	// Push samples
	PushSampleInterval    time.Duration
	PushSampleToAmount    int
	PushSampleRetryAmount int
	PushSampleRetryDelay  time.Duration

	// Clean samples
	CleanSampleInterval time.Duration
	SampleMaxAge        time.Duration

	// Sample: RTT
	RttInterval time.Duration
}

type SetupConfiguration struct {
	// remote target
	Targets []string

	// local config
	Name          string
	JoinAddress   string
	ListenAddress string
	ListenPort    int64

	// API
	ApiPort int64

	// TLS server side
	ServerCertPath string
	ServerKeyPath  string
	ServerCert     []byte
	ServerKey      []byte
	// TLS client side
	CaCertPath []string
	CaCert     []byte

	//Auth API
	Tokens []string

	//Logging
	Debug     bool
	DebugGrpc bool
}

func StandardProductionRoutineConfig() *RoutineConfiguration {
	return &RoutineConfiguration{
		RequestTimeout:        time.Second * 3,
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
	}
}

func (setupConfig *SetupConfiguration) setDefaults(logger *zap.SugaredLogger) {
	// get IP of this node if no bind-address and/or domain set
	externalIP, err := h.ExternalIP()
	if setupConfig.ListenAddress == "" {
		if err != nil {
			logger.Fatalln("Could not get external IP, please use bind-address flag")
		}
		logger.Info("Bind-address flag not set - using external interface IP")
		setupConfig.ListenAddress = externalIP
	}

	if setupConfig.JoinAddress == "" {
		if err != nil {
			logger.Fatalln("Could not get external IP, please use join-address flag")
		}
		logger.Info("JoinAddress flag not set - using external interface IP")
		setupConfig.JoinAddress = externalIP + strconv.FormatInt(setupConfig.ListenPort, 10)
	}

	// get tokens; generate one if none is set
	if len(setupConfig.Tokens) == 0 {
		newToken := h.GenerateRandomToken(64)
		setupConfig.Tokens = append(setupConfig.Tokens, newToken)
		logger.Infow("No API tokens set - generated new token", "token", newToken)
		logger.Info("Please use and set new token as environment variable")

	}
}

func (setupConfig *SetupConfiguration) checkDefaults(logger *zap.SugaredLogger) {
	// check TLS mode
	if setupConfig.CaCert != nil || len(setupConfig.CaCertPath) > 0 {
		if (setupConfig.ServerCert != nil || setupConfig.ServerCertPath != "") &&
			(setupConfig.ServerKey != nil || setupConfig.ServerKeyPath != "") {
			logger.Info("Mesh is set to mutal TLS mode")
		} else {
			logger.Info("Mesh is set to edge-terminated TLS mode")
		}
	} else {
		logger.Info("Mesh is set to unsecure mode - no TLS used")
	}

	// validate if name is set
	if setupConfig.Name == "" {
		logger.Fatalln("Please set a name for the creating node. It has to be unique in the mesh.")
	}

	// validate if target(s) is/are set
	if len(setupConfig.Targets) == 0 {
		logger.Fatal("No target(s) set, please set to join a (future) mesh")
	}
}
