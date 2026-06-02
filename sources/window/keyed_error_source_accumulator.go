// Copyright (c) 2025 Palantir Technologies. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package window

import (
	"context"
	"maps"
	"sync"

	"github.com/palantir/witchcraft-go-health/v2/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/v2/sources"
	"github.com/palantir/witchcraft-go-health/v2/sources/tree"
	"github.com/palantir/witchcraft-go-health/v2/status"
)

// KeyedErrorSourceAccumulator extends the KeyedErrorHealthCheckSource interface
// It includes an additional method AddHealthCheckSource which allows health sources to be added into the base health sources
// These sources are queried in conjunction with the given KeyedErrorHealthCheckSource
// This can be useful if you are working in legacy code bases and are in the process of converting health check, but need to add existing checks into a KeyedErrorHealthCheckSource
type KeyedErrorSourceAccumulator interface {
	AddHealthCheckSource(healthCheckSource status.HealthCheckSource)
	RootUnderTree(healthCheckSource status.HealthCheckSource)
	KeyedErrorHealthCheckSource
}

type defaultKeyedErrorSourceAccumulator struct {
	keyedErrorHealthCheckSource KeyedErrorHealthCheckSource
	allSources                  []status.HealthCheckSource
}

type defaultKeyedErrorSourceAccumulatorParent struct {
	mutex                              sync.Mutex
	defaultKeyedErrorSourceAccumulator *defaultKeyedErrorSourceAccumulator
	rootedHealthCheck                  status.HealthCheckSource
}

// NewDefaultKeyedErrorSourceAccumulator is the default implementation of KeyedErrorSourceAccumulator
func NewDefaultKeyedErrorSourceAccumulator(keyedErrorHealthCheckSource KeyedErrorHealthCheckSource) KeyedErrorSourceAccumulator {
	defaultKeyedErrorSourceAccumulator := &defaultKeyedErrorSourceAccumulator{
		keyedErrorHealthCheckSource: keyedErrorHealthCheckSource,
		allSources:                  []status.HealthCheckSource{keyedErrorHealthCheckSource},
	}
	staticParent := sources.StaticHealthCheckSource(health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{},
	})
	defaultKeyedErrorSourceAccumulatorParent := &defaultKeyedErrorSourceAccumulatorParent{
		defaultKeyedErrorSourceAccumulator: defaultKeyedErrorSourceAccumulator,
	}
	defaultKeyedErrorSourceAccumulatorParent.RootUnderTree(staticParent)
	return defaultKeyedErrorSourceAccumulatorParent
}

func (n *defaultKeyedErrorSourceAccumulatorParent) AddHealthCheckSource(healthCheckSource status.HealthCheckSource) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.defaultKeyedErrorSourceAccumulator.allSources = append(n.defaultKeyedErrorSourceAccumulator.allSources, healthCheckSource)
}

func (n *defaultKeyedErrorSourceAccumulatorParent) RootUnderTree(healthCheckSource status.HealthCheckSource) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.rootedHealthCheck = tree.NewHealthCheckSourceTree(healthCheckSource, []status.HealthCheckSource{n.defaultKeyedErrorSourceAccumulator})
}

func (n *defaultKeyedErrorSourceAccumulatorParent) Submit(ctx context.Context, key string, err error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.defaultKeyedErrorSourceAccumulator.keyedErrorHealthCheckSource.Submit(ctx, key, err)
}

func (n *defaultKeyedErrorSourceAccumulatorParent) HealthStatus(ctx context.Context) health.HealthStatus {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.rootedHealthCheck.HealthStatus(ctx)
}

func (n *defaultKeyedErrorSourceAccumulator) HealthStatus(ctx context.Context) health.HealthStatus {
	result := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{},
	}
	for _, healthCheckSource := range n.allSources {
		maps.Copy(result.Checks, healthCheckSource.HealthStatus(ctx).Checks)
	}
	return result

}
