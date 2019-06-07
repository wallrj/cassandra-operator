package com.sky.core.operators.fake;

import fi.iki.elonen.NanoHTTPD;

import java.io.IOException;
import java.io.PrintWriter;
import java.io.StringWriter;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

public class FakeCassandra {
    private static final int[] DUMMY_PORTS = new int[]{7000, 7199, 9042};
    public static String nodeListenAddress;

    public static void main(String[] args) {
        nodeListenAddress = System.getenv("NODE_LISTEN_ADDRESS");
        if (nodeListenAddress == null) {
            nodeListenAddress = "127.0.0.1";
        }

        final FakeCassandra fc = new FakeCassandra();
        try {
            fc.run();
        } catch (IOException e) {
            printError("FakeCassandra startup failed", e);
        }
    }

    public void run() throws IOException {
        final ExecutorService exec = Executors.newFixedThreadPool(DUMMY_PORTS.length);
        final ServerSocket[] serverSockets = new ServerSocket[DUMMY_PORTS.length];
        for (int i = 0; i < DUMMY_PORTS.length; i++) {
            serverSockets[i] = createSocket(DUMMY_PORTS[i]);
        }

        Arrays.stream(serverSockets).forEach(s -> exec.execute(() -> startListening(s)));

        final FakeJolokiaServer fakeJolokiaServer = new FakeJolokiaServer();
        System.out.println("Starting fake Jolokia server");
        fakeJolokiaServer.start();

        final FakeMetricsServer fakeMetricsServer = new FakeMetricsServer();
        System.out.println("Starting fake metrics server");
        fakeMetricsServer.start();

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            System.out.println("Shutting down");
            Arrays.stream(serverSockets).forEach(this::closeQuietly);
            fakeJolokiaServer.stop();
            fakeMetricsServer.stop();
            exec.shutdown();
            try {
                exec.awaitTermination(10, TimeUnit.SECONDS);
            } catch (InterruptedException e) {
                printError("Failed while waiting for executor termination", e);
            }
        }));
    }

    private void closeQuietly(final ServerSocket serverSocket) {
        try {
            serverSocket.close();
        } catch (IOException ex) {
            printError(String.format("Error while closing server socket on port %d", serverSocket.getLocalPort()), ex);
        }
    }

    private ServerSocket createSocket(final int port) throws IOException {
        return new ServerSocket(port);
    }

    private void startListening(final ServerSocket serverSocket) {
        System.out.printf("Listening on port %d\n", serverSocket.getLocalPort());
        while (!serverSocket.isClosed()) {
            try {
                final Socket client = serverSocket.accept();
                System.out.printf("Accepted connection from %s on port %d\n", client.getInetAddress().toString(), serverSocket.getLocalPort());
                client.close();
            } catch (IOException ex) {
                printError(String.format("Error while handling client connection on port %d", serverSocket.getLocalPort()), ex);
            }
        }
    }

    private static void printError(String message, Throwable t) {
        StringWriter stringWriter = new StringWriter();
        try(PrintWriter printWriter = new PrintWriter(stringWriter)) {
            printWriter.write(message);
            printWriter.write(": ");
            t.printStackTrace(printWriter);
        }
        System.err.print(stringWriter.toString());
    }
}

class FakeMetricsServer extends NanoHTTPD {
    public FakeMetricsServer() {
        super(7070);
    }

    @Override
    public Response serve(final IHTTPSession session) {
        return newFixedLengthResponse("cassandra_clientrequest_write_latency_count");
    }
}

class FakeJolokiaServer extends NanoHTTPD {
    private static Map<String, String> PERMITTED_PATHS;
    static {
        PERMITTED_PATHS = new HashMap<>();
        PERMITTED_PATHS.put("/jolokia/exec/org.apache.cassandra.db:type=EndpointSnitchInfo/getRack/" + FakeCassandra.nodeListenAddress,
                            "{\"status\":200}");
        PERMITTED_PATHS.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/LiveNodes",
                            "{\"status\":200, \"value\":[\"" + FakeCassandra.nodeListenAddress+ "\"]}");
        PERMITTED_PATHS.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/UnreachableNodes",
                            "{\"status\":200}");
        PERMITTED_PATHS.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/JoiningNodes",
                            "{\"status\":200}");
        PERMITTED_PATHS.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/LeavingNodes",
                            "{\"status\":200}");
        PERMITTED_PATHS.put("/jolokia/read/org.apache.cassandra.db:type=StorageService/MovingNodes",
                            "{\"status\":200}");
    }
    public FakeJolokiaServer() {
        super(7777);
    }

    @Override
    public Response serve(final IHTTPSession session) {
        if (session.getMethod() == Method.POST) {
            return newFixedLengthResponse(Response.Status.FORBIDDEN, "application/text", "HTTP method post is not allowed according to the installed security policy\",\"status\":403");
        } else if(PERMITTED_PATHS.containsKey(session.getUri())) {
            return newFixedLengthResponse(PERMITTED_PATHS.get(session.getUri()));
        } else {
            return newFixedLengthResponse("status\":403");
        }
    }
}
