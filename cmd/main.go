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

package main

import (
	"context"
	"flag"
	"os"
	"sync"

	"openebs.io/metac/controller/generic"
	"openebs.io/metac/start"

	"mayadata.io/e2e-metrics/controller/coverage"
	"mayadata.io/e2e-metrics/metrics"
	ctx "mayadata.io/e2e-metrics/pkg/context"
	logf "mayadata.io/e2e-metrics/pkg/logs"
	"mayadata.io/e2e-metrics/pkg/signal"
)

var (
	metricsAddr = flag.String(
		"e2e-metrics-addr",
		":9898",
		"The address to bind the http endpoint to be scraped by prometheus",
	)
)

// main function is the entry point of this binary.
//
// This registers various controller (i.e. kubernetes reconciler)
// handler functions. Each handler function gets triggered due
// to any changes (add, update or delete) to configured watch
// resource.
//
// NOTE:
// 	These functions will also be triggered in case this binary
// gets deployed or redeployed (due to restarts, etc.).
//
// NOTE:
//	One can consider each registered function as an independent
// kubernetes controller & this project as the operator.
func main() {
	logf.InitLogs()
	defer logf.FlushLogs()

	stopCh := signal.SetupSignalHandler()
	rootCtx := ctx.ContextWithStopCh(context.Background(), stopCh)
	rootCtx = logf.NewContext(rootCtx, nil, "operator")
	log := logf.FromContext(rootCtx)

	m := metrics.New(log)
	mserver, err := m.Start(*metricsAddr)
	if err != nil {
		log.Error(
			err,
			"failed to listen on prometheus address",
			"address",
			metricsAddr,
		)
		os.Exit(1)
	}

	syncer := coverage.NewSyncer(log, m)
	generic.AddToInlineRegistry("sync/pipelinecoverage", syncer.Sync)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		start.Start()
	}()
	wg.Wait()

	log.Info("e2e metrics operator loops exited")
	m.Shutdown(mserver)
	os.Exit(0)
}
