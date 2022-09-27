package data

import (
	h "canary-bot/helper"
	meshv1 "canary-bot/proto/mesh/v1"
	"testing"
	"time"

	"github.com/go-test/deep"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger
var nodes []*Node
var samples []*Sample

func init() {
	logger, _ := zap.NewDevelopment()
	log = logger.Sugar()

	nodes = []*Node{
		{Id: 1, Name: "node_1", Target: "target_1", State: 1, LastSampleTs: 0},
		{Id: 2, Name: "node_2", Target: "target_2", State: 2, LastSampleTs: 0},
		{Id: 3, Name: "node_3", Target: "target_3", State: 2, LastSampleTs: 0},
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
				Id:           h.Hash("tegraT"),
				Name:         "test",
				State:        1,
				Target:       "tegraT",
				LastSampleTs: 0,
			},
		},
		{
			name: "Node to MeshNode",
			inputNode: &Node{
				Id:           h.Hash("tegraT"),
				Name:         "test",
				State:        12,
				Target:       "tegraT",
				LastSampleTs: time.Now().Unix(),
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
				if diff := deep.Equal(result, tt.expectedNode); diff != nil {
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
			expectedId: h.Hash("tegraT"),
		},
		{
			name:       "Node without target",
			node:       &Node{},
			expectedId: h.Hash(""),
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
				To:   "Gose",
				Key:  1,
			},
			expectedId: h.Hash("EagleGose1"),
		},
		{
			name:       "Empty samples",
			sample:     &Sample{},
			expectedId: h.Hash("0"),
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
