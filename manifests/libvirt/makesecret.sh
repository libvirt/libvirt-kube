#!/bin/sh

certtool --generate-privkey > cakey.pem

cat >ca.info <<EOF
cn = libvirt kubelet demo
ca
cert_signing_key
EOF

certtool --generate-self-signed --load-privkey cakey.pem \
  --template ca.info --outfile cacert.pem

CAKEY=`base64 -w 0 cakey.pem`
CACERT=`base64 -w 0 cacert.pem`

cp sec-virtdx509ca.yaml.in sec-virtdx509ca.yaml
echo "  cakey.pem: \"$CAKEY\"" >> sec-virtdx509ca.yaml
echo "  cacert.pem: \"$CACERT\"" >> sec-virtdx509ca.yaml
