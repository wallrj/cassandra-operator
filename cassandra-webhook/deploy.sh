#!/bin/bash -e

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

function deploy() {
    local image=$1
    local context=$2
    local namespace=$3
    local ingressHost=$4
    local deployment=cassandra-webhook
    local tmpDir=$(mktemp -d)
    trap '{ CODE=$?; rm -rf ${tmpDir} ; exit ${CODE}; }' EXIT

    k8Resources="apiservice.yaml rbac.yaml service.yaml deployment.yaml"
    for k8Resource in ${k8Resources}
    do
        sed -e "s@\$TARGET_NAMESPACE@$namespace@g" \
            -e "s@\$IMAGE@$image@g" \
            -e "s@\$ARGS@$args@g" \
            -e "s@\$INGRESS_HOST@$ingressHost@g" \
            ${resourcesDir}/${k8Resource} > ${tmpDir}/${k8Resource}
        kubectl --context ${context} apply -f ${tmpDir}/${k8Resource}
    done

    waitForDeployment ${context} ${namespace} ${deployment}
}

usage="Usage: CONTEXT=<ctx> IMAGE=<dockerImage> NAMESPACE=<namespace> INGRESS_HOST=<ingressHost> $0"
: ${IMAGE?${usage}}
: ${CONTEXT?${usage}}
: ${NAMESPACE?${usage}}
: ${INGRESS_HOST?${usage}}

deploy ${IMAGE} ${CONTEXT} ${NAMESPACE} ${INGRESS_HOST}
