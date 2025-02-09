#!/bin/bash

CLUSTER_NAME=${1:-remote}
echo "Build Docker image"
docker build -t workloadstealagent:alpha1 .

echo "Set the kubectl context to $CLUSTER_NAME cluster"
kubectl cluster-info --context kind-$CLUSTER_NAME
kubectl config use-context kind-$CLUSTER_NAME

echo "Load Image to Kind cluster named '$CLUSTER_NAME'"
kind load docker-image --name $CLUSTER_NAME workloadstealagent:alpha1

echo "Restarting Agent Deployment"
kubectl rollout restart deployment mutation-webhook-server -n hiro