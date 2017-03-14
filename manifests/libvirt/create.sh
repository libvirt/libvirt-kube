#!/bin/sh

if ! test -f cacert.pem
then
    ./makesecret.sh
fi
kubectl create -f ns-libvirt-kube.yaml
kubectl create -f sec-virtdx509ca.yaml
kubectl create -f ds-libvirt.yaml
kubectl create -f pod-virtkubevmshim.yaml
