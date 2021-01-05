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

type keyErrorPair struct {
	key string
	err error
}

func TestKeyedUnhealthyIfAtLeastOneErrorSource(t *testing.T) {
	checkMessage := "message in case of error"
	for _, testCase := range []struct {
		name          string
		keyErrorPairs []keyErrorPair
		expectedCheck health.HealthCheckResult
	}{
		{
			name:          "healthy when there are no items",
			keyErrorPairs: nil,
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "healthy when all keys are completely healthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1"},
				{key: "1"},
				{key: "2"},
				{key: "3"},
			},
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "unhealthy when some keys are partially healthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1"},
				{key: "1", err: werror.Error("Error #1 for key 1")},
				{key: "1"},
				{key: "2", err: werror.Error("Error #1 for key 2")},
				{key: "2"},
				{key: "3"},
			},
			expectedCheck: health.HealthCheckResult{
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "Error #1 for key 1",
					"2": "Error #1 for key 2",
				},
			},
		},
		{
			name: "unhealthy when all keys are completely unhealthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1", err: werror.Error("Error #1 for key 1")},
				{key: "2", err: werror.Error("Error #1 for key 2")},
				{key: "2", err: werror.Error("Error #2 for key 2")},
				{key: "3", err: werror.Error("Error #1 for key 3")},
			},
			expectedCheck: health.HealthCheckResult{
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "Error #1 for key 1",
					"2": "Error #2 for key 2",
					"3": "Error #1 for key 3",
				},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			source, err := NewKeyedErrorHealthCheckSource(testCheckType, UnhealthyIfAtLeastOneError,
				WithWindowSize(time.Hour),
				WithCheckMessage(checkMessage))
			require.NoError(t, err)
			for _, keyErrorPair := range testCase.keyErrorPairs {
				source.Submit(keyErrorPair.key, keyErrorPair.err)
			}
			expectedStatus := health.HealthStatus{
				Checks: map[health.CheckType]health.HealthCheckResult{
					testCheckType: testCase.expectedCheck,
				},
			}
			actualStatus := source.HealthStatus(context.Background())
			assert.Equal(t, expectedStatus, actualStatus)
		})
	}
}

func TestKeyedHealthyIfNotAllErrorsSource_OutsideStartWindow(t *testing.T) {
	checkMessage := "message in case of error"
	for _, testCase := range []struct {
		name          string
		keyErrorPairs []keyErrorPair
		expectedCheck health.HealthCheckResult
	}{
		{
			name:          "healthy when there are no items",
			keyErrorPairs: nil,
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "healthy when all keys are completely healthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1"},
				{key: "1"},
				{key: "2"},
				{key: "3"},
			},
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "healthy when all keys are partially healthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1"},
				{key: "1", err: werror.Error("Error #1 for key 1")},
				{key: "1"},
				{key: "2", err: werror.Error("Error #1 for key 2")},
				{key: "2"},
				{key: "3"},
				{key: "3", err: werror.Error("Error #1 for key 3")},
				{key: "3", err: werror.Error("Error #2 for key 3")},
			},
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "unhealthy when some keys are completely unhealthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1", err: werror.Error("Error #1 for key 1")},
				{key: "2", err: werror.Error("Error #1 for key 2")},
				{key: "2", err: werror.Error("Error #2 for key 2")},
				{key: "3"},
				{key: "3", err: werror.Error("Error #1 for key 3")},
			},
			expectedCheck: health.HealthCheckResult{
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "Error #1 for key 1",
					"2": "Error #2 for key 2",
				},
			},
		},
		{
			name: "unhealthy when all keys are completely unhealthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1", err: werror.Error("Error #1 for key 1")},
				{key: "2", err: werror.Error("Error #1 for key 2")},
				{key: "2", err: werror.Error("Error #2 for key 2")},
				{key: "3", err: werror.Error("Error #1 for key 3")},
			},
			expectedCheck: health.HealthCheckResult{
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "Error #1 for key 1",
					"2": "Error #2 for key 2",
					"3": "Error #1 for key 3",
				},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			timeProvider := &offsetTimeProvider{}
			source, err := NewKeyedErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
				WithWindowSize(time.Hour),
				WithCheckMessage(checkMessage),
				WithRequireFullWindow(),
				WithTimeProvider(timeProvider))
			require.NoError(t, err)

			// sleep puts all tests outside the required healthy start window
			timeProvider.RestlessSleep(time.Hour)

			for _, keyErrorPair := range testCase.keyErrorPairs {
				source.Submit(keyErrorPair.key, keyErrorPair.err)
			}
			expectedStatus := health.HealthStatus{
				Checks: map[health.CheckType]health.HealthCheckResult{
					testCheckType: testCase.expectedCheck,
				},
			}
			actualStatus := source.HealthStatus(context.Background())
			assert.Equal(t, expectedStatus, actualStatus)
		})
	}
}

