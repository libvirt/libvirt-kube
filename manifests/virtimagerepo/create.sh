#!/bin/sh

kubectl create -f host-localhost.yaml
kubectl create -f shared-images.yaml
kubectl create -f rc-shared-images.yaml
