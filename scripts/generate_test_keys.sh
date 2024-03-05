#!/bin/sh
#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e


echo "Generating test PKI for VC services"

cd /opt/workspace/wallet-sdk
mkdir -p test/integration/fixtures/keys/tls
tmp=$(mktemp)
echo "subjectKeyIdentifier=hash
authorityKeyIdentifier = keyid,issuer
extendedKeyUsage = serverAuth
keyUsage = Digital Signature, Key Encipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = localhost
DNS.2 = testnet.orb.local
DNS.4 = file-server.trustbloc.local
DNS.5 = mock-login-consent.example.com
DNS.6 = api-gateway.trustbloc.local
DNS.7 = mock-trust-registry.example.com
" >> "$tmp"

#create CA
openssl ecparam -name prime256v1 -genkey -noout -out test/integration/fixtures/keys/tls/ec-cakey.pem
openssl req -new -x509 -key test/integration/fixtures/keys/tls/ec-cakey.pem -subj "/C=CA/ST=ON/O=Example Internet CA Inc.:CA Sec/OU=CA Sec" -out test/integration/fixtures/keys/tls/ec-cacert.pem

#create TLS creds
openssl ecparam -name prime256v1 -genkey -noout -out test/integration/fixtures/keys/tls/ec-key.pem
openssl req -new -key test/integration/fixtures/keys/tls/ec-key.pem -subj "/C=CA/ST=ON/O=Example Inc.:vcs/OU=vcs/CN=localhost" -out test/integration/fixtures/keys/tls/ec-key.csr
openssl x509 -req -in test/integration/fixtures/keys/tls/ec-key.csr -CA test/integration/fixtures/keys/tls/ec-cacert.pem -CAkey test/integration/fixtures/keys/tls/ec-cakey.pem -CAcreateserial -extfile "$tmp" -out test/integration/fixtures/keys/tls/ec-pubCert.pem -days 365

#create master key for kms secret lock
openssl rand 32 | base64 | sed 's/+/-/g; s/\//_/g' > test/integration/fixtures/keys/tls/secret-lock.key

echo "done generating PKI"
