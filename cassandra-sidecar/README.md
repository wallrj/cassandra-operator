# Cassandra Sidecar

A golang server which runs in a sidecar container on each cassandra node.
It checks the status of the local C* node by connecting to the local Jolokia service.
It reports readiness liveness via the endpoints:
 * http://localhost:8080/ready
 * http://localhost:8080/live

These are consumed by Kubernetes HTTP readiness and liveness probes.

This approach avoids having to run Cassandra `nodetool` which,
because it is written in Java,
is resource intensive and slow to start.
