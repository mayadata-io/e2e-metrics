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
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"openebs.io/metac/controller/generic"

	"mayadata.io/e2e-metrics/config"
	"mayadata.io/e2e-metrics/pkg/metac"
	"mayadata.io/e2e-metrics/types"
)

type errHandler struct {
	watch    *unstructured.Unstructured
	response *generic.SyncHookResponse
}

func (e *errHandler) handle(err error) {
	if err == nil {
		// do nothing
		return
	}
	glog.Errorf(
		"Failed to sync 'Namespace' %q: %+v", e.watch.GetName(), err,
	)
	// this will stop further reconciliation at metac
	e.response.SkipReconcile = true
	// set error to nil to avoid panic
	err = nil
}

// Sync implements the idempotent logic to reconcile Namespace
//
// NOTE:
// 	SyncHookRequest is the payload received as part of reconcile
// request. Similarly, SyncHookResponse is the payload sent as a
// response as part of reconcile request.
//
// NOTE:
//	SyncHookRequest uses Namespace as the watched resource.
// SyncHookResponse has PipelineCoverage that forms the desired
// state w.r.t this watched resource.
//
// NOTE:
//	Returning error will panic this process. We would rather want
// this controller to run continuously. Hence, the errors are handled.
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	if request == nil {
		// this will panic
		return errors.Errorf("Failed to sync 'Namespace': Nil request")
	}
	if request.Watch == nil || request.Watch.Object == nil {
		// this will panic
		return errors.Errorf("Failed to sync 'Namespace': Nil watch")
	}
	if response == nil {
		// this will panic
		return errors.Errorf("Failed to sync 'Namespace': Nil response")
	}

	podNS := os.Getenv("MY_POD_NAMESPACE")
	if request.Watch.GetName() != podNS {
		glog.V(4).Infof(
			"Will skip 'Namespace' %q: Want %q", request.Watch.GetName(), podNS,
		)
		response.SkipReconcile = true
		return nil
	}

	glog.V(3).Infof("Will sync 'Namespace' %q", request.Watch.GetName())

	// construct the error handler
	errHandler := &errHandler{
		watch:    request.Watch,
		response: response,
	}

	var err error
	defer errHandler.handle(err)

	var observedCoverage *unstructured.Unstructured
	for _, attachment := range request.Attachments.List() {
		if attachment.GetKind() == "PipelineCoverage" &&
			attachment.GetNamespace() == podNS {
			observedCoverage = attachment
		} else {
			// Add un required attachments to response.
			// Metac in turn ignores them
			response.Attachments = append(response.Attachments, attachment)
		}
	}

	reconciler := NewReconciler(observedCoverage)
	desired := reconciler.Reconcile()
	response.Attachments = append(response.Attachments, desired)

	glog.V(2).Infof(
		"Sync 'Namespace' %q was successful: %s",
		request.Watch.GetName(), metac.GetDetailsFromResponse(response),
	)

	return nil
}

// Percentage helps in formating a float value into
// percent notation
type Percentage float32

// String returns percentage notation of coverage
func (c Percentage) String() string {
	var pi int
	pi = int(math.Round(float64(c) * 100))
	return fmt.Sprintf("%d%%", pi)
}

// Reconciler enables reconciliation of Namespace
type Reconciler struct {
	ObservedPipelineCoverage *unstructured.Unstructured

	metrics *config.MetricsConfig

	// actual & valid test case names
	validTests []string

	// actual test case names that are not registered as desired
	invalidTests []string

	coverage float32
	warnings []string
	err      error
}

// NewReconciler returns a new instance of reconciler
func NewReconciler(pipelineCoverage *unstructured.Unstructured) *Reconciler {
	return &Reconciler{
		ObservedPipelineCoverage: pipelineCoverage,
	}
}

func (r *Reconciler) getPhase() string {
	if r.err != nil {
		return string(types.PipelineCoverageFailed)
	}
	return string(types.PipelineCoveragePassed)
}

func (r *Reconciler) getErrOrEmpty() string {
	if r.err != nil {
		return r.err.Error()
	}
	return ""
}

func (r *Reconciler) getWarnOrEmpty() string {
	if len(r.warnings) == 0 {
		return ""
	}
	wcount := fmt.Sprintf("%d warnings", len(r.warnings))
	return fmt.Sprintf("%s: %s", wcount, strings.Join(r.warnings, ": "))
}

