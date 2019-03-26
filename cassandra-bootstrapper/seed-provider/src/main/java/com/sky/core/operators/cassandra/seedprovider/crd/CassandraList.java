package com.sky.core.operators.cassandra.seedprovider.crd;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import io.fabric8.kubernetes.client.CustomResourceList;

@JsonDeserialize
@JsonIgnoreProperties(ignoreUnknown = true)
public class CassandraList extends CustomResourceList<Cassandra> {
}
