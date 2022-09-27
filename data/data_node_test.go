package data

import (
	"testing"
	"time"

	"github.com/go-test/deep"
)

func Test_SetNode(t *testing.T) {
	db, _ := NewMemDB(log)
	db.SetNode(nodes[1])
	txn := db.Txn(false)
	raw, err := txn.First("node", "id", nodes[1].Id)
	if err != nil {
		t.Errorf("error occured: %v", err)
	}
	result := raw.(*Node)
	if diff := deep.Equal(result, nodes[1]); diff != nil {
		t.Error(diff)
	}
}

func Test_GetNode(t *testing.T) {
	tests := []struct {
		name         string
		emptyDb      bool
		state        int
		expected     *Node
		expectedList []*Node
	}{
		{name: "nodes in db", emptyDb: false, expected: nodes[0], expectedList: nil},
		{name: "empty db", emptyDb: true, expected: &Node{}, expectedList: nil},

		{name: "list/nodes in db", emptyDb: false, expected: nil, expectedList: nodes},
		{name: "list/empty db", emptyDb: true, expected: nil, expectedList: []*Node{}},

		{name: "list/byState/nodes in db state success 1 node", emptyDb: false, state: 1, expected: nil, expectedList: []*Node{nodes[0]}},
		{name: "list/byState/nodes in db state success 2 nodes", emptyDb: false, state: 2, expected: nil, expectedList: []*Node{nodes[1], nodes[2]}},
		{name: "list/byState/nodes in db no result", emptyDb: false, state: 9, expected: nil, expectedList: []*Node{}},
		{name: "list/byState/nodes in db no result", emptyDb: true, state: 9, expected: nil, expectedList: []*Node{}},
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
func Test_DeleteNode(t *testing.T) {
	db, _ := NewMemDB(log)
	for _, node := range nodes {
		db.SetNode(node)
	}
	for _, node := range nodes {
		db.DeleteNode(node.Id)
	}
	nodes_result := db.GetNodeList()
	if len(nodes_result) > 0 {
		t.Errorf("still some nodes in db, should be empty. Amount of nodes that are left in result: %v", len(nodes_result))
	}
}

func Test_SetNodeTsNow(t *testing.T) {
	db, _ := NewMemDB(log)
	for _, node := range nodes {
		db.SetNode(node)
	}
	tsBefore := time.Now().Unix()
	time.Sleep(time.Second)
	for _, node := range nodes {
		db.SetNodeTsNow(node.Id)
	}
	time.Sleep(time.Second)
	tsAfter := time.Now().Unix()
	nodes_result := db.GetNodeList()
	for _, node := range nodes_result {
		if node.LastSampleTs <= tsBefore || node.LastSampleTs >= tsAfter {
			t.Errorf("timestamp not correct (currently: %v), should be between %v and %v", node.LastSampleTs, tsBefore, tsAfter)
		}
	}
	if len(nodes_result) == 0 {
		t.Errorf("no nodes set in db, %v nodes should be set", len(nodes))
	}
}
