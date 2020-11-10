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

package wlog

import (
	"fmt"
	"strings"
)

type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

func (l *LogLevel) UnmarshalText(b []byte) error {
	switch strings.ToLower(string(b)) {
	case string(DebugLevel):
		*l = DebugLevel
		return nil
	case "", string(InfoLevel):
		*l = InfoLevel
		return nil
	case string(WarnLevel):
		*l = WarnLevel
		return nil
	case string(ErrorLevel):
		*l = ErrorLevel
		return nil
	case string(FatalLevel):
		*l = FatalLevel
		return nil
	default:
		return fmt.Errorf("invalid log level: %q", string(b))
	}
}
