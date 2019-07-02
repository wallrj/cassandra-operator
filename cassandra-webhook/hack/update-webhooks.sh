#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT="$(cd "$(dirname "$0")" && pwd -P)"/..
cd "${REPO_ROOT}"

webhook_path="${REPO_ROOT}/kubernetes-resources/cassandra-webhook.yml"
webhook_path="${REPO_ROOT}/kubernetes-resources/manifests.yaml"
output="$(mktemp -d)"

controller-gen paths=./... output:webhook:dir="${output}"
mv "${output}"/manifests.yaml "${webhook_path}"
