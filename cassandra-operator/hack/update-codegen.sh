#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

PROJECT_ROOT="github.com/sky-uk/cassandra-operator/cassandra-operator"
API_PACKAGE="cassandra:v1alpha1"
SCRIPT_DIR=$(dirname ${BASH_SOURCE})
VENDOR_DIR=${SCRIPT_DIR}/../vendor/k8s.io/code-generator

${VENDOR_DIR}/generate-groups.sh all \
  ${PROJECT_ROOT}/pkg/client ${PROJECT_ROOT}/pkg/apis \
  ${API_PACKAGE} \
  --go-header-file ${SCRIPT_DIR}/empty-boilerplate.txt
