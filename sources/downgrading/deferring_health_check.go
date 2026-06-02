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
	"github.com/palantir/witchcraft-go-health/v2/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/v2/status"
)

// NewDeferringHealthCheck returns a new DowngradingHealthCheck with HealthState_DEFERRING set.
// This is a convenience function for the common case of deferring unhealthy checks.
func NewDeferringHealthCheck(healthCheckSource status.HealthCheckSource) status.HealthCheckSource {
	return NewDowngradingHealthCheck(healthCheckSource, health.HealthState_DEFERRING)
}
