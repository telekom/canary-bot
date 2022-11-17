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

package api

import (
	"context"
	"time"

	"github.com/telekom/canary-bot/data"

	apiv1 "github.com/telekom/canary-bot/proto/api/v1"

	connect "github.com/bufbuild/connect-go"
	"go.uber.org/zap"
)

// Api implements the protobuf interface
type Api struct {
	data   data.Database
	config *Configuration
	log    *zap.SugaredLogger
}

type Configuration struct {
	NodeName       string
	Address        string
	Port           int64
	Tokens         []string
	DebugGrpc      bool
	ServerCertPath string
	ServerKeyPath  string
	ServerCert     []byte
	ServerKey      []byte
	CaCertPath     []string
	CaCert         []byte
}

// List all measured samples
func (b *Api) ListSamples(ctx context.Context, req *connect.Request[apiv1.ListSampleRequest]) (*connect.Response[apiv1.ListSampleResponse], error) {
	samples := []*apiv1.Sample{}

	for _, sample := range b.data.GetSampleList() {
		samples = append(samples, &apiv1.Sample{
			From:  sample.From,
			To:    sample.To,
			Type:  data.SampleName[sample.Key],
			Value: sample.Value,
			Ts:    time.Unix(sample.Ts, 0).String(),
		})
	}

	return connect.NewResponse(&apiv1.ListSampleResponse{
		Samples: samples,
	}), nil
}

// List all known nodes in mesh
func (b *Api) ListNodes(ctx context.Context, req *connect.Request[apiv1.ListNodesRequest]) (*connect.Response[apiv1.ListNodesResponse], error) {
	nodes := []string{b.config.NodeName}

	for _, node := range b.data.GetNodeList() {
		nodes = append(nodes, node.Name)
	}

	return connect.NewResponse(&apiv1.ListNodesResponse{
		Nodes: nodes,
	}), nil
}
