// Copyright (c) 2019 Palantir Technologies. All rights reserved.
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
	"testing"
	"time"

	werror "github.com/palantir/witchcraft-go-error"
	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCheckType health.CheckType = "TEST_CHECK"
	windowSize                     = 100 * time.Millisecond
)

func TestUnhealthyIfAtLeastOneErrorSource(t *testing.T) {
	checkMessage := "found an error"
	for _, testCase := range []struct {
		name          string
		errors        []error
		expectedCheck health.HealthCheckResult
	}{
		{
			name:          "healthy when there are no items",
			errors:        nil,
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "healthy when there are only nil items",
			errors: []error{
				nil,
				nil,
				nil,
			},
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "unhealthy when there is at least one err",
			errors: []error{
				nil,
				werror.ErrorWithContextParams(context.Background(), "Error #1"),
				nil,
				werror.ErrorWithContextParams(context.Background(), "Error #2", werror.SafeParam("foo", "bar")),
				nil,
			},
			expectedCheck: sources.UnhealthyHealthCheckResult(testCheckType, checkMessage, map[string]interface{}{
				"error": "Error #2",
			}),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			timeProvider := &offsetTimeProvider{}
			source, err := NewErrorHealthCheckSource(testCheckType, UnhealthyIfAtLeastOneError,
				WithWindowSize(time.Hour),
				WithCheckMessage(checkMessage),
				WithTimeProvider(timeProvider))
			require.NoError(t, err)

			for _, err := range testCase.errors {
				source.Submit(err)
			}
			actualStatus := source.HealthStatus(context.Background())
			expectedStatus := health.HealthStatus{
				Checks: map[health.CheckType]health.HealthCheckResult{
					testCheckType: testCase.expectedCheck,
				},
			}
			assert.Equal(t, expectedStatus, actualStatus)
		})
	}
}

func TestHealthyIfNotAllErrorsSource(t *testing.T) {
	checkMessage := "found an error"
	for _, testCase := range []struct {
		name          string
		errors        []error
		expectedCheck health.HealthCheckResult
	}{
		{
			name:          "healthy when there are no items",
			errors:        nil,
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "healthy when there are only nil items",
			errors: []error{
				nil,
				nil,
				nil,
			},
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "healthy when there is at least one non nil err",
			errors: []error{
				nil,
				werror.ErrorWithContextParams(context.Background(), "Error #1"),
				nil,
				werror.ErrorWithContextParams(context.Background(), "Error #2"),
				nil,
			},
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "unhealthy when there are only non nil items",
			errors: []error{
				werror.ErrorWithContextParams(context.Background(), "Error #1"),
				werror.ErrorWithContextParams(context.Background(), "Error #2", werror.SafeParam("foo", "bar")),
			},
			expectedCheck: sources.UnhealthyHealthCheckResult(testCheckType, checkMessage, map[string]interface{}{
				"error": "Error #2",
			}),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			timeProvider := &offsetTimeProvider{}
			source, err := NewErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
				WithWindowSize(time.Hour),
				WithCheckMessage(checkMessage),
				WithTimeProvider(timeProvider))

			require.NoError(t, err)
			for _, err := range testCase.errors {
				source.Submit(err)
			}
			actualStatus := source.HealthStatus(context.Background())
			expectedStatus := health.HealthStatus{
				Checks: map[health.CheckType]health.HealthCheckResult{
					testCheckType: testCase.expectedCheck,
				},
			}
			assert.Equal(t, expectedStatus, actualStatus)
		})
	}
}

// TestHealthyIfNotAllErrorsSource_ErrorInInitialWindowWhenFirstFullWindowRequired validates that error in the first window
// causes the check to report as repairing when first window is required.
func TestHealthyIfNotAllErrorsSource_RequireFullWindow_ErrorInInitialWindow(t *testing.T) {
	timeProvider := &offsetTimeProvider{}
	source, err := NewErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRequireFullWindow(),
		WithTimeProvider(timeProvider))
	assert.NoError(t, err)

	source.Submit(werror.ErrorWithContextParams(context.Background(), "an error"))
	healthStatus := source.HealthStatus(context.Background())
	checkResult, ok := healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_REPAIRING, checkResult.State.Value())
}

// TestHealthyIfNotAllErrorsSource_RepairingGracePeriod_ErrorInInitialAnchoredWindow validates that error in the first window
// does not cause the health status to become unhealthy when anchored as well.
func TestHealthyIfNotAllErrorsSource_RepairingGracePeriod_ErrorInInitialAnchoredWindow(t *testing.T) {
	timeProvider := &offsetTimeProvider{}
	source, err := NewErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRepairingGracePeriod(windowSize),
		WithTimeProvider(timeProvider))
	assert.NoError(t, err)

	source.Submit(werror.ErrorWithContextParams(context.Background(), "an error"))
	healthStatus := source.HealthStatus(context.Background())
	checkResult, ok := healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_REPAIRING, checkResult.State.Value())
}

