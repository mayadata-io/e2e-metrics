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

package coverage

import (
	"math"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"mayadata.io/e2e-metrics/config"
	"mayadata.io/e2e-metrics/types"
	"openebs.io/metac/controller/generic"
)

func TestSync(t *testing.T) {
	var tests = map[string]struct {
		request               *generic.SyncHookRequest
		response              *generic.SyncHookResponse
		expectAttachmentCount int
		isSkipReconcile       bool
		isErr                 bool
	}{
		"nil request": {
			request:  nil,
			response: &generic.SyncHookResponse{},
			isErr:    true,
		},
		"nil response": {
			request:  &generic.SyncHookRequest{},
			response: nil,
			isErr:    true,
		},
		"nil watch": {
			request: &generic.SyncHookRequest{
				Watch: nil,
			},
			response: &generic.SyncHookResponse{},
			isErr:    true,
		},
		"nil watch object": {
			request: &generic.SyncHookRequest{
				Watch: &unstructured.Unstructured{},
			},
			response: &generic.SyncHookResponse{},
			isErr:    true,
		},
		"namespace != pod namespace": {
			request: &generic.SyncHookRequest{
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"name": "Hi",
						},
					},
				},
			},
			response:              &generic.SyncHookResponse{},
			expectAttachmentCount: 0,
			isSkipReconcile:       true,
			isErr:                 false,
		},
		"namespace == pod namespace": {
			request: &generic.SyncHookRequest{
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
			},
			response:              &generic.SyncHookResponse{},
			expectAttachmentCount: 1,
			isSkipReconcile:       false,
			isErr:                 false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			err := Sync(mock.request, mock.response)
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got [%+v]", err)
			}
			if mock.response != nil && mock.isSkipReconcile != mock.response.SkipReconcile {
				t.Fatalf(
					"Expected skip reconcile %t got %t",
					mock.isSkipReconcile, mock.response.SkipReconcile,
				)
			}
			if mock.response != nil && mock.expectAttachmentCount != len(mock.response.Attachments) {
				t.Fatalf(
					"Expected response attachments %d got %d",
					mock.expectAttachmentCount, len(mock.response.Attachments),
				)
			}
		})
	}
}

func TestPercentageString(t *testing.T) {
	var tests = map[string]struct {
		value  float32
		expect string
	}{
		"0 to 0%": {
			value:  0,
			expect: "0%",
		},
		".1 to 10%": {
			value:  .1,
			expect: "10%",
		},
		".5 to 50%": {
			value:  .5,
			expect: "50%",
		},
		".54 to 54%": {
			value:  .54,
			expect: "54%",
		},
		".55 to 55%": {
			value:  .55,
			expect: "55%",
		},
		".56 to 56%": {
			value:  .56,
			expect: "56%",
		},
		".59 to 59%": {
			value:  .59,
			expect: "59%",
		},
		".594 to 59%": {
			value:  .594,
			expect: "59%",
		},
		".595 to 60%": {
			value:  .595,
			expect: "60%",
		},
		".596 to 60%": {
			value:  .596,
			expect: "60%",
		},
		".599 to 60%": {
			value:  .599,
			expect: "60%",
		},
		".5955 to 60%": {
			value:  .5955,
			expect: "60%",
		},
		".5956 to 60%": {
			value:  .5956,
			expect: "60%",
		},
		".5999 to 60%": {
			value:  .5999,
			expect: "60%",
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			p := Percentage(mock.value)
			if mock.expect != p.String() {
				t.Fatalf("Expected %s got %s", mock.expect, p.String())
			}
		})
	}
}

