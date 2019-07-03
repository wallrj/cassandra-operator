#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR=$(dirname ${BASH_SOURCE})
ROOT_DIR="$(cd ${SCRIPT_DIR}/.. && pwd)"

find "${ROOT_DIR}" -type f -name go.mod | while read f; do
    gomodroot="$(dirname $f)"
    pushd "${gomodroot}" > /dev/null
    go mod tidy
    popd > /dev/null
done
