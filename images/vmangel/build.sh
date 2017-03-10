#!/bin/sh

set -e
set -v

cp $GOPATH/bin/virtkubevmangel .

docker build -t libvirtkubevmangel .
