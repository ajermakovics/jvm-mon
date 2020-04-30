package jvmmon;

import static java.lang.System.out;

public class Main {
    /** Just for development purposes **/
    public static void main(String[] args) throws Exception {

        for(int i = 0; i < 12; i++) {
            new Thread(() -> {
                try {
                    out.println("Started thread " + Thread.currentThread());
                    Thread.sleep(300_000);
                } catch (InterruptedException e) {
                    e.printStackTrace();
                }
                out.println("Finished thread " + Thread.currentThread());
            }, "testThread_" + i + "_with_potentially_long_name").start();
        }

        out.println("Waiting for input => ");
        System.in.read();
    }
}
