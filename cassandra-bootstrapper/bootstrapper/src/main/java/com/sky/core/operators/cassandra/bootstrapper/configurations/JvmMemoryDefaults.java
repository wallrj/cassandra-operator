package com.sky.core.operators.cassandra.bootstrapper.configurations;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import static java.lang.Math.max;
import static java.lang.Math.min;

public class JvmMemoryDefaults extends ConfigurationAction {
    @Override
    public void apply(final Context context) {
        final List<String> jvmOptionsLines = readLines(context.getJvmOptions());
        final List<String> linesToAdd = new ArrayList<>();

        String podMemoryInBytes = context.getEnvironmentReader().readMandatory("POD_MEMORY_BYTES");
        final long heapSizeInBytes = max(1, Long.parseLong(podMemoryInBytes) / 2);

        if (jvmOptionsLines.stream().noneMatch(line -> line.startsWith("-Xmx"))) {
            linesToAdd.add(String.format("-Xmx%d", heapSizeInBytes));
        }

        if (jvmOptionsLines.stream().noneMatch(line -> line.startsWith("-Xms"))) {
            linesToAdd.add(String.format("-Xms%d", heapSizeInBytes));
        }

        String podCpu = context.getEnvironmentReader().readMandatory("POD_CPU_MILLICORES");
        if (jvmOptionsLines.stream().noneMatch(line -> line.startsWith("-Xmn"))) {

            long youngGenInMB = Math.max(Long.parseLong(podCpu) / 10, 100);
            Optional<String> memoryInBytes = context.getEnvironmentReader().read("POD_MEMORY_BYTES");
            if (memoryInBytes.isPresent()) {
                long memoryInMB = bytesToMegabytes(Long.parseLong(memoryInBytes.get()));
                youngGenInMB = min(youngGenInMB, memoryInMB / 8);
            }
            linesToAdd.add(String.format("-Xmn%dM", max(youngGenInMB, 1)));
        }

        appendLines(context.getJvmOptions(), linesToAdd);
    }

    private long bytesToMegabytes(final long byteVal) {
        return byteVal / 1048576L;
    }
}
