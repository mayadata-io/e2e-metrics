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

package config

import (
	"testing"

	"mayadata.io/e2e-metrics/metrics"
	logstesting "mayadata.io/e2e-metrics/pkg/logs/testing"
)

func TestConfigLoad(t *testing.T) {
	var expectDesiredTestNames = []string{
		"TCID-DIR-INSTALL-ON-LOCAL-PV",
		"TCID-DIR-HEALTH-CHECK-V2",
		"TCID-DIR-HEALTH-CHECK",
	}
	var expectActualTestNames = []string{
		"TCID-DIR-HEALTH-CHECK-V2",
		"TCID-DIR-HEALTH-CHECK",
	}
	var expectDeprecatedTestNames = []string{
		"tcid-dir-health-check-v2",
		"tcid-DIR-HEALTH-CHECK",
	}

	log := &logstesting.TestLogger{
		T: t,
	}
	prom := metrics.New(log)
	config := New(LoadableConfig{
		Path: "testdata/",
		Log:  log,
		Prom: prom,
	})
	metrics, err := config.Load()
	if err != nil {
		t.Fatalf("Expected no error: Got %v", err)
	}
	// Desired
	if len(metrics.DesiredTestCases) != len(expectDesiredTestNames) {
		t.Fatalf(
			"Expected desired test case count %d got %d: %+v",
			len(expectDesiredTestNames),
			len(metrics.DesiredTestCases),
			metrics.DesiredTestCases,
		)
	}
	for _, eDesiredTestName := range expectDesiredTestNames {
		if !metrics.DesiredTestCases[eDesiredTestName] {
			t.Fatalf("Expected desired test name %q got %#v",
				eDesiredTestName,
				metrics.DesiredTestCases,
			)
		}
	}
	// Actuals
	if len(metrics.ActualTestCases) != len(expectActualTestNames) {
		t.Fatalf(
			"Expected actual test case count %d got %d: %+v",
			len(expectActualTestNames),
			len(metrics.ActualTestCases),
			metrics.ActualTestCases,
		)
	}
	for _, eActualTestName := range expectActualTestNames {
		if !metrics.ActualTestCases[eActualTestName] {
			t.Fatalf("Expected actual test name %q got %#v",
				eActualTestName,
				metrics.ActualTestCases,
			)
		}
	}
	// Deprecated
	if len(metrics.DeprecatedTestCases) != len(expectDeprecatedTestNames) {
		t.Fatalf(
			"Expected deprecated test case count %d got %d: %+v",
			len(expectDeprecatedTestNames),
			len(metrics.DeprecatedTestCases),
			metrics.DeprecatedTestCases,
		)
	}
	for _, eDeprecatedTestName := range expectDeprecatedTestNames {
		var found bool
		for _, gotDeprecatedTestName := range metrics.DeprecatedTestCases {
			if eDeprecatedTestName == gotDeprecatedTestName {
				found = true
			}
		}
		if !found {
			t.Fatalf("Expected deprecated test name %q got %#v",
				eDeprecatedTestName,
				metrics.DeprecatedTestCases,
			)
		}
	}
}
