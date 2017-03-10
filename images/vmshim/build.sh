#!/bin/sh

set -e
set -v

cp $GOPATH/bin/virtkubevmshim .

docker build -t libvirtkubevmshim .
