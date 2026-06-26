// Package metrics is the go104 observability surface: Prometheus series plus the
// /metrics handler. It is a leaf package — it imports only the Prometheus client,
// never any go104 package, so every other package can call into it freely.
package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	buildInfo = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "go104_build_info",
		Help: "Build info; constant 1 with version/commit/go_version labels.",
	}, []string{"version", "commit", "go_version"})

	// Field acquisition — labelled by line + protocol (NOT per-signal: cardinality).
	ingestSamples = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "go104_ingest_samples_total",
		Help: "Datapoints ingested from the field (the throughput SLI).",
	}, []string{"line", "protocol"})

	ingestQuality = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "go104_ingest_quality_total",
		Help: "Ingested datapoints by quality (good vs IV/BL).",
	}, []string{"line", "quality"})

	lineState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "go104_line_state",
		Help: "Line connection state: 0 disconnected, 1 connecting, 2 active, 3 stopped.",
	}, []string{"line", "protocol"})

	// Control (audit-grade).
	commandsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "go104_commands_total",
		Help: "Control commands issued to the field.",
	}, []string{"line", "type_id", "result"})

	// Store / RTDB.
	snapshotDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "go104_snapshot_duration_seconds",
		Help:    "Duration of the periodic current-value snapshot flush.",
		Buckets: prometheus.DefBuckets,
	})
	snapshotRows = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "go104_snapshot_rows",
		Help: "Rows written in the most recent snapshot.",
	})
	snapshotErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "go104_snapshot_errors_total",
		Help: "Current-value snapshot flush errors.",
	})
	rtdbSignals = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "go104_rtdb_signals",
		Help: "Signals held in the in-memory RTDB (current values).",
	})
	historyPoints = promauto.NewCounter(prometheus.CounterOpts{
		Name: "go104_history_ring_points_total",
		Help: "Points appended to the in-memory history ring.",
	})

	// Transport.
	wsClients = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "go104_ws_clients",
		Help: "Connected WebSocket clients.",
	})
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "go104_http_requests_total",
		Help: "HTTP API requests.",
	}, []string{"route", "method", "code"})
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "go104_http_request_duration_seconds",
		Help:    "HTTP API request duration.",
		Buckets: prometheus.DefBuckets,
	}, []string{"route", "method"})
)

// Handler serves the Prometheus exposition (default registry → Go runtime +
// process collectors are included automatically).
func Handler() http.Handler { return promhttp.Handler() }

// SetBuildInfo records the build_info series once at startup.
func SetBuildInfo(version, commit, goVersion string) {
	buildInfo.WithLabelValues(version, commit, goVersion).Set(1)
}

// IngestSample counts one ingested datapoint and its quality.
func IngestSample(line, protocol string, good bool) {
	ingestSamples.WithLabelValues(line, protocol).Inc()
	q := "good"
	if !good {
		q = "bad"
	}
	ingestQuality.WithLabelValues(line, q).Inc()
}

// LineState records a connection-state change.
func LineState(line, protocol, state string) {
	lineState.WithLabelValues(line, protocol).Set(stateCode(state))
}

func stateCode(s string) float64 {
	switch s {
	case "DISCONNECTED":
		return 0
	case "CONNECTING":
		return 1
	case "ACTIVE":
		return 2
	case "STOPPED":
		return 3
	}
	return -1
}

// Command records one field-control command and its result.
func Command(line string, typeID int, ok bool) {
	r := "ok"
	if !ok {
		r = "error"
	}
	commandsTotal.WithLabelValues(line, strconv.Itoa(typeID), r).Inc()
}

// Snapshot records one current-value snapshot flush.
func Snapshot(seconds float64, rows int, failed bool) {
	snapshotDuration.Observe(seconds)
	snapshotRows.Set(float64(rows))
	if failed {
		snapshotErrors.Inc()
	}
}

func RTDBSignals(n int) { rtdbSignals.Set(float64(n)) }
func HistoryAppend()    { historyPoints.Inc() }
func WSClients(n int)   { wsClients.Set(float64(n)) }

// HTTPRequest records one served HTTP request.
func HTTPRequest(route, method string, code int, seconds float64) {
	httpRequests.WithLabelValues(route, method, strconv.Itoa(code)).Inc()
	httpDuration.WithLabelValues(route, method).Observe(seconds)
}
