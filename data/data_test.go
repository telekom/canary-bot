package data

import (
	"testing"
	"time"

	meshv1 "gitlab.devops.telekom.de/caas/canary-bot/proto/mesh/v1"

	h "gitlab.devops.telekom.de/caas/canary-bot/helper"

	"github.com/go-test/deep"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger
var nodes []*DbNode

func init() {
	logger, _ := zap.NewDevelopment()
	log = logger.Sugar()

	nodes = []*DbNode{
		{Id: 1, Name: "node_1", Target: "target_1", State: 1, LastSampleTs: 0},
		{Id: 2, Name: "node_2", Target: "target_2", State: 2, LastSampleTs: 0},
		{Id: 3, Name: "node_3", Target: "target_3", State: 2, LastSampleTs: 0},
	}
}

func Test_SetNode(t *testing.T) {
	db, _ := NewMemDB(log)
	db.SetNode(nodes[1])
	txn := db.Txn(false)
	raw, err := txn.First("node", "id", nodes[1].Id)
	if err != nil {
		t.Errorf("error occured: %v", err)
	}
	result := raw.(*DbNode)
	if diff := deep.Equal(result, nodes[1]); diff != nil {
		t.Error(diff)
	}
}

func Test_GetNode(t *testing.T) {
	tests := []struct {
		name         string
		emptyDb      bool
		state        int
		expected     *DbNode
		expectedList []*DbNode
	}{
		{name: "nodes in db", emptyDb: false, expected: nodes[0], expectedList: nil},
		{name: "empty db", emptyDb: true, expected: &DbNode{}, expectedList: nil},

		{name: "list/nodes in db", emptyDb: false, expected: nil, expectedList: nodes},
		{name: "list/empty db", emptyDb: true, expected: nil, expectedList: []*DbNode{}},

		{name: "list/byState/nodes in db state success 1 node", emptyDb: false, state: 1, expected: nil, expectedList: []*DbNode{nodes[0]}},
		{name: "list/byState/nodes in db state success 2 nodes", emptyDb: false, state: 2, expected: nil, expectedList: []*DbNode{nodes[1], nodes[2]}},
		{name: "list/byState/nodes in db no result", emptyDb: false, state: 9, expected: nil, expectedList: []*DbNode{}},
		{name: "list/byState/nodes in db no result", emptyDb: true, state: 9, expected: nil, expectedList: []*DbNode{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _ := NewMemDB(log)
			if !tt.emptyDb {
				// fill db
				for _, node := range nodes {
					db.SetNode(node)
				}
			}
			if tt.expected != nil {
				// GetNode(ByName)
				result := db.GetNode(nodes[0].Id)
				if diff := deep.Equal(result, tt.expected); diff != nil {
					t.Error(diff)
				}
				result = db.GetNodeByName(nodes[0].Name)
				if diff := deep.Equal(result, tt.expected); diff != nil {
					t.Error(diff)
				}
			} else {
				if tt.state == 0 {
					// GetNodeList
					result := db.GetNodeList()
					if len(result) != len(tt.expectedList) {
						t.Errorf("the amount returned nodes is incorrect: %v but expected %v", len(result), len(tt.expectedList))
					}
					for i, node := range result {
						if diff := deep.Equal(node, tt.expectedList[i]); diff != nil {
							t.Error(diff)
						}

					}

				} else {
					// GetNodeListByState
					result := db.GetNodeListByState(tt.state)
					if len(result) != len(tt.expectedList) {
						t.Errorf("the amount returned nodes is incorrect: %v but expected %v", len(result), len(tt.expectedList))
					}
					for i, node := range result {
						if diff := deep.Equal(node, tt.expectedList[i]); diff != nil {
							t.Error(diff)
						}

					}

				}

			}
		})
	}
}

func Test_Convert(t *testing.T) {
	tests := []struct {
		name           string
		inputMeshNode  *meshv1.Node
		state          int
		expectedDbNode *DbNode

		inputDbNode      *DbNode
		expectedMeshNode *meshv1.Node
	}{
		{
			name: "MeshNode to DbNode",
			inputMeshNode: &meshv1.Node{
				Name:   "test",
				Target: "tegraT",
			},
			state: 1,
			expectedDbNode: &DbNode{
				Id:           h.Hash("tegraT"),
				Name:         "test",
				State:        1,
				Target:       "tegraT",
				LastSampleTs: 0,
			},
		},
		{
			name: "DbNode to MeshNode",
			inputDbNode: &DbNode{
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
				if diff := deep.Equal(result, tt.expectedDbNode); diff != nil {
					t.Error(diff)
				}
			} else {
				result := tt.inputDbNode.Convert()
				if diff := deep.Equal(result, tt.expectedMeshNode); diff != nil {
					t.Error(diff)
				}
			}
		})
	}
}

// func GetId(n *DbNode) uint32 {
// 	return h.Hash(n.Target)
// }

// func GetSampleId(p *Sample) uint32 {
// 	return h.Hash(p.From + p.To + strconv.FormatInt(p.Key, 10))
// }

// func (db *Database) SetNodeTsNow(id uint32) {}

// func (db *Database) SetSample(sample *Sample) {}

// func (db *Database) GetSample(id uint32) *Sample {}

// func (db *Database) DeleteSample(id uint32) {}

// func (db *Database) GetSampleTs(id uint32) int64 {}

// func (db *Database) GetSampleList() []*Sample {}

// func (db *Database) DeleteNode(id uint32) {}
