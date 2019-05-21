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

## Notes

Clusters can not be deleted until they have been fully initialized.
Operator does not appear to perform any further actions until it has completed the creation.

Document that Rack > Zone == node zone

```
Exception (org.apache.cassandra.exceptions.ConfigurationException) encountered during startup: com.sky.core.operators.cassandra.seedprovider.KubernetesSeedProvider
Fatal configuration error; unable to start server.  See log for stacktrace.
org.apache.cassandra.exceptions.ConfigurationException: com.sky.core.operators.cassandra.seedprovider.KubernetesSeedProvider
Fatal configuration error; unable to start server.  See log for stacktrace.
       at org.apache.cassandra.config.DatabaseDescriptor.applySeedProvider(DatabaseDescriptor.java:901)
       at org.apache.cassandra.config.DatabaseDescriptor.applyAll(DatabaseDescriptor.java:330)
       at org.apache.cassandra.config.DatabaseDescriptor.daemonInitialization(DatabaseDescriptor.java:148)
       at org.apache.cassandra.config.DatabaseDescriptor.daemonInitialization(DatabaseDescriptor.java:132)
       at org.apache.cassandra.service.CassandraDaemon.applyConfig(CassandraDaemon.java:665)
       at org.apache.cassandra.service.CassandraDaemon.activate(CassandraDaemon.java:609)
       at org.apache.cassandra.service.CassandraDaemon.main(CassandraDaemon.java:732)
ERROR [main] 2019-05-21 13:41:45,902 CassandraDaemon.java:749 - Exception encountered during startup
org.apache.cassandra.exceptions.ConfigurationException: com.sky.core.operators.cassandra.seedprovider.KubernetesSeedProvider
Fatal configuration error; unable to start server.  See log for stacktrace.
       at org.apache.cassandra.config.DatabaseDescriptor.applySeedProvider(DatabaseDescriptor.java:901) ~[apache-cassandra-3.11.4.jar:3.11.4]
       at org.apache.cassandra.config.DatabaseDescriptor.applyAll(DatabaseDescriptor.java:330) ~[apache-cassandra-3.11.4.jar:3.11.4]
       at org.apache.cassandra.config.DatabaseDescriptor.daemonInitialization(DatabaseDescriptor.java:148) ~[apache-cassandra-3.11.4.jar:3.11.4]
       at org.apache.cassandra.config.DatabaseDescriptor.daemonInitialization(DatabaseDescriptor.java:132) ~[apache-cassandra-3.11.4.jar:3.11.4]
       at org.apache.cassandra.service.CassandraDaemon.applyConfig(CassandraDaemon.java:665) [apache-cassandra-3.11.4.jar:3.11.4]
       at org.apache.cassandra.service.CassandraDaemon.activate(CassandraDaemon.java:609) [apache-cassandra-3.11.4.jar:3.11.4]
       at org.apache.cassandra.service.CassandraDaemon.main(CassandraDaemon.java:732) [apache-cassandra-3.11.4.jar:3.11.4]
```

```
LOCAL_JMX="no"
CLUSTER_USE_DEFAULT_SEED_PROVIDER="true"
```

```
INFO   [2019-05-21 14:12:34,843] [dw-15 - POST /cluster?seedHost=10.16.2.17&jmxPort=7199] i.c.j.JmxConnectionFactory - Unreachable host: Failure when establishing JMX connection to 10.16.2.17:7
199: java.lang.SecurityException: Authentication failed! Credentials required
ERROR  [2019-05-21 14:12:34,846] [dw-15 - POST /cluster?seedHost=10.16.2.17&jmxPort=7199] i.c.r.ClusterResource - failed to find cluster with seed hosts: [10.16.2.17]
io.cassandrareaper.ReaperException: no host could be reached through JMX
```


```
kubectl -n test-cassandra-operator exec mycluster-a-0 -- touch /etc/cassandra/jmxremote.password
```
