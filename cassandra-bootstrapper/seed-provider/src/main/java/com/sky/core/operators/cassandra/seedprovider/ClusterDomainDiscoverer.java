package com.sky.core.operators.cassandra.seedprovider;

import java.net.UnknownHostException;

public class ClusterDomainDiscoverer {
    private final PodNameDnsResolver resolver;

    ClusterDomainDiscoverer(PodNameDnsResolver resolver) {
        this.resolver = resolver;
    }

    public String getClusterDomain() throws UnknownHostException {
        final String canonicalName = resolver.localhostCanonicalName();
        final int clusterDomainStart = canonicalName.lastIndexOf(".svc.") + 5;
        return canonicalName.substring(clusterDomainStart);
    }
}
