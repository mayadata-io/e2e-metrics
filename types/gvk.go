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

package types

const (
	// MayadataGroup represents mayadata org as group
	MayadataGroup string = "mayadata.io"

	// E2EMetricsMayadataGroup represent e2e-metrics within mayadata org
	E2EMetricsMayadataGroup string = "e2e-metrics" + "." + MayadataGroup

	// E2EMetricsMayadataV1Alpha1 represents v1alpha1 api version for e2e-metrics
	E2EMetricsMayadataV1Alpha1 string = E2EMetricsMayadataGroup + "/" + "v1alpha1"
)

const (
	// KindPipelineCoverage represent custom resource of kind
	// PipelineCoverage
	KindPipelineCoverage string = "PipelineCoverage"
)
