package com.sky.core.operators.cassandra.seedprovider;

import java.net.InetAddress;
import java.net.UnknownHostException;

public class PodNameDnsResolver {
    public InetAddress resolve(String hostname) throws UnknownHostException {
        return InetAddress.getByName(hostname);
    }

    public String localhostCanonicalName() throws UnknownHostException {
        return InetAddress.getLocalHost().getCanonicalHostName();
    }
}
