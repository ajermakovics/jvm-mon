package jvmmon;

import jvmmon.core.JvmMon;
import jvmmon.util.SocketWriter;

import java.lang.instrument.Instrumentation;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;

import static java.lang.System.out;

public class Agent implements Runnable {

    private final int port;
    private final JvmMon jvmMon;
    private static boolean debug = false;

    public Agent(int port) {
        this.port = port;
        this.jvmMon = new JvmMon();
    }

    public static void agentmain(String args, Instrumentation instrumentation) throws Exception {
        println("Loaded jvm-mon agent. Args: " + args);

        int port = Integer.valueOf(args);

        new Thread(new Agent(port), Agent.class.getName()).start();
    }

    @Override
    public void run() {
        SocketWriter socketWriter = new SocketWriter(port, jvmMon::getMetricsJson);
        socketWriter.run();
    }

    /** For development. Starts in this process and sends metrics to running jvm-mon */
    public static void main(String[] args) throws Exception {
        List<String> log = Files.readAllLines(Paths.get("log"));
        Optional<String> port = log.stream().filter(line -> line.contains("port:"))
                .flatMap(line -> Arrays.stream(line.split(":")))
                .map(String::trim)
                .filter(p -> p.matches("[0-9]{4,5}"))
                .findFirst();
        out.println("Server port: " + port);
        debug = true;
        agentmain(port.get(), null);
    }

    private static void println(String msg) {
        if(debug) out.println(msg);
    }
}