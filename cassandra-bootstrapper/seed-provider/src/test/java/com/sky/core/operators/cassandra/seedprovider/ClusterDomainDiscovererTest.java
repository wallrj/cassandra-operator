package com.sky.core.operators.cassandra.seedprovider;

import com.sky.core.operators.cassandra.seedprovider.ClusterDomainDiscoverer;
import com.sky.core.operators.cassandra.seedprovider.PodNameDnsResolver;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.mockito.Mock;
import org.mockito.runners.MockitoJUnitRunner;

import java.net.UnknownHostException;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

@RunWith(MockitoJUnitRunner.class)
public class ClusterDomainDiscovererTest {
    @Mock
    private PodNameDnsResolver resolver;

    private ClusterDomainDiscoverer discoverer;

    @Before
    public void setUp() {
        discoverer = new ClusterDomainDiscoverer(resolver);
    }

    @Test
    public void discoversClusterNameFromKubernetesDefaultService() throws UnknownHostException {
        when(resolver.localhostCanonicalName()).thenReturn("mycluster-0.mycluster.namespace.svc.dev");

        final String clusterName = discoverer.getClusterDomain();

        assertThat(clusterName).isEqualTo("dev");
    }
}
