#!/usr/bin/env bash
shopt -s nullglob globstar
set -x # have bash print command been ran
set -e # fail if any command fails

setup_certs(){
  { # create CA.
    openssl \
      req \
      -new \
      -newkey rsa:4096 \
      -days 1024 \
      -nodes \
      -x509 \
      -subj "/C=US/ST=CA/O=MyOrg/CN=myOrgCA" \
      -keyout confs/rootCA.key \
      -out confs/rootCA.crt

  }

  { # create server certs.
    openssl \
      req \
      -new \
      -newkey rsa:2048 \
      -days 372 \
      -nodes \
      -keyout confs/server.key \
      -out confs/server.cst \
      -subj "/C=US/ST=CA/O=MyOrg/CN=myOrgCA" \
      -addext "subjectAltName=DNS:example.com,DNS:example.net,DNS:otel_collector,DNS:localhost" \

    # sign the cert
    openssl \
      x509 \
      -req \
      -in confs/server.cst -CA confs/rootCA.crt -CAkey confs/rootCA.key -CAcreateserial -out confs/server.crt
  }

  { # create client certs.
    openssl \
      req \
      -new \
      -newkey rsa:2048 \
      -days 372 \
      -nodes \
      -keyout confs/client.key \
      -out confs/client.cst \
      -subj "/C=US/ST=CA/O=MyOrg/CN=myOrgCA" \
      -addext "subjectAltName=DNS:example.com,DNS:example.net,DNS:otel_collector,DNS:localhost" \

    # sign the cert
    openssl \
      x509 \
      -req \
      -in confs/client.cst -CA confs/rootCA.crt -CAkey confs/rootCA.key -CAcreateserial -out confs/client.crt
  }

  { # clean
    #rm -rf confs/*.cst
    rm -rf confs/*.srl

    chmod 666 confs/server.crt confs/server.key confs/rootCA.crt
  }
}
setup_certs
