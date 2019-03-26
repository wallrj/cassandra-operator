package com.sky.core.operators.cassandra.bootstrapper.configurations;

import java.io.File;
import java.util.ArrayList;
import java.util.List;

import static java.lang.String.format;

public class ClusterProperties extends ConfigurationAction {
    @Override
    public void apply(final Context context) {
        configureSnitch(context);
        configureClusterName(context);
        configureRpcAndListenAddress(context);
    }

    private void configureSnitch(final Context context) {
        setPropertyInCassandraYaml(context.getCassandraYaml(),"endpoint_snitch", "GossipingPropertyFileSnitch");
    }

    private void configureClusterName(final Context context) {
        final String clusterName = context.getEnvironmentReader().readMandatory("CLUSTER_NAME");
        setPropertyInCassandraYaml(context.getCassandraYaml(), "cluster_name", clusterName);
    }

    private void configureRpcAndListenAddress(final Context context) {
        final String podIp = context.getEnvironmentReader().readMandatory("NODE_LISTEN_ADDRESS");
        setPropertyInCassandraYaml(context.getCassandraYaml(), "listen_address", podIp);
        setPropertyInCassandraYaml(context.getCassandraYaml(), "rpc_address", podIp);
    }

    private void setPropertyInCassandraYaml(final File cassandraYaml, final String propertyName, final String replacementValue) {
        final String propertyAndReplacementValue = format("%s: %s", propertyName, replacementValue);

        final List<String> originalFile = readLines(cassandraYaml);
        final List<String> modifiedFile = new ArrayList<>();
        boolean propertyWasUpdated = false;
        for (String line: originalFile) {
            if (line.startsWith(format("%s: ", propertyName))) {
                modifiedFile.add(propertyAndReplacementValue);
                propertyWasUpdated = true;
            } else {
                modifiedFile.add(line);
            }
        }

        if (!propertyWasUpdated) {
            modifiedFile.add(propertyAndReplacementValue);
        }

        writeLines(cassandraYaml, modifiedFile);
    }

}
