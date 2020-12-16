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

package window

import (
	"context"
	"sync"
	"time"

	werror "github.com/palantir/witchcraft-go-error"
	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/sources"
	"github.com/palantir/witchcraft-go-health/status"
)

// ErrorSubmitter allows components whose functionality dictates a portion of health status to only consume this interface.
type ErrorSubmitter interface {
	Submit(error)
}

// ErrorHealthCheckSource is a health check source with statuses determined by submitted errors.
type ErrorHealthCheckSource interface {
	ErrorSubmitter
	status.HealthCheckSource
}

// MustNewErrorHealthCheckSource creates a new ErrorHealthCheckSource panicking in case of error.
// Should only be used in instances where the inputs are statically defined and known to be valid.
func MustNewErrorHealthCheckSource(checkType health.CheckType, errorMode ErrorMode, options ...ErrorOption) ErrorHealthCheckSource {
	source, err := NewErrorHealthCheckSource(checkType, errorMode, options...)
	if err != nil {
		panic(err)
	}
	return source
}

// NewErrorHealthCheckSource creates a new ErrorHealthCheckSource.
func NewErrorHealthCheckSource(checkType health.CheckType, errorMode ErrorMode, options ...ErrorOption) (ErrorHealthCheckSource, error) {
	conf := defaultErrorSourceConfig(checkType, errorMode)
	conf.apply(options...)

	switch errorMode {
	case UnhealthyIfAtLeastOneError, HealthyIfNotAllErrors:
		return newErrorHealthCheckSource(conf)
	default:
		return nil, werror.Error("unknown or unsupported error mode", werror.SafeParam("errorMode", errorMode))
	}
}

// MustNewUnhealthyIfAtLeastOneErrorSource returns the result of calling NewUnhealthyIfAtLeastOneErrorSource, but panics if it returns an error.
// Should only be used in instances where the inputs are statically defined and known to be valid.
// DEPRECATED: please use MustNewErrorHealthCheckSource.
func MustNewUnhealthyIfAtLeastOneErrorSource(checkType health.CheckType, windowSize time.Duration) ErrorHealthCheckSource {
	return MustNewErrorHealthCheckSource(checkType, UnhealthyIfAtLeastOneError,
		WithWindowSize(windowSize))
}

// NewUnhealthyIfAtLeastOneErrorSource creates an unhealthyIfAtLeastOneErrorSource
// with a sliding window of size windowSize and uses the checkType.
// windowSize must be a positive value, otherwise returns error.
// NewUnhealthyIfAtLeastOneErrorSource creates an unhealthyIfAtLeastOneErrorSource
// with a sliding window of size windowSize and uses the checkType.
// windowSize must be a positive value, otherwise returns error.
// DEPRECATED: please use NewErrorHealthCheckSource.
func NewUnhealthyIfAtLeastOneErrorSource(checkType health.CheckType, windowSize time.Duration) (ErrorHealthCheckSource, error) {
	return NewErrorHealthCheckSource(checkType, UnhealthyIfAtLeastOneError,
		WithWindowSize(windowSize))
}

// MustNewHealthyIfNotAllErrorsSource returns the result of calling NewHealthyIfNotAllErrorsSource, but panics if it returns an error.
// Should only be used in instances where the inputs are statically defined and known to be valid.
// DEPRECATED: please use MustNewErrorHealthCheckSource.
func MustNewHealthyIfNotAllErrorsSource(checkType health.CheckType, windowSize time.Duration) ErrorHealthCheckSource {
	return MustNewErrorHealthCheckSource(checkType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRequireFullWindow())
}

// NewHealthyIfNotAllErrorsSource creates an healthyIfNotAllErrorsSource
// with a sliding window of size windowSize and uses the checkType.
// windowSize must be a positive value, otherwise returns error.
// Errors submitted in the first time window cause the health check to go to REPAIRING instead of ERROR.
// DEPRECATED: please use NewErrorHealthCheckSource.
func NewHealthyIfNotAllErrorsSource(checkType health.CheckType, windowSize time.Duration) (ErrorHealthCheckSource, error) {
	return NewErrorHealthCheckSource(checkType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRequireFullWindow())
}

// MustNewAnchoredHealthyIfNotAllErrorsSource returns the result of calling
// NewAnchoredHealthyIfNotAllErrorsSource but panics if that call returns an error
// Should only be used in instances where the inputs are statically defined and known to be valid.
// Care should be taken in considering health submission rate and window size when using anchored
// windows. Windows too close to service emission frequency may cause errors to not surface.
// DEPRECATED: please use MustNewErrorHealthCheckSource.
func MustNewAnchoredHealthyIfNotAllErrorsSource(checkType health.CheckType, windowSize time.Duration) ErrorHealthCheckSource {
	return MustNewErrorHealthCheckSource(checkType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRepairingGracePeriod(windowSize),
		WithRequireFullWindow())
}

