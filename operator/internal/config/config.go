package config

import (
	"os"
	"strconv"
	"time"
)

// OperatorConfig holds all operator configuration
type OperatorConfig struct {
	// Chart configuration
	WordPressChartPath string
	BaseDomain         string

	// Timing configuration (for reconciliation intervals)
	NamespaceRequeueInterval  time.Duration
	HelmFailureRetryInterval  time.Duration
	PodReadinessCheckInterval time.Duration
	DeletionRequeueInterval   time.Duration

	// Helm values configuration
	PersistenceEnabled         bool
	LivenessProbeInitialDelay  int
	LivenessProbePeriod        int
	ReadinessProbeInitialDelay int
	ReadinessProbePeriod       int
}

// Load reads configuration from environment variables with sensible defaults
func Load() *OperatorConfig {
	return &OperatorConfig{
		// Chart path: default to embedded charts in container
		WordPressChartPath: getEnv("WORDPRESS_CHART_PATH", "/charts/engine-woo"),

		// Base domain for store URLs
		BaseDomain: getEnv("BASE_DOMAIN", "127.0.0.1.nip.io"),

		// Reconciliation timing
		NamespaceRequeueInterval:  parseDuration(getEnv("NAMESPACE_REQUEUE_INTERVAL", "1s")),
		HelmFailureRetryInterval:  parseDuration(getEnv("HELM_RETRY_INTERVAL", "20s")),
		PodReadinessCheckInterval: parseDuration(getEnv("POD_CHECK_INTERVAL", "5s")),
		DeletionRequeueInterval:   parseDuration(getEnv("DELETION_REQUEUE_INTERVAL", "5s")),

		// Helm values defaults
		PersistenceEnabled:         parseBool(getEnv("PERSISTENCE_ENABLED", "false")),
		LivenessProbeInitialDelay:  parseInt(getEnv("LIVENESS_INITIAL_DELAY", "120")),
		LivenessProbePeriod:        parseInt(getEnv("LIVENESS_PERIOD", "20")),
		ReadinessProbeInitialDelay: parseInt(getEnv("READINESS_INITIAL_DELAY", "60")),
		ReadinessProbePeriod:       parseInt(getEnv("READINESS_PERIOD", "10")),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// parseInt parses an integer from a string, returning 0 on error
func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

// parseBool parses a boolean from a string, returning false on error
func parseBool(s string) bool {
	val, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return val
}

// parseDuration parses a time.Duration from a string, returning 0 on error
func parseDuration(s string) time.Duration {
	val, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return val
}
