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

type ComponentUnderTest string

const (
	ComponentUnderTestDAO ComponentUnderTest = "dao"

	ComponentUnderTestDirector ComponentUnderTest = "director"

	ComponentUnderTestOpenEBS ComponentUnderTest = "openebs"
)

type FeatureUnderTest string

const (
	FeatureUnderTestDMAAS FeatureUnderTest = "dmaas"

	FeatureUnderTestAuth FeatureUnderTest = "auth"

	FeatureUnderTestTeaming FeatureUnderTest = "teaming"
)

type KindUnderTest string

const (
	KindUnderTestBackup KindUnderTest = "backup"

	KindUnderTestRestore KindUnderTest = "restore"

	KindUnderTestGoogleAuth KindUnderTest = "googleauth"

	KindUnderTestLocalAuth KindUnderTest = "localauth"
)

type TestImplementationType string

const (
	TestImplementationTypeLitmus TestImplementationType = "litmus"

	TestImplementationTypeDope TestImplementationType = "dope"
)

var (
	TestCountMetricLblNames = []string{"component", "feature", "kind", "testimpltype"}
)

// BaseTestCount structure to populate metrics
//
// It exposes following metrics:
// 	planned_test_count{"component", "feature", "kind", "testimpl"}
// 	actual_test_count{"component", "feature", "kind", "testimpl"}
// where
// - component="director|dao|openebs"
// - feature="dmaas|auth|teaming"
// - kind="backup|restore|googleauth|localauth"
// - testimpl="litmus|dope"
type BaseTestCount struct {
	Value                  float64
	Component              ComponentUnderTest
	Feature                FeatureUnderTest
	Kind                   KindUnderTest
	TestImplementationType TestImplementationType
}
