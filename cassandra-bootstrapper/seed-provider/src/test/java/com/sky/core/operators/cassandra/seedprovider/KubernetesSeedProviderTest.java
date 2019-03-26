package com.sky.core.operators.cassandra.seedprovider;

import com.google.common.collect.Lists;
import com.sky.core.operators.cassandra.seedprovider.KubernetesSeedProvider;
import com.sky.core.operators.cassandra.seedprovider.PodNameDnsResolver;
import com.sky.core.operators.cassandra.seedprovider.crd.Cassandra;
import com.sky.core.operators.cassandra.seedprovider.crd.CassandraSpec;
import com.sky.core.operators.cassandra.seedprovider.crd.DoneableCassandra;
import com.sky.core.operators.cassandra.seedprovider.crd.Rack;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.dsl.MixedOperation;
import io.fabric8.kubernetes.client.dsl.Resource;
import junitparams.JUnitParamsRunner;
import junitparams.Parameters;
import org.junit.Before;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.ExpectedException;
import org.junit.runner.RunWith;

import java.net.InetAddress;
import java.net.UnknownHostException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static java.lang.String.format;
import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Matchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

@RunWith(JUnitParamsRunner.class)
public class KubernetesSeedProviderTest {

    private static final String NAMESPACE = "namespace";
    private static final String CLUSTER_NAME = "mycluster";
    private KubernetesClient client;
    private MixedOperation customResourceApi;
    private PodNameDnsResolver podNameDnsResolver;
    private Resource<Cassandra, DoneableCassandra> cassandraResource;
    private KubernetesSeedProvider kubernetesSeedProvider;

    @Rule
    public ExpectedException expected = ExpectedException.none();

    @Before
    public void setup() {
        client = mock(KubernetesClient.class);
        customResourceApi = mock(MixedOperation.class);
        podNameDnsResolver = mock(PodNameDnsResolver.class);
        cassandraResource = mock(Resource.class);
    }

    @Test
    @Parameters({"1, 1", "2, 1", "3, 1", "4, 2", "5, 2", "6, 3", "7, 3", "8, 3"})
    public void findsTheCorrectNumberOfSeedsWhenThereIsOnlyOneRack(int clusterSize, int expectedSeedNodes) throws UnknownHostException {
        givenASeedProviderWithParameters(NAMESPACE, CLUSTER_NAME);
        givenAClusterDefinition(NAMESPACE, CLUSTER_NAME, rack("a", clusterSize));

        List<InetAddress> expectedSeedIps = expectIpAddressResolutionForPods("a", expectedSeedNodes);

        List<InetAddress> seeds = kubernetesSeedProvider.getSeeds();
        assertThat(seeds).containsExactlyElementsOf(expectedSeedIps);
    }

    @Test
    public void findsOneSeedPerRack() throws UnknownHostException {
        givenASeedProviderWithParameters(NAMESPACE, CLUSTER_NAME);
        givenAClusterDefinition(NAMESPACE, CLUSTER_NAME, rack("a", 4), rack("b", 3), rack("c", 3));

        when(podNameDnsResolver.localhostCanonicalName()).thenReturn(format("mycluster-a-0.mycluster.%s.svc.cluster", NAMESPACE));
        InetAddress rackaSeedIp = mock(InetAddress.class, "racka");
        InetAddress rackbSeedIp = mock(InetAddress.class, "rackb");
        InetAddress rackcSeedIp = mock(InetAddress.class, "rackc");
        when(podNameDnsResolver.resolve(format("mycluster-a-0.mycluster.%s.svc.cluster", NAMESPACE))).thenReturn(rackaSeedIp);
        when(podNameDnsResolver.resolve(format("mycluster-b-0.mycluster.%s.svc.cluster", NAMESPACE))).thenReturn(rackbSeedIp);
        when(podNameDnsResolver.resolve(format("mycluster-c-0.mycluster.%s.svc.cluster", NAMESPACE))).thenReturn(rackcSeedIp);

        List<InetAddress> seeds = kubernetesSeedProvider.getSeeds();
        assertThat(seeds).containsExactly(rackaSeedIp, rackbSeedIp, rackcSeedIp);
    }

