package jvmmon.model;

import jvmmon.util.Json;

public class Metrics implements Jsonable {

    public long Used;
    public long Max;
    public double Load;
    public long GcUsage;
    public JThreads Threads;

    @Override
    public String toJson() {
        return Json.toJson("Used", Used,
                "Max", Max,
                "Load", Load,
                "GcUsage", GcUsage,
                "Threads", Threads);
    }
}
