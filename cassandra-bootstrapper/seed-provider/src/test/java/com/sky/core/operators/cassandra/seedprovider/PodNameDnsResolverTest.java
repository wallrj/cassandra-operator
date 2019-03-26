package com.sky.core.operators.cassandra.seedprovider;

import com.sky.core.operators.cassandra.seedprovider.PodNameDnsResolver;
import org.junit.Before;
import org.junit.Test;

import java.net.InetAddress;
import java.net.UnknownHostException;

import static org.assertj.core.api.Assertions.assertThat;

public class PodNameDnsResolverTest {

    private PodNameDnsResolver podNameDnsResolver;

    @Before
    public void setUp() {
        podNameDnsResolver = new PodNameDnsResolver();
    }

    @Test
    public void canResolveAHost() throws UnknownHostException {
        InetAddress localhost = podNameDnsResolver.resolve("localhost");
        assertThat(localhost.getAddress()).isEqualTo(new byte[]{127, 0, 0, 1});
    }
}

