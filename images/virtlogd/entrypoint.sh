#!/bin/sh

set -v
set -e

bind() {
    mkdir -p $1
    mkdir -p $2
    mount --bind $1 $2
}

bind /srv/libvirt/log /var/log/libvirt
bind /srv/libvirt/run /run/libvirt

exec "$@"
