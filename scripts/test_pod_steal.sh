#!/bin/bash

echo "Deploy a sample pod to the remote cluster"
kubectl config use-context kind-remote
kubectl apply -f ../examples/pod_hiroworksteal_ns.yaml