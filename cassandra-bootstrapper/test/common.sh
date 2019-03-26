#!/bin/bash -e

CASSANDRA_IMAGE=cassandra:3.11

CONFIG_EMPTY_DIR=fake-empty-dir-configuration
EXTRA_LIB_EMPTY_DIR=fake-empty-dir-extra-lib
USER_CONFIG_MAP=fake-user-config-map

function createDgossVolumes {
    echo == Creating empty-dir volumes
    docker volume create ${CONFIG_EMPTY_DIR}
    docker volume create ${EXTRA_LIB_EMPTY_DIR}
    docker volume create ${USER_CONFIG_MAP}
}

function deleteDgossVolumes {
    echo == Cleaning up empty-dir volumes
    docker volume rm ${CONFIG_EMPTY_DIR} ${EXTRA_LIB_EMPTY_DIR} ${USER_CONFIG_MAP}
}

function copyCassandraConfiguration {
    echo == Copying cassandra config from cassandra image
    docker run --rm \
        -v ${CONFIG_EMPTY_DIR}:/configuration \
        --entrypoint=sh \
        ${CASSANDRA_IMAGE} \
        -c "cp -r /etc/cassandra/* /configuration"
}

function runCommonChecks {
    local script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
    GOSS_FILES_PATH=${script_dir}/common-checks dgoss run \
        -v ${CONFIG_EMPTY_DIR}:/etc/cassandra \
        -v ${EXTRA_LIB_EMPTY_DIR}:/extra-lib \
        ${CASSANDRA_IMAGE}
}

trap deleteDgossVolumes EXIT SIGTERM SIGKILL
