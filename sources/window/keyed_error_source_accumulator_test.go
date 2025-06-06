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
	"errors"
	"testing"

	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
	"github.com/stretchr/testify/assert"
)

func TestKeyedErrorSourceAccumulatorCanError(t *testing.T) {
	keyedErrorSourceAccumulator := NewDefaultKeyedErrorSourceAccumulator(MustNewKeyedErrorHealthCheckSource("check", UnhealthyIfAtLeastOneError))
	keyedErrorSourceAccumulator.Submit("key", errors.New("uhoh"))
	str := ""
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"check": {
				Type:    "check",
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &str,
				Params: map[string]interface{}{
					"key": "uhoh",
				},
			},
		},
	}, keyedErrorSourceAccumulator.HealthStatus(context.Background()))
}

func TestKeyedErrorSourceAccumulatorCanAdd(t *testing.T) {
	keyedErrorSourceAccumulator := NewDefaultKeyedErrorSourceAccumulator(MustNewKeyedErrorHealthCheckSource("check", UnhealthyIfAtLeastOneError))
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"check": {
				Type:  "check",
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
		},
	}, keyedErrorSourceAccumulator.HealthStatus(context.Background()))
	anotherCheck := MustNewKeyedErrorHealthCheckSource("check2", UnhealthyIfAtLeastOneError)
	keyedErrorSourceAccumulator.AddHealthCheckSource(anotherCheck)
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"check": {
				Type:  "check",
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
			"check2": {
				Type:  "check2",
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
		},
	}, keyedErrorSourceAccumulator.HealthStatus(context.Background()))
	anotherCheck.Submit("key", errors.New("uhoh"))
	str := ""
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"check": {
				Type:  "check",
				State: health.New_HealthState(health.HealthState_HEALTHY),
			},
			"check2": {
				Type:    "check2",
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &str,
				Params: map[string]interface{}{
					"key": "uhoh",
				},
			},
		},
	}, keyedErrorSourceAccumulator.HealthStatus(context.Background()))
}
