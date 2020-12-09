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

package diag1log

import (
	"github.com/palantir/witchcraft-go-logging/conjure/witchcraft/api/logging"
	"github.com/palantir/witchcraft-go-logging/wlog"
)

type defaultLogger struct {
	logger wlog.Logger
}

func (l *defaultLogger) Diagnostic(diagnostic logging.Diagnostic, params ...Param) {
	l.logger.Log(toParams(diagnostic, params)...)
}

func toParams(diagnostic logging.Diagnostic, inParams []Param) []wlog.Param {
	outParams := make([]wlog.Param, len(defaultTypeParam)+1+len(inParams))
	copy(outParams, defaultTypeParam)
	outParams[len(defaultTypeParam)] = wlog.NewParam(diagnosticParam(diagnostic).apply)
	for idx := range inParams {
		outParams[len(defaultTypeParam)+1+idx] = wlog.NewParam(inParams[idx].apply)
	}
	return outParams
}

var defaultTypeParam = []wlog.Param{
	wlog.NewParam(func(entry wlog.LogEntry) {
		entry.StringValue(wlog.TypeKey, TypeValue)
	}),
}