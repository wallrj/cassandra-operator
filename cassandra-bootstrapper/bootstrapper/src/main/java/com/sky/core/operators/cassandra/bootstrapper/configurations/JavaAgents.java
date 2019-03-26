package com.sky.core.operators.cassandra.bootstrapper.configurations;

import com.sky.core.operators.cassandra.bootstrapper.ConfigurerException;

import java.io.File;
import java.util.Collections;
import java.util.List;
import java.util.stream.Collectors;

import static java.lang.String.format;

public class JavaAgents extends ConfigurationAction {

    @Override
    public void apply(final Context context) {
        String libPath = context.getTargetLibDir().getAbsolutePath();
        String configPath = context.getTargetConfDir().getAbsolutePath();
        String jmxPrometheusJar = libPath + "/jmx_prometheus_javaagent.jar";
        String jolokiaJar = libPath + "/jolokia-jvm-agent.jar";

        addJavaAgentUnlessPresent(context.getJvmOptions(), jmxPrometheusJar, format("-javaagent:%s=7070:%s/prometheus.yml", jmxPrometheusJar, configPath));
        addJavaAgentUnlessPresent(context.getJvmOptions(), jolokiaJar, format("-javaagent:%s=port=7777,host=*,policyLocation=file://%s/jolokia-policy.xml", jolokiaJar, configPath));
    }

    private void addJavaAgentUnlessPresent(final File jvmOptions, final String javaAgentPath, final String lineToAdd) {
        final List<String> jvmOptionsLines = readLines(jvmOptions);
        final List<String> potentialMatchingLines = jvmOptionsLines.stream().filter(line -> line.startsWith(format("-javaagent:%s", javaAgentPath))).collect(Collectors.toList());
        if (potentialMatchingLines.isEmpty()) {
            appendLines(jvmOptions, Collections.singletonList(lineToAdd));
        } else if (!potentialMatchingLines.contains(lineToAdd)) {
            throw new ConfigurerException("Conflicting java agent definition found in jvm.options");
        }
    }
}
