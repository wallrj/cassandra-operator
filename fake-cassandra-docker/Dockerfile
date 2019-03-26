FROM openjdk:8-slim

RUN apt-get update && \
    apt-get install -y curl lsof && \
    mkdir -p /etc/cassandra && \
    touch /etc/cassandra/jvm.options /etc/cassandra/cassandra.yaml

COPY conf/fake-cassandra.yaml /etc/cassandra/cassandra.yaml
COPY conf/fake-cassandra-run /fake-cassandra-run
COPY conf/fake-nodetool /usr/local/bin/nodetool
COPY build/libs/fake-cassandra.jar /

ENTRYPOINT ["/fake-cassandra-run"]
