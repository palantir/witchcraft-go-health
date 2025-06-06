package window

import (
	"context"
	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/status"
	"sync"
)

// KeyedErrorSourceAccumulator extends the KeyedErrorHealthCheckSource interface
// It includes an additional method AddHealthCheckSource which allows health sources to be added into the base health sources
// These sources are queried in conjunction with the given KeyedErrorHealthCheckSource
// This can be useful if you are working in legacy code bases and are in the process of converting health check, but need to add existing checks into a KeyedErrorHealthCheckSource
type KeyedErrorSourceAccumulator interface {
	AddHealthCheckSource(healthCheckSource status.HealthCheckSource)
	KeyedErrorHealthCheckSource
}

type defaultKeyedErrorSourceAccumulator struct {
	mutex                       sync.Mutex
	keyedErrorHealthCheckSource KeyedErrorHealthCheckSource
	allSources                  []status.HealthCheckSource
}

// NewDefaultKeyedErrorSourceAccumulator is the default implementation of KeyedErrorSourceAccumulator
func NewDefaultKeyedErrorSourceAccumulator(keyedErrorHealthCheckSource KeyedErrorHealthCheckSource) KeyedErrorSourceAccumulator {
	return &defaultKeyedErrorSourceAccumulator{
		keyedErrorHealthCheckSource: keyedErrorHealthCheckSource,
		allSources:                  []status.HealthCheckSource{keyedErrorHealthCheckSource},
	}
}

func (n *defaultKeyedErrorSourceAccumulator) AddHealthCheckSource(healthCheckSource status.HealthCheckSource) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.allSources = append(n.allSources, healthCheckSource)
}

func (n *defaultKeyedErrorSourceAccumulator) Submit(key string, err error) {
	n.keyedErrorHealthCheckSource.Submit(key, err)
}

func (n *defaultKeyedErrorSourceAccumulator) HealthStatus(ctx context.Context) health.HealthStatus {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	result := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{},
	}
	for _, healthCheckSource := range n.allSources {
		for k, v := range healthCheckSource.HealthStatus(ctx).Checks {
			result.Checks[k] = v
		}
	}
	return result

}
