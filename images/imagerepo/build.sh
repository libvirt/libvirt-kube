#!/bin/sh

set -e
set -v

cp $GOPATH/bin/virtkubeimagerepo .

docker build -t libvirtkubeimagerepo .
