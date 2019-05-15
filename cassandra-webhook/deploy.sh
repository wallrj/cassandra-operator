#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

usage="Usage: CONTEXT=<ctx> IMAGE=<dockerImage> NAMESPACE=<namespace> $0"

: ${IMAGE?${usage}}
: ${CONTEXT?${usage}}
: ${NAMESPACE?${usage}}

scriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
scriptPath="${scriptDir}/$(basename ${BASH_SOURCE[0]})"
templatesDir="${scriptDir}/kubernetes-resources"
resourcesDir="${scriptPath}.files"
name="cassandra-webhook"


function create_certificates() {
    local fqdn="${name}.${NAMESPACE}.svc"
    pushd "${resourcesDir}"
    # See https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/
    cat <<EOF > "pki/cfssl.json"
{
  "hosts": [
      "$fqdn"
  ],
  "CN": "${fqdn}",
  "key": {
    "algo": "ecdsa",
    "size": 256
  }
}
EOF

    cfssl genkey "pki/cfssl.json" \
        | cfssljson -bare "pki/server"

    cat <<EOF > manifests/csr.yaml
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: ${fqdn}
spec:
  groups:
  - system:authenticated
  request: $(cat "pki/server.csr" | base64 | tr -d '\n')
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF

    kubectl apply -f manifests/csr.yaml

    kubectl certificate approve "${fqdn}"

    while ! test -s pki/server.crt; do
        kubectl get csr "${fqdn}" -o jsonpath='{.status.certificate}' \
            | base64 --decode > pki/server.crt
    done
}


function create_resources() {
    pushd "${templatesDir}"
    find . -type f -iname '*.yaml' | while read relPath; do
        envsubst < $relPath > "${resourcesDir}/manifests/${relPath}"
    done
    popd
}

function deploy() {
    kubectl apply -f ${resourcesDir}/manifests
}


if test -d $resourcesDir; then
    echo "ERROR: $resourcesDir already exists. Cleanup first." >&2
    exit 1
fi
mkdir -p "${resourcesDir}/manifests" "${resourcesDir}/pki"

create_certificates
create_resources
deploy
