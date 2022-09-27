package data

import (
	h "canary-bot/helper"
	meshv1 "canary-bot/proto/mesh/v1"
	"strconv"

	"github.com/hashicorp/go-memdb"
	"go.uber.org/zap"
)

// Sample keys
const (
	STATE       = 1
	RTT_TOTAL   = 2
	RTT_REQUEST = 3
)

// Sample keys map for mapping back to string
var SampleName = map[int64]string{
	STATE:       "state",
	RTT_TOTAL:   "rtt_total",
	RTT_REQUEST: "rtt_request",
}

// Database that is used by the mesh.
// It will hold node and sample data.
// It is a in-memory database. A logger
// is provided.
type Database struct {
	*memdb.MemDB
	log *zap.SugaredLogger
}

// A database node will have an Id
// witch is a unique integer.
// The name is unique in the node db
// schema. The target defines the address:port
// of the node. The state is the status of
// the node. LastSampleTs will be updated if
// a new sample gets measured. Is used by
// the clean up routine.
type Node struct {
	Id           uint32
	Name         string
	Target       string
	State        int
	LastSampleTs int64
}

// A sample represents a measurement
// sample from a node to another node (e.g. round-trip-time).
// The key is the sample name and the value
// the measurement.
type Sample struct {
	Id    uint32
	From  string
	To    string
	Key   int64
	Value string
	Ts    int64
}

// Will create a in-memory database and
// a looger. The database will be created with
// 2 schemas: node, sample
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
					"lastSampleTs": {
						Name:         "lastSampleTs",
						Unique:       false,
						AllowMissing: true,
						Indexer:      &memdb.IntFieldIndex{Field: "LastSampleTs"},
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
	return &Node{
		Id:           h.Hash(n.Target),
		Name:         n.Name,
		Target:       n.Target,
		State:        state,
		LastSampleTs: 0,
	}
}

// Get the id of a database node.
// The id is a hash integer
func GetId(n *Node) uint32 {
	return h.Hash(n.Target)
}

// Get the id of a given sample.
// The id is a hash integer
func GetSampleId(p *Sample) uint32 {
	return h.Hash(p.From + p.To + strconv.FormatInt(p.Key, 10))
}
