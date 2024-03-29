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

func Test_shuffleNodes(t *testing.T) {
	maxTries := 20
	firstId := nodes[0].Id
	for i := 0; i < maxTries; i++ {
		shuffledNodes := shuffleNodes(nodes)
		if shuffledNodes[0].Id != firstId {
			return
		}
	}

	t.Errorf("Tried to shuffel nodes %v times. No effect.", maxTries)
}

func Test_removeNodeByIdFromSlice(t *testing.T) {
	testId := nodes[3].Id

	var nodesCopy []*Node
	for _, node := range nodes {
		nodesCopy = append(nodesCopy, node)
	}
	nodesCopy = removeNodeByIdFromSlice(nodesCopy, testId)
	for _, node := range nodesCopy {
		if node.Id == testId {
			t.Errorf("Id is still in slice. ID: %+v", testId)
		}
	}
}

func Test_GetRandomNodeListByState(t *testing.T) {
	tests := []struct {
		name                  string
		amountOfNodesInDb     int
		amountOfRandomNodes   int
		expectedAmountOfNodes int
		expectedWithout       int
	}{
		{name: "no nodes in db - no requested", amountOfNodesInDb: 0, amountOfRandomNodes: 0, expectedAmountOfNodes: 0},
		{name: "requested is higher than nodes in db - zero nodes in db", amountOfNodesInDb: 0, amountOfRandomNodes: 3, expectedAmountOfNodes: 0},
		{name: "requested is higher than nodes in db", amountOfNodesInDb: 2, amountOfRandomNodes: 3, expectedAmountOfNodes: 2},
		{name: "requested is less then amount of nodes in db", amountOfNodesInDb: 5, amountOfRandomNodes: 3, expectedAmountOfNodes: 3},
		{name: "requested is same as amount of nodes in db", amountOfNodesInDb: 3, amountOfRandomNodes: 3, expectedAmountOfNodes: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _ := NewMemDB(log)
			for a := 0; a < tt.amountOfNodesInDb; a++ {
				// fill db
				db.SetNode(nodesSameState[a])
			}

			result := db.GetRandomNodeListByState(0, tt.amountOfRandomNodes)
			if len(result) != tt.expectedAmountOfNodes {
				t.Errorf("The amount of expected randome nodes is not right. Expected amount: %+v, result amount %+v", tt.expectedAmountOfNodes, len(result))
			}

			result = db.GetRandomNodeListByState(0, tt.amountOfRandomNodes, nodesSameState[0].Id)
			for _, node := range result {
				if node.Id == nodesSameState[0].Id {
					t.Errorf("The 'without' node(s) are still present in random nodes array")
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
		if node.StateChangeTs <= tsBefore || node.StateChangeTs >= tsAfter {
			t.Errorf("timestamp not correct (currently: %v), should be between %v and %v", node.StateChangeTs, tsBefore, tsAfter)
		}
	}
	if len(nodes_result) == 0 {
		t.Errorf("no nodes set in db, %v nodes should be set", len(nodes))
	}
}
