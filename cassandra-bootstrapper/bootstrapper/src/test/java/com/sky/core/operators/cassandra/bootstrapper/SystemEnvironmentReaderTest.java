package com.sky.core.operators.cassandra.bootstrapper;

import com.sky.core.operators.cassandra.bootstrapper.ConfigurerException;
import com.sky.core.operators.cassandra.bootstrapper.SystemEnvironmentReader;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.ExpectedException;

import java.util.UUID;

import static java.lang.String.format;
import static org.assertj.core.api.Assertions.assertThat;

public class SystemEnvironmentReaderTest {

    @Rule
    public ExpectedException expected = ExpectedException.none();

    @Test
    public void throwsExceptionWhenMissingMandatoryVariable() {
        String variableName = UUID.randomUUID().toString();
        expected.expect(ConfigurerException.class);
        expected.expectMessage(format("Mandatory environment variable %s not defined", variableName));

        new SystemEnvironmentReader().readMandatory(variableName);
    }

    @Test
    public void returnTheVariableValue() {
        assertThat(new SystemEnvironmentReader().readMandatory("PATH")).isNotEmpty();
    }
}
