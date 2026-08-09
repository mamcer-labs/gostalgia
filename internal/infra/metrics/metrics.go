package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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

	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gostalgia_http_requests_total",
		Help: "Total HTTP requests processed, labeled by method, route and status",
	}, []string{"method", "route", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gostalgia_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds, labeled by method and route",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)

// GinMiddleware records RED metrics (rate, errors, duration) for every
// request. Uses c.FullPath() as the route label to avoid high cardinality
// from path parameters (e.g. /files/:id, not /files/123).
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(method, route, status).Inc()
		httpRequestDuration.WithLabelValues(method, route).Observe(time.Since(start).Seconds())
	}
}
