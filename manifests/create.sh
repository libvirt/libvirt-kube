#!/bin/sh

for dir in tpr docker-registry virtimagerepo virtimagefile libvirt
do
    echo
    echo $dir
    cd $dir
    ./create.sh
    cd ..
done
