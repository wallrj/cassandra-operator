package com.sky.core.operators.cassandra.seedprovider.crd;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import io.fabric8.kubernetes.api.model.KubernetesResource;

import java.util.Arrays;

@JsonDeserialize
@JsonIgnoreProperties(ignoreUnknown = true)
public class CassandraSpec implements KubernetesResource {
    private Rack[] racks;

    public Rack[] getRacks() {
        return racks;
    }

    public void setRacks(Rack[] racks) {
        this.racks = racks;
    }

    @Override
    public String toString() {
        return "CassandraSpec{" +
                "racks=" + Arrays.toString(racks) +
                '}';
    }
}
