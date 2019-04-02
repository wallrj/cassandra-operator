package com.sky.core.operators.cassandra.bootstrapper.configurations;

import org.yaml.snakeyaml.Yaml;

import java.util.*;

public class SeedProvider extends ConfigurationAction {
    @Override
    public void apply(final Context context) {
        final boolean useDefaultSeedProvider = Boolean.valueOf(context.getEnvironmentReader().read("CLUSTER_USE_DEFAULT_SEED_PROVIDER").orElse("false"));
        if (useDefaultSeedProvider) {
            return;
        }

        final String namespace = context.getEnvironmentReader().readMandatory("CLUSTER_NAMESPACE");
        final String clusterName = context.getEnvironmentReader().readMandatory("CLUSTER_NAME");

        final List<String> seedProviderSection = buildSeedProviderYamlSection(namespace, clusterName);

        final List<String> originalFile = readLines(context.getCassandraYaml());
        final List<String> modifiedFile = new ArrayList<>();
        boolean replacingSeedProvider = false;
        for (String line : originalFile) {
            if (line.equals("seed_provider:")) {
                replacingSeedProvider = true;
                modifiedFile.addAll(seedProviderSection);
            }

            if (!replacingSeedProvider || line.startsWith("#")) {
                modifiedFile.add(line);
                replacingSeedProvider = false;
            }
        }

        writeLines(context.getCassandraYaml(), modifiedFile);
    }

    private List<String> buildSeedProviderYamlSection(final String namespace, final String clusterName) {
        Map<String, String> customSeedProviderParams = new HashMap<>();
        customSeedProviderParams.put("namespace", namespace);
        customSeedProviderParams.put("clusterName", clusterName);

        Map<String, Object> customSeedProvider = new HashMap<>();
        customSeedProvider.put("class_name", "com.sky.core.operators.cassandra.seedprovider.KubernetesSeedProvider");
        customSeedProvider.put("parameters", Collections.singletonList(customSeedProviderParams));

        Yaml yaml = new Yaml();
        Map<String, Object> cassandraConfig = new HashMap<>();
        cassandraConfig.put("seed_provider", Collections.singletonList(customSeedProvider));

        return Arrays.asList(yaml.dump(cassandraConfig).split("\\n"));
    }

}
