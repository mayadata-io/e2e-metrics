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

type MasterPlanYmlLoadDurationStatus string

const (
	MasterPlanYmlLoadDurationStatusFailed MasterPlanYmlLoadDurationStatus = "failed"

	MasterPlanYmlLoadDurationStatusPassed MasterPlanYmlLoadDurationStatus = "passed"
)

const (
	MasterPlanYMLLoadDurationSecondsMetricName string = "masterplan_yml_load_duration_seconds"

	MasterPlanYMLLoadDurationSecondsMetricHelp string = "Time taken in seconds to load .masterplan.yml."
)

var (
	MasterPlanYMLLoadDurationSecondsMetricLblNames = []string{"status"}

	MasterPlanYMLLoadDurationSecondsMetricObjs = map[float64]float64{
		0.5:  0.05,
		0.9:  0.01,
		0.99: 0.001,
	}
)

// MasterPlanYmlLoadDuration structure to populate metrics
//
// It exposes following metrics:
//
// masterplan_yml_load_duration_seconds{"status"}
// - where status="loaded|failed"
type MasterPlanYmlLoadDuration struct {
	ValueInSeconds float64
	Status         MasterPlanYmlLoadDurationStatus
}

// ObserveMasterPlanYmlLoadDuration sets the master plan yaml's load duration
func (m *Metrics) ObserveMasterPlanYmlLoadDuration(load *MasterPlanYmlLoadDuration) {
	m.MasterPlanYMLLoadDurationSeconds.
		With(
			prometheus.Labels{
				"status": string(load.Status),
			},
		).
		Observe(load.ValueInSeconds)
}