func (r *Reconciler) getDeprecatedOrEmpty() string {
	if len(r.metrics.DeprecatedTestCases) == 0 {
		return ""
	}
	dcount := fmt.Sprintf(
		"%d deprecations",
		len(r.metrics.DeprecatedTestCases),
	)
	return fmt.Sprintf(
		"%s: %s",
		dcount,
		strings.Join(r.metrics.DeprecatedTestCases, ": "),
	)
}

// calculateCoverage has the real business logic of calculating
// test coverage percentage including setting warnings if any
func (r *Reconciler) calculateCoverage() {
	for tcid := range r.metrics.ActualTestCases {
		if r.metrics.DesiredTestCases[tcid] {
			// the gitlab-ci.yml test case(s) that are registered
			// in .master-plan.yml are valid
			r.validTests = append(r.validTests, tcid)
		} else {
			r.invalidTests = append(r.invalidTests, tcid)
		}
	}
	if len(r.invalidTests) > 0 {
		r.warnings = append(
			r.warnings,
			fmt.Sprintf("Invalid tests [%s]", strings.Join(r.invalidTests, ",")),
		)
	}

	validTestCount := len(r.validTests)
	desiredTestCount := len(r.metrics.DesiredTestCases)
	if desiredTestCount == 0 {
		r.warnings = append(r.warnings, fmt.Sprintf("Missing desired tests"))
		// return to avoid divide-by-0 error
		return
	}

	glog.V(2).Infof("Coverage calc = %d/%d*100", validTestCount, desiredTestCount)
	actual := float32(validTestCount)
	desired := float32(desiredTestCount)
	r.coverage = actual / desired
}

// loadConfigOrEmpty loads the config or empty if config
// is not found
func (r *Reconciler) loadConfigOrEmpty() {
	c := config.New("/etc/config/e2e-metrics/")
	r.metrics, r.err = c.LoadOrEmpty()
}

// Reconcile observed state of CStorClusterPlan to its desired
// state
func (r *Reconciler) Reconcile() *unstructured.Unstructured {
	var fns = []func(){
		r.loadConfigOrEmpty,
		r.calculateCoverage,
	}
	for _, fn := range fns {
		fn()
		if r.err != nil {
			// we log & stop executing remaining functions
			glog.Errorf("Reconcile failed: %+v", r.err)
			break
		}
	}
	return r.getDesiredPipelineCoverage()
}

// getDesiredPipelineCoverage returns the desired PipelineCoverage
// instance
//
// NOTE:
//	The returned instance is idempotent and hence can be used during
// create & update operations
func (r *Reconciler) getDesiredPipelineCoverage() *unstructured.Unstructured {
	coverage := &unstructured.Unstructured{}
	coverage.SetUnstructuredContent(map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      os.Getenv("E2E_METRICS_COVERAGE_NAME"),
			"namespace": os.Getenv("MY_POD_NAMESPACE"),
		},
		"spec": map[string]interface{}{
			"pipeline": map[string]interface{}{
				"id": os.Getenv("E2E_METRICS_PIPELINE_ID"),
			},
			"test": map[string]interface{}{
				"count": int64(len(r.metrics.DesiredTestCases)),
			},
		},
		// since metac does not sync the attachment's status
		// we are renaming status -> result
		//
		// NOTE:
		//	metac does not reconcile status since it can lead
		// to hot loop reconciliations. Once metac exposes a
		// new tunable or starts supporting reconcilining
		// attachment's status, we may rename result -> status.
		//
		// ref - https://github.com/AmitKumarDas/metac/issues/100
		"result": map[string]interface{}{
			"phase":            r.getPhase(),
			"reason":           r.getErrOrEmpty(),
			"warning":          r.getWarnOrEmpty(),
			"deprecated":       r.getDeprecatedOrEmpty(),
			"runid":            os.Getenv("E2E_METRICS_RUN_ID"),
			"validTestCount":   int64(len(r.validTests)),
			"invalidTestCount": int64(len(r.invalidTests)),
			"coverage":         Percentage(r.coverage).String(),
		},
	})
	// below is the right way to set APIVersion & Kind
	coverage.SetAPIVersion(string(types.E2EMetricsMayadataV1Alpha1))
	coverage.SetKind(string(types.KindPipelineCoverage))
	return coverage
}
