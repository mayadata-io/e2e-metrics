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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type ControllerType string

const (
	ControllerTypeSync ControllerType = "sync"

	ControllerTypeFinalize ControllerType = "finalize"
)

type ControllerStatus string

const (
	ControllerStatusPassed ControllerStatus = "passed"

	ControllerStatusFailed ControllerStatus = "failed"

	ControllerStatusSkipped ControllerStatus = "skipped"
)

const (
	ControllerMetricName string = "controller_sync_call_count"

	ControllerMetricHelp string = "The number of sync() calls received by the controller."
)

var (
	ControllerMetricLblNames = []string{"name", "type", "status"}
)

// Controller structure to populate metrics
//
// It exposes following metrics:
// 	controller_sync_call_count{"name", "type", "status"}
// where
// - name="coverage"
// - type="sync|finalize"
// - status="passed|failed|skipped"
type Controller struct {
	Name   string
	Type   ControllerType
	Status ControllerStatus
	Error  error
}

// IncrementControllerSyncCount increments controller sync metric
func (m *Metrics) IncrementControllerSyncCount(ctrl *Controller) {
	if ctrl.Status == "" && ctrl.Error != nil {
		ctrl.Status = ControllerStatusFailed
	}
	m.ControllerSyncCallCount.
		With(
			prometheus.Labels{
				"name":   ctrl.Name,
				"type":   string(ctrl.Type),
				"status": string(ctrl.Status),
			},
		).
		Inc()
}
