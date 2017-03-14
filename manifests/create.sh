#!/bin/sh

for dir in tpr docker-registry libvirt virtimagerepo virtimagefile virtnode
do
    echo
    echo $dir
    cd $dir
    ./create.sh
    cd ..
done
