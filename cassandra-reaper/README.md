# Cassandra Reaper

Cassandra Reaper is deployed using the `SIDECAR` deployment method.

## Architecture

Cassandra Reaper is deployed in every C* node pod,
as a [Sidecar container](https://kubernetes.io/docs/tasks/access-application-cluster/communicate-containers-same-pod-shared-volume/),
alongside the main Cassandra container.

The Cassandra Reaper side car process will interact with Cassandra JMX port via a loop back IP address,
so there is no need to configure remote JMX access Cassandra.
This is possible, because all containers in a pod share the same network namespace.

In `SIDECAR` mode, the Cassandra Reaper *must* be configured to use [Cassandra Storage Backend](http://cassandra-reaper.io/docs/backends/cassandra/).
The Cassandra Operator automatically configures Cassandra Reaper to use the local Cassandra cluster.
And it will automatically create a `Keyspace` called `reaper_db` in the local Cassandra cluster when it first starts up.

## Building

`SIDECAR` mode is an unreleased feature of Cassandra Reaper, so you need to compile Cassandra Reaper from the
[PR 622: Implement the sidecar mode](https://github.com/thelastpickle/cassandra-reaper/pull/622) feature branch
and compile as follows:

```
mvn package
mvn -B -pl src/server/ docker:build -Ddocker.directory=src/server/src/main/docker
```

Then re-tag the generated image, so that it can be pushed to your cluster. E.g.

```
gcloud docker -a
docker tag cassandra-reaper:latest gcr.io/jetstack-richard/cassandra-reaper:latest
docker push gcr.io/jetstack-richard/cassandra-reaper:latest
```

## Debugging

* View reaper logs
  `kubectl -n test-cassandra-operator logs mycluster-a-0 reaper`

* View cassandra logs
  `kubectl -n test-cassandra-operator logs mycluster-a-0 cassandra`
