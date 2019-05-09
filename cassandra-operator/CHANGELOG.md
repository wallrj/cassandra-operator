# Changes

## 0.70.0-alpha
- [ANNOUNCEMENT] This is the first release available to the general public

  The operator is alpha-status software and can be used in development environments.
  It is not yet recommended for use in production environments.

- [FEATURE] New properties to specify docker images for snapshot and bootstrapping

  Adds 2 new properties to the cassandra cluster definition:
  - `pod.bootstrapperImage`: docker image used to bootstrap cassandra pods - see [cassandra-bootstrapper](../cassandra-bootstrapper/README.md)
  - `snapshot.image`: docker image used to trigger snapshot creation and cleanup - see [cassandra-snapshot](../cassandra-snapshot/README.md)

  These properties default to the latest version of each image.
  They are exposed for transparency and flexibility and mostly intended to be used during upgrades
  and to allow testing the operator against custom docker images version
  while working on features or improvements.

## 0.67.0
- [FEATURE] [Scheduled backups of cluster data](https://github.com/sky-uk/cassandra-operator/issues/20)

  This change introduces snapshot creation and cleanup by means of cronjobs,
  and so requires the Operator to have additional permissions to list, create, modify and delete cronjobs.
  For more details refer to [cassandra-snapshot](../cassandra-snapshot/README.md)

## 0.66.0
- [BUGFIX] [Wait indefinitely for rack changes](https://github.com/sky-uk/cassandra-operator/issues/19)

  Rather than waiting a fixed time for rack changes to be applied before proceeding to apply changes to subsequent racks,
  the operator will now wait indefinitely for a change to complete. This ensures that for scaling-up operations on
  heavily-loaded clusters, new nodes will not be added until all previously-added nodes have completed initalisation.

  This change requires the operator's service account to have the `patch` permission on the `events` resource.

## 0.65.0
- [FEATURE] [Make it possible to upgrade configurer without destroying clusters](https://github.com/sky-uk/cassandra-operator/issues/18)

  **Note: This version is not backwards-compatible with previous Cassandra clusters deployed by the operator.**

  This version introduces a major rework in the way that configuration is injected into Cassandra pods. Where previously
  we provided custom-built Cassandra Docker images which baked in code for configuration injection, this is now done by
  a pair of init-containers, which is a more scalable and future-proof means of doing this.

  The upshot of this is that the [official Cassandra Docker images](https://hub.docker.com/_/cassandra/) version 3.x can
  be used instead.

## 0.56.0
- [BUGFIX] [Scaling up a rack causes a full rolling update of the underlying StatefulSet](https://github.com/sky-uk/cassandra-operator/issues/17)

## 0.54.0
- [FEATURE] [Allow configurable readiness and liveness check parameters](https://github.com/sky-uk/cassandra-operator/issues/16)

  Note: This change is backwards incompatible with previous configurations which used the `PodLivenessProbeTimeoutSeconds` or `PodReadinessProbeTimeoutSeconds` pod settings.
  These have been removed and replaced with `PodLivenessProbe` and `PodReadinessProbe` which are nested objects which allow the following properties to be set:
  - `FailureThreshold`
  - `InitialDelaySeconds`
  - `PeriodSeconds`
  - `SuccessThreshold`
  - `TimeoutSeconds`

  These properties mirror those in Kubernetes' [Probe V1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.10/#probe-v1-core) and are all optional values.

## 0.48.0
- [FEATURE] [Split CRD bootstrapping out of operator](https://github.com/sky-uk/cassandra-operator/issues/14)

## 0.47.0
- [BUGFIX] [Operator internal state left inconsistent if cluster definition is invalid](https://github.com/sky-uk/cassandra-operator/issues/13)

## 0.42.0
- [BUGFIX] [Changing pod memory updates requests.memory but not limits.memory](https://github.com/sky-uk/cassandra-operator/issues/12)

## 0.41.0
- [BUGFIX] [Remove cluster check functionality from operator](https://github.com/sky-uk/cassandra-operator/issues/21)

## 0.37.0
- [FEATURE] [Apply updates on custom configuration changes](https://github.com/sky-uk/cassandra-operator/issues/5)

## 0.36.0
- [BUGFIX] [Liveness and readiness probes timeout after 1s](https://github.com/sky-uk/cassandra-operator/issues/10)
- [BUGFIX] [Metrics gathering should not attempt to access Jolokia on a pod which doesn't have an IP address](https://github.com/sky-uk/cassandra-operator/22)

## 0.33.0
- [FEATURE] [Apply updates on configuration changes](https://github.com/sky-uk/cassandra-operator/issues/8)

  The following updates will be applied by the operator:
  - Changes to PodMemory and PodCPU values
  - Increase in the number of racks
  - Increase in the number of replicas in a rack

## 0.29.0
- [FEATURE] Introduce `Cassandra` custom resource and change the operator from listening to changes on specially-labelled configmaps
to listening to changes on a dedicated custom resource, `core.sky.uk/cassandras`, currently set to version `v1alpha1`.

## 0.28.0
- [FEATURE] Allow the Cassandra docker image used in a cluster to be specified in the cluster ConfigMap.

## 0.24.0
- [BUGFIX] [Fetch metrics from a random cluster node each time](https://github.com/sky-uk/cassandra-operator/issues/3)
Fetch node metrics from a random node in the cluster rather than requesting via the service name,
so we avoid reporting inaccurate status when the target node is unavailable.

## 0.23.0
- [BUGFIX] [Remove empty racks](https://github.com/sky-uk/cassandra-operator/issues/2)
So we avoid creating unnecessary StatefulSet and invalid seed nodes.

## 0.21.0
TODO: Find out why this feature was added. If I remove it, will tests be prohibitively slow?

- [FEATURE] Added `--allow-empty-dir` flag to store Cassandra data on emptyDir volumes      (intended for test use only)

## 0.18.0
- [BUGFIX] [Metrics collection does not work for multi-clusters](https://github.com/sky-uk/cassandra-operator/issues/1)
Fix to support multiple clusters

## 0.13.0
- Pre-alpha release that creates a cluster based on a ConfigMap with a limited set of properties
to configure the number of racks, name and size of the cluster.
