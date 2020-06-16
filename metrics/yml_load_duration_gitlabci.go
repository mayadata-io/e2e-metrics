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

type GitlabCIYmlLoadDurationStatus string

const (
	GitlabCIYmlLoadDurationStatusFailed GitlabCIYmlLoadDurationStatus = "failed"

	GitlabCIYmlLoadDurationStatusPassed GitlabCIYmlLoadDurationStatus = "passed"
)

const (
	GitlabCIYmlLoadDurationSecondsMetricName string = "gitlab_ci_yml_load_duration_seconds"

	GitlabCIYmlLoadDurationSecondsMetricHelp string = "Time taken in seconds to load .gitlab-ci.yml."
)

var (
	GitlabCIYmlLoadDurationSecondsMetricLblNames = []string{"status"}

	GitlabCIYmlLoadDurationSecondsMetricObjs = map[float64]float64{
		0.5:  0.05,
		0.9:  0.01,
		0.99: 0.001,
	}
)

// GitlabCIYmlLoadDuration structure to populate metrics
//
// It exposes following metrics:
//
// gitlab_ci_yml_load_duration_seconds{"status"}
// - where status="loaded|failed"
type GitlabCIYmlLoadDuration struct {
	ValueInSeconds float64
	Status         GitlabCIYmlLoadDurationStatus
}

// ObserveGitlabCIYmlLoadDuration sets the gitlab ci yaml's load duration
func (m *Metrics) ObserveGitlabCIYmlLoadDuration(load *GitlabCIYmlLoadDuration) {
	m.MasterPlanYMLLoadDurationSeconds.
		With(
			prometheus.Labels{
				"status": string(load.Status),
			},
		).
		Observe(load.ValueInSeconds)
}