// NewAnchoredHealthyIfNotAllErrorsSource creates an healthyIfNotAllErrorsSource
// with supplied checkType, using sliding window of size windowSize, which will
// anchor (force the window to be at least the grace period) by defining a repairing deadline
// at the end of the initial window or one window size after the end of a gap.
// If all errors happen before the repairing deadline, the health check returns REPAIRING instead of ERROR.
// windowSize must be a positive value, otherwise returns error.
// Care should be taken in considering health submission rate and window size when using anchored
// windows. Windows too close to service emission frequency may cause errors to not surface.
// DEPRECATED: please use NewErrorHealthCheckSource.
func NewAnchoredHealthyIfNotAllErrorsSource(checkType health.CheckType, windowSize time.Duration) (ErrorHealthCheckSource, error) {
	return NewErrorHealthCheckSource(checkType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRepairingGracePeriod(windowSize),
		WithRequireFullWindow())
}

// errorHealthCheckSource is a HealthCheckSource that polls a TimeWindowedStore.
// It returns, if there are only non-nil errors, the latest non-nil error as an unhealthy check.
// If there are no items, returns healthy.
type errorHealthCheckSource struct {
	errorMode ErrorMode
	timeProvider         TimeProvider
	windowSize           time.Duration
	lastErrorTime        time.Time
	lastError            error
	lastSuccessTime      time.Time
	sourceMutex          sync.RWMutex
	checkType            health.CheckType
	repairingGracePeriod time.Duration
	repairingDeadline    time.Time
}

func newErrorHealthCheckSource(conf errorSourceConfig) (ErrorHealthCheckSource, error) {
	if conf.windowSize <= 0 {
		return nil, werror.Error("windowSize must be positive",
			werror.SafeParam("windowSize", conf.windowSize.String()))
	}
	if conf.repairingGracePeriod < 0 {
		return nil, werror.Error("repairingGracePeriod must be non negative",
			werror.SafeParam("repairingGracePeriod", conf.repairingGracePeriod.String()))
	}

	source := &errorHealthCheckSource{
		errorMode: conf.errorMode,
		timeProvider:         conf.timeProvider,
		windowSize:           conf.windowSize,
		checkType:            conf.checkType,
		repairingGracePeriod: conf.repairingGracePeriod,
		repairingDeadline:    conf.timeProvider.Now(),
	}

	// If requireFirstFullWindow, extend the repairing deadline to one windowSize from now.
	if conf.requireFirstFullWindow {
		source.repairingDeadline = conf.timeProvider.Now().Add(conf.windowSize)
	}

	return source, nil
}

// Submit submits an error.
func (e *errorHealthCheckSource) Submit(err error) {
	e.sourceMutex.Lock()
	defer e.sourceMutex.Unlock()

	// If using anchored windows when last submit is greater than the window
	// it will re-anchor the next window with a new repairing deadline.
	if !e.hasSuccessInWindow() && !e.hasErrorInWindow() {
		newRepairingDeadline := e.timeProvider.Now().Add(e.repairingGracePeriod)
		if newRepairingDeadline.After(e.repairingDeadline) {
			e.repairingDeadline = newRepairingDeadline
		}
	}

	if err != nil {
		e.lastError = err
		e.lastErrorTime = e.timeProvider.Now()
	} else {
		e.lastSuccessTime = e.timeProvider.Now()
	}
}

// HealthStatus polls the items inside the window and creates the HealthStatus.
func (e *errorHealthCheckSource) HealthStatus(ctx context.Context) health.HealthStatus {
	e.sourceMutex.RLock()
	defer e.sourceMutex.RUnlock()

	var healthCheckResult health.HealthCheckResult
	if e.hasSuccessInWindow() && e.errorMode == HealthyIfNotAllErrors {
		healthCheckResult = sources.HealthyHealthCheckResult(e.checkType)
	} else if e.hasErrorInWindow() {
		if e.lastErrorTime.Before(e.repairingDeadline) {
			healthCheckResult = sources.RepairingHealthCheckResult(e.checkType, e.lastError.Error(), sources.SafeParamsFromError(e.lastError))
		} else {
			healthCheckResult = sources.UnhealthyHealthCheckResult(e.checkType, e.lastError.Error(), sources.SafeParamsFromError(e.lastError))
		}
	} else {
		healthCheckResult = sources.HealthyHealthCheckResult(e.checkType)
	}

	return health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			e.checkType: healthCheckResult,
		},
	}
}

func (e *errorHealthCheckSource) hasSuccessInWindow() bool {
	return !e.lastSuccessTime.IsZero() && e.timeProvider.Now().Sub(e.lastSuccessTime) <= e.windowSize
}

func (e *errorHealthCheckSource) hasErrorInWindow() bool {
	return !e.lastErrorTime.IsZero() && e.timeProvider.Now().Sub(e.lastErrorTime) <= e.windowSize
}
