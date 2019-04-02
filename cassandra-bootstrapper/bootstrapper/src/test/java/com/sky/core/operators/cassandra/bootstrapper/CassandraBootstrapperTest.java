package com.sky.core.operators.cassandra.bootstrapper;

import com.sky.core.operators.cassandra.seedprovider.KubernetesSeedProvider;
import junitparams.JUnitParamsRunner;
import junitparams.Parameters;
import org.apache.cassandra.config.Config;
import org.apache.cassandra.config.YamlConfigurationLoader;
import org.assertj.core.api.Condition;
import org.assertj.core.util.diff.Delta;
import org.assertj.core.util.diff.DiffUtils;
import org.junit.Before;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.ExpectedException;
import org.junit.rules.TemporaryFolder;
import org.junit.runner.RunWith;

import java.io.File;
import java.io.IOException;
import java.net.MalformedURLException;
import java.net.URISyntaxException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;

import static java.lang.String.format;
import static org.assertj.core.api.Assertions.assertThat;

@RunWith(JUnitParamsRunner.class)
public class CassandraBootstrapperTest {
    private static final long GIGABYTE = 1024 * 1024 * 1024;
    private static final String NAMESPACE = "some-namespace";
    private static final String CLUSTER = "some-cluster";
    private static final String CURRENT_RACK = "racka";
    private static final String DC = "my-dc";
    private static final String REFERENCE_CASSANDRA_YAML = "/cassandra.yaml";

    private static final List<String> SEED_PROVIDER_BLOCK = Arrays.asList(
            format("- class_name: %s", KubernetesSeedProvider.class.getName()),
            "  parameters:",
            "  - {clusterName: some-cluster, namespace: some-namespace}"
    );

    private final Condition<Delta<String>> SEED_PROVIDER_CHANGE = new Condition<>(
            d -> d.getRevised().getLines().equals(SEED_PROVIDER_BLOCK),
            "seed_provider change"
    );

    private static final Condition<Delta<String>> ENDPOINT_SNITCH_CHANGE = new Condition<>(
            d -> d.getRevised().getLines().get(0).startsWith("endpoint_snitch: "),
            "endpoint_snitch change"
    );

    private static final Condition<Delta<String>> CLUSTER_NAME_CHANGE = new Condition<>(
            d -> d.getRevised().getLines().get(0).startsWith("cluster_name: "),
            "cluster_name change"
    );

    private static final Condition<Delta<String>> LISTEN_ADDRESS_CHANGE = new Condition<>(
            d -> d.getRevised().getLines().get(0).startsWith("listen_address: "),
            "listen_address change"
    );

    private static final Condition<Delta<String>> RPC_ADDRESS_CHANGE = new Condition<>(
            d -> d.getRevised().getLines().get(0).startsWith("rpc_address: "),
            "rpc_address change"
    );

    @Rule
    public TemporaryFolder cassandraConfigFolder = new TemporaryFolder();

    private final File targetConfDir = new File("/etc/cassandra");
    private final File targetLibDir = new File("/extra-lib");

    @Rule
    public ExpectedException expected = ExpectedException.none();

    private StubbedEnvironmentReader environmentReader;
    private File jvmOptions;
    private File cassandraYaml;

    @Before
    public void setUp() throws IOException {
        environmentReader =  new StubbedEnvironmentReader(cassandraConfigFolder);
        jvmOptions = cassandraConfigFolder.newFile("jvm.options");
        cassandraYaml = new File(cassandraConfigFolder.getRoot(), "cassandra.yaml");
        Files.copy(getClass().getResourceAsStream(REFERENCE_CASSANDRA_YAML), cassandraYaml.toPath());
    }

    @Test
    public void addsAgentsDefinitionToJvmOptions() throws IOException {
        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);

