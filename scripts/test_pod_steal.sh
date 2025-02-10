#!/bin/bash

CLUSTER_NAME=${1:-remote}
POD_FILE=${2:-pod1}
echo "Deploy a sample pod to the $CLUSTER_NAME cluster"
kubectl config use-context kind-$CLUSTER_NAME
kubectl apply -f examples/$POD_FILE.yaml