func TestKeyedHealthyIfNotAllErrorsSource_InsideStartWindow(t *testing.T) {
	checkMessage := "message in case of error"
	for _, testCase := range []struct {
		name          string
		keyErrorPairs []keyErrorPair
		expectedCheck health.HealthCheckResult
	}{
		{
			name: "healthy when all keys are partially healthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1"},
				{key: "1", err: werror.Error("Error #1 for key 1")},
				{key: "1"},
				{key: "2", err: werror.Error("Error #1 for key 2")},
				{key: "2"},
				{key: "3"},
				{key: "3", err: werror.Error("Error #1 for key 3")},
				{key: "3", err: werror.Error("Error #2 for key 3")},
			},
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "healthy when some keys are completely unhealthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1", err: werror.Error("Error #1 for key 1")},
				{key: "2", err: werror.Error("Error #1 for key 2")},
				{key: "2", err: werror.Error("Error #2 for key 2")},
				{key: "3"},
				{key: "3", err: werror.Error("Error #1 for key 3")},
			},
			expectedCheck: health.HealthCheckResult{
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "Error #1 for key 1",
					"2": "Error #2 for key 2",
				},
			},
		},
		{
			name: "healthy when all keys are completely unhealthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1", err: werror.Error("Error #1 for key 1")},
				{key: "2", err: werror.Error("Error #1 for key 2")},
				{key: "2", err: werror.Error("Error #2 for key 2")},
				{key: "3", err: werror.Error("Error #1 for key 3")},
			},
			expectedCheck: health.HealthCheckResult{
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "Error #1 for key 1",
					"2": "Error #2 for key 2",
					"3": "Error #1 for key 3",
				},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			timeProvider := &offsetTimeProvider{}
			source, err := NewKeyedErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
				WithWindowSize(time.Hour),
				WithCheckMessage(checkMessage),
				WithRequireFullWindow(),
				WithTimeProvider(timeProvider))
			require.NoError(t, err)

			for _, keyErrorPair := range testCase.keyErrorPairs {
				source.Submit(keyErrorPair.key, keyErrorPair.err)
			}
			expectedStatus := health.HealthStatus{
				Checks: map[health.CheckType]health.HealthCheckResult{
					testCheckType: testCase.expectedCheck,
				},
			}
			actualStatus := source.HealthStatus(context.Background())
			assert.Equal(t, expectedStatus, actualStatus)
		})
	}
}

func TestKeyedHealthyIfNotAllErrorsSource_InitialWindowErrorsReturnRepairing(t *testing.T) {
	ctx := context.Background()
	checkMessage := "message in case of error"
	const timeWindow = time.Minute

	timeProvider := &offsetTimeProvider{}
	source, err := NewKeyedErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithWindowSize(timeWindow),
		WithCheckMessage(checkMessage),
		WithRequireFullWindow(),
		WithTimeProvider(timeProvider))
	require.NoError(t, err)

	// move partially into the initial health check window
	timeProvider.RestlessSleep(3 * timeWindow / 4)
	source.Submit("1", werror.ErrorWithContextParams(ctx, "error for key: 1"))

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
				},
			},
		},
	}, source.HealthStatus(ctx))

	// move out of the initial health check window but keep error inside sliding window
	timeProvider.RestlessSleep(timeWindow / 2)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
				},
			},
		},
	}, source.HealthStatus(ctx))
}

