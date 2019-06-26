FROM phusion/baseimage:0.11

COPY build/bin/cassandra-sidecar /cassandra-sidecar

ENTRYPOINT ["/sbin/my_init", "--quiet", "--", "/cassandra-sidecar"]
