#!/usr/bin/env bash

# based off https://github.com/jetstack/cert-manager/blob/master/hack/verify-crds.sh

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT="$(cd "$(dirname "$0")" && pwd -P)"/..

output="$(mktemp -d)"

cleanup() {
  rm -rf "${output}"
}
trap "cleanup" EXIT SIGINT

tmp="${output}/cassandra-operator"


rsync -avvL "${REPO_ROOT}"/ "${tmp}" >/dev/null
cd "${tmp}"
"./hack/update-crds.sh"

echo "diffing against freshly generated crd (${tmp})"
ret=0
diff -Naupr "${REPO_ROOT}/kubernetes-resources/cassandra-operator-crd.yml" "${tmp}/kubernetes-resources/cassandra-operator-crd.yml" || ret=$?
if [[ $ret -eq 0 ]]
then
  echo "${REPO_ROOT} up to date."
else
  echo "${REPO_ROOT} is out of date. Please run ./hack/update-crds.sh"
  exit 1
fi
