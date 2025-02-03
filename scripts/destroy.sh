#!/bin/bash
echo "Delete and Create a 'kind' cluster with name 'local'"
kind delete cluster --name remote
kind create cluster --name remote