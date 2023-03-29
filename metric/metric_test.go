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

package metric

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/telekom/canary-bot/data"
	"go.uber.org/zap"
)

func TestInitMetric(t *testing.T) {
	m := InitMetrics()
	registry := m.GetRegistry()
	if registry == nil {
		t.Error("registry is nil")
	}
}

func TestGetNodes(t *testing.T) {
	m := InitMetrics()
	nodes := m.GetNodes()
	if nodes == nil {
		t.Error("nodes is nil")
	}
}

func TestGetRtt(t *testing.T) {
	m := InitMetrics()
	rtt := m.GetRtt()
	if rtt == nil {
		t.Error("rtt is nil")
	}
}

func TestHandler(t *testing.T) {
	m := InitMetrics()
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Error("could not create logger")
	}
	db, err := data.NewMemDB(logger.Sugar())
	if err != nil {
		t.Error("could not create db")
	}
	handlerEmpty := m.Handler(db, nil)
	if handlerEmpty == nil {
		t.Error("handler is nil")
	}

	// test handler
	finalTestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("test ok - final test handler called")
		return
	})
	handler := m.Handler(db, finalTestHandler)
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
}
