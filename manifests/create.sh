#!/bin/sh

for dir in tpr libvirt virtimagerepo virtimagefile virtnode
do
    echo
    echo $dir
    cd $dir
    ./create.sh
    cd ..
done
