// Copyright (c) 2021 Palantir Technologies. All rights reserved.
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

package heartbeat

import (
	"time"

	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
)

// Option is an option for a heartbeat based health check source.
type Option func(conf *heartbeatSourceConfig)

type heartbeatSourceConfig struct {
	checkType          health.CheckType
	startupGracePeriod time.Duration
}

func defaultHeartbeatSourceConfig(checkType health.CheckType) heartbeatSourceConfig {
	return heartbeatSourceConfig{
		checkType:          checkType,
		startupGracePeriod: 0,
	}
}

func (h *heartbeatSourceConfig) apply(options ...Option) {
	for _, option := range options {
		option(h)
	}
}

// WithStartupGracePeriod configures an initial grace period that allows the health check source to remain healthy
// despite no heartbeats within the configured heartbeat timeout. This grace period is only applied once on startup.
func WithStartupGracePeriod(period time.Duration) Option {
	return func(conf *heartbeatSourceConfig) {
		conf.startupGracePeriod = period
	}
}
