# Cassandra Snapshot

An application which executes `nodetool snapshot` operations on Cassandra pods running within a Kubernetes cluster.
Packaged as a Docker container, it is intended to be used from Kubernetes [CronJobs](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
in order to automate creation and cleanup of Cassandra snapshots.

It assumes that it is running under a service account with `exec` access to pods in the required namespaces.

The Cassandra pods on which to manipulate snapshots are selected by a comma-separated list of labels which identify
those pods. Typically, the set of labels supplied would identify all of the pods of a single Cassandra cluster, and no
other pods.

You can find information on how to manage snapshots on the [WIKI](https://github.com/sky-uk/cassandra-operator/wiki).
