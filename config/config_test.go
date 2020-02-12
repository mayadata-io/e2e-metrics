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
)

func TestConfigLoad(t *testing.T) {
	var expectDesiredTestNames = []string{}
	var expectActualTestNames = []string{
		"miot1x", "blabb",
	}

	config := New("testdata/")
	metrics, err := config.Load()
	if err != nil {
		t.Fatalf("Expected no error: Got %v", err)
	}
	// Desired
	if len(metrics.DesiredTestCases) != len(expectDesiredTestNames) {
		t.Fatalf(
			"Expected desired test case count %d got %d",
			len(expectDesiredTestNames), len(metrics.DesiredTestCases),
		)
	}
	for _, eDesiredTestName := range expectDesiredTestNames {
		if !metrics.DesiredTestCases[eDesiredTestName] {
			t.Fatalf("Expected desired test name %q got %+v",
				eDesiredTestName, metrics.DesiredTestCases,
			)
		}
	}
	// Actuals
	if len(metrics.ActualTestCases) != len(expectActualTestNames) {
		t.Fatalf(
			"Expected actual test case count %d got %d",
			len(expectActualTestNames), len(metrics.ActualTestCases),
		)
	}
	for _, eActualTestName := range expectActualTestNames {
		if !metrics.ActualTestCases[eActualTestName] {
			t.Fatalf("Expected actual test name %q got %#v",
				eActualTestName, metrics.ActualTestCases,
			)
		}
	}
}
