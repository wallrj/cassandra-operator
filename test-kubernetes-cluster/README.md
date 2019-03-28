# Test Kubernetes cluster

Creates a multi-node Kubernetes cluster based on [Kubernetes Docker-in-Docker](https://github.com/kubernetes-sigs/kubeadm-dind-cluster).
The nodes are spread across multiple availability zones: `a` and `b`.
Support for persistent volumes is achieved via local volumes that are lazily created by a static [local provisioner](https://github.com/kubernetes-incubator/external-storage/tree/master/local-volume/provisioner).
Persistent volume claims can target volumes in specific AZ by using the corresponding storage class - e.g. `standard-zone-a`

This Kubernetes cluster is used in our end-to-end tests during CI, but is also very handy when testing locally.   

A local registry is started on port 5000 to allow DIND to use local docker images.
 
## Usage

```
./recreate-dind-cluster.sh
```

Optional environment variable | Meaning | Default 
---|---|---
DIND_VERSION | The dind version to install. | 1.10
ZONE_A_NODES | The number of nodes to create in the availability zone A. | 2
ZONE_B_NODES | The number of nodes to create in the availability zone B. | 2
