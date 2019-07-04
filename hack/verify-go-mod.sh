#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# based off https://github.com/jetstack/cert-manager/blob/master/hack/verify-crds.sh

SCRIPT_DIR=$(dirname ${BASH_SOURCE})
REPO_ROOT="$(cd ${SCRIPT_DIR}/.. && pwd)"

tmpdir="$(mktemp -d)"

cleanup() {
    rm -rf "${tmpdir}"
}
trap "cleanup" EXIT SIGINT

rsync -avvL "${REPO_ROOT}"/ "${tmpdir}" >/dev/null
cd "${tmpdir}"

./hack/update-go-mod.sh

if diff -Naupr "${REPO_ROOT}" "${tmpdir}"; then
    echo "${REPO_ROOT} up to date."
else
    echo "${REPO_ROOT} is out of date. Please run ./hack/update-go-mod.sh"
    exit 1
fi
