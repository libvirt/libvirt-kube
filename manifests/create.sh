#!/bin/sh

for dir in tpr docker-registry libvirt virtimagerepo virtimagefile
do
    echo
    echo $dir
    cd $dir
    ./create.sh
    cd ..
done
