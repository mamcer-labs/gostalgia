package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	FilesScannedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nostalgia_files_scanned_total",
		Help: "The total number of files scanned by the system",
	})

	ScanDurationSummary = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "nostalgia_scan_duration_seconds",
		Help: "Summary of scan durations in seconds",
	})
)
