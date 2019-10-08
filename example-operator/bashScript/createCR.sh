#!/bin/bash
printf "\n"
echo "CR is created ***************"
vim /root/go/src/github.com/pod-operator/tony-operator/deploy/crds/tony_v1alpha1_server_cr.yaml
kubectl apply -f /root/go/src/github.com/pod-operator/tony-operator/deploy/crds/tony_v1alpha1_server_cr.yaml
echo "CR is generated *************"
printf "\n"
