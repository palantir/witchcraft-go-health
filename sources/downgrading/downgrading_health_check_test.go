// Copyright (c) 2026 Palantir Technologies. All rights reserved.
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

package downgrading

import (
	"context"
	"testing"

	"github.com/palantir/witchcraft-go-health/v2/conjure/witchcraft/api/health"
	"github.com/stretchr/testify/assert"
)

type healthStatusFn func(ctx context.Context) health.HealthStatus

func (fn healthStatusFn) HealthStatus(ctx context.Context) health.HealthStatus {
	return fn(ctx)
}

func TestDeferringHealthCheck(t *testing.T) {
	expected := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"TEST_CHECK": {
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
		},
	}
	healthCheckSource := healthStatusFn(func(ctx context.Context) health.HealthStatus {
		return expected
	})
	healthCheck := NewDowngradingHealthCheck(healthCheckSource, health.HealthState_DEFERRING)
	actual := healthCheck.HealthStatus(context.Background())
	assert.Equal(t, expected, actual)
}

func TestDeferringHealthCheckDowngrades(t *testing.T) {
	toReturn := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"TEST_CHECK": {
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
			"TEST_CHECK_BAD": {
				State: health.New_HealthState(health.HealthState_ERROR),
			},
		},
	}
	expected := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"TEST_CHECK": {
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
			"TEST_CHECK_BAD": {
				State: health.New_HealthState(health.HealthState_DEFERRING),
			},
		},
	}
	healthCheckSource := healthStatusFn(func(ctx context.Context) health.HealthStatus {
		return toReturn
	})
	healthCheck := NewDowngradingHealthCheck(healthCheckSource, health.HealthState_DEFERRING)
	actual := healthCheck.HealthStatus(context.Background())
	assert.Equal(t, expected, actual)
}
