// Copyright (c) 2020 Palantir Technologies. All rights reserved.
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

package transform

import (
	"context"

	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/status"
)

type HealthStatusTransformerFunc func(health.HealthStatus) health.HealthStatus

type source struct {
	healthStatusTransformerFunc HealthStatusTransformerFunc
	healthCheckSource           status.HealthCheckSource
}

func NewSource(healthCheckSource status.HealthCheckSource, healthStatusTransformerFunc HealthStatusTransformerFunc) status.HealthCheckSource {
	return source{
		healthStatusTransformerFunc: healthStatusTransformerFunc,
		healthCheckSource:           healthCheckSource,
	}
}

func (s source) HealthStatus(ctx context.Context) health.HealthStatus {
	if s.healthCheckSource != nil {
		raw := s.healthCheckSource.HealthStatus(ctx)
		if s.healthStatusTransformerFunc != nil {
			return s.healthStatusTransformerFunc(raw)
		}
		return raw
	}
	return health.HealthStatus{}
}
