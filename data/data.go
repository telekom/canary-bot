package data

import (
	h "canary-bot/helper"
	meshv1 "canary-bot/proto/mesh/v1"
	"strconv"
	"time"

	"github.com/hashicorp/go-memdb"
	"go.uber.org/zap"
)

const (
	STATE = 1
	RTT   = 2
)

var SampleName = map[int64]string{
	STATE: "state",
	RTT:   "rtt",
}

type Database struct {
	*memdb.MemDB
	log *zap.SugaredLogger
}

type DbNode struct {
	Id           uint32
	Name         string
	Target       string
	State        int
	LastSampleTs int64
}
type Sample struct {
	id    uint32
	From  string
	To    string
	Key   int64
	Value string
	Ts    int64
}

func NewMemDB(logger *zap.SugaredLogger) (Database, error) {
	defer logger.Sync()

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
						Indexer:      &memdb.UintFieldIndex{Field: "id"},
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

func (db *Database) SetNode(node *DbNode) {
	// Create a write transaction
	txn := db.Txn(true)
	err := txn.Insert("node", node)
	if err != nil {
		panic(err)
	}

	// Commit the transaction
	txn.Commit()
}

func (db *Database) SetNodeTsNow(id uint32) {
	txn := db.Txn(true)

	node := db.GetNode(id)
	if node.Id == 0 {
		txn.Abort()
		return
	}

	node.LastSampleTs = time.Now().Unix()
	err := txn.Insert("node", node)
	if err != nil {
		panic(err)
	}

	// Commit the transaction
	txn.Commit()
}

func (db *Database) SetSample(sample *Sample) {
	// Create a write transaction
	txn := db.Txn(true)
	sample.id = GetSampleId(sample)
	err := txn.Insert("sample", sample)
	if err != nil {
		panic(err)
	}

	// Commit the transaction
	txn.Commit()
}

func (db *Database) GetSampleTs(id uint32) int64 {
	txn := db.Txn(false)
	raw, err := txn.First("sample", "id", id)
	if err != nil {
		panic(err)
	}
	if raw == nil {
		return 0
	}
	return raw.(*Sample).Ts
}

func (db *Database) GetSampleList() []*Sample {
	txn := db.Txn(false)
	it, err := txn.Get("sample", "id")
	if err != nil {
		panic(err)
	}
	var samples []*Sample
	for obj := it.Next(); obj != nil; obj = it.Next() {
		samples = append(samples, obj.(*Sample))
	}
	return samples
}

func (db *Database) DeleteNode(id uint32) {
	txn := db.Txn(true)
	err := txn.Delete("node", db.GetNode(id))
	if err != nil {
		db.log.Debugf("Could not delete Node")
	}
	// Commit the transaction
	txn.Commit()
}

func (db *Database) GetNode(id uint32) *DbNode {
	txn := db.Txn(false)
	raw, err := txn.First("node", "id", id)
	if err != nil {
		panic(err)
	}
	if raw == nil {
		return &DbNode{}
	}
	return raw.(*DbNode)
}

func (db *Database) GetNodeByName(name string) *DbNode {
	txn := db.Txn(false)
	raw, err := txn.First("node", "name", name)
	if err != nil {
		panic(err)
	}
	if raw == nil {
		return &DbNode{}
	}
	return raw.(*DbNode)
}

func (db *Database) GetNodeList() []*DbNode {
	txn := db.Txn(false)
	it, err := txn.Get("node", "id")
	if err != nil {
		panic(err)
	}
	var nodes []*DbNode
	for obj := it.Next(); obj != nil; obj = it.Next() {
		nodes = append(nodes, obj.(*DbNode))
	}
	return nodes
}

func (db *Database) GetNodeListByState(byState int) []*DbNode {
	txn := db.Txn(false)
	it, err := txn.Get("node", "id")
	if err != nil {
		panic(err)
	}
	var nodes []*DbNode
	for obj := it.Next(); obj != nil; obj = it.Next() {
		if obj.(*DbNode).State == byState {
			nodes = append(nodes, obj.(*DbNode))
		}
	}
	return nodes
}

func (n *DbNode) Convert() *meshv1.Node {
	return &meshv1.Node{
		Name:   n.Name,
		Target: n.Target,
	}
}

func Convert(n *meshv1.Node, state int) *DbNode {
	return &DbNode{
		Id:           h.Hash(n.Target),
		Name:         n.Name,
		Target:       n.Target,
		State:        state,
		LastSampleTs: 0,
	}
}

func GetId(n *DbNode) uint32 {
	return h.Hash(n.Target)
}

func GetSampleId(p *Sample) uint32 {
	return h.Hash(p.From + p.To + strconv.FormatInt(p.Key, 10))
}
