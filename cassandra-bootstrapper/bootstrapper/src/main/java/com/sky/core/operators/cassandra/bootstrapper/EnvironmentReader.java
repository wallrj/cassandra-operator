package com.sky.core.operators.cassandra.bootstrapper;

import java.util.Optional;

public interface EnvironmentReader {
    String readMandatory(String variableName);

    Optional<String> read(String variableName);
}