func TestReconcilerCalculateCoverage(t *testing.T) {
	var tests = map[string]struct {
		reconciler  *Reconciler
		expect      float64
		expectWarns int
	}{
		"0/0 coverage": {
			reconciler: &Reconciler{
				metrics: &config.MetricsConfig{},
			},
			expect:      0,
			expectWarns: 1,
		},
		"1/2 coverage": {
			reconciler: &Reconciler{
				metrics: &config.MetricsConfig{
					ActualTestCases: map[string]bool{
						"101": true,
					},
					DesiredTestCases: map[string]bool{
						"101": true,
						"201": true,
					},
				},
			},
			expect: .5,
		},
		"1/3 coverage": {
			reconciler: &Reconciler{
				metrics: &config.MetricsConfig{
					ActualTestCases: map[string]bool{
						"101": true,
					},
					DesiredTestCases: map[string]bool{
						"101": true,
						"201": true,
						"301": true,
					},
				},
			},
			expect: .33,
		},
		"2/2 coverage": {
			reconciler: &Reconciler{
				metrics: &config.MetricsConfig{
					ActualTestCases: map[string]bool{
						"101": true,
						"201": true,
					},
					DesiredTestCases: map[string]bool{
						"101": true,
						"201": true,
					},
				},
			},
			expect: 1,
		},
		"1/2 coverage - actuals != desired": {
			reconciler: &Reconciler{
				metrics: &config.MetricsConfig{
					ActualTestCases: map[string]bool{
						"101": true,
						"301": true, // not registered in desired
					},
					DesiredTestCases: map[string]bool{
						"101": true,
						"201": true,
					},
				},
			},
			expect:      .5,
			expectWarns: 1,
		},
		"1/3 coverage - actuals != desired": {
			reconciler: &Reconciler{
				metrics: &config.MetricsConfig{
					ActualTestCases: map[string]bool{
						"101": true,
						"401": true, // not registered in desired
						"501": true, // not registered in desired
					},
					DesiredTestCases: map[string]bool{
						"101": true,
						"201": true,
						"301": true,
					},
				},
			},
			expect:      .33,
			expectWarns: 1,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			mock.reconciler.calculateCoverage()
			got := float64(mock.reconciler.coverage)
			if math.Round(got) != math.Round(mock.expect) {
				t.Fatalf("Expected %f got %f", mock.expect, got)
			}
			if mock.expectWarns != len(mock.reconciler.warnings) {
				t.Fatalf("Expected warns %d got %d",
					mock.expectWarns, len(mock.reconciler.warnings),
				)
			}
		})
	}
}

func TestReconcilerReconcile(t *testing.T) {
	var tests = map[string]struct {
		reconciler *Reconciler
		expect     *unstructured.Unstructured
	}{
		"empty metrics": {
			reconciler: &Reconciler{
				metrics: &config.MetricsConfig{},
			},
			expect: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": string(types.E2EMetricsMayadataV1Alpha1),
					"kind":       string(types.KindPipelineCoverage),
					"metadata": map[string]interface{}{
						"name":      "",
						"namespace": "",
					},
					"spec": map[string]interface{}{
						"pipeline": map[string]interface{}{
							"id": "",
						},
						"test": map[string]interface{}{
							"count": int64(0),
						},
					},
					"result": map[string]interface{}{
						"phase":            "Failed",
						"reason":           "open /etc/config/e2e-metrics/: no such file or directory",
						"warning":          "",
						"deprecated":       "",
						"runid":            "",
						"validTestCount":   int64(0),
						"invalidTestCount": int64(0),
						"coverage":         "0%",
					},
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			got := mock.reconciler.Reconcile()
			if !reflect.DeepEqual(got, mock.expect) {
				t.Fatalf("Expected no diff got\n%s", cmp.Diff(mock.expect, got))
			}
		})
	}
}

func TestReconcilerGetDesiredPipelineCoverage(t *testing.T) {
	var tests = map[string]struct {
		reconciler *Reconciler
		expect     *unstructured.Unstructured
	}{
		"empty metrics": {
			reconciler: &Reconciler{
				metrics: &config.MetricsConfig{},
			},
			expect: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": string(types.E2EMetricsMayadataV1Alpha1),
					"kind":       string(types.KindPipelineCoverage),
					"metadata": map[string]interface{}{
						"name":      "",
						"namespace": "",
					},
					"spec": map[string]interface{}{
						"pipeline": map[string]interface{}{
							"id": "",
						},
						"test": map[string]interface{}{
							"count": int64(0),
						},
					},
					"result": map[string]interface{}{
						"phase":            "Passed",
						"reason":           "",
						"warning":          "",
						"deprecated":       "",
						"runid":            "",
						"validTestCount":   int64(0),
						"invalidTestCount": int64(0),
						"coverage":         "0%",
					},
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			got := mock.reconciler.getDesiredPipelineCoverage()
			if !reflect.DeepEqual(got, mock.expect) {
				t.Fatalf("Expected no diff got\n%s", cmp.Diff(mock.expect, got))
			}
		})
	}
}
