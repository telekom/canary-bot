package data

import "time"

// SetSample inserts a measurement sample in the db
func (db *Database) SetSample(sample *Sample) {
	// Create a write transaction
	txn := db.Txn(true)
	defer txn.Abort()

	sample.Id = GetSampleId(sample)
	err := txn.Insert("sample", sample)
	if err != nil {
		panic(err)
	}

	// Commit the transaction
	txn.Commit()
}

// SetSampleNaN sets a sample to not a number "NaN"
// E.g. a ping failed, RTT has to be set to NaN
func (db *Database) SetSampleNaN(id uint32) {
	// Create a write transaction
	txn := db.Txn(true)
	defer txn.Abort()

	sample := *db.GetSample(id)
	if sample.Id == 0 {
		return
	}

	sample.Value = "NaN"
	sample.Ts = time.Now().Unix()
	err := txn.Insert("sample", &sample)
	if err != nil {
		panic(err)
	}

	// Commit the transaction
	txn.Commit()
}

// GetSample returns a measurement sample by id
func (db *Database) GetSample(id uint32) *Sample {
	txn := db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("sample", "id", id)
	if err != nil {
		panic(err)
	}
	if raw == nil {
		return &Sample{}
	}
	return raw.(*Sample)
}

// DeleteSample deletes a measurement sample by id
func (db *Database) DeleteSample(id uint32) {
	txn := db.Txn(true)
	defer txn.Abort()

	err := txn.Delete("sample", db.GetSample(id))
	if err != nil {
		db.log.Debugf("Could not delete sample")
	}
	// Commit the transaction
	txn.Commit()
}

// GetSampleTs returns the timestamp from a measurement sample by id
func (db *Database) GetSampleTs(id uint32) int64 {
	txn := db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("sample", "id", id)
	if err != nil {
		panic(err)
	}
	if raw == nil {
		return 0
	}
	return raw.(*Sample).Ts
}

// GetSampleList returns all measurement samples in db
func (db *Database) GetSampleList() []*Sample {
	txn := db.Txn(false)
	defer txn.Abort()

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
g