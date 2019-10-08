#!/bin/bash
printf "\n"
echo "Operattor is creating...."
kubectl apply -f /root/go/src/github.com/pod-operator/tony-operator/deploy/operator.yaml
echo "Operator is ready........"
printf "\n"

