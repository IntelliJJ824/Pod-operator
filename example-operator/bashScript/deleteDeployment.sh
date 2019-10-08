#!/bin/bash
printf "\n"
echo "Deployments are deleting..............."
kubectl delete -f ~/go/src/github.com/pod-operator/tony-operator/deploy/operator.yaml
kubectl delete -f ~/go/src/github.com/pod-operator/tony-operator/deploy/role.yaml
kubectl delete -f ~/go/src/github.com/pod-operator/tony-operator/deploy/role_binding.yaml
kubectl delete -f ~/go/src/github.com/pod-operator/tony-operator/deploy/service_account.yaml
kubectl delete -f ~/go/src/github.com/pod-operator/tony-operator/deploy/crds/tony_v1alpha1_server_cr.yaml
kubectl delete -f ~/go/src/github.com/pod-operator/tony-operator/deploy/crds/tony_v1alpha1_server_crd.yaml
echo "Deployments are clear................."
printf "\n"
kubectl delete pods --all
echo "Pods are clear........................"
printf "\n"
