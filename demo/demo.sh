#!/bin/bash

cleanup() {
  set +e
  
  echo ""
  echo "--------------------------"
  echo "++ Clean up started"
  echo "--------------------------"

  kubectl delete -f ../deploy/operator.yaml || true
  kubectl delete configmap metac-config-test -n e2e-metrics || true
  kubectl delete -f ../deploy/crd.yaml || true
  kubectl delete -f ../deploy/rbac.yaml || true
  kubectl delete -f ../deploy/namespace.yaml || true

  echo "--------------------------"
  echo "++ Clean up completed"
  echo "--------------------------"
  echo ""
}
#trap cleanup EXIT

# Uncomment this if you want to run this script in debug mode
#set -ex

echo -e "\n++ Installing e2e-metrics operator"

kubectl apply -f ../deploy/namespace.yaml
kubectl apply -f ../deploy/rbac.yaml
kubectl apply -f ../deploy/crd.yaml
kubectl create configmap metac-config-test -n e2e-metrics --from-file="../deploy/metac-config.yaml"
kubectl apply -f ../deploy/operator.yaml

echo -e "\n++ Installed e2e-metrics operator successfully"