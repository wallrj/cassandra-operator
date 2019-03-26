# Cassandra operator

A simple operator for management of Cassandra clusters, which creates and monitors Cassandra instances deployed in StatefulSets.

You can find information on how to deploy the operator and provision Cassandra clusters on the [wiki](https://github.com/sky-uk/cassandra-operator/wiki).

# Contributing

## Build

```
make setup install
```

## Tests

An end-to-end testing approach is used wherever possible. End-to-end tests can be run against either the AWS dev cluster or a local [Docker-in-Docker](https://github.com/kubernetes-sigs/kubeadm-dind-cluster/) cluster (default). In both cases, the tests and the operator will run in the `test-cassandra-operator` namespace.

The operator must first be deployed into the target namespace.
The end-to-end tests are run in parallel in order to the reduce build time as much as possible.
The number of parallel tests is dictated by a hardcoded value in the end-to-end test suite, which has been chosen to reflect the namespace resource quota in AWS Dev.

### Running locally using Docker-in-Docker (DIND)

The tests can be run locally inside a [Kubernetes Docker-in-Docker](https://github.com/kubernetes-sigs/kubeadm-dind-cluster/) cluster. This cluster may be created by running the `cd-scripts/create-dind-cluster.sh` script. By default the script will create a Kubernetes 1.10 cluster, but if a 1.8 or 1.9 cluster is required, the `DIND_VERSION` environment variable should be set to 1.8 or 1.9 respectively.

The cluster bootstrapping process will create a Kubernetes context named `dind` which can be used to interact with the local cluster.

The `cd-scripts/create-dind-cluster.sh` script will also create all the necessary permissions, namespaces and CRDs required to run the operator. However, the operator itself must be still be deployed into the `test-cassandra-operator` namespace. The `./cd-scripts/localBuildAndDeployToDind.sh` script can be used to build the local codebase, create a docker image and then deploy into the local DIND cluster. 

Finally, the tests can be run by running `make check KUBE_CONTEXT=dind`.
