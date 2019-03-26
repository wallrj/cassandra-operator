package com.sky.core.operators.cassandra.bootstrapper.configurations;

import com.sky.core.operators.cassandra.bootstrapper.EnvironmentReader;

import java.io.File;

public class Context {
    private final File jvmOptions;
    private final File cassandraYaml;
    private final EnvironmentReader environmentReader;
    private final File stagingDir;
    private final File targetConfDir;
    private final File targetLibDir;

    public Context(final File jvmOptions, final File cassandraYaml, final EnvironmentReader environmentReader, final File stagingDir, final File targetConfDir, final File targetLibDir) {
        this.jvmOptions = jvmOptions;
        this.cassandraYaml = cassandraYaml;
        this.environmentReader = environmentReader;
        this.stagingDir = stagingDir;
        this.targetConfDir = targetConfDir;
        this.targetLibDir = targetLibDir;
    }

    public File getJvmOptions() {
        return jvmOptions;
    }

    public File getCassandraYaml() {
        return cassandraYaml;
    }

    public EnvironmentReader getEnvironmentReader() {
        return environmentReader;
    }

    public File getStagingDir() {
        return stagingDir;
    }

    public File getTargetConfDir() {
        return targetConfDir;
    }

    public File getTargetLibDir() {
        return targetLibDir;
    }
}
