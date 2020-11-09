// This file was generated by Conjure and should not be manually edited.

package logging

import (
	"regexp"
	"strings"

	"github.com/palantir/conjure-go-runtime/v2/conjure-go-contract/errors"
	werror "github.com/palantir/witchcraft-go-error"
	wparams "github.com/palantir/witchcraft-go-params"
)

var enumValuePattern = regexp.MustCompile("^[A-Z][A-Z0-9]*(_[A-Z0-9]+)*$")

type AuditResult string

const (
	AuditResultSuccess      AuditResult = "SUCCESS"
	AuditResultUnauthorized AuditResult = "UNAUTHORIZED"
	AuditResultError        AuditResult = "ERROR"
)

func (e *AuditResult) UnmarshalText(data []byte) error {
	switch v := strings.ToUpper(string(data)); v {
	default:
		if !enumValuePattern.MatchString(v) {
			return werror.Convert(errors.NewInvalidArgument(wparams.NewSafeAndUnsafeParamStorer(map[string]interface{}{"enumType": "AuditResult", "message": "enum value must match pattern ^[A-Z][A-Z0-9]*(_[A-Z0-9]+)*$"}, map[string]interface{}{"enumValue": string(data)})))
		}
		*e = AuditResult(v)
	case "SUCCESS":
		*e = AuditResultSuccess
	case "UNAUTHORIZED":
		*e = AuditResultUnauthorized
	case "ERROR":
		*e = AuditResultError
	}
	return nil
}

type LogLevel string

const (
	LogLevelFatal LogLevel = "FATAL"
	LogLevelError LogLevel = "ERROR"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelTrace LogLevel = "TRACE"
)

func (e *LogLevel) UnmarshalText(data []byte) error {
	switch v := strings.ToUpper(string(data)); v {
	default:
		if !enumValuePattern.MatchString(v) {
			return werror.Convert(errors.NewInvalidArgument(wparams.NewSafeAndUnsafeParamStorer(map[string]interface{}{"enumType": "LogLevel", "message": "enum value must match pattern ^[A-Z][A-Z0-9]*(_[A-Z0-9]+)*$"}, map[string]interface{}{"enumValue": string(data)})))
		}
		*e = LogLevel(v)
	case "FATAL":
		*e = LogLevelFatal
	case "ERROR":
		*e = LogLevelError
	case "WARN":
		*e = LogLevelWarn
	case "INFO":
		*e = LogLevelInfo
	case "DEBUG":
		*e = LogLevelDebug
	case "TRACE":
		*e = LogLevelTrace
	}
	return nil
}
