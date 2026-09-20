package insights

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func serve(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}

func probesHandler(t *testing.T, funcs ...IsReadyFunc) http.Handler {
	t.Helper()
	impl, ok := NewProbes(funcs...).(*probes)
	require.True(t, ok)
	return impl.server.Handler
}

func ready() bool    { return true }
func notReady() bool { return false }
func crashing() bool { panic("database handle is nil") }

func TestProbes_LiveIsAlwaysOK(t *testing.T) {
	handler := probesHandler(t, notReady, crashing)

	assert.Equal(t, http.StatusOK, serve(t, handler, "/live").Code, "liveness ignores readiness checks")
}

func TestProbes_ReadyRequiresEveryCheck(t *testing.T) {
	cases := []struct {
		name     string
		funcs    []IsReadyFunc
		expected int
	}{
		{"no checks", nil, http.StatusOK},
		{"all ready", []IsReadyFunc{ready, ready}, http.StatusOK},
		{"one not ready", []IsReadyFunc{ready, notReady, ready}, http.StatusServiceUnavailable},
		{"a crashing check counts as not ready", []IsReadyFunc{ready, crashing}, http.StatusServiceUnavailable},
		{"crash is recovered even when first", []IsReadyFunc{crashing, ready}, http.StatusServiceUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, serve(t, probesHandler(t, tc.funcs...), "/ready").Code)
		})
	}
}

func TestProbes_ReadyShortCircuitsAfterFirstFailure(t *testing.T) {
	calls := 0
	counting := func() bool {
		calls++
		return true
	}

	serve(t, probesHandler(t, notReady, counting), "/ready")
	assert.Zero(t, calls, "once unready, remaining checks are skipped")

	serve(t, probesHandler(t, counting, counting), "/ready")
	assert.Equal(t, 2, calls)
}

func TestProbes_UnknownPathIsNotFound(t *testing.T) {
	assert.Equal(t, http.StatusNotFound, serve(t, probesHandler(t), "/metrics").Code)
}

func TestProbes_ListenOnConfiguredPort(t *testing.T) {
	previous := viper.Get(constants.ProbePort)
	t.Cleanup(func() { viper.Set(constants.ProbePort, previous) })
	viper.Set(constants.ProbePort, 9191)

	impl, ok := NewProbes().(*probes)
	require.True(t, ok)
	assert.Equal(t, ":9191", impl.server.Addr)
}

func TestProbes_ShutdownClosesServerForGood(t *testing.T) {
	impl, ok := NewProbes(ready).(*probes)
	require.True(t, ok)

	impl.Shutdown()

	assert.ErrorIs(t, impl.server.ListenAndServe(), http.ErrServerClosed)
	assert.NotPanics(t, (&probes{}).Shutdown, "a probe without server has nothing to close")
}

func TestPrometheusMetrics_ExposesMetricsEndpoint(t *testing.T) {
	previous := viper.Get(constants.MetricPort)
	t.Cleanup(func() { viper.Set(constants.MetricPort, previous) })
	viper.Set(constants.MetricPort, 2113)

	impl, ok := NewPrometheusMetrics().(*prom)
	require.True(t, ok)
	assert.Equal(t, ":2113", impl.server.Addr)

	recorder := serve(t, impl.server.Handler, "/metrics")
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "go_goroutines", "default Go collectors are exported")

	assert.Equal(t, http.StatusNotFound, serve(t, impl.server.Handler, "/live").Code)
}

func TestPrometheusMetrics_ShutdownClosesServerForGood(t *testing.T) {
	impl, ok := NewPrometheusMetrics().(*prom)
	require.True(t, ok)

	impl.Shutdown()

	assert.ErrorIs(t, impl.server.ListenAndServe(), http.ErrServerClosed)
	assert.NotPanics(t, (&prom{}).Shutdown, "metrics without server have nothing to close")
}
