#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT="$(cd "$(dirname "$0")" && pwd -P)"/..
cd "${REPO_ROOT}"

controller-gen paths=./... output:webhook:dir="./kubernetes-resources"
