/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package metrics contains global structures related to metrics
// collection. Metrics that are exposed are described in respective
// files in this package.
package metrics

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// Namespace is the namespace for e2e metric names
	// This is not Kubernetes namespace but a Prometheus thing
	namespace                              = "e2emet"
	prometheusMetricsServerShutdownTimeout = 5 * time.Second
	prometheusMetricsServerReadTimeout     = 8 * time.Second
	prometheusMetricsServerWriteTimeout    = 8 * time.Second
	prometheusMetricsServerMaxHeaderBytes  = 1 << 20 // 1 MiB
)

// Metrics is a shared instance used to update metrics exposed
// in this project
type Metrics struct {
	log      logr.Logger
	registry *prometheus.Registry

	GitlabCIYMLLoadDurationSeconds   *prometheus.SummaryVec
	MasterPlanYMLLoadDurationSeconds *prometheus.SummaryVec

	PlannedTestsTotal *prometheus.GaugeVec
	ActualTestsTotal  *prometheus.GaugeVec

	ControllerSyncCallCount *prometheus.CounterVec
}

// New returns a new instance of Metrics
func New(log logr.Logger) *Metrics {
	var (
		GitlabCIYMLLoadDurationSeconds = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  namespace,
				Name:       GitlabCIYmlLoadDurationSecondsMetricName,
				Help:       GitlabCIYmlLoadDurationSecondsMetricHelp,
				Objectives: GitlabCIYmlLoadDurationSecondsMetricObjs,
			},
			GitlabCIYmlLoadDurationSecondsMetricLblNames,
		)

		MasterPlanYMLLoadDurationSeconds = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  namespace,
				Name:       MasterPlanYMLLoadDurationSecondsMetricName,
				Help:       MasterPlanYMLLoadDurationSecondsMetricHelp,
				Objectives: MasterPlanYMLLoadDurationSecondsMetricObjs,
			},
			MasterPlanYMLLoadDurationSecondsMetricLblNames,
		)

		PlannedTestCount = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      PlannedTestCountMetricName,
				Help:      PlannedTestCountMetricHelp,
			},
			TestCountMetricLblNames,
		)

		ActualTestCount = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      ActualTestCountMetricName,
				Help:      ActualTestCountMetricHelp,
			},
			TestCountMetricLblNames,
		)

		controllerSyncCallCount = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      ControllerMetricName,
				Help:      ControllerMetricHelp,
			},
			ControllerMetricLblNames,
		)
	)

	// Create server and register Prometheus metrics handler
	m := &Metrics{
		log:      log.WithName("metrics"),
		registry: prometheus.NewRegistry(),

		MasterPlanYMLLoadDurationSeconds: MasterPlanYMLLoadDurationSeconds,
		GitlabCIYMLLoadDurationSeconds:   GitlabCIYMLLoadDurationSeconds,
		ActualTestsTotal:                 ActualTestCount,
		PlannedTestsTotal:                PlannedTestCount,
		ControllerSyncCallCount:          controllerSyncCallCount,
	}

	return m
}

// Start will register the Prometheu metrics, and start the Prometheus server
func (m *Metrics) Start(listenAddress string) (*http.Server, error) {
	m.registry.MustRegister(
		m.ActualTestsTotal,
		m.PlannedTestsTotal,
		m.GitlabCIYMLLoadDurationSeconds,
		m.MasterPlanYMLLoadDurationSeconds,
		m.ControllerSyncCallCount,
	)

	router := mux.NewRouter()
	router.Handle(
		"/metrics/e2e",
		promhttp.HandlerFor(
			m.registry,
			promhttp.HandlerOpts{},
		),
	)

	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return nil, err
	}

	server := &http.Server{
		Addr:           ln.Addr().String(),
		ReadTimeout:    prometheusMetricsServerReadTimeout,
		WriteTimeout:   prometheusMetricsServerWriteTimeout,
		MaxHeaderBytes: prometheusMetricsServerMaxHeaderBytes,
		Handler:        router,
	}

	go func() {
		log := m.log.WithValues("address", ln.Addr())
		log.Info("Starting prometheus metrics server")

		if err := server.Serve(ln); err != nil {
			log.Error(err, "error running prometheus metrics server")
			return
		}
	}()

	return server, nil
}

// Shutdown prometheus metrics server
func (m *Metrics) Shutdown(server *http.Server) {
	m.log.Info("Stopping prometheus metrics server")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		prometheusMetricsServerShutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		m.log.Error(err, "prometheus metrics server shutdown failed", err)
		return
	}

	m.log.Info("prometheus metrics server gracefully stopped")
}
