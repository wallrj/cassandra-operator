package com.sky.core.operators.cassandra.bootstrapper;

import java.util.Optional;

import static java.lang.String.format;

public class SystemEnvironmentReader implements EnvironmentReader {
    @Override
    public String readMandatory(final String variableName) {
        return read(variableName).orElseThrow(() -> new ConfigurerException(format("Mandatory environment variable %s not defined", variableName)));
    }

    @Override
    public Optional<String> read(final String variableName) {
        return Optional.ofNullable(System.getenv(variableName));
    }
}
