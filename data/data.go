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
	l "log"
	"strconv"
	"time"

	h "github.com/telekom/canary-bot/helper"
	meshv1 "github.com/telekom/canary-bot/proto/mesh/v1"

	"github.com/hashicorp/go-memdb"
	"go.uber.org/zap"
)

// Sample keys
const (
	State      = 1
	RttTotal   = 2
	RttRequest = 3
)

// SampleName holds the mapping of the sample keys
var SampleName = map[int64]string{
	State:      "state",
	RttTotal:   "rtt_total",
	RttRequest: "rtt_request",
}

// Database that is used by the mesh.
// It will hold node and sample data.
// It is an in-memory database. A logger
// is provided.
type Database struct {
	*memdb.MemDB
	log *zap.SugaredLogger
}

// Node represents a member of the canary mesh.
type Node struct {
	// Id is the unique ID of the node
	Id uint32
	// Name is the unique name of the node
	Name string
	// Target defines the address:port of the node
	Target string
	// State is the state of the node
	State int
	// StateChangeTs is the timestamp when the state changed last time
	StateChangeTs int64
}

// Sample represents a measurement of the canary mesh.
type Sample struct {
	Id   uint32
	From string
	To   string
	// Key is the sample name
	Key int64
	// Value is the measurement value
	Value string
	Ts    int64
}

// NewMemDB Will create an in-memory database and a logger.
// The database will be created with 2 schemas: node, sample
func NewMemDB(logger *zap.SugaredLogger) (Database, error) {
	defer logger.Sync()

	// 2 tables: node, sample
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"node": {
				Name: "node",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.UintFieldIndex{Field: "Id"},
					},
					"name": {
						Name:    "name",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
					"target": {
						Name:    "target",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Target"},
					},
					"state": {
						Name:    "state",
						Unique:  false,
						Indexer: &memdb.IntFieldIndex{Field: "State"},
					},
					"stateChangeTs": {
						Name:         "stateChangeTs",
						Unique:       false,
						AllowMissing: true,
						Indexer:      &memdb.IntFieldIndex{Field: "StateChangeTs"},
					},
				},
			},
			"sample": {
				Name: "sample",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:         "id",
						Unique:       true,
						AllowMissing: false,
						Indexer:      &memdb.UintFieldIndex{Field: "Id"},
					},
					"from": {
						Name:         "from",
						Unique:       false,
						AllowMissing: false,
						Indexer:      &memdb.StringFieldIndex{Field: "From"},
					},
					"to": {
						Name:         "to",
						Unique:       false,
						AllowMissing: false,
						Indexer:      &memdb.StringFieldIndex{Field: "To"},
					},
					"key": {
						Name:         "key",
						Unique:       false,
						AllowMissing: false,
						Indexer:      &memdb.IntFieldIndex{Field: "Key"},
					},
					"value": {
						Name:         "value",
						Unique:       false,
						AllowMissing: false,
						Indexer:      &memdb.StringFieldIndex{Field: "Value"},
					},
					"ts": {
						Name:         "ts",
						Unique:       false,
						AllowMissing: false,
						Indexer:      &memdb.IntFieldIndex{Field: "Ts"},
					},
				},
			},
		},
	}
	// Create new database
	db, err := memdb.NewMemDB(schema)
	return Database{db, logger}, err
}

// Convert a given database node to a mesh node
func (n *Node) Convert() *meshv1.Node {
	return &meshv1.Node{
		Name:   n.Name,
		Target: n.Target,
	}
}

// Convert a given mesh node to a database node
// with a given state of the node
func Convert(n *meshv1.Node, state int) *Node {
	id, err := h.Hash(n.Target)
	if err != nil {
		l.Printf("Could not get the hash value of the ID, please check the hash function")
	}

	return &Node{
		Id:            id,
		Name:          n.Name,
		Target:        n.Target,
		State:         state,
		StateChangeTs: time.Now().Unix(),
	}
}

// GetId returns the id of a database node.
func GetId(n *Node) uint32 {
	id, err := h.Hash(n.Target)
	if err != nil {
		l.Printf("Could not get the hash value of the ID, please check the hash function")
	}

	return id
}

// GetSampleId returns the id of a given sample.
func GetSampleId(p *Sample) uint32 {
	id, err := h.Hash(p.From + p.To + strconv.FormatInt(p.Key, 10))
	if err != nil {
		l.Printf("Could not get the hash value of the sample, please check the hash function")
	}
	return id
}
