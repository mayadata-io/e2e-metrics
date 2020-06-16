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

const (
	ActualTestCountMetricName string = "actual_test_count"

	ActualTestCountMetricHelp string = "Total number of actual test cases."
)

// ActualTestCount structure to populate metrics
//
// It exposes following metrics:
// 	actual_test_count{"component", "feature", "kind", "testimpltype"}
// where
// - component="director|dao|openebs"
// - feature="dmaas|auth|teaming"
// - kind="backup|restore|googleauth"
// - testimpltype="litmus|dope"
type ActualTestCount struct {
	BaseTestCount
}

// SetActualTestCount sets the planned test count metric
func (m *Metrics) SetActualTestCount(atc *ActualTestCount) {
	m.ActualTestsTotal.
		With(
			prometheus.Labels{
				"component":    string(atc.Component),
				"feature":      string(atc.Feature),
				"kind":         string(atc.Kind),
				"testimpltype": string(atc.TestImplementationType),
			},
		).
		Set(atc.Value)
}
