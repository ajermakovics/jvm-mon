package jvmmon.util;

import java.io.Closeable;
import java.io.OutputStreamWriter;
import java.net.Socket;
import java.net.SocketException;
import java.time.Duration;
import java.util.concurrent.Callable;

import static java.nio.charset.StandardCharsets.UTF_8;

public class SocketWriter implements Runnable {

    private String host = "127.0.0.1";
    private int port;
    private Duration sampleInterval = Duration.ofSeconds(2);
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

            while (!socket.isClosed() && !Thread.currentThread().isInterrupted()) {
                String message;
                try {
                    message = messageSupplier.call();
                } catch (Exception e) {
                    Thread.sleep(sampleInterval.toMillis());
                    continue;
                }
                osw.write(message + "\n");
                osw.flush();
                Thread.sleep(sampleInterval.toMillis());
            }

        } catch(SocketException socketEx) {
            // jvm-mon disconnected; exit quietly without polluting target app output
        } catch (InterruptedException ex) {
            Thread.currentThread().interrupt();
        } catch (Exception ex) {
            // ignore: never disturb the monitored application
        } finally {
            close(socket);
        }
    }

    static void close(Closeable cl) {
        if (cl == null)
            return;
        try {
            cl.close();
        } catch (Exception ignored) {
        }
    }

}
