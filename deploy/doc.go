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

// Package deploy provides the required resources that are required
// to run e2e metrics controller.
//
// Setup KIND (with local docker registry listening on 5000 port) by
// running following:
// sudo ./testing/kind-with-registry.sh
// kubectl cluster-info --context kind-kind
//
// Run following commands from root of this project:
// # change REGISTRY to match yours
// REGISTRY=amitnist make push
//
// Run the following kubectl commands in the Kubernetes setup in the
// following order to **test** this controller
//
// kubectl apply -f crd.yaml
// kubectl apply -f namespace.yaml
// kubectl apply -f rbac.yaml
// kubectl create configmap metac-config-test -n e2e-metrics --from-file=metac-config.yaml
// kubectl create configmap metrics-config-test -n e2e-metrics --from-file=testing/.master-plan.yml --from-file=testing/.gitlab-ci.yml
// kubectl apply -f testing/test-operator.yaml
// kubectl apply -f service.yaml
//
// # Setup prometheus & grafana
// kubectl apply -f testing/rbac.yaml
// kubectl apply -f testing/prometheus-config.yaml
// kubectl apply -f testing/prom.yaml
// kubectl apply -f testing/grafana.yaml
//
// # Temporarily forward Grafana to localhost
// kubectl port-forward deployments/grafana 8080:3000
// kubectl port-forward deployments/prometheus 8181:9090
//
// Now go to http://localhost:8080 in your browser
// login using admin:admin.
// Navigate to Configuration > Data Sources > Add data source,
// choose Prometheus as type and enter http://prometheus:9090 as URL.
// Hit Save & Test which should yield a big green bar.
package deploy
