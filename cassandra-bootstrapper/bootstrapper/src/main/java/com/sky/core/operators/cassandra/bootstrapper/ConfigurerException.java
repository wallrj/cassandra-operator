package com.sky.core.operators.cassandra.bootstrapper;

public class ConfigurerException extends RuntimeException {
    public ConfigurerException(String message) {
        super(message);
    }

    public ConfigurerException(final String message, final Throwable cause) {
        super(message, cause);
    }
}
