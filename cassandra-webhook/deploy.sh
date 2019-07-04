#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

scriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
resourcesDir="${scriptDir}/kubernetes-resources"

function waitForDeployment {
    local count=0
    local sleepBetweenRetries=2
    local maxRetry=150 # 5mins max, as corresponds to: maxRetry * sleepBetweenRetries
    local context=$1
    local namespace=$2
    local deployment=$3

    local desiredReplicas=1
    local updatedReplicas=""
    local readyReplicas=""
    until ([[ "$desiredReplicas" = "$updatedReplicas" ]] && [[ "$desiredReplicas" = "$readyReplicas" ]]) || (( "$count" >= "$maxRetry" )); do
        count=$((count+1))
        echo "Waiting for ${namespace}.${deployment} to have ${desiredReplicas} updated replicas. Attempt: $count"
        readyReplicas=$(kubectl --context ${context} -n ${namespace} get deployment ${deployment} -o go-template="{{.status.readyReplicas}}")
        updatedReplicas=$(kubectl --context ${context} -n ${namespace} get deployment ${deployment} -o go-template="{{.status.updatedReplicas}}")

        sleep ${sleepBetweenRetries}
    done

    if [[ "$desiredReplicas" != "$updatedReplicas" ]] || [[ "$desiredReplicas" != "$readyReplicas" ]]; then
        echo "Deployment failed to become ready after ${maxRetry} retries"
        exit 1
    fi
    echo "Deployment is ready"
}

tmpDir=$(mktemp -d)
trap '{ CODE=$?; rm -rf ${tmpDir} ; exit ${CODE}; }' EXIT


function create_certificates() {
    local fqdn="${1}"
    mkdir "${tmpDir}/pki"
    # See https://github.com/cloudflare/cfssl/wiki/Creating-a-new-CSR#creating-a-self-signed-cert
    cat <<EOF > "${tmpDir}/pki/cfssl.json"
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

    cfssl selfsign "${fqdn}" "${tmpDir}/pki/cfssl.json" \
        | cfssljson -bare "${tmpDir}/pki/server"

    kubectl create secret tls \
            --dry-run \
            --output yaml \
            --namespace test-cassandra-operator \
            --cert "${tmpDir}/pki/server.pem" \
            --key="${tmpDir}/pki/server-key.pem" \
            cassandra-operator-webhook-tls > "${tmpDir}/secret.yaml"
}

function deploy() {
    local image=$1
    local context=$2
    local namespace=$3
    local deployment=cassandra-webhook

    create_certificates "webhook-service.${namespace}.svc"

    TARGET_NAMESPACE=$namespace \
        CA_BUNDLE="$(cat ${tmpDir}/pki/server.pem)" \
        go run ./hack/munge-webhook.go ./kubernetes-resources/cassandra-webhook-configuration.yaml > $tmpDir/cassandra-webhook-configuration.yaml

    k8Resources="cassandra-webhook-deployment.yaml"
    for k8Resource in ${k8Resources}
    do
        sed -e "s@\$TARGET_NAMESPACE@$namespace@g" \
            -e "s@\$WEBHOOK_IMAGE@$image@g" \
            ${resourcesDir}/${k8Resource} > ${tmpDir}/${k8Resource}
    done

    kubectl --context ${context} -n ${namespace} apply -f ${tmpDir}/

    waitForDeployment ${context} ${namespace} ${deployment}
}

usage="Usage: CONTEXT=<ctx> IMAGE=<dockerImage> NAMESPACE=<namespace> $0"
: ${IMAGE?${usage}}
: ${CONTEXT?${usage}}
: ${NAMESPACE?${usage}}

deploy ${IMAGE} ${CONTEXT} ${NAMESPACE}
