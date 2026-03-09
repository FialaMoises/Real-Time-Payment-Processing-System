package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latencies in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Transaction Metrics
	TransactionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transactions_total",
			Help: "Total number of transactions processed",
		},
		[]string{"type", "status"},
	)

	TransactionsFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transactions_failed_total",
			Help: "Total number of failed transactions",
		},
		[]string{"type", "reason"},
	)

	TransactionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "transaction_duration_seconds",
			Help:    "Transaction processing duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"type"},
	)

	TransactionsInProgress = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "transactions_in_progress",
			Help: "Number of transactions currently being processed",
		},
	)

	TransactionAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "transaction_amount_brl",
			Help:    "Transaction amounts in BRL",
			Buckets: []float64{10, 50, 100, 500, 1000, 5000, 10000, 50000, 100000},
		},
		[]string{"type"},
	)

	// Database Metrics
	DatabaseConnectionsFailed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "database_connections_failed_total",
			Help: "Total number of failed database connections",
		},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation"},
	)

	DatabaseConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
	)

	// Account Metrics
	AccountsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "accounts_total",
			Help: "Total number of accounts in the system",
		},
	)

	AccountsCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "accounts_created_total",
			Help: "Total number of accounts created",
		},
	)

	// Balance Metrics
	TotalBalanceAmount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "total_balance_amount_brl",
			Help: "Total balance amount across all accounts in BRL",
		},
	)

	// Authentication Metrics
	LoginAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "login_attempts_total",
			Help: "Total number of login attempts",
		},
		[]string{"status"},
	)

	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_sessions",
			Help: "Number of active user sessions",
		},
	)

	// Concurrency Metrics
	ConcurrentRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "concurrent_requests",
			Help: "Number of concurrent HTTP requests",
		},
	)

	MutexWaitDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mutex_wait_duration_seconds",
			Help:    "Time spent waiting for mutex locks",
			Buckets: []float64{.0001, .0005, .001, .005, .01, .05, .1, .5, 1},
		},
		[]string{"resource"},
	)

	// Idempotency Metrics
	IdempotencyHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "idempotency_hits_total",
			Help: "Number of idempotent request hits",
		},
	)

	// Error Metrics
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "component"},
	)

	// System Metrics
	UptimeSeconds = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "uptime_seconds",
			Help: "Application uptime in seconds",
		},
	)
)

// RecordHTTPRequest records an HTTP request with its metadata
func RecordHTTPRequest(method, path, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordTransaction records a transaction with its metadata
func RecordTransaction(txType, status string, duration, amount float64) {
	TransactionsTotal.WithLabelValues(txType, status).Inc()
	TransactionDuration.WithLabelValues(txType).Observe(duration)
	TransactionAmount.WithLabelValues(txType).Observe(amount)
}

// RecordTransactionFailure records a failed transaction
func RecordTransactionFailure(txType, reason string) {
	TransactionsFailed.WithLabelValues(txType, reason).Inc()
}

// IncrementTransactionsInProgress increments the in-progress transaction gauge
func IncrementTransactionsInProgress() {
	TransactionsInProgress.Inc()
}

// DecrementTransactionsInProgress decrements the in-progress transaction gauge
func DecrementTransactionsInProgress() {
	TransactionsInProgress.Dec()
}

// RecordDatabaseQuery records a database query with its duration
func RecordDatabaseQuery(operation string, duration float64) {
	DatabaseQueryDuration.WithLabelValues(operation).Observe(duration)
}

// RecordLogin records a login attempt
func RecordLogin(success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	LoginAttempts.WithLabelValues(status).Inc()
}

// RecordError records an error
func RecordError(errorType, component string) {
	ErrorsTotal.WithLabelValues(errorType, component).Inc()
}
