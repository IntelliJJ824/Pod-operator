#!/bin/bash
printf "\n"
echo "Deployments are createing ############################"
kubectl apply -f ~/go/src/github.com/pod-operator/tony-operator/deploy/crds/tony_v1alpha1_server_crd.yaml
kubectl apply -f ~/go/src/github.com/pod-operator/tony-operator/deploy/role.yaml
kubectl apply -f ~/go/src/github.com/pod-operator/tony-operator/deploy/service_account.yaml
kubectl apply -f ~/go/src/github.com/pod-operator/tony-operator/deploy/role_binding.yaml
echo "Deployments are complete #############################"
printf "\n"

