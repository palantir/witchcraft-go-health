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

	"github.com/palantir/witchcraft-go-health/v2/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/v2/sources"
	"github.com/stretchr/testify/assert"
)

func TestKeyedErrorSourceAccumulatorCanError(t *testing.T) {
	keyedErrorSourceAccumulator := NewDefaultKeyedErrorSourceAccumulator(MustNewKeyedErrorHealthCheckSource("check", UnhealthyIfAtLeastOneError))
	keyedErrorSourceAccumulator.Submit(context.Background(), "key", errors.New("uhoh"))
	str := ""
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"check": {
				Type:    "check",
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &str,
				Params: map[string]any{
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
	anotherCheck.Submit(context.Background(), "key", errors.New("uhoh"))
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
				Params: map[string]any{
					"key": "uhoh",
				},
			},
		},
	}, keyedErrorSourceAccumulator.HealthStatus(context.Background()))
}

func TestKeyedErrorSourceAccumulatorRootUnderTreeMergesHealthyParentChecks(t *testing.T) {
	keyedErrorSourceAccumulator := NewDefaultKeyedErrorSourceAccumulator(MustNewKeyedErrorHealthCheckSource("check", UnhealthyIfAtLeastOneError))
	keyedErrorSourceAccumulator.RootUnderTree(sources.StaticHealthCheckSource(health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"PARENT": sources.HealthyHealthCheckResult("PARENT"),
		},
	}))
	keyedErrorSourceAccumulator.Submit(context.Background(), "key", errors.New("uhoh"))
	str := ""
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"PARENT": sources.HealthyHealthCheckResult("PARENT"),
			"check": {
				Type:    "check",
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &str,
				Params: map[string]any{
					"key": "uhoh",
				},
			},
		},
	}, keyedErrorSourceAccumulator.HealthStatus(context.Background()))
}

func TestKeyedErrorSourceAccumulatorRootUnderTreeUnhealthyParentBlocksChildren(t *testing.T) {
	keyedErrorSourceAccumulator := NewDefaultKeyedErrorSourceAccumulator(MustNewKeyedErrorHealthCheckSource("check", UnhealthyIfAtLeastOneError))
	unhealthyParent := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"PARENT": {
				Type:  "PARENT",
				State: health.New_HealthState(health.HealthState_ERROR),
			},
		},
	}
	keyedErrorSourceAccumulator.RootUnderTree(sources.StaticHealthCheckSource(unhealthyParent))
	keyedErrorSourceAccumulator.Submit(context.Background(), "key", errors.New("uhoh"))
	assert.Equal(t, unhealthyParent, keyedErrorSourceAccumulator.HealthStatus(context.Background()))
}

func TestKeyedErrorSourceAccumulatorRootUnderTreeReplacesPriorRoot(t *testing.T) {
	keyedErrorSourceAccumulator := NewDefaultKeyedErrorSourceAccumulator(MustNewKeyedErrorHealthCheckSource("check", UnhealthyIfAtLeastOneError))
	keyedErrorSourceAccumulator.RootUnderTree(sources.StaticHealthCheckSource(health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"FIRST": sources.HealthyHealthCheckResult("FIRST"),
		},
	}))
	keyedErrorSourceAccumulator.RootUnderTree(sources.StaticHealthCheckSource(health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"SECOND": sources.HealthyHealthCheckResult("SECOND"),
		},
	}))
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"SECOND": sources.HealthyHealthCheckResult("SECOND"),
			"check":  sources.HealthyHealthCheckResult("check"),
		},
	}, keyedErrorSourceAccumulator.HealthStatus(context.Background()))
}
