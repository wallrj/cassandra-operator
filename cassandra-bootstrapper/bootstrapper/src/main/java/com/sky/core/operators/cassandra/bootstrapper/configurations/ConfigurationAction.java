package com.sky.core.operators.cassandra.bootstrapper.configurations;

import com.sky.core.operators.cassandra.bootstrapper.ConfigurerException;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.StandardOpenOption;
import java.util.List;

import static java.lang.String.format;

public abstract class ConfigurationAction {
    public abstract void apply(Context context);

    protected void writeLines(final File file, final List<String> lines) {
        if (lines.isEmpty()) {
            return;
        }

        try {
            Files.write(file.toPath(), lines);
        } catch (IOException ex) {
            throw new ConfigurerException(
                    format("Unable to write file at: %s", file.getAbsolutePath()),
                    ex
            );
        }
    }

    protected void appendLines(final File file, final List<String> lines) {
        if (lines.isEmpty()) {
            return;
        }

        try {
            Files.write(file.toPath(), lines, StandardOpenOption.APPEND);
        } catch (IOException ex) {
            throw new ConfigurerException(
                    format("Unable to append to file at: %s", file.getAbsolutePath()),
                    ex
            );
        }
    }

    protected List<String> readLines(final File file) {
        try {
            return Files.readAllLines(file.toPath());
        } catch (IOException ex) {
            throw new ConfigurerException(
                    format("Unable to read file at: %s", file.getAbsolutePath()),
                    ex
            );
        }
    }

}
