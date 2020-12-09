package transform

import (
	"context"
	"testing"

	werror "github.com/palantir/witchcraft-go-error"
	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/sources/store"
	"github.com/stretchr/testify/assert"
)

func TestSource(t *testing.T) {
	expected := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			"a": {},
		},
	}
	mapper := func(in health.HealthStatus) health.HealthStatus {
		return expected
	}
	keyed := store.NewKeyedErrorHealthCheckSource("foo", "bar")
	keyed.Submit("foo", werror.Error("err"))
	source, err := NewSource(mapper, keyed)
	assert.NoError(t, err)
	status := source.HealthStatus(context.Background())
	assert.Equal(t, expected, status)
}

func TestSourceNilChecks(t *testing.T) {
	mapper := func(in health.HealthStatus) health.HealthStatus {
		return health.HealthStatus{}
	}
	keyed := store.NewKeyedErrorHealthCheckSource("foo", "bar")
	_, err := NewSource(nil, keyed)
	assert.Error(t, err)
	_, err = NewSource(mapper, nil)
	assert.Error(t, err)
}
