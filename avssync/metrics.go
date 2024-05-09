package avssync

import (
	"net/http"

	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const metricsNamespace = "avssync"

type UpdateStakeStatus string

const (
	UpdateStakeStatusError   UpdateStakeStatus = "error"
	UpdateStakeStatusSucceed UpdateStakeStatus = "succeed"
)

var (
	updateStakeAttempt = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "update_stake_attempt",
		Help:      "Result from an update stake attempt. Either succeed or error (either tx was mined but reverted, or failed to get processed by chain).",
	}, []string{"status", "quorum"})
	txRevertedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "tx_reverted_total",
		Help:      "The total number of transactions that made it onchain but reverted (most likely because out of gas)",
	})
	operatorsUpdated = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "operators_updated",
		Help:      "The total number of operators updated (during the last quorum sync)",
	}, []string{"quorum"})
)

type Metrics struct {
	registry *prometheus.Registry
	addr     string
	logger   logging.Logger
}

func NewMetrics(addr string, logger logging.Logger) *Metrics {
	registry := prometheus.NewRegistry()
	registry.MustRegister(updateStakeAttempt, txRevertedTotal, operatorsUpdated)

	return &Metrics{
		registry: registry,
		addr:     addr,
		logger:   logger,
	}
}

func (m *Metrics) Start() {
	go func() {
		log := m.logger
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			m.registry,
			promhttp.HandlerOpts{},
		))
		err := http.ListenAndServe(m.addr, mux)
		log.Error("Prometheus server failed", "err", err)
	}()
}
