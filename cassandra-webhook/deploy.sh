#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

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

function deploy() {
    local image=$1
    local context=$2
    local namespace=$3
    local deployment=cassandra-webhook

    # kind load docker-image "${image}"
    kubectl delete namespace "${namespace}" || true
    until kubectl create namespace "${namespace}"; do sleep 1; done

    kubectl -n "${namespace}" create secret tls --key tls.key --cert tls.crt webhook-tls
    k8Resources="webhook-deployment.yaml manifests.yaml"
    for k8Resource in ${k8Resources}
    do
        sed -e "s@\$TARGET_NAMESPACE@$namespace@g" \
            -e "s@\$WEBHOOK_IMAGE@$image@g" \
            ${resourcesDir}/${k8Resource} > ${tmpDir}/${k8Resource}
        kubectl --context ${context} -n ${namespace} apply -f ${tmpDir}/${k8Resource}
    done

    waitForDeployment ${context} ${namespace} ${deployment}
}

usage="Usage: CONTEXT=<ctx> IMAGE=<dockerImage> NAMESPACE=<namespace> $0"
: ${IMAGE?${usage}}
: ${CONTEXT?${usage}}
: ${NAMESPACE?${usage}}

deploy ${IMAGE} ${CONTEXT} ${NAMESPACE}
