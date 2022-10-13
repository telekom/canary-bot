package data

import (
	"testing"
	"time"

	"github.com/go-test/deep"
)

func Test_SetSample(t *testing.T) {
	db, _ := NewMemDB(log)
	for _, sample := range samples {
		db.SetSample(sample)
	}
	txn := db.Txn(false)
	raw, _ := txn.Get("sample", "id")
	counter := 0
	for raw.Next() != nil {
		counter++
	}
	if counter != len(samples) {
		t.Errorf("the amount of samples (amount: %v) in the db is not as expected: %v", counter, len(samples))
	}
}

func Test_SetSampleNaN(t *testing.T) {
	db, _ := NewMemDB(log)
	for _, sample := range samples {
		db.SetSample(sample)
		db.SetSampleNaN(GetSampleId(sample))
	}
	txn := db.Txn(false)
	raw, _ := txn.Get("sample", "id")
	for obj := raw.Next(); obj != nil; obj = raw.Next() {
		if obj.(*Sample).Value != "NaN" {
			t.Errorf("The sample value is not Nan as expected. Sample value: %v", obj.(*Sample).Value)
		}
	}
}

func Test_GetSample(t *testing.T) {
	tests := []struct {
		name         string
		emptyDb      bool
		state        int
		expected     *Sample
		expectedTs   int64
		expectedList []*Sample
	}{
		{name: "samples in db", emptyDb: false, expected: samples[0], expectedList: nil},
		{name: "empty db", emptyDb: true, expected: &Sample{}, expectedList: nil},

		{name: "ts/samples in db", emptyDb: false, expectedTs: samples[0].Ts, expectedList: nil},
		{name: "ts/empty db", emptyDb: true, expectedTs: 0, expectedList: nil},

		{name: "list/nodes in db", emptyDb: false, expected: nil, expectedList: samples},
		{name: "list/empty db", emptyDb: true, expected: nil, expectedList: []*Sample{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _ := NewMemDB(log)
			if !tt.emptyDb {
				// fill db
				for _, sample := range samples {
					db.SetSample(sample)
				}
			}
			if tt.expected != nil {
				// GetSample(Ts)
				result := db.GetSample(samples[0].Id)
				if diff := deep.Equal(result, tt.expected); diff != nil {
					t.Error(diff)
				}
			} else if tt.expectedTs != 0 {
				resultTs := db.GetSampleTs(samples[0].Id)
				if diff := deep.Equal(resultTs, tt.expectedTs); diff != nil {
					t.Error(diff)
				}
			} else {
				// GetSampleList
				result := db.GetSampleList()
				if len(result) != len(tt.expectedList) {
					t.Errorf("the amount returned samples is incorrect: %v but expected %v", len(result), len(tt.expectedList))
				}
				for i, sample := range result {
					if diff := deep.Equal(sample, tt.expectedList[i]); diff != nil {
						t.Error(diff)
					}
				}
			}
		})
	}
}
func Test_DeleteSample(t *testing.T) {
	db, _ := NewMemDB(log)
	for _, sample := range samples {
		db.SetSample(sample)
	}
	for _, sample := range samples {
		db.DeleteSample(sample.Id)
	}
	samples_result := db.GetSampleList()
	if len(samples_result) > 0 {
		t.Errorf("still some samples in db, should be empty. Amount of nodes that are left in result: %v", len(samples_result))
	}
}

func Test_SetSampleTsNow(t *testing.T) {
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

// func (db *Database) DeleteSample(id uint32) {}
