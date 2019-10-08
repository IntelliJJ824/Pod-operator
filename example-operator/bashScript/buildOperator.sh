#!/bin/bash
printf "\n"
echo "Building the operator #########################"
operator-sdk build docker.io/zz35/pod-operator
docker push docker.io/zz35/pod-operator
echo "The operator is built #########################"
printf "\n"
