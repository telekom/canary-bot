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

package mesh

import (
	"strconv"
	"time"

	h "github.com/telekom/canary-bot/helper"

	"go.uber.org/zap"
)

// Configuration for the timer- and channelRoutines
type RoutineConfiguration struct {
	// Timeout for every grpc request
	RequestTimeout time.Duration

	// Join config
	JoinInterval time.Duration

	// Ping config
	PingInterval    time.Duration
	PingRetryAmount int
	PingRetryDelay  time.Duration

	// Node discovery
	BroadcastToAmount int

	// Push samples
	PushSampleInterval    time.Duration
	PushSampleToAmount    int
	PushSampleRetryAmount int
	PushSampleRetryDelay  time.Duration

	// Clean nodes & samples
	CleanupInterval time.Duration
	CleanupMaxAge   time.Duration

	// Sample: RTT
	RttInterval time.Duration
}

// Configuration how the bot can connect to the mesh etc.
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

	// Clean nodes & samples
	CleanupNodes   bool
	CleanupSamples bool

	//Logging
	Debug     bool
	DebugGrpc bool
}

// Use standard configuration parameters for your production
func StandardProductionRoutineConfig() *RoutineConfiguration {
	return &RoutineConfiguration{
		RequestTimeout:        time.Second * 3,
		JoinInterval:          time.Second * 3,
		PingInterval:          time.Second * 10,
		PingRetryAmount:       3,
		PingRetryDelay:        time.Second * 5,
		BroadcastToAmount:     2,
		PushSampleInterval:    time.Second * 5,
		PushSampleToAmount:    2,
		PushSampleRetryAmount: 2,
		PushSampleRetryDelay:  time.Second * 10,
		CleanupInterval:       time.Minute,
		CleanupMaxAge:         time.Hour * 24,

		RttInterval: time.Second * 3,
	}
}

// Default setter method
// - get external IP as listenAddress & joinAdress
// - generate API token
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
		setupConfig.JoinAddress = externalIP + ":" + strconv.FormatInt(setupConfig.ListenPort, 10)
	}

	// get tokens; generate one if none is set
	if len(setupConfig.Tokens) == 0 {
		newToken := h.GenerateRandomToken(64)
		setupConfig.Tokens = append(setupConfig.Tokens, newToken)
		logger.Infow("No API tokens set - generated new token", "token", newToken)
		logger.Info("Please use and set new token as environment variable")

	}
}

// Check the default configuration to discover TLS mode.
// Check if name and target(s) are set in config.
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
