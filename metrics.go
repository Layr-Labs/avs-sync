package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	erroredTxs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "errored_txs_total",
		Help: "The total number of transactions that errored (failed to get processed by chain)",
	})
	revertedTxs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "reverted_txs_total",
		Help: "The total number of transactions that reverted (processed by chain but reverted)",
	})
	operatorsUpdated = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "operators_updated",
		Help: "The total number of operators updated (during the last quorum sync)",
	}, []string{"quorum"})
)

func StartMetricsServer(metricsAddr string) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(erroredTxs, revertedTxs, operatorsUpdated)
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	// not sure if we need to handle this error, since if metric server errors, then we will get alerts from grafana
	go http.ListenAndServe(metricsAddr, nil)
}
