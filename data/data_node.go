package data

import (
	"math/rand"
	"time"
)

// Insert node in database
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

// Set timestamp of a node to now.
// The node will be selected by id.
func (db *Database) SetNodeTsNow(id uint32) {
	txn := db.Txn(true)
	defer txn.Abort()

	node := *db.GetNode(id)
	if node.Id == 0 {
		return
	}

	node.LastSampleTs = time.Now().Unix()
	err := txn.Insert("node", &node)
	if err != nil {
		panic(err)
	}

	// Commit the transaction
	txn.Commit()
}

// Delete a node by its id
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

// Get a node by its id
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

//G Get a node by its name
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

// Get all nodes
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

// Get all nodes with a specific state
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

// Get a specific amount of random nodes by state
// Use a list of node ids (without) that should be removed from the list
func (db *Database) GetRandomNodeListByState(byState int, amountOfNodes int, without ...uint32) []*Node {
	nodes := db.GetNodeListByState(byState)

	if len(nodes) == 0 {
		return nodes
	}

	// remove "without" nodes from result list
	for _, id := range without {
		nodes = removeNodeByIdFromSlice(nodes, id)
	}

	if len(nodes) == 0 {
		return nodes
	}

	// shuffel nodes array randomly
	nodes = shuffleNodes(nodes)

	// check if list is already smaller or equalt to requested amout
	if len(nodes) <= amountOfNodes {
		return nodes
	}

	return nodes[:amountOfNodes]
}

// shuffel a slice of nodes
func shuffleNodes(nodes []*Node) []*Node {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})
	return nodes
}

// Remove a node from a given slice by node.id
func removeNodeByIdFromSlice(nodes []*Node, id uint32) []*Node {
	rmIndex := 0

	// get the index of node with node.id
	for index, node := range nodes {
		if node.Id == id {
			rmIndex = index
			break
		}
	}

	// replacement
	nodes[rmIndex] = nodes[len(nodes)-1]
	nodes[len(nodes)-1] = nil
	nodes = nodes[:len(nodes)-1]

	return nodes
}
