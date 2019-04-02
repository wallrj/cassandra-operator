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

## Testing 

[dgoss](https://github.com/aelsabbahy/goss/tree/master/extras/dgoss) is used to validate that the Cassandra image 
can be started with the Cassandra config modified by the bootstrapper.  
The tests setup intermediate docker volumes that are modified by the `cassandra-bootstrapper` 
and then used as configuration by the Cassandra image. 
For practical reasons, the tests use the default Cassandra Seed Provider as testing the custom Kubernetes Seed Provider 
would require integrating with a Kubernetes cluster.
 
All test scenarios share a common set of specifications, but can also specify their own via a custom `run.sh`.
For instance `test-with-user-config` scenario has an additional `goss.yaml` to validate additional files 
can be provided via the user's ConfigMap. 
 
To check a single test scenario (e.g. `test-with-user-config`) with additional logging run the following:
```
make install

GOSS_WAIT_OPTS="-r 30s -s 1s" IMAGE_TO_TEST=<image-just-built> test/test-with-user-config/run.sh
```
