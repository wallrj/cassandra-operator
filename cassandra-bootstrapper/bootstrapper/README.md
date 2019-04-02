# Bootstrapper

The bootstrapper is responsible for modifying the Cassandra configuration so it can be used inside Kubernetes.

It preserves as much as possible of the user's provided Cassandra configuration by:

- copying the default Cassandra configuration from the user's provided Cassandra image
- applying the user's ConfigMap custom configuration (if any) on top of the default configuration
- modifying the Cassandra configuration to make it suitable to run Cassandra as a pod, this involves:

  - updating the `rack` and `dc` properties in `cassandra-rackdc.properties`
  - registering the custom [Kubernetes seed provider](../seed-provider/README.md) in `cassandra.yaml` 
  - configuring the Java heap in `jvm.options` to 1/2 the pod requested memory, similarly to how `cassandra-env.sh` would do on standard VM nodes 
  - configuring the Java young generation in `jvm.options` based on the number of pod requested cpu, similarly to how `cassandra-env.sh` would do on standard VM nodes
  - add the JMX Prometheus and Jolokia java agents definition to the `jvm.options` and copies the JAR to the extra libraries area
