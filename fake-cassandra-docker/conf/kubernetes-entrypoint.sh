#!/bin/bash -e

# Initialize cassandra volume - needed when the volume is mounted via kubernetes volume
mkdir -p ${CASSANDRA_VOLUME}/commitlog ${CASSANDRA_VOLUME}/saved_caches ${CASSANDRA_VOLUME}/data
chown -R cassandra:cassandra ${CASSANDRA_VOLUME}

java -jar /usr/share/cassandra/operator/cassandra-configurer.jar
exec /sbin/setuser cassandra cassandra -f