func TestKeyedHealthyIfNotAllErrorsSource_RepairingGracePeriod_GapThenRepairingThenHealthy(t *testing.T) {
	ctx := context.Background()
	checkMessage := "message in case of error"
	const timeWindow = time.Minute

	timeProvider := &offsetTimeProvider{}
	source, err := NewKeyedErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithWindowSize(timeWindow),
		WithCheckMessage(checkMessage),
		WithRepairingGracePeriod(timeWindow),
		WithRequireFullWindow(),
		WithTimeProvider(timeProvider))
	require.NoError(t, err)

	// move out of the initial health check window
	timeProvider.RestlessSleep(2 * timeWindow)

	source.Submit("1", werror.ErrorWithContextParams(ctx, "error for key: 1"))
	timeProvider.RestlessSleep(timeWindow / 2)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
				},
			},
		},
	}, source.HealthStatus(ctx))

	source.Submit("2", werror.ErrorWithContextParams(ctx, "error for key: 2"))
	timeProvider.RestlessSleep(timeWindow / 4)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
					"2": "error for key: 2",
				},
			},
		},
	}, source.HealthStatus(ctx))

	source.Submit("1", nil)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"2": "error for key: 2",
				},
			},
		},
	}, source.HealthStatus(ctx))

	source.Submit("2", nil)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: sources.HealthyHealthCheckResult(testCheckType),
		},
	}, source.HealthStatus(ctx))
}

func TestKeyedHealthyIfNotAllErrorsSource_RepairingGracePeriod_GapThenRepairingThenGap(t *testing.T) {
	ctx := context.Background()
	checkMessage := "message in case of error"
	const timeWindow = time.Minute

	timeProvider := &offsetTimeProvider{}
	source, err := NewKeyedErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithWindowSize(timeWindow),
		WithCheckMessage(checkMessage),
		WithRepairingGracePeriod(timeWindow),
		WithRequireFullWindow(),
		WithTimeProvider(timeProvider))
	require.NoError(t, err)

	// move out of the initial health check window
	timeProvider.RestlessSleep(2 * timeWindow)

	source.Submit("1", werror.ErrorWithContextParams(ctx, "error for key: 1"))
	timeProvider.RestlessSleep(timeWindow / 2)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
				},
			},
		},
	}, source.HealthStatus(ctx))

	source.Submit("2", werror.ErrorWithContextParams(ctx, "error for key: 2"))

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
					"2": "error for key: 2",
				},
			},
		},
	}, source.HealthStatus(ctx))

	timeProvider.RestlessSleep(3 * timeWindow / 4)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"2": "error for key: 2",
				},
			},
		},
	}, source.HealthStatus(ctx))

	timeProvider.RestlessSleep(timeWindow / 2)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: sources.HealthyHealthCheckResult(testCheckType),
		},
	}, source.HealthStatus(ctx))
}

func TestKeyedHealthyIfNotAllErrorsSource_RepairingGracePeriod_GapThenRepairingThenError(t *testing.T) {
	ctx := context.Background()
	checkMessage := "message in case of error"
	const timeWindow = time.Minute

	timeProvider := &offsetTimeProvider{}
	source, err := NewKeyedErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithWindowSize(timeWindow),
		WithCheckMessage(checkMessage),
		WithRepairingGracePeriod(timeWindow),
		WithRequireFullWindow(),
		WithTimeProvider(timeProvider))
	require.NoError(t, err)

	// move out of the initial health check window
	timeProvider.RestlessSleep(2 * timeWindow)

	source.Submit("1", werror.ErrorWithContextParams(ctx, "error for key: 1"))
	timeProvider.RestlessSleep(timeWindow / 2)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
				},
			},
		},
	}, source.HealthStatus(ctx))

	source.Submit("2", werror.ErrorWithContextParams(ctx, "error for key: 2"))
	timeProvider.RestlessSleep(timeWindow / 4)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
					"2": "error for key: 2",
				},
			},
		},
	}, source.HealthStatus(ctx))

	source.Submit("1", werror.ErrorWithContextParams(ctx, "error for key: 1"))
	timeProvider.RestlessSleep(timeWindow / 2)
	source.Submit("1", werror.ErrorWithContextParams(ctx, "error for key: 1"))

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
					"2": "error for key: 2",
				},
			},
		},
	}, source.HealthStatus(ctx))

	timeProvider.RestlessSleep(timeWindow / 2)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
				},
			},
		},
	}, source.HealthStatus(ctx))
}

