package com.sky.core.operators.fake;

import fi.iki.elonen.NanoHTTPD;
import org.yaml.snakeyaml.Yaml;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.InputStream;
import java.io.IOException;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

class FakeJolokiaServerConfig {
    Map<String, String> PERMITTED_PATHS;

    public FakeJolokiaServerConfig(String nodeListenAddress) {
        PERMITTED_PATHS = new HashMap<>();
        PERMITTED_PATHS.put("/jolokia/exec/org.apache.cassandra.db:type=EndpointSnitchInfo/getRack/" + nodeListenAddress,
                            "{\"status\":200}");
        PERMITTED_PATHS.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/LiveNodes",
                            "{\"status\":200, \"value\":[\"" + nodeListenAddress+ "\"]}");
        PERMITTED_PATHS.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/UnreachableNodes",
                            "{\"status\":200}");
        PERMITTED_PATHS.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/JoiningNodes",
                            "{\"status\":200}");
        PERMITTED_PATHS.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/LeavingNodes",
                            "{\"status\":200}");
        PERMITTED_PATHS.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/MovingNodes",
                            "{\"status\":200}");
     }

    public String toString() {
        return String.format("PERMITTED_PATHS: %s", PERMITTED_PATHS);
    }

}

class FakeCassandraConfig {
    private static final Logger logger = LoggerFactory.getLogger(FakeCassandraConfig.class);

    static final int[] DUMMY_PORTS = new int[]{7000, 7199, 9042};
    static final String CASSANDRA_YAML_PATH = "/etc/cassandra/cassandra.yaml";
    String NODE_LISTEN_ADDRESS = "localhost";

    public static final FakeCassandraConfig Load() throws FileNotFoundException {
        FakeCassandraConfig config = new FakeCassandraConfig();
        config.NODE_LISTEN_ADDRESS = readNodeListenAddressFromCassandraYaml(config.NODE_LISTEN_ADDRESS);
        return config;
    }

    private static final String readNodeListenAddressFromCassandraYaml(String defaultNodeListenAddress) throws FileNotFoundException {
        InputStream input = new FileInputStream(new File(CASSANDRA_YAML_PATH));
        Yaml yaml = new Yaml();
        Map<String, Object> obj = (Map<String, Object>) yaml.load(input);
        logger.info("{}: {}", CASSANDRA_YAML_PATH, obj);
        String nodeListenAddress = (String) obj.get("listen_address");
        if (nodeListenAddress == null) {
            logger.warn("{}:listen_address:null. Using default instead.", CASSANDRA_YAML_PATH, obj);
            nodeListenAddress = defaultNodeListenAddress;
        }
        return nodeListenAddress;
    }

    public String toString() {
        return String.format("DUMMY_PORTS: %s, CASSANDRA_YAML_PATH: %s, NODE_LISTEN_ADDRESS: %s",
                             Arrays.toString(DUMMY_PORTS),
                             CASSANDRA_YAML_PATH,
                             NODE_LISTEN_ADDRESS);
    }
}

public class FakeCassandra {
    private static final Logger logger = LoggerFactory.getLogger(FakeCassandra.class);

    private FakeCassandraConfig config;
    private FakeJolokiaServer jolokiaServer;
    private FakeMetricsServer metricsServer;

    public static void main(String[] args) {
        FakeCassandraConfig cassandraConfig;
        try {
            cassandraConfig = FakeCassandraConfig.Load();
        } catch (FileNotFoundException e) {
            logger.error("failed to load configuration", e);
            return;
        }
        logger.info("FakeCassandraConfig: {}", cassandraConfig);

        final FakeJolokiaServerConfig jolokiaConfig = new FakeJolokiaServerConfig(cassandraConfig.NODE_LISTEN_ADDRESS);
        logger.info("FakeJolokiaServerConfig: {}", jolokiaConfig);

        final FakeCassandra fc = new FakeCassandra(cassandraConfig,
                                                   new FakeJolokiaServer(jolokiaConfig),
                                                   new FakeMetricsServer());
        try {
            fc.run();
        } catch (IOException e) {
            logger.error("FakeCassandra startup failed", e);
        }
    }

