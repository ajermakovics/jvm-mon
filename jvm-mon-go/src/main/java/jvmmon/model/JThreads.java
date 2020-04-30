package jvmmon.model;

import jvmmon.util.Json;

import java.util.List;

public class JThreads implements Jsonable {

    int Count;
    List<JThread> Threads;

    public JThreads(int count, List<JThread> threads) {
        Count = count;
        Threads = threads;
    }

    @Override
    public String toJson() {
        return Json.toJson("Count", Count, "Threads", Threads);
    }
}
