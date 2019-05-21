## Cassandra Reaper

### Plan

Aim: Deploy one cassandra reaper per cluster

#### Prototype
1. Containerize the reaper (or use existing)
2. Deploy in Kuberetes
3. Maybe statefulset to ensure only one running
4. Connect it to Cassandra clusters in Kubernetes
5. Generate / inject credentials into the Cassandra clsuter using JMX
6. Configure access to web UI
7. Credentials for the Web UI
8. Document and review

#### Operator and Tests

In `cassandra-operator/cassandra-reaper`

1. CRD: CassandraReaper with reference to Secret containing access credentials
2. CassandraReaper Operator
   1. Create Statefulset, Service, Secret
   1. Owner refs ensure proper deletion of

Tests:
1. Write E2E tests framework
2. Real cassandra?

## Deploy on GKE

```
kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole cluster-admin \
    --user $(gcloud config get-value account)

make \
   KUBE_CONTEXT=gke_jetstack-richard_europe-west1-b_richardw-cassandra-1 \
   TEST_REGISTRY=gcr.io/jetstack-richard \
   install

make \
   KUBE_CONTEXT=gke_jetstack-richard_europe-west1-b_richardw-cassandra-1 \
   TEST_REGISTRY=gcr.io/jetstack-richard \
   deploy-operator
```
