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

package downgrading_test

import (
	"context"
	"testing"

	"github.com/palantir/witchcraft-go-health/v2/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/v2/sources/downgrading"
	"github.com/stretchr/testify/assert"
)

type healthStatusFn func(ctx context.Context) health.HealthStatus

func (fn healthStatusFn) HealthStatus(ctx context.Context) health.HealthStatus {
	return fn(ctx)
}

func TestDowngradingHealthCheck_HealthyPassesThrough(t *testing.T) {
	delegate := healthStatusFn(func(ctx context.Context) health.HealthStatus {
		return health.HealthStatus{
			Checks: map[health.CheckType]health.HealthCheckResult{
				"TEST_CHECK": {
					Type:  "TEST_CHECK",
					State: health.New_HealthState(health.HealthState_HEALTHY),
				},
			},
		}
	})
	healthCheck := downgrading.NewDowngradingHealthCheck(delegate, health.HealthState_DEFERRING)

	expected := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"TEST_CHECK": {
				Type:  "TEST_CHECK",
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
		},
	}
	actual := healthCheck.HealthStatus(context.Background())
	assert.Equal(t, expected, actual)
}

func TestDowngradingHealthCheck_UnhealthyDowngrades(t *testing.T) {
	delegate := healthStatusFn(func(ctx context.Context) health.HealthStatus {
		return health.HealthStatus{
			Checks: map[health.CheckType]health.HealthCheckResult{
				"TEST_CHECK": {
					Type:  "TEST_CHECK",
					State: health.New_HealthState(health.HealthState_HEALTHY),
				},
				"TEST_CHECK_BAD": {
					Type:  "TEST_CHECK_BAD",
					State: health.New_HealthState(health.HealthState_ERROR),
				},
			},
		}
	})
	healthCheck := downgrading.NewDowngradingHealthCheck(delegate, health.HealthState_DEFERRING)

	expected := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"TEST_CHECK": {
				Type:  "TEST_CHECK",
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
			"TEST_CHECK_BAD": {
				Type:  "TEST_CHECK_BAD",
				State: health.New_HealthState(health.HealthState_DEFERRING),
			},
		},
	}
	actual := healthCheck.HealthStatus(context.Background())
	assert.Equal(t, expected, actual)
}

func TestDeferringHealthCheck(t *testing.T) {
	delegate := healthStatusFn(func(ctx context.Context) health.HealthStatus {
		return health.HealthStatus{
			Checks: map[health.CheckType]health.HealthCheckResult{
				"TEST_CHECK": {
					Type:  "TEST_CHECK",
					State: health.New_HealthState(health.HealthState_ERROR),
				},
			},
		}
	})
	healthCheck := downgrading.NewDeferringHealthCheck(delegate)

	expected := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"TEST_CHECK": {
				Type:  "TEST_CHECK",
				State: health.New_HealthState(health.HealthState_DEFERRING),
			},
		},
	}
	actual := healthCheck.HealthStatus(context.Background())
	assert.Equal(t, expected, actual)
}

func TestDowngradingHealthCheck_PreservesMessageAndParams(t *testing.T) {
	message := "something went wrong"
	params := map[string]interface{}{"key": "value"}
	delegate := healthStatusFn(func(ctx context.Context) health.HealthStatus {
		return health.HealthStatus{
			Checks: map[health.CheckType]health.HealthCheckResult{
				"TEST_CHECK": {
					Type:    "TEST_CHECK",
					State:   health.New_HealthState(health.HealthState_ERROR),
					Message: &message,
					Params:  params,
				},
			},
		}
	})
	healthCheck := downgrading.NewDowngradingHealthCheck(delegate, health.HealthState_WARNING)

	actual := healthCheck.HealthStatus(context.Background())
	assert.Equal(t, health.HealthState_WARNING, actual.Checks["TEST_CHECK"].State.Value())
	assert.Equal(t, &message, actual.Checks["TEST_CHECK"].Message)
	assert.Equal(t, params, actual.Checks["TEST_CHECK"].Params)
}
