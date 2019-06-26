#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT="$(cd "$(dirname "$0")" && pwd -P)"/..
cd "${REPO_ROOT}"

code_path="${REPO_ROOT}/pkg/apis/cassandra/v1alpha1/zz_generated.deepcopy.go"

output="$(mktemp -d)"
controller-gen object:headerFile=./hack/empty-boilerplate.txt  paths=./pkg/apis/... output:object:dir="${output}"
cp "${output}"/zz_generated.deepcopy.go "${code_path}"
