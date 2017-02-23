#!/bin/sh

# This is used by "virsh console" to create lock files.  Technically,
# "virsh console" uses /var/lock, but that is a symlink to /run/lock.
mkdir -p /run/lock

# Make sure permissions on /dev/kvm are correct.
if [ -c /dev/kvm ]; then
	chmod a+rw /dev/kvm
fi

CADIR=/etc/pki/CA/libvirt

cd /etc/pki
mkdir -p libvirt
cd libvirt

# The code here is totally insane from a security POV and
# will have to be thrown away. We need to get unique certs
# into each libvirtd pod. We cant pull them from k8s secret
# objects since the DaemonSet has no way to alter which
# secret is pulled per POD it spawns, and we don't know the
# IP addr that will be allocated to the POD either so could
# not pre-populate secrets even if we wanted to.
#
# A real solution will likely involve the bare metal node
# being bootstrapped with an identity it can use to talk
# to FreeIPA. Then, this entrypoint could use FreeIPA tools
# to request a cert for libvirtd and QEMU
#
# As a temporary hack for proof of concept demos, we store
# the CA cert + private key in a secret and just generate
# a cert now

rpm -qf /usr/bin/certtool

cat > server.info <<EOF
organization = Libvirt Kubelet
EOF

echo "cn = `hostname`" >> server.info
echo "dns_name = `hostname -f`" >> server.info
echo "dns_name = `hostname -s`" >> server.info
for ip in `hostname -I`
do
    echo "ip_address = $ip" >> server.info
done

cat server.info

cp server.info client.info

cat >> server.info <<EOF
tls_www_server
encryption_key
signing_key
EOF

cat >> client.info <<EOF
tls_www_client
encryption_key
signing_key
EOF

certtool --generate-privkey > clientkey.pem
certtool --generate-privkey > serverkey.pem

certtool --generate-certificate --load-privkey serverkey.pem \
  --load-ca-certificate $CADIR/cacert.pem --load-ca-privkey $CADIR/cakey.pem \
  --template server.info --outfile servercert.pem

certtool --generate-certificate --load-privkey clientkey.pem \
  --load-ca-certificate $CADIR/cacert.pem --load-ca-privkey $CADIR/cakey.pem \
  --template client.info --outfile clientcert.pem

ln -s $CADIR/cacert.pem cacert.pem

exec "$@"
