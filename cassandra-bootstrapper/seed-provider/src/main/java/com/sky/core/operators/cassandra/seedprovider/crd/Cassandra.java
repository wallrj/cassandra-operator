package com.sky.core.operators.cassandra.seedprovider.crd;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import io.fabric8.kubernetes.client.CustomResource;

@JsonDeserialize
@JsonIgnoreProperties(ignoreUnknown = true)
public class Cassandra extends CustomResource {
    private CassandraSpec spec;

    public CassandraSpec getSpec() {
        return spec;
    }

    public void setSpec(CassandraSpec spec) {
        this.spec = spec;
    }

    @Override
    public String toString() {
        return String.format("Cassandra{apiVersion='%s', metadata='%s', spec='%s'}", getApiVersion(), getMetadata(), spec);
    }
}