    @Test
    public void throwsExceptionWhenNamespaceNotDefinedInSeedProviderConfig() {
        expected.expect(RuntimeException.class);
        expected.expectMessage("Missing mandatory namespace parameter in Cassandra config");
        givenASeedProviderWithParameters(null, CLUSTER_NAME);

        kubernetesSeedProvider.getSeeds();
    }

    @Test
    public void throwsExceptionWhenClusterNameNotDefinedInSeedProviderConfig() {
        expected.expect(RuntimeException.class);
        expected.expectMessage("Missing mandatory clusterName parameter in Cassandra config");
        givenASeedProviderWithParameters(NAMESPACE, null);

        kubernetesSeedProvider.getSeeds();
    }

    @Test
    public void throwsExceptionWhenCassandraResourceHasBeenDeleted() {
        expected.expect(RuntimeException.class);
        expected.expectMessage(format("Unable to find cassandra resource definition for %s.%s - Resource may have been deleted", NAMESPACE, CLUSTER_NAME));

        givenASeedProviderWithParameters(NAMESPACE, CLUSTER_NAME);
        when(client.customResources(any(), any(), any(), any())).thenReturn(customResourceApi);
        when(customResourceApi.inNamespace(NAMESPACE)).thenReturn(customResourceApi);
        when(customResourceApi.withName(CLUSTER_NAME)).thenReturn(cassandraResource);
        when(cassandraResource.get()).thenReturn(null);

        kubernetesSeedProvider.getSeeds();
    }

    @Test
    public void throwsExceptionWhenUnableToLookupClusterDomain() throws UnknownHostException {
        expected.expect(RuntimeException.class);
        expected.expectMessage("Unable to lookup cluster domain");

        givenASeedProviderWithParameters(NAMESPACE, CLUSTER_NAME);
        when(podNameDnsResolver.localhostCanonicalName()).thenThrow(new UnknownHostException("thrown for test purposes"));
        givenAClusterDefinition(NAMESPACE, CLUSTER_NAME, rack("a", 2), rack("b", 2), rack("c", 2));

        kubernetesSeedProvider.getSeeds();
    }

    private List<InetAddress> expectIpAddressResolutionForPods(String rack, int expectedSeedNodes) throws UnknownHostException {
        when(podNameDnsResolver.localhostCanonicalName()).thenReturn(format("mycluster-%s-0.mycluster.%s.svc.cluster", rack, NAMESPACE));
        List<InetAddress> expectedSeedIps = Lists.newArrayList();
        for (int i = 0; i < expectedSeedNodes; i++) {
            InetAddress podIpAddress = mock(InetAddress.class, format("pod%d-ipaddress", i));
            expectedSeedIps.add(podIpAddress);
            when(podNameDnsResolver.resolve(format("mycluster-%s-%d.mycluster.%s.svc.cluster", rack, i, NAMESPACE))).thenReturn(podIpAddress);
        }
        return expectedSeedIps;
    }

    private void givenAClusterDefinition(String namespace, String clusterName, Rack... racks) {
        Cassandra cassandra = new Cassandra();
        cassandra.getMetadata().setNamespace(namespace);
        cassandra.getMetadata().setName(clusterName);
        CassandraSpec spec = new CassandraSpec();
        cassandra.setSpec(spec);
        spec.setRacks(racks);

        when(client.customResources(any(), any(), any(), any())).thenReturn(customResourceApi);
        when(customResourceApi.inNamespace("namespace")).thenReturn(customResourceApi);
        when(customResourceApi.withName("mycluster")).thenReturn(cassandraResource);
        when(cassandraResource.get()).thenReturn(cassandra);
    }

    private void givenASeedProviderWithParameters(String namespace, String clusterName) {
        final Map<String, String> seedParameters = new HashMap<>();
        seedParameters.put("namespace", namespace);
        seedParameters.put("clusterName", clusterName);

        kubernetesSeedProvider = new KubernetesSeedProvider(client, podNameDnsResolver, seedParameters);
    }

    private Rack rack(final String name, final int replicas) {
        return new Rack(name, replicas);
    }
}
