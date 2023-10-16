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
