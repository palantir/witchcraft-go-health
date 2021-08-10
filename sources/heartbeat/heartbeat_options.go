package heartbeat

import (
	"time"

	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
)

// HeartbeatOption is an option for a heartbeat based health check source.
type HeartbeatOption func(conf *heartbeatSourceConfig)

type heartbeatSourceConfig struct {
	checkType          health.CheckType
	startupGracePeriod time.Duration
}

func defaultHeartbeatSourceConfig(checkType health.CheckType) heartbeatSourceConfig {
	return heartbeatSourceConfig{
		checkType:          checkType,
		startupGracePeriod: 0,
	}
}

func (h *heartbeatSourceConfig) apply(options ...HeartbeatOption) {
	for _, option := range options {
		option(h)
	}
}

// WithStartupGracePeriod configures an initial grace period that allows the health check source to remain healthy
// despite no heartbeats within the configured heartbeat timeout. This grace period is only applied once on startup.
func WithStartupGracePeriod(period time.Duration) HeartbeatOption {
	return func(conf *heartbeatSourceConfig) {
		conf.startupGracePeriod = period
	}
}
