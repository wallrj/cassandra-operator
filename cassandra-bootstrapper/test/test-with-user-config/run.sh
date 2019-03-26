#!/bin/bash -e

testScriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd ${testScriptDir}

source ../common.sh

createDgossVolumes
copyCassandraConfiguration

# Add test file to user config map
docker run \
        --rm \
        -v ${USER_CONFIG_MAP}:/user-config-map \
        --entrypoint=touch \
        ${CASSANDRA_IMAGE} \
        /user-config-map/test-file

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
        -v ${EXTRA_LIB_EMPTY_DIR}:/extra-lib \
        -v ${USER_CONFIG_MAP}:/custom-config \
        ${IMAGE_TO_TEST}

runCommonChecks

GOSS_SLEEP=0 dgoss run \
    -v ${CONFIG_EMPTY_DIR}:/etc/cassandra \
    -v ${EXTRA_LIB_EMPTY_DIR}:/extra-lib \
    ${CASSANDRA_IMAGE}
