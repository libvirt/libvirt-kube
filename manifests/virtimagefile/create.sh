#!/bin/sh

for file in *.yaml
do
    kubectl create -f $file
done
