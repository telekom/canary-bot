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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/telekom/canary-bot/data"
)

//go:generate moq -out metric_test_moq.go . Metrics
type Metrics interface {
	GetRegistry() *prometheus.Registry
	Handler(data data.Database, h http.Handler) http.Handler
	GetNodes() prometheus.Gauge
	GetRtt() *prometheus.HistogramVec
}

type PrometheusMetrics struct {
	registry *prometheus.Registry
	nodes    prometheus.Gauge
	rtt      *prometheus.HistogramVec
}

// InitMetrics initializes the metrics and returns the PrometheusMetrics
func InitMetrics() *PrometheusMetrics {
	m := &PrometheusMetrics{
		registry: prometheus.NewRegistry(),
		rtt: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "rtt",
				Help: "Round-trip-time to a mesh node",
			},
			[]string{"type", "to"},
		),
		nodes: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "node_count",
			Help: "Total number of nodes",
		}),
	}

	// register metrics
	m.registry.MustRegister(
		m.rtt,
		m.nodes,
	)

	return m
}

// GetRegistry returns the registry to register prometheus metrics
func (m *PrometheusMetrics) GetRegistry() *prometheus.Registry {
	return m.registry
}

// Handler is a middleware to collect metrics
func (m *PrometheusMetrics) Handler(data data.Database, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set node count
		m.nodes.Set(float64(len(data.GetNodeList())))
		h.ServeHTTP(w, r)
	})
}

// GetNodes returns the node count metric
func (m *PrometheusMetrics) GetNodes() prometheus.Gauge {
	return m.nodes
}

// GetRtt returns the rtt metric
func (m *PrometheusMetrics) GetRtt() *prometheus.HistogramVec {
	return m.rtt
}
