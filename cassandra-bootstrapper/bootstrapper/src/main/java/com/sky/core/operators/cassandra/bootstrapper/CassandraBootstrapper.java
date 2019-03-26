package com.sky.core.operators.cassandra.bootstrapper;

import com.sky.core.operators.cassandra.bootstrapper.configurations.*;

import java.io.File;

public class CassandraBootstrapper {
    private final EnvironmentReader environmentReader;

    CassandraBootstrapper(EnvironmentReader environmentReader) {
        this.environmentReader = environmentReader;
    }

    public static void main(String[] args) {
        new CassandraBootstrapper(new SystemEnvironmentReader()).configure();
    }

    public void configure() {
        configure(new File("/configuration"), new File("/etc/cassandra"), new File("/extra-lib"));
    }

    public void configure(File stagingDir, File targetConfDir, File targetLibDir) {
        final ConfigurationAction[] actions = new ConfigurationAction[] {
                new SeedProvider(),
                new ClusterProperties(),
                new JavaAgents(),
                new JvmMemoryDefaults(),
                new RackDC()
        };

        final Context context = new Context(
                new File(stagingDir, "jvm.options"),
                new File(stagingDir, "cassandra.yaml"),
                environmentReader,
                stagingDir,
                targetConfDir,
                targetLibDir
        );

        for (ConfigurationAction action: actions) {
            action.apply(context);
        }
    }
}
