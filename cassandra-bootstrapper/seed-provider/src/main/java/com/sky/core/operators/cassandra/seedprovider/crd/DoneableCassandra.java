package com.sky.core.operators.cassandra.seedprovider.crd;

import io.fabric8.kubernetes.api.builder.Function;
import io.fabric8.kubernetes.client.CustomResourceDoneable;

public class DoneableCassandra extends CustomResourceDoneable<Cassandra> {
    public DoneableCassandra(Cassandra resource, Function<Cassandra, Cassandra> function) {
        super(resource, function);
    }
}
