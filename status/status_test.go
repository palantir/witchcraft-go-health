// Copyright (c) 2018 Palantir Technologies. All rights reserved.
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

package status

import (
	"context"
	"net/http"
	"testing"

	"github.com/palantir/pkg/refreshable/v2"
	"github.com/palantir/witchcraft-go-health/v2/conjure/witchcraft/api/health"
	"github.com/stretchr/testify/assert"
)

type testHealthCheckSource struct {
	healthStatus health.HealthStatus
}

func (t *testHealthCheckSource) HealthStatus(_ context.Context) health.HealthStatus {
	return t.healthStatus
}

func TestCombinedHealthCheckSource(t *testing.T) {
	sourceA := &testHealthCheckSource{
		healthStatus: health.HealthStatus{
			Checks: map[health.CheckType]health.HealthCheckResult{
				"a": {
					State: health.New_HealthState(health.HealthState_HEALTHY),
				},
				"b": {
					State: health.New_HealthState(health.HealthState_HEALTHY),
				},
			},
		},
	}
	sourceB := &testHealthCheckSource{
		healthStatus: health.HealthStatus{
			Checks: map[health.CheckType]health.HealthCheckResult{
				"a": {
					State: health.New_HealthState(health.HealthState_ERROR),
				},
				"c": {
					State: health.New_HealthState(health.HealthState_HEALTHY),
				},
			},
		},
	}
	combined := NewCombinedHealthCheckSource(sourceA, sourceB, nil)
	actual := combined.HealthStatus(context.Background())
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"a": {
				State: health.New_HealthState(health.HealthState_ERROR),
			},
			"b": {
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
			"c": {
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
		},
	}, actual)
}

func TestCombinedHealthCheckSourceWithRefreshable(t *testing.T) {
	sourceA := &testHealthCheckSource{
		healthStatus: health.HealthStatus{
			Checks: map[health.CheckType]health.HealthCheckResult{
				"a": {
					State: health.New_HealthState(health.HealthState_HEALTHY),
				},
			},
		},
	}
	sourceB := &testHealthCheckSource{
		healthStatus: health.HealthStatus{
			Checks: map[health.CheckType]health.HealthCheckResult{
				"b": {
					State: health.New_HealthState(health.HealthState_HEALTHY),
				},
			},
		},
	}
	sources := refreshable.New([]HealthCheckSource{sourceA})
	combined := NewCombinedHealthCheckSourceWithRefresh(sources)
	actual := combined.HealthStatus(context.Background())
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"a": {
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
		},
	}, actual)
	sources.Update([]HealthCheckSource{sourceA, sourceB})
	actual = combined.HealthStatus(context.Background())
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"a": {
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
			"b": {
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
		},
	}, actual)
}

func TestHealthBasedReadinessSource(t *testing.T) {
	tests := []struct {
		name           string
		healthSources  []HealthCheckSource
		expectedStatus int
		expectMetadata bool
	}{
		{
			name: "all checks healthy, deferring or warning returns OK",
			healthSources: []HealthCheckSource{
				&testHealthCheckSource{
					healthStatus: health.HealthStatus{
						Checks: map[health.CheckType]health.HealthCheckResult{
							"check1": {
								State: health.New_HealthState(health.HealthState_HEALTHY),
							},
						},
					},
				},
				&testHealthCheckSource{
					healthStatus: health.HealthStatus{
						Checks: map[health.CheckType]health.HealthCheckResult{
							"check2": {
								State: health.New_HealthState(health.HealthState_DEFERRING),
							},
						},
					},
				},
				&testHealthCheckSource{
					healthStatus: health.HealthStatus{
						Checks: map[health.CheckType]health.HealthCheckResult{
							"check3": {
								State: health.New_HealthState(health.HealthState_WARNING),
							},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectMetadata: false,
		},
		{
			name: "mixed ready states returns OK",
			healthSources: []HealthCheckSource{
				&testHealthCheckSource{
					healthStatus: health.HealthStatus{
						Checks: map[health.CheckType]health.HealthCheckResult{
							"check1": {
								State: health.New_HealthState(health.HealthState_HEALTHY),
							},
							"check2": {
								State: health.New_HealthState(health.HealthState_DEFERRING),
							},
							"check3": {
								State: health.New_HealthState(health.HealthState_WARNING),
							},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectMetadata: false,
		},
		{
			name: "error state returns error status code",
			healthSources: []HealthCheckSource{
				&testHealthCheckSource{
					healthStatus: health.HealthStatus{
						Checks: map[health.CheckType]health.HealthCheckResult{
							"check1": {
								State: health.New_HealthState(health.HealthState_ERROR),
							},
						},
					},
				},
			},
			expectedStatus: 522,
			expectMetadata: true,
		},
		{
			name: "suspended state returns suspended status code",
			healthSources: []HealthCheckSource{
				&testHealthCheckSource{
					healthStatus: health.HealthStatus{
						Checks: map[health.CheckType]health.HealthCheckResult{
							"check1": {
								State: health.New_HealthState(health.HealthState_SUSPENDED),
							},
						},
					},
				},
			},
			expectedStatus: 519,
			expectMetadata: true,
		},
		{
			name: "repairing state returns repairing status code",
			healthSources: []HealthCheckSource{
				&testHealthCheckSource{
					healthStatus: health.HealthStatus{
						Checks: map[health.CheckType]health.HealthCheckResult{
							"check1": {
								State: health.New_HealthState(health.HealthState_REPAIRING),
							},
						},
					},
				},
			},
			expectedStatus: 520,
			expectMetadata: true,
		},
		{
			name: "terminal state returns terminal status code",
			healthSources: []HealthCheckSource{
				&testHealthCheckSource{
					healthStatus: health.HealthStatus{
						Checks: map[health.CheckType]health.HealthCheckResult{
							"check1": {
								State: health.New_HealthState(health.HealthState_TERMINAL),
							},
						},
					},
				},
			},
			expectedStatus: 523,
			expectMetadata: true,
		},
		{
			name: "multiple sources with one non-ready returns error",
			healthSources: []HealthCheckSource{
				&testHealthCheckSource{
					healthStatus: health.HealthStatus{
						Checks: map[health.CheckType]health.HealthCheckResult{
							"check1": {
								State: health.New_HealthState(health.HealthState_HEALTHY),
							},
						},
					},
				},
				&testHealthCheckSource{
					healthStatus: health.HealthStatus{
						Checks: map[health.CheckType]health.HealthCheckResult{
							"check2": {
								State: health.New_HealthState(health.HealthState_SUSPENDED),
							},
						},
					},
				},
			},
			expectedStatus: 519,
			expectMetadata: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := HealthBasedReadinessSource(refreshable.New(tt.healthSources))
			status, metadata := source.Status()
			assert.Equal(t, tt.expectedStatus, status)
			if tt.expectMetadata {
				assert.NotNil(t, metadata)
			} else {
				assert.Nil(t, metadata)
			}
		})
	}
}

func TestHealthCheckSourceFunc(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "sentinel")
	expected := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"check": {
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
		},
	}

	var source HealthCheckSource = HealthCheckSourceFunc(func(gotCtx context.Context) health.HealthStatus {
		assert.Equal(t, ctx, gotCtx)
		return expected
	})

	assert.Equal(t, expected, source.HealthStatus(ctx))
}

func TestHealthCheckSourceFuncNil(t *testing.T) {
	var source HealthCheckSource = HealthCheckSourceFunc(nil)
	assert.Equal(t, health.HealthStatus{}, source.HealthStatus(context.Background()))
}
