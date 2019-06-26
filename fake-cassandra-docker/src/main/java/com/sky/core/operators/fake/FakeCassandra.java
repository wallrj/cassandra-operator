package com.sky.core.operators.fake;

import fi.iki.elonen.NanoHTTPD;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.yaml.snakeyaml.Yaml;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

public class FakeCassandra {
    private static final Logger LOGGER = LoggerFactory.getLogger(FakeCassandra.class);

    private final FakeCassandraConfig config;
    private final FakeJolokiaServer jolokiaServer;
    private final FakeMetricsServer metricsServer;

    public static void main(String[] args) {
        try {
            FakeCassandraConfig cassandraConfig = new FakeCassandraConfig();
            FakeCassandra fc = new FakeCassandra(
                    cassandraConfig,
                    new FakeJolokiaServer(new FakeJolokiaServerConfig(cassandraConfig.nodeListenAddress)),
                    new FakeMetricsServer()
            );
            fc.run();
        } catch (IOException e) {
            LOGGER.error("FakeCassandra startup failed", e);
            System.exit(1);
        }
    }

    private FakeCassandra(FakeCassandraConfig config,
                          FakeJolokiaServer jolokiaServer,
                          FakeMetricsServer metricsServer) {
        this.config = config;
        this.jolokiaServer = jolokiaServer;
        this.metricsServer = metricsServer;
    }

    private void run() throws IOException {
        LOGGER.info("Starting fake cassandra with config {}", config);

        final ExecutorService exec = Executors.newFixedThreadPool(this.config.dummyPorts.length);
        final ServerSocket[] serverSockets = new ServerSocket[this.config.dummyPorts.length];
        for (int i = 0; i < this.config.dummyPorts.length; i++) {
            serverSockets[i] = createSocket(this.config.dummyPorts[i]);
        }

        Arrays.stream(serverSockets).forEach(s -> exec.execute(() -> startListening(s)));

        this.jolokiaServer.start();
        this.metricsServer.start();

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            LOGGER.info("Shutting down");
            Arrays.stream(serverSockets).forEach(this::closeQuietly);
            this.jolokiaServer.stop();
            this.metricsServer.stop();
            exec.shutdown();
            try {
                exec.awaitTermination(10, TimeUnit.SECONDS);
            } catch (InterruptedException e) {
                LOGGER.error("Failed while waiting for executor termination", e);
            }
        }));
    }

    private void closeQuietly(final ServerSocket serverSocket) {
        try {
            serverSocket.close();
        } catch (IOException ex) {
            LOGGER.error(String.format("Error while closing server socket on port %d", serverSocket.getLocalPort()), ex);
        }
    }

    private ServerSocket createSocket(final int port) throws IOException {
        return new ServerSocket(port);
    }

    private void startListening(final ServerSocket serverSocket) {
        LOGGER.info("Listening on port {}", serverSocket.getLocalPort());
        while (!serverSocket.isClosed()) {
            try {
                final Socket client = serverSocket.accept();
                LOGGER.info("Accepted connection from {} on port {}", client.getInetAddress().toString(), serverSocket.getLocalPort());
                client.close();
            } catch (IOException ex) {
                LOGGER.error(String.format("Error while handling client connection on port %d", serverSocket.getLocalPort()), ex);
            }
        }
    }
}

class FakeJolokiaServerConfig {
    final Map<String, String> permittedPaths;

    FakeJolokiaServerConfig(String nodeListenAddress) {
        permittedPaths = new HashMap<>();
        permittedPaths.put("/jolokia/exec/org.apache.cassandra.db:type=EndpointSnitchInfo/getRack/" + nodeListenAddress,
                "{\"status\":200}");
        permittedPaths.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/LiveNodes",
                "{\"status\":200, \"value\":[\"" + nodeListenAddress + "\"]}");
        permittedPaths.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/UnreachableNodes",
                "{\"status\":200}");
        permittedPaths.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/JoiningNodes",
                "{\"status\":200}");
        permittedPaths.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/LeavingNodes",
                "{\"status\":200}");
        permittedPaths.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/MovingNodes",
                "{\"status\":200}");
    }

    @Override
    public String toString() {
        return "FakeJolokiaServerConfig{" +
                "permittedPaths=" + permittedPaths +
                '}';
    }
}

class FakeCassandraConfig {
    private static final Logger LOGGER = LoggerFactory.getLogger(FakeCassandraConfig.class);
    private static final String CASSANDRA_YAML_PATH = "/etc/cassandra/cassandra.yaml";
    private static final String DEFAULT_NODE_ADDRESS = "localhost";

    final int[] dummyPorts = new int[]{7000, 7199, 9042};
    final String nodeListenAddress;

    FakeCassandraConfig() throws FileNotFoundException {
        this(readNodeListenAddressFromCassandraYaml());
    }

    private FakeCassandraConfig(String nodeListenAddress) {
        this.nodeListenAddress = nodeListenAddress;
    }

    private static String readNodeListenAddressFromCassandraYaml() throws FileNotFoundException {
        Map<String, Object> cassandraConfigAsMap = new Yaml().load(new FileInputStream(new File(CASSANDRA_YAML_PATH)));
        LOGGER.info("{}: {}", CASSANDRA_YAML_PATH, cassandraConfigAsMap);

        String nodeListenAddress = (String) cassandraConfigAsMap.get("listen_address");
        if (nodeListenAddress == null) {
            LOGGER.warn("{}:listen_address:null. Using default instead.", CASSANDRA_YAML_PATH);
            return DEFAULT_NODE_ADDRESS;
        }
        return nodeListenAddress;
    }

    @Override
    public String toString() {
        return "FakeCassandraConfig{" +
                "dummyPorts=" + Arrays.toString(dummyPorts) +
                ", nodeListenAddress='" + nodeListenAddress + '\'' +
                '}';
    }
}

class FakeMetricsServer extends NanoHTTPD {
    private static final Logger LOGGER = LoggerFactory.getLogger(FakeMetricsServer.class);
    private static final int PORT = 7070;

    FakeMetricsServer() {
        super(PORT);
    }

    @Override
    public void start() throws IOException {
        LOGGER.info("Starting fake metrics server on port {}", PORT);
        super.start();
    }

    @Override
    public Response serve(final IHTTPSession session) {
        LOGGER.debug("serve: {} {}", session.getMethod(), session.getUri());
        return newFixedLengthResponse("cassandra_clientrequest_write_latency_count");
    }
}

class FakeJolokiaServer extends NanoHTTPD {
    private static final Logger LOGGER = LoggerFactory.getLogger(FakeJolokiaServer.class);
    private static final int PORT = 7777;

    private final FakeJolokiaServerConfig config;

    FakeJolokiaServer(FakeJolokiaServerConfig config) {
        super(PORT);
        this.config = config;
    }

    @Override
    public void start() throws IOException {
        LOGGER.info("Starting fake Jolokia server on port {} with config: {}", PORT, config);
        super.start();
    }

    @Override
    public Response serve(final IHTTPSession session) {
        LOGGER.info("serve: {} {}", session.getMethod(), session.getUri());
        if (session.getMethod() == Method.POST) {
            return newFixedLengthResponse(Response.Status.FORBIDDEN, "application/text", "HTTP method post is not allowed according to the installed security policy\",\"{\"status\":403}");
        }
        return newFixedLengthResponse(this.config.permittedPaths.getOrDefault(session.getUri(), "{\"status\":403}"));
    }
}
