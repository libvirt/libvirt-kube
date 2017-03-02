#!/bin/sh

kubectl create -f pv-docker-registry.yaml
kubectl create -f pvc-docker-registry.yaml
kubectl create -f rc-docker-registry.yaml
kubectl create -f ds-docker-registry-proxy.yaml
kubectl create -f svc-docker-registry.yaml
