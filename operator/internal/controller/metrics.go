package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	storeCreatedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "store_created_total",
		Help: "Total number of stores created",
	})

	storeDeletionTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "store_deletion_total",
		Help: "Total number of stores deleted",
	})

	storeProvisioningSeconds = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "store_provisioning_seconds",
		Help:    "Time taken for a store to go from Provisioning to Ready",
		Buckets: prometheus.ExponentialBuckets(5, 2, 8), // 5s, 10s, 20s, ... 640s
	})
)

func init() {
	metrics.Registry.MustRegister(
		storeCreatedTotal,
		storeDeletionTotal,
		storeProvisioningSeconds,
	)
}
