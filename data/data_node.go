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
	"math/rand"
	"time"
)

// SetNode inserts a node in database
func (db *Database) SetNode(node *Node) {
	// Create a write transaction
	txn := db.Txn(true)
	defer txn.Abort()

	err := txn.Insert("node", node)
	if err != nil {
		panic(err)
	}

	// Commit the transaction
	txn.Commit()
}

// SetNodeTsNow sets the timestamp of a node to now.
// Id will select the node.
func (db *Database) SetNodeTsNow(id uint32) {
	txn := db.Txn(true)
	defer txn.Abort()

	node := *db.GetNode(id)
	if node.Id == 0 {
		return
	}

	node.StateChangeTs = time.Now().Unix()
	err := txn.Insert("node", &node)
	if err != nil {
		panic(err)
	}

	// Commit the transaction
	txn.Commit()
}

// DeleteNode deletes a node by its id
func (db *Database) DeleteNode(id uint32) {
	txn := db.Txn(true)
	defer txn.Abort()

	err := txn.Delete("node", db.GetNode(id))
	if err != nil {
		db.log.Debugf("Could not delete Node")
	}
	// Commit the transaction
	txn.Commit()
}

// GetNode returns a node by its id
func (db *Database) GetNode(id uint32) *Node {
	txn := db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("node", "id", id)
	if err != nil {
		panic(err)
	}
	if raw == nil {
		return &Node{}
	}
	return raw.(*Node)
}

// GetNodeByName returns a node by its name
func (db *Database) GetNodeByName(name string) *Node {
	txn := db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("node", "name", name)
	if err != nil {
		panic(err)
	}
	if raw == nil {
		return &Node{}
	}
	return raw.(*Node)
}

// GetNodeList returns all nodes
func (db *Database) GetNodeList() []*Node {
	txn := db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("node", "id")
	if err != nil {
		panic(err)
	}
	var nodes []*Node
	for obj := it.Next(); obj != nil; obj = it.Next() {
		nodes = append(nodes, obj.(*Node))
	}
	return nodes
}

// GetNodeListByState get all nodes with a specific state
func (db *Database) GetNodeListByState(byState int) []*Node {
	txn := db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("node", "id")
	if err != nil {
		panic(err)
	}
	var nodes []*Node
	for obj := it.Next(); obj != nil; obj = it.Next() {
		if obj.(*Node).State == byState {
			nodes = append(nodes, obj.(*Node))
		}
	}
	return nodes
}

// GetRandomNodeListByState get a specific number of random nodes by state
// The without defines which nodes should be excluded from the selection
func (db *Database) GetRandomNodeListByState(byState int, amountOfNodes int, without ...uint32) []*Node {
	nodes := db.GetNodeListByState(byState)

	if len(nodes) == 0 {
		return nodes
	}

	// remove "without" nodes from the result
	for _, id := range without {
		nodes = removeNodeByIdFromSlice(nodes, id)
	}

	if len(nodes) == 0 {
		return nodes
	}

	// shuffle the nodes-array randomly
	nodes = shuffleNodes(nodes)

	// check if the list is already smaller or equal to the requested amount
	if len(nodes) <= amountOfNodes {
		return nodes
	}

	return nodes[:amountOfNodes]
}

// shuffle a slice of nodes
func shuffleNodes(nodes []*Node) []*Node {
	rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})
	return nodes
}

// Remove a node from a given slice by node.id
func removeNodeByIdFromSlice(nodes []*Node, id uint32) []*Node {
	rmIndex := -1

	// get the index of node with node.id
	for index, node := range nodes {
		if node.Id == id {
			rmIndex = index
			break
		}
	}

	// replacement
	if rmIndex != -1 {
		nodes[rmIndex] = nodes[len(nodes)-1]
		nodes[len(nodes)-1] = nil
		nodes = nodes[:len(nodes)-1]
	}

	return nodes
}
