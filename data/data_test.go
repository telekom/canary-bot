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

package data

import (
	"testing"
	"time"

	h "github.com/telekom/canary-bot/helper"
	meshv1 "github.com/telekom/canary-bot/proto/mesh/v1"

	"github.com/go-test/deep"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger
var nodes, nodesSameState []*Node
var samples []*Sample

func init() {
	logger, _ := zap.NewDevelopment()
	log = logger.Sugar()

	nodes = []*Node{
		{Id: 1, Name: "node_1", Target: "target_1", State: 1, StateChangeTs: 0},
		{Id: 2, Name: "node_2", Target: "target_2", State: 2, StateChangeTs: 0},
		{Id: 3, Name: "node_3", Target: "target_3", State: 2, StateChangeTs: 0},
		{Id: 4, Name: "node_4", Target: "target_4", State: 3, StateChangeTs: 0},
		{Id: 5, Name: "node_5", Target: "target_5", State: 3, StateChangeTs: 0},
	}

	nodesSameState = []*Node{
		{Id: 1, Name: "node_1", Target: "target_1", State: 0, StateChangeTs: 0},
		{Id: 2, Name: "node_2", Target: "target_2", State: 0, StateChangeTs: 0},
		{Id: 3, Name: "node_3", Target: "target_3", State: 0, StateChangeTs: 0},
		{Id: 4, Name: "node_4", Target: "target_4", State: 0, StateChangeTs: 0},
		{Id: 5, Name: "node_5", Target: "target_5", State: 0, StateChangeTs: 0},
	}

	samples = []*Sample{
		{Id: 1, From: "node_1", To: "node_2", Key: 1, Value: "12345", Ts: 1},
		{Id: 2, From: "node_1", To: "node_3", Key: 1, Value: "454545", Ts: 2},
		{Id: 3, From: "node_2", To: "node_3", Key: 2, Value: "8910", Ts: 3},
	}
}

func Test_NewMemDb(t *testing.T) {
	_, err := NewMemDB(log)
	if err != nil {
		t.Errorf("error during creation of memDb")
	}
}

// helper function
func value(n uint32, _ error) uint32 {
	return n
}
func Test_Convert(t *testing.T) {
	tests := []struct {
		name          string
		inputMeshNode *meshv1.Node
		state         int
		expectedNode  *Node

		inputNode        *Node
		expectedMeshNode *meshv1.Node
	}{
		{
			name: "MeshNode to Node",
			inputMeshNode: &meshv1.Node{
				Name:   "test",
				Target: "tegraT",
			},
			state: 1,
			expectedNode: &Node{
				Id:            value(h.Hash("tegraT")),
				Name:          "test",
				State:         1,
				Target:        "tegraT",
				StateChangeTs: 0,
			},
		},
		{
			name: "Node to MeshNode",
			inputNode: &Node{
				Id:            value(h.Hash("tegraT")),
				Name:          "test",
				State:         12,
				Target:        "tegraT",
				StateChangeTs: time.Now().Unix(),
			},
			expectedMeshNode: &meshv1.Node{
				Name:   "test",
				Target: "tegraT",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.inputMeshNode != nil {
				result := Convert(tt.inputMeshNode, tt.state)
				if result.Id != tt.expectedNode.Id ||
					result.Name != tt.expectedNode.Name ||
					result.Target != tt.expectedNode.Target ||
					result.State != tt.expectedNode.State {
					diff := deep.Equal(result, tt.expectedNode)
					t.Error(diff)
				}
			} else {
				result := tt.inputNode.Convert()
				if diff := deep.Equal(result, tt.expectedMeshNode); diff != nil {

					t.Error(diff)
				}
			}
		})
	}
}

func Test_GetId(t *testing.T) {
	tests := []struct {
		name       string
		node       *Node
		expectedId uint32
	}{
		{
			name:       "Node with target",
			node:       &Node{Target: "tegraT"},
			expectedId: value(h.Hash("tegraT")),
		},
		{
			name:       "Node without target",
			node:       &Node{},
			expectedId: value(h.Hash("")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetId(tt.node)
			if result != tt.expectedId {
				t.Errorf("The result (%v) is not as expected: %v", result, tt.expectedId)
			}
		})
	}
}

func Test_GetSampleId(t *testing.T) {
	tests := []struct {
		name       string
		sample     *Sample
		expectedId uint32
	}{
		{
			name: "Normal sample",
			sample: &Sample{
				From: "Eagle",
				To:   "Goose",
				Key:  1,
			},
			expectedId: value(h.Hash("EagleGoose1")),
		},
		{
			name:       "Empty samples",
			sample:     &Sample{},
			expectedId: value(h.Hash("0")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSampleId(tt.sample)
			if result != tt.expectedId {
				t.Errorf("The result (%v) is not as expected: %v", result, tt.expectedId)
			}
		})
	}
}
