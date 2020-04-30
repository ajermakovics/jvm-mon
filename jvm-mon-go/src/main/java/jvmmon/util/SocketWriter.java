package jvmmon.util;

import java.io.Closeable;
import java.io.OutputStreamWriter;
import java.net.Socket;
import java.net.SocketException;
import java.time.Duration;
import java.util.concurrent.Callable;

import static java.lang.System.out;
import static java.nio.charset.StandardCharsets.UTF_8;

public class SocketWriter implements Runnable {

    private String host = "127.0.0.1";
    private int port;
    private Duration sampleInterval = Duration.ofSeconds(1);
    private Callable<String> messageSupplier;

    public SocketWriter(int port, Callable<String> messageSupplier) {
        this.port = port;
        this.messageSupplier = messageSupplier;
    }

    @Override
    public void run() {
        Socket socket = null;
        OutputStreamWriter osw;

        try {
            socket = new Socket(host, port);
            osw = new OutputStreamWriter(socket.getOutputStream(), UTF_8);

            while(socket.isConnected()) {
                String message;
                try {
                    message = messageSupplier.call();
                } catch (Exception e) {
                    System.err.println("Error getting message: " + e.getClass() + " " + e.getMessage());
                    e.printStackTrace();
                    Thread.sleep(sampleInterval.toMillis());
                    continue;
                }
                osw.write(message + "\n");
                osw.flush();
                Thread.sleep(sampleInterval.toMillis());
            }

        } catch(SocketException socketEx) {
            out.println("Disconnected. " + socketEx.getMessage());

        } catch(Exception ex) {
            System.err.println("Socket writer error: " + ex.getClass() + " - " + ex.getMessage());
        } finally {
            close(socket);
        }
    }

    static void close(Closeable cl) {
        try {
            cl.close();
        } catch (Exception e) {
            System.err.println("Socket close error: " + e.getMessage());
        }
    }

}
