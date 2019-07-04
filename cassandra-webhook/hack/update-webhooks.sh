#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT="$(cd "$(dirname "$0")" && pwd -P)"/..
cd "${REPO_ROOT}"

kubernetes_resources="${REPO_ROOT}/kubernetes-resources"
webhook_path="${kubernetes_resources}/cassandra-webhook-configuration.yaml"

controller-gen paths=./... "output:webhook:dir=${kubernetes_resources}"
mv "${kubernetes_resources}/manifests.yaml" "${webhook_path}"
