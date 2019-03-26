# Cassandra Snapshot

An application which executes `nodetool snapshot` operations on Cassandra pods running within a Kubernetes cluster.
Packaged as a Docker container, it is intended to be used from Kubernetes [CronJobs](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
in order to automate creation and cleanup of Cassandra snapshots.

## Usage

`cassandra-snapshot` allows the creation and cleanup of Cassandra snapshots on pods running within a Kubernetes cluster.
It assumes that it is running within the same cluster under a service account with `exec` access to pods in the required
namespaces.

The Cassandra pods on which to manipulate snapshots are selected by a comma-separated list of labels which identify
those pods. Typically, the set of labels supplied would identify all of the pods of a single Cassandra cluster, and no
other pods.

### Creating snapshots

When the `cassandra-snapshot create` command is run, snapshots will be created for the specified keyspaces on all
selected Cassandra pods. The name of the each snapshot is the time of the command invocation, measured in seconds since
the epoch.

Run ...

`cassandra-snapshot create <flags>`

... where the available flags are:

Flag | Meaning | Required? | Default
---|---|---|---
`-l`, `--pod-label`        | Comma-separated list of labels which identifies the pods to create snapshots on | Yes
`-n`, `--namespace`        | Kubernetes namespace which the pods belong to                                   | Yes 
`-k`, `--keyspace`         | Comma-separated list of Cassandra keyspaces to snapshot                         | No  | All keyspaces
`-t`, `--snapshot-timeout` | Maximum time to wait for snapshot operation to complete                         | No  | `10s`
`-L`, `--log-level`        | Log level, can be one of `debug`, `info`, `warn`, `error`, `fatal`, `panic`     | No  | `info`

### Cleaning up snapshots

Snapshots are cleaned up based upon a retention period, with all snapshots which fall outside that retention period
being removed. Only snapshots named in line with the "seconds since epoch" naming convention will be correctly cleaned
up.

Run ...

`cassandra-snapshot cleanup <flags>`

... where the available flags are:

Flag | Meaning | Required? | Default
---|---|---|---
`-l`, `--pod-label`        | Comma-separated list of labels which identifies the pods to clean up snapshots on | Yes
`-n`, `--namespace`        | Kubernetes namespace which the pods belong to                                     | Yes 
`-k`, `--keyspace`         | Comma-separated list of Cassandra keyspaces to cleanup snapshots for              | No  | All keyspaces
`-r`, `--retention-period` | Snapshot retention period                                                         | No  | `7d`
`-t`, `--cleanup-timeout`  | Maximum time to wait for cleanup operation to complete                            | No  | `10s`
`-L`, `--log-level`        | Log level, can be one of `debug`, `info`, `warn`, `error`, `fatal`, `panic`       | No  | `info`

### Running with kubectl

It's also possible to trigger snapshot creation and cleanup within a Kubernetes cluster via the `kubectl` tool. Run
the following command to take a snapshot of keyspace `my_keyspace` on all nodes of cluster `my_cluster` in Kubernetes
namespace `my_namespace`.

The assumption is that all Cassandra pods in the cluster are identified by the labels `sky.uk/cassandra-operator=my_cluster,app=my_cluster`. 

```bash
kubectl run my-snapshot-taker  \
  -n my_namespace \
  --serviceaccount=cassandra-snapshot \
  --rm \
  --attach \
  --image=skyuk/cassandra-operator:cassandra-snapshot-latest \
  -- \
  create -n my_namespace -l sky.uk/cassandra-operator=my_cluster,app=my_cluster -k my_keyspace
```
