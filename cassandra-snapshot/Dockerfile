FROM phusion/baseimage:0.11

COPY build/bin/cassandra-snapshot /cassandra-snapshot

ENTRYPOINT ["/sbin/my_init", "--quiet", "--", "/cassandra-snapshot"]
