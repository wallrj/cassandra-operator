FROM phusion/baseimage:0.10.1

COPY build/bin/cassandra-webhook /cassandra-webhook

ENTRYPOINT ["/sbin/my_init", "--quiet", "--", "/cassandra-webhook"]