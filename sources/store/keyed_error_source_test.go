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

package store

import (
	"context"
	"fmt"
	"testing"

	werror "github.com/palantir/witchcraft-go-error"
	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
	"github.com/stretchr/testify/assert"
)

var (
	testMessage = "test message"
)

func TestKeyedMessengerHealthStateError(t *testing.T) {
	keyedErrorSource := NewKeyedErrorHealthCheckSource("TEST", testMessage)
	keyedErrorSource.Submit("1", werror.Error("error message 1", werror.SafeParam("foo", "baz")))
	keyedErrorSource.Submit("2", werror.Error("error message 2", werror.SafeParam("foo", "bar")))
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"TEST": {
				Message: &testMessage,
				Params: map[string]interface{}{
					"1":     "error message 1",
					"2":     "error message 2",
					"1-foo": "baz",
					"2-foo": "bar",
				},
				State: health.New_HealthState(health.HealthState_ERROR),
				Type:  "TEST",
			},
		},
	}, keyedErrorSource.HealthStatus(context.Background()))
}

func TestKeyedMessengerHealthStateHealthy(t *testing.T) {
	keyedErrorSource := NewKeyedErrorHealthCheckSource("TEST", testMessage)
	keyedErrorSource.Submit("1", fmt.Errorf("error message 1"))
	keyedErrorSource.Submit("1", nil)
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"TEST": {
				Message: &testMessage,
				State:   health.New_HealthState(health.HealthState_HEALTHY),
				Type:    "TEST",
			},
		},
	}, keyedErrorSource.HealthStatus(context.Background()))
}

func TestKeyedMessengerHealthStateErrorAndPreserve(t *testing.T) {
	keyedErrorSource := NewKeyedErrorHealthCheckSource("TEST", testMessage)
	keyedErrorSource.Submit("1", werror.Error("error message 1", werror.SafeParam("foo", "baz")))
	keyedErrorSource.Submit("2", nil)
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"TEST": {
				Message: &testMessage,
				Params: map[string]interface{}{
					"1":     "error message 1",
					"1-foo": "baz",
				},
				State: health.New_HealthState(health.HealthState_ERROR),
				Type:  "TEST",
			},
		},
	}, keyedErrorSource.HealthStatus(context.Background()))
	keyedErrorSource.PreserveKeys([]string{"2"})
	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"TEST": {
				Message: &testMessage,
				State:   health.New_HealthState(health.HealthState_HEALTHY),
				Type:    "TEST",
			},
		},
	}, keyedErrorSource.HealthStatus(context.Background()))
}
