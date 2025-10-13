package metrics

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const (
	DEFAULT_PORT  = 9090
	DEFAULT_ROUTE = "/metrics"
	// Should be more than the scraping time and less that the global shutdown
	// timeout.
	SHUTDOWN_TIMEOUT_SEC = 5 * time.Second
)

// Empty prometheus server
var server *http.Server = nil

// StateChecker interface for accessing controller state
type StateChecker interface {
	IsProcessing() bool
	IsShutdownRequested() bool
}

// StartServerWithReadiness starts the metrics server with readiness and health endpoints
func StartServerWithReadiness(worker_id string, route string, port int, state StateChecker) {
	StartServer(worker_id, route, port)

	// Add readiness and health endpoints
	if server != nil {
		mux, ok := server.Handler.(*http.ServeMux)
		if !ok {
			logrus.Warnln("Could not add readiness endpoint: server handler is not *http.ServeMux")
			return
		}

		// Readiness endpoint - returns 503 if processing or shutting down
		mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
			// Not ready if shutdown requested
			if state.IsShutdownRequested() {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("shutting down"))
				return
			}

			// Not ready if processing (busy with a job)
			if state.IsProcessing() {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("processing job"))
				return
			}

			// Ready if idle
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ready"))
		})

		// Health endpoint - always returns 200 unless server is dead
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("healthy"))
		})

		logrus.Infof("Added readiness endpoint: /ready and health endpoint: /health on port %v", port)
	}
}

// The server will
func StartServer(worker_id string, route string, port int) {

	// If the endpoint is left empty, uses the standard /metrics endpoint
	if len(route) == 0 {
		route = "/metrics"
	}

	// If the route is not prefixed with a `/`, prefix it
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}

	// If the port is not specified, uses the standard 9090. We consider "0" to
	// be an unset value.
	if port <= 0 || port >= 1<<16 {
		port = DEFAULT_PORT
	}

	// If the port is specifi

	// If no worker_id is specified, then use a default string. This relevant
	// only for testing purposes since the application already sets a default
	// for the worker_id.
	if len(worker_id) == 0 {
		worker_id = "default-worker-id"
	}

	// Initialize the registry
	initRegistry(worker_id)

	// We don't use the default server and mux because we don't want to this
	// server to be the main server and we particularly don't want this route
	// to leak outside of this package. Also, we want to be able to shutdown the
	// server gracefully which the global server does not allow.
	server = &http.Server{
		Addr:              fmt.Sprintf(":%v", port),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		Handler: func() http.Handler {
			mux := http.NewServeMux()
			mux.Handle(route, promhttp.Handler())
			return mux
		}(),
	}

	// Starts the server
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logrus.Errorf("Metric server closed with error %v", err)
		}
	}()

	logrus.Infof("Started the metric server on port %v and route %v", port, route)
}

// Shutdown the metric server gracefully
func ShutdownServer(ctx context.Context) {
	// nothing to shutdown
	if server == nil {
		logrus.Infof("no metric server to shutdown")
		return
	}

	// Give it 2 sec to finish before forcing it down
	shutdownCtx, cancel := context.WithTimeout(ctx, SHUTDOWN_TIMEOUT_SEC)
	server.Shutdown(shutdownCtx)
	cancel()
	logrus.Infof("metric server successfully shutdown")
}