func TestKeyedHealthyIfNotAllErrorsSource_MaximumErrorAge(t *testing.T) {
	ctx := context.Background()
	checkMessage := "message in case of error"
	const timeWindow = time.Minute

	timeProvider := &offsetTimeProvider{}
	source, err := NewKeyedErrorHealthCheckSource(testCheckType, HealthyIfNotAllErrors,
		WithWindowSize(timeWindow),
		WithCheckMessage(checkMessage),
		WithMaximumErrorAge(timeWindow/2),
		WithTimeProvider(timeProvider))
	require.NoError(t, err)

	source.Submit("1", werror.ErrorWithContextParams(ctx, "error for key: 1"))
	timeProvider.RestlessSleep(timeWindow / 4)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
				},
			},
		},
	}, source.HealthStatus(ctx))

	timeProvider.RestlessSleep(timeWindow / 2)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
				},
			},
		},
	}, source.HealthStatus(ctx))

	source.Submit("2", werror.ErrorWithContextParams(ctx, "error for key: 2"))

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "error for key: 1",
					"2": "error for key: 2",
				},
			},
		},
	}, source.HealthStatus(ctx))

	timeProvider.RestlessSleep(timeWindow / 2)

	assert.Equal(t, health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			testCheckType: {
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_REPAIRING),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"2": "error for key: 2",
				},
			},
		},
	}, source.HealthStatus(ctx))
}

func TestKeyedUnhealthyIfNoRecentErrorsSource(t *testing.T) {
	checkMessage := "message in case of error"
	for _, testCase := range []struct {
		name                     string
		keyErrorPairs            []keyErrorPair
		expectedCheck            health.HealthCheckResult
		durationAfterSubmissions time.Duration
	}{
		{
			name:          "healthy when there are no items",
			keyErrorPairs: nil,
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "healthy when all keys are completely healthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1"},
				{key: "1"},
				{key: "2"},
				{key: "3"},
			},
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "unhealthy when some keys are unhealthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1"},
				{key: "1", err: werror.Error("Error #1 for key 1")},
				{key: "1"},
				{key: "2", err: werror.Error("Error #1 for key 2")},
				{key: "3"},
			},
			expectedCheck: health.HealthCheckResult{
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"2": "Error #1 for key 2",
				},
			},
		},
		{
			name: "healthy when outside window",
			keyErrorPairs: []keyErrorPair{
				{key: "1", err: werror.Error("Error #1 for key 1")},
			},
			expectedCheck:            sources.HealthyHealthCheckResult(testCheckType),
			durationAfterSubmissions: time.Hour,
		},
		{
			name: "unhealthy when inside window",
			keyErrorPairs: []keyErrorPair{
				{key: "1", err: werror.Error("Error #1 for key 1")},
				{key: "1", err: werror.Error("Error #2 for key 1")},
			},
			expectedCheck: health.HealthCheckResult{
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "Error #2 for key 1",
				},
			},
		},
		{
			name: "healthy when last keys are healthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1"},
				{key: "1", err: werror.Error("Error #1 for key 1")},
				{key: "1"},
				{key: "2", err: werror.Error("Error #1 for key 2")},
				{key: "2"},
				{key: "3"},
			},
			expectedCheck: sources.HealthyHealthCheckResult(testCheckType),
		},
		{
			name: "unhealthy when all keys are completely unhealthy",
			keyErrorPairs: []keyErrorPair{
				{key: "1", err: werror.Error("Error #1 for key 1")},
				{key: "2", err: werror.Error("Error #1 for key 2")},
				{key: "2", err: werror.Error("Error #2 for key 2")},
				{key: "3", err: werror.Error("Error #1 for key 3")},
			},
			expectedCheck: health.HealthCheckResult{
				Type:    testCheckType,
				State:   health.New_HealthState(health.HealthState_ERROR),
				Message: &checkMessage,
				Params: map[string]interface{}{
					"1": "Error #1 for key 1",
					"2": "Error #2 for key 2",
					"3": "Error #1 for key 3",
				},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			timeProvider := &offsetTimeProvider{}
			source, err := NewKeyedErrorHealthCheckSource(testCheckType, HealthyIfNoRecentErrors,
				WithWindowSize(time.Minute),
				WithCheckMessage(checkMessage),
				WithTimeProvider(timeProvider))
			require.NoError(t, err)
			for _, keyErrorPair := range testCase.keyErrorPairs {
				source.Submit(keyErrorPair.key, keyErrorPair.err)
			}
			timeProvider.RestlessSleep(testCase.durationAfterSubmissions)
			expectedStatus := health.HealthStatus{
				Checks: map[health.CheckType]health.HealthCheckResult{
					testCheckType: testCase.expectedCheck,
				},
			}
			actualStatus := source.HealthStatus(context.Background())
			assert.Equal(t, expectedStatus, actualStatus)
		})
	}
}