    public FakeCassandra(FakeCassandraConfig config,
                         FakeJolokiaServer jolokiaServer,
                         FakeMetricsServer metricsServer) {
        this.config = config;
        this.jolokiaServer = jolokiaServer;
        this.metricsServer = metricsServer;
    }

    public void run() throws IOException {
        final ExecutorService exec = Executors.newFixedThreadPool(this.config.DUMMY_PORTS.length);
        final ServerSocket[] serverSockets = new ServerSocket[this.config.DUMMY_PORTS.length];
        for (int i = 0; i < this.config.DUMMY_PORTS.length; i++) {
            serverSockets[i] = createSocket(this.config.DUMMY_PORTS[i]);
        }

        Arrays.stream(serverSockets).forEach(s -> exec.execute(() -> startListening(s)));

        logger.info("Starting fake Jolokia server");
        this.jolokiaServer.start();

        logger.info("Starting fake metrics server");
        this.metricsServer.start();

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            logger.info("Shutting down");
            Arrays.stream(serverSockets).forEach(this::closeQuietly);
            this.jolokiaServer.stop();
            this.metricsServer.stop();
            exec.shutdown();
            try {
                exec.awaitTermination(10, TimeUnit.SECONDS);
            } catch (InterruptedException e) {
                logger.error("Failed while waiting for executor termination", e);
            }
        }));
    }

    private void closeQuietly(final ServerSocket serverSocket) {
        try {
            serverSocket.close();
        } catch (IOException ex) {
            logger.error(String.format("Error while closing server socket on port %d", serverSocket.getLocalPort()), ex);
        }
    }

    private ServerSocket createSocket(final int port) throws IOException {
        return new ServerSocket(port);
    }

    private void startListening(final ServerSocket serverSocket) {
        logger.info("Listening on port {}", serverSocket.getLocalPort());
        while (!serverSocket.isClosed()) {
            try {
                final Socket client = serverSocket.accept();
                logger.info("Accepted connection from {} on port {}", client.getInetAddress().toString(), serverSocket.getLocalPort());
                client.close();
            } catch (IOException ex) {
                logger.error(String.format("Error while handling client connection on port %d", serverSocket.getLocalPort()), ex);
            }
        }
    }
}

class FakeMetricsServer extends NanoHTTPD {
    private static final Logger logger = LoggerFactory.getLogger(FakeMetricsServer.class);

    public FakeMetricsServer() {
        super(7070);
    }

    @Override
    public Response serve(final IHTTPSession session) {
        logger.info("serve: {} {}", session.getMethod(), session.getUri());
        return newFixedLengthResponse("cassandra_clientrequest_write_latency_count");
    }
}

class FakeJolokiaServer extends NanoHTTPD {
    private static final Logger logger = LoggerFactory.getLogger(FakeJolokiaServer.class);

    private FakeJolokiaServerConfig config;

    public FakeJolokiaServer(FakeJolokiaServerConfig config) {
        super(7777);
        this.config = config;
    }

    @Override
    public Response serve(final IHTTPSession session) {
        logger.info("serve: {} {}", session.getMethod(), session.getUri());
        if (session.getMethod() == Method.POST) {
            return newFixedLengthResponse(Response.Status.FORBIDDEN, "application/text", "HTTP method post is not allowed according to the installed security policy\",\"{\"status\":403}");
        } else if(this.config.PERMITTED_PATHS.containsKey(session.getUri())) {
            return newFixedLengthResponse(this.config.PERMITTED_PATHS.get(session.getUri()));
        } else {
            return newFixedLengthResponse("{\"status\":403}");
        }
    }
}
