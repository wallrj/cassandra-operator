FROM phusion/baseimage:0.11

COPY build/bin/cassandra-manager /cassandra-manager

ENTRYPOINT ["/sbin/my_init", "--quiet", "--", "/cassandra-manager"]
