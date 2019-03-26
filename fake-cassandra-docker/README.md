# Fake Cassandra Docker image

## About

A Docker image which "fakes" the properties of the [real cassandra-docker image](../cassandra-docker/README.md)
sufficiently to pass the Goss tests and be usable in place of a real Cassandra image within operator managed clusters.

The motivation for this image is to allow certain of the end-to-end tests which don't actually test the properties of
the Cassandra cluster to run more quickly than using a real Cassandra image allows. We have found that a Cassandra
pod can take up to a minute to become ready when running end-to-end tests. If we're simply checking that the operator
has successfully adjusted the number of pods in a cluster, or some other property of the cluster topology, there's no
need for the containers running in the pods to actually be functioning Cassandra containers. It's enough for them
just to do enough to look like they are Cassandra containers.