// TestHealthyIfNotAllErrorsSource_RepairingGracePeriod_GapThenRepairing validates that error in the first window
// does not cause the health status to become unhealthy when anchored as well.
func TestHealthyIfNotAllErrorsSource_RepairingGracePeriod_GapThenRepairing(t *testing.T) {
	timeProvider := &offsetTimeProvider{}
	source, err := NewErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRepairingGracePeriod(windowSize),
		WithRequireFullWindow(),
		WithTimeProvider(timeProvider))
	assert.NoError(t, err)

	timeProvider.RestlessSleep(2 * windowSize)
	source.Submit(werror.ErrorWithContextParams(context.Background(), "an error"))
	timeProvider.RestlessSleep(windowSize / 2)

	healthStatus := source.HealthStatus(context.Background())
	checkResult, ok := healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_REPAIRING, checkResult.State.Value())
}

// TestHealthyIfNotAllErrorsSource_RepairingGracePeriod_GapThenRepairingThenError validates that in a constant stream of errors, the health
// check initially reports repairing and then reports error after the time window.
func TestHealthyIfNotAllErrorsSource_RepairingGracePeriod_GapThenRepairingThenError(t *testing.T) {
	timeProvider := &offsetTimeProvider{}
	source, err := NewErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRepairingGracePeriod(windowSize),
		WithRequireFullWindow(),
		WithTimeProvider(timeProvider))
	assert.NoError(t, err)

	timeProvider.RestlessSleep(2 * windowSize)
	source.Submit(werror.ErrorWithContextParams(context.Background(), "an error"))
	timeProvider.RestlessSleep(windowSize / 2)
	source.Submit(werror.ErrorWithContextParams(context.Background(), "an error"))

	healthStatus := source.HealthStatus(context.Background())
	checkResult, ok := healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_REPAIRING, checkResult.State.Value())

	timeProvider.RestlessSleep(windowSize / 2)
	source.Submit(werror.ErrorWithContextParams(context.Background(), "an error"))

	healthStatus = source.HealthStatus(context.Background())
	checkResult, ok = healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_ERROR, checkResult.State.Value())
}

// TestHealthyIfNotAllErrorsSource_RepairingGracePeriod_GapThenRepairingThenHealthy validates that if a success is submitted during repairing phase,
// the health check recovers.
func TestHealthyIfNotAllErrorsSource_RepairingGracePeriod_GapThenRepairingThenHealthy(t *testing.T) {
	timeProvider := &offsetTimeProvider{}
	source, err := NewErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRepairingGracePeriod(windowSize),
		WithRequireFullWindow(),
		WithTimeProvider(timeProvider))
	assert.NoError(t, err)

	timeProvider.RestlessSleep(2 * windowSize)
	source.Submit(werror.ErrorWithContextParams(context.Background(), "an error"))
	timeProvider.RestlessSleep(windowSize / 2)
	source.Submit(werror.ErrorWithContextParams(context.Background(), "an error"))

	healthStatus := source.HealthStatus(context.Background())
	checkResult, ok := healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_REPAIRING, checkResult.State.Value())

	timeProvider.RestlessSleep(windowSize / 2)
	source.Submit(nil)

	healthStatus = source.HealthStatus(context.Background())
	checkResult, ok = healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_HEALTHY, checkResult.State.Value())
}

// TestHealthyIfNotAllErrorsSource_RepairingGracePeriod_RepairingThenGap validates if no more errors happen beyond the repairing phase,
// the health check recovers.
func TestHealthyIfNotAllErrorsSource_RepairingGracePeriod_RepairingThenGap(t *testing.T) {
	timeProvider := &offsetTimeProvider{}
	source, err := NewErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRepairingGracePeriod(windowSize),
		WithRequireFullWindow(),
		WithTimeProvider(timeProvider))
	assert.NoError(t, err)

	timeProvider.RestlessSleep(2 * windowSize)
	source.Submit(werror.ErrorWithContextParams(context.Background(), "an error"))
	timeProvider.RestlessSleep(windowSize / 2)
	source.Submit(werror.ErrorWithContextParams(context.Background(), "an error"))

	healthStatus := source.HealthStatus(context.Background())
	checkResult, ok := healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_REPAIRING, checkResult.State.Value())

	timeProvider.RestlessSleep(3 * windowSize / 2)

	healthStatus = source.HealthStatus(context.Background())
	checkResult, ok = healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_HEALTHY, checkResult.State.Value())
}

