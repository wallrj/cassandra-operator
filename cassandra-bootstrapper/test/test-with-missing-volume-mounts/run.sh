#!/bin/bash

testScriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd ${testScriptDir}

source ../common.sh

createDgossVolumes
copyCassandraConfiguration

docker run \
        --rm \
        -e CLUSTER_USE_DEFAULT_SEED_PROVIDER=true \
        -e CLUSTER_DATA_CENTER=dc \
        -e CLUSTER_CURRENT_RACK=rack1 \
        -e CLUSTER_NAME=mycluster \
        -e CLUSTER_NAMESPACE=mycluster \
        -e NODE_LISTEN_ADDRESS=localhost \
        -e POD_CPU_MILLICORES=12 \
        -e POD_MEMORY_BYTES=2147483648 \
        ${IMAGE_TO_TEST}

if [[ $? == 0 ]] ; then
    echo Expected bootstrapper to fail due to missing /configuration volume, but it succeeded
    exit 1
fi

docker run \
        --rm \
        -e CLUSTER_USE_DEFAULT_SEED_PROVIDER=true \
        -e CLUSTER_DATA_CENTER=dc \
        -e CLUSTER_CURRENT_RACK=rack1 \
        -e CLUSTER_NAME=mycluster \
        -e CLUSTER_NAMESPACE=mycluster \
        -e NODE_LISTEN_ADDRESS=localhost \
        -e POD_CPU_MILLICORES=12 \
        -e POD_MEMORY_BYTES=2147483648 \
        -v ${CONFIG_EMPTY_DIR}:/configuration \
        ${IMAGE_TO_TEST}

if [[ $? == 0 ]] ; then
    echo Expected bootstrapper to fail due to missing /extra-lib volume, but it succeeded
    exit 1
fi
