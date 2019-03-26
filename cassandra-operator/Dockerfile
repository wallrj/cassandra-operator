FROM phusion/baseimage:0.10.1

COPY build/bin/cassandra-operator /cassandra-operator

ENTRYPOINT ["/sbin/my_init", "--quiet", "--", "/cassandra-operator"]