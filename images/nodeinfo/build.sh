#!/bin/sh

set -e
set -v

cp $GOPATH/bin/virtkubenodeinfo .

docker build -t libvirtkubenodeinfo .
