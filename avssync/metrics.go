package avssync

import (
	"net/http"

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

type Metrics struct {
	updateStakeAttempts *prometheus.CounterVec
	txRevertedTotal     prometheus.Counter
	operatorsUpdated    *prometheus.GaugeVec

	registry *prometheus.Registry
}

func NewMetrics(reg *prometheus.Registry) *Metrics {
	metrics := &Metrics{
		updateStakeAttempts: promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "update_stake_attempt",
			Help:      "Result from an update stake attempt. Either succeed or error (either tx was mined but reverted, or failed to get processed by chain).",
		}, []string{"status", "quorum"}),

		txRevertedTotal: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "tx_reverted_total",
			Help:      "The total number of transactions that made it onchain but reverted (most likely because out of gas)",
		}),

		operatorsUpdated: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "operators_updated",
			Help:      "The total number of operators updated (during the last quorum sync)",
		}, []string{"quorum"}),

		registry: reg,
	}

	return metrics
}

func (g *Metrics) UpdateStakeAttemptInc(status UpdateStakeStatus, quorum string) {
	g.updateStakeAttempts.WithLabelValues(string(status), quorum).Inc()
}

func (g *Metrics) TxRevertedTotalInc() {
	g.txRevertedTotal.Inc()
}

func (g *Metrics) OperatorsUpdatedSet(quorum string, operators int) {
	g.operatorsUpdated.WithLabelValues(quorum).Set(float64(operators))
}

func (g *Metrics) Start(metricsAddr string) {
	http.Handle("/metrics", promhttp.HandlerFor(g.registry, promhttp.HandlerOpts{}))
	// not sure if we need to handle this error, since if metric server errors, then we will get alerts from grafana
	go func() {
		_ = http.ListenAndServe(metricsAddr, nil)
	}()
}
