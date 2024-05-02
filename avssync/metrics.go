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

func StartMetricsServer(metricsAddr string) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(updateStakeAttempt, txRevertedTotal, operatorsUpdated)
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	// not sure if we need to handle this error, since if metric server errors, then we will get alerts from grafana
	go func() {
		_ = http.ListenAndServe(metricsAddr, nil)
	}()
}