// TestHealthyIfNotAllErrorsSource_MaximumErrorAge validates that errors that happened more than the max age ago
// only cause the health check to report repairing.
func TestHealthyIfNotAllErrorsSource_MaximumErrorAge(t *testing.T) {
	timeProvider := &offsetTimeProvider{}
	source, err := NewErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithMaximumErrorAge(windowSize/2),
		WithTimeProvider(timeProvider))
	assert.NoError(t, err)

	source.Submit(werror.ErrorWithContextParams(context.Background(), "an error"))
	timeProvider.RestlessSleep(windowSize / 4)

	healthStatus := source.HealthStatus(context.Background())
	checkResult, ok := healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_ERROR, checkResult.State.Value())

	timeProvider.RestlessSleep(windowSize / 2)

	healthStatus = source.HealthStatus(context.Background())
	checkResult, ok = healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_REPAIRING, checkResult.State.Value())

	source.Submit(werror.ErrorWithContextParams(context.Background(), "an error"))

	healthStatus = source.HealthStatus(context.Background())
	checkResult, ok = healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_ERROR, checkResult.State.Value())
}

func TestHealthyIfNoRecentErrorsSource(t *testing.T) {
	checkMessage := "found an error"
	for _, testCase := range []struct {
		name          string
		errors        []error
		expectedCheck health.HealthCheckResult
	}{
		{
			name:          "healthy when there are no items",
			errors:        nil,
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "healthy when there are only nil items",
			errors: []error{
				nil,
				nil,
				nil,
			},
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "healthy when latest submission is a nil err",
			errors: []error{
				nil,
				werror.ErrorWithContextParams(context.Background(), "Error #1"),
				nil,
				werror.ErrorWithContextParams(context.Background(), "Error #2"),
				nil,
			},
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "healthy when latest submission is a non nil err",
			errors: []error{
				werror.ErrorWithContextParams(context.Background(), "Error #1"),
				nil,
				werror.ErrorWithContextParams(context.Background(), "Error #2", werror.SafeParam("foo", "bar")),
			},
			expectedCheck: sources.UnhealthyHealthCheckResult(testCheckType, checkMessage, map[string]interface{}{
				"error": "Error #2",
			}),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			timeProvider := &offsetTimeProvider{}
			source, err := NewErrorHealthCheckSource(testCheckType, HealthyIfNoRecentErrors,
				WithWindowSize(time.Hour),
				WithCheckMessage(checkMessage),
				WithTimeProvider(timeProvider))

			require.NoError(t, err)
			for _, err := range testCase.errors {
				source.Submit(err)
			}
			actualStatus := source.HealthStatus(context.Background())
			expectedStatus := health.HealthStatus{
				Checks: map[health.CheckType]health.HealthCheckResult{
					testCheckType: testCase.expectedCheck,
				},
			}
			assert.Equal(t, expectedStatus, actualStatus)
		})
	}
}

// TestHealthStateValue asserts the behavior of overriding the default ERROR health state.
func TestHealthStateValue(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		healthState health.HealthState_Value
	}{
		{health.HealthState_HEALTHY},
		{health.HealthState_DEFERRING},
		{health.HealthState_SUSPENDED},
		{health.HealthState_REPAIRING},
		{health.HealthState_WARNING},
		{health.HealthState_ERROR},
		{health.HealthState_TERMINAL},
	} {
		t.Run(string(tc.healthState), func(t *testing.T) {
			source, err := NewErrorHealthCheckSource(testCheckType, HealthyIfNoRecentErrors,
				WithHealthStateValue(tc.healthState),
				WithWindowSize(time.Hour),
				WithTimeProvider(&offsetTimeProvider{}))
			require.NoError(t, err)

			// submit the error and validate the health state value
			source.Submit(werror.ErrorWithContextParams(ctx, "an error"))
			healthStatus := source.HealthStatus(ctx)
			checkResult, ok := healthStatus.Checks[testCheckType]
			assert.True(t, ok)
			assert.Equal(t, tc.healthState, checkResult.State.Value())
		})
	}
}

// TestHealthStateOverrideWithMaxAge ensures overriding the health state to something other than the default
// does not affect the REPAIRING health state being set when using the maximum error age option.
func TestHealthStateOverrideWithMaxAge(t *testing.T) {
	ctx := context.Background()
	timeProvider := &offsetTimeProvider{}
	source, err := NewErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithHealthStateValue(health.HealthState_WARNING),
		WithWindowSize(windowSize),
		WithMaximumErrorAge(windowSize/2),
		WithTimeProvider(timeProvider))
	assert.NoError(t, err)

	source.Submit(werror.ErrorWithContextParams(ctx, "an error"))

	// Initial state before maximum error age is reached is expected to be WARNING
	timeProvider.RestlessSleep(windowSize / 4)
	healthStatus := source.HealthStatus(ctx)
	checkResult, ok := healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_WARNING, checkResult.State.Value())

	// Sleeping for the maximum error age should now produce a REPAIRING health check
	timeProvider.RestlessSleep(windowSize / 2)
	healthStatus = source.HealthStatus(ctx)
	checkResult, ok = healthStatus.Checks[testCheckType]
	assert.True(t, ok)
	assert.Equal(t, health.HealthState_REPAIRING, checkResult.State.Value())

}
