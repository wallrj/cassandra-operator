# Cassandra Bootstrapper

The bootstrapper image runs as an init-container on each Cassandra pod. 

It is responsible for:

- Making the custom [Kubernetes seed provider](seed-provider/README.md) available to Cassandra
- Invoking the [bootstrapper](bootstrapper/README.md) to modify the Cassandra config to reflect the configuration specified through Kubernetes (e.g. DC name, cluster name etc) 

It works as follows:

- It expects two empty-dir volumes to be present:
  - `configuration` - this should already have been populated with a default Cassandra configuration, typically by
    a prior init-container
  - `extra-lib` - this should be empty and is used for the transfer of extra JAR files required by Cassandra at
    runtime (e.g. Jolokia and Prometheus agents, and the Kubernetes seed provider)
- The bootstrapper copies config files for Jolokia and Prometheus into `configuration`, which will overwrite any
  existing files of the same name.
- If the cluster administrator has supplied config overrides through a Kubernetes ConfigMap, these files will also be
  copied into `configuration`, replacing any existing files of the same name.
- Finally, it also modifies `cassandra.yaml` to reflect the cluster configuration specified through Kubernetes, and to
  install the Kubernetes seed provider.
