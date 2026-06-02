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

	"github.com/palantir/witchcraft-go-health/v2/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/v2/status"
)

type downgradingHealthCheck struct {
	healthCheckSource status.HealthCheckSource
	downgradeTo       health.HealthState_Value
}

// NewDowngradingHealthCheck returns a new downgradingHealthCheck which implements status.HealthCheckSource.
// This health check wraps another health check source and will return its result if healthy.
// If it is not healthy, it downgrades the health state to the specified downgradeTo value.
func NewDowngradingHealthCheck(healthCheckSource status.HealthCheckSource, downgradeTo health.HealthState_Value) status.HealthCheckSource {
	return &downgradingHealthCheck{
		healthCheckSource: healthCheckSource,
		downgradeTo:       downgradeTo,
	}
}

func (d *downgradingHealthCheck) HealthStatus(ctx context.Context) health.HealthStatus {
	healthStatus := d.healthCheckSource.HealthStatus(ctx)
	for _, v := range healthStatus.Checks {
		if v.State.Value() != health.HealthState_HEALTHY {
			return d.downgradeToTarget(healthStatus)
		}
	}
	return healthStatus
}

func (d *downgradingHealthCheck) downgradeToTarget(originalHealthStatus health.HealthStatus) health.HealthStatus {
	healthStatus := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{},
	}
	for healthCheckType, healthCheckResult := range originalHealthStatus.Checks {
		healthStatus.Checks[healthCheckType] = d.downgradeHealthCheckResult(healthCheckResult)
	}
	return healthStatus
}

func (d *downgradingHealthCheck) downgradeHealthCheckResult(result health.HealthCheckResult) health.HealthCheckResult {
	return health.HealthCheckResult{
		Type:    result.Type,
		State:   d.getHealthCheckResultStatus(result.State),
		Message: result.Message,
		Params:  result.Params,
	}
}

func (d *downgradingHealthCheck) getHealthCheckResultStatus(state health.HealthState) health.HealthState {
	if state.Value() == health.HealthState_HEALTHY {
		return state
	}
	return health.New_HealthState(d.downgradeTo)
}