        List<String> lines = Files.readAllLines(jvmOptions.toPath());
        assertThat(lines).contains(
                format("-javaagent:%s/jmx_prometheus_javaagent.jar=7070:%s/prometheus.yml", targetLibDir, targetConfDir),
                format("-javaagent:%s/jolokia-jvm-agent.jar=port=7777,host=*,policyLocation=file://%s/jolokia-policy.xml", targetLibDir, targetConfDir)
        );
    }

    @Parameters({"1100,550", "0, 1"})
    @Test
    public void setsDefaultHeapSizeToHalfPodRequestedMemoryInJvmOptions(final int memory, final int heap) throws IOException {
        environmentReader.addEnvironmentVariable("POD_MEMORY_BYTES", String.valueOf(memory));

        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);

        List<String> lines = Files.readAllLines(jvmOptions.toPath());
        assertThat(lines).contains(String.format("-Xmx%d", heap));
        assertThat(lines).contains(String.format("-Xms%d", heap));
    }

    @SuppressWarnings("Duplicates")
    @Test
    @Parameters({"1000,100", "2000, 200", "2100, 210", "2115, 211"})
    public void setsYoungGenerationSizeTo100MBPerCpu(final int cpuMillis, final int youngGenSize) throws IOException {
        environmentReader.addEnvironmentVariable("POD_CPU_MILLICORES", String.valueOf(cpuMillis));
        environmentReader.addEnvironmentVariable("POD_MEMORY_BYTES", String.valueOf(10 * GIGABYTE));

        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);

        List<String> lines = Files.readAllLines(jvmOptions.toPath());
        assertThat(lines).contains(String.format("-Xmn%dM", youngGenSize));
    }

    @SuppressWarnings("Duplicates")
    @Test
    @Parameters({"100, 100", "0, 100"})
    public void setsYoungGenerationSizeToAMinimumOf100MB(final int cpuMillis, final int youngGenSize) throws IOException {
        environmentReader.addEnvironmentVariable("POD_CPU_MILLICORES", String.valueOf(cpuMillis));
        environmentReader.addEnvironmentVariable("POD_MEMORY_BYTES", String.valueOf(10 * GIGABYTE));

        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);

        List<String> lines = Files.readAllLines(jvmOptions.toPath());
        assertThat(lines).contains(String.format("-Xmn%dM", youngGenSize));
    }

    @Test
    @Parameters({"10000,1,128", "20000,1,128"})
    public void setsYoungGenerationSizeToMaxQuarterOfHeap(final int cpuMillis, final int podMemoryGbs, final int youngGenSize) throws IOException {
        environmentReader.addEnvironmentVariable("POD_CPU_MILLICORES", String.valueOf(cpuMillis));
        environmentReader.addEnvironmentVariable("POD_MEMORY_BYTES", String.valueOf(podMemoryGbs * GIGABYTE));

        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);

        List<String> lines = Files.readAllLines(jvmOptions.toPath());
        assertThat(lines).contains(String.format("-Xmn%dM", youngGenSize));
    }

    @Test
    public void doesNotAddDuplicateAgentDefinitionsToJvmOptions() throws IOException {
        final List<String> agentDefinitions = Arrays.asList(
            format("-javaagent:/somepath/jmx.prometheus.jar=7070:%s/prometheus.yml", cassandraConfigFolder.getRoot().getAbsolutePath()),
            format("-javaagent:/somepath/jolokia.jar=port=7777,host=*,policyLocation=file://%s/operator/jolokia-policy.xml", cassandraConfigFolder.getRoot().getAbsolutePath())
        );
        Files.write(jvmOptions.toPath(), agentDefinitions);

        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);

        final List<String> modifiedFile = Files.readAllLines(jvmOptions.toPath());
        for (String agentDefinition: agentDefinitions) {
            assertThat(modifiedFile).containsOnlyOnce(agentDefinition);
        }
    }

    @Test
    public void failsIfConflictingAgentDefinitionExists() throws IOException {
        Files.write(jvmOptions.toPath(), format("-javaagent:%s/jmx_prometheus_javaagent.jar=7070:/otherpath/prometheus.yml", targetLibDir).getBytes());

        expected.expect(ConfigurerException.class);
        expected.expectMessage("Conflicting java agent definition found in jvm.options");
        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);
    }

    @Test
    public void setsClusterNameInCassandraYaml() throws Exception {
        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);

        final YamlConfigurationLoader loader = new YamlConfigurationLoader();
        final Config modifiedConfig = loader.loadConfig(cassandraYaml.toURI().toURL());

        assertThat(modifiedConfig.cluster_name).isEqualTo(CLUSTER);
    }

    @Test
    public void addsRackAwareSeedProviderDefinitionToCassandraYaml() throws IOException, URISyntaxException {
        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);

        final YamlConfigurationLoader loader = new YamlConfigurationLoader();
        final Config modifiedConfig = loader.loadConfig(cassandraYaml.toURI().toURL());

        assertThat(modifiedConfig.seed_provider.class_name).isEqualTo(KubernetesSeedProvider.class.getName());
        assertThat(modifiedConfig.seed_provider.parameters.get("namespace")).isEqualTo(NAMESPACE);
        assertThat(modifiedConfig.seed_provider.parameters.get("clusterName")).isEqualTo(CLUSTER);
        assertThatAllRequiredChangesHaveBeenMade(Paths.get(getClass().getResource(REFERENCE_CASSANDRA_YAML).toURI()));
    }

    @Test
    public void doesNotChangeSeedProviderDefinitionIfDefaultIsRequested() throws IOException, URISyntaxException {
        environmentReader.addEnvironmentVariable("CLUSTER_USE_DEFAULT_SEED_PROVIDER", "true");

        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);

        List<String> sourceLines = Files.readAllLines(Paths.get(getClass().getResource(REFERENCE_CASSANDRA_YAML).toURI()));
        List<String> updatedLines = Files.readAllLines(cassandraYaml.toPath());

        final List<Delta<String>> changes = DiffUtils.diff(sourceLines, updatedLines).getDeltas();

        assertThat(changes).hasSize(4);
        assertThat(changes).areExactly(1, ENDPOINT_SNITCH_CHANGE);
        assertThat(changes).areExactly(1, CLUSTER_NAME_CHANGE);
        assertThat(changes).areExactly(1, LISTEN_ADDRESS_CHANGE);
        assertThat(changes).areExactly(1, RPC_ADDRESS_CHANGE);
    }

    @Test
    public void setsGossipingPropertyFileSnitchAsTheChosenSnitch() throws MalformedURLException {
        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);

        final YamlConfigurationLoader loader = new YamlConfigurationLoader();
        final Config modifiedConfig = loader.loadConfig(cassandraYaml.toURI().toURL());

        assertThat(modifiedConfig.endpoint_snitch).isEqualTo("GossipingPropertyFileSnitch");
    }

    @Test
    public void setsPodIpAsBothRpcAndListenAddress() throws MalformedURLException {
        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);

        final YamlConfigurationLoader loader = new YamlConfigurationLoader();
        final Config modifiedConfig = loader.loadConfig(cassandraYaml.toURI().toURL());

        assertThat(modifiedConfig.listen_address).isEqualTo("some-pod-ip");
        assertThat(modifiedConfig.rpc_address).isEqualTo("some-pod-ip");
    }

    @Test
    public void failsIfJvmOptionsFileIsNotPresent() {
        expected.expect(ConfigurerException.class);
        expected.expectMessage(format("Unable to read file at: %s/jvm.options", cassandraConfigFolder.getRoot().getAbsolutePath()));

        jvmOptions.delete();
        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);
    }

    @Test
    public void failsIfCassandraYamlNotPresent() {
        expected.expect(ConfigurerException.class);
        expected.expectMessage(format("Unable to read file at: %s/cassandra.yaml", cassandraConfigFolder.getRoot().getAbsolutePath()));

        cassandraYaml.delete();
        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);
    }

    @Test
    public void addsRackAndDcInformationToRackDefinitionFile() throws IOException {
        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);

        List<String> rackDcLines = Files.readAllLines(new File(cassandraConfigFolder.getRoot(), "cassandra-rackdc.properties").toPath());
        assertThat(rackDcLines).contains(
            format("dc=%s", DC),
            format("rack=%s", CURRENT_RACK)
        );
    }

    @Parameters({"CLUSTER_NAMESPACE", "CLUSTER_NAME", "CLUSTER_CURRENT_RACK", "CLUSTER_DATA_CENTER", "NODE_LISTEN_ADDRESS", "POD_MEMORY_BYTES", "POD_CPU_MILLICORES",})
    @Test
    public void failsWhenMandatoryEnvVariablesAreNotProvided(String missingEnvVariable) {
        expected.expect(ConfigurerException.class);
        expected.expectMessage(format("Mandatory environment variable %s not defined", missingEnvVariable));

        environmentReader.removeEnvironmentVariable(missingEnvVariable);
        new CassandraBootstrapper(environmentReader).configure(cassandraConfigFolder.getRoot(), targetConfDir, targetLibDir);
    }

    private void assertThatAllRequiredChangesHaveBeenMade(Path sourceFilePath) throws IOException {
        final List<String> sourceLines = Files.readAllLines(sourceFilePath);
        final List<String> updatedLines = Files.readAllLines(new File(cassandraConfigFolder.getRoot(), "cassandra.yaml").toPath());

        final List<Delta<String>> changes = DiffUtils.diff(sourceLines, updatedLines).getDeltas();
        assertThat(changes).hasSize(5);
        assertThat(changes).areExactly(1, ENDPOINT_SNITCH_CHANGE);
        assertThat(changes).areExactly(1, SEED_PROVIDER_CHANGE);
        assertThat(changes).areExactly(1, CLUSTER_NAME_CHANGE);
        assertThat(changes).areExactly(1, LISTEN_ADDRESS_CHANGE);
        assertThat(changes).areExactly(1, RPC_ADDRESS_CHANGE);
    }

    private class StubbedEnvironmentReader extends SystemEnvironmentReader {
        private final Map<String, String> envVariables;

        StubbedEnvironmentReader(TemporaryFolder cassandraConfigFolder) {
            envVariables = new HashMap<String, String>() {{
                put("JMX_PROMETHEUS_JAR", "/somepath/jmx.prometheus.jar");
                put("JOLOKIA_JAR", "/somepath/jolokia.jar");
                put("CASSANDRA_CONFIG", cassandraConfigFolder.getRoot().getAbsolutePath());
                put("CLUSTER_NAMESPACE", NAMESPACE);
                put("CLUSTER_NAME", CLUSTER);
                put("CLUSTER_CURRENT_RACK", CURRENT_RACK);
                put("CLUSTER_DATA_CENTER", DC);
                put("NODE_LISTEN_ADDRESS", "some-pod-ip");
                put("POD_MEMORY_BYTES", "0");
                put("POD_CPU_MILLICORES", "0");
            }};
        }

        void addEnvironmentVariable(String name, String value) {
            envVariables.put(name, value);
        }

        void removeEnvironmentVariable(String name) {
            envVariables.remove(name);
        }

        @Override
        public Optional<String> read(String variableName) {
            return Optional.ofNullable(envVariables.get(variableName));
        }
    }

}
