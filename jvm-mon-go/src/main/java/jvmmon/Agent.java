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

        Thread thread = new Thread(new Agent(port), Agent.class.getName());
        thread.setDaemon(true);
        thread.start();
    }

    @Override
    public void run() {
        SocketWriter socketWriter = new SocketWriter(port, jvmMon::getMetricsJson);
        socketWriter.run();
    }

    /**
     * For development. Starts in this process and sends metrics to a running jvm-mon.
     * Usage: java -jar jvm-mon-go.jar [port | path/to/jvm-mon.log]
     * Defaults to the log jvm-mon writes in the user cache dir.
     */
    public static void main(String[] args) throws Exception {
        String arg = args.length > 0 ? args[0] : defaultLogPath();
        String port;
        if (arg.matches("[0-9]{1,5}")) {
            port = arg;
        } else {
            List<String> log = Files.readAllLines(Paths.get(arg));
            Optional<String> found = log.stream().filter(line -> line.contains("port:"))
                    .flatMap(line -> Arrays.stream(line.split(":")))
                    .map(String::trim)
                    .filter(p -> p.matches("[0-9]{1,5}"))
                    .reduce((a, b) -> b); // last started server
            if (!found.isPresent())
                throw new IllegalArgumentException("No server port found in " + arg);
            port = found.get();
        }
        out.println("Server port: " + port);
        debug = true;
        agentmain(port, null);
        Thread.currentThread().join(); // agent thread is a daemon; keep dev process alive
    }

    /** Mirrors Go's os.UserCacheDir() + "/jvm-mon/jvm-mon.log" */
    static String defaultLogPath() {
        String home = System.getProperty("user.home");
        String os = System.getProperty("os.name", "").toLowerCase();
        String cache;
        if (os.contains("mac")) {
            cache = Paths.get(home, "Library", "Caches").toString();
        } else {
            String xdg = System.getenv("XDG_CACHE_HOME");
            cache = (xdg != null && !xdg.isEmpty()) ? xdg : Paths.get(home, ".cache").toString();
        }
        return Paths.get(cache, "jvm-mon", "jvm-mon.log").toString();
    }

    private static void println(String msg) {
        if(debug) out.println(msg);
    }
}