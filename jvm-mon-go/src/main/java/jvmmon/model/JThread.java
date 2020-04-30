package jvmmon.model;

import jvmmon.util.Json;

import java.lang.management.ThreadInfo;

public class JThread implements Jsonable {
    public long Id;
    public String Name;
    public Thread.State State;
    public long CpuTime;
    public long prevCpuTime;

    public JThread(ThreadInfo ti) {
        Id = ti.getThreadId();
        Name = ti.getThreadName();
        State = ti.getThreadState();

        if(Name.length() > 25)
            Name = Name.substring(0, 24);
    }

    public JThread withCpuTime(long cpuTime) {
        if(prevCpuTime != 0) {
            CpuTime = cpuTime - prevCpuTime;
            prevCpuTime = cpuTime;
        }
        prevCpuTime = cpuTime;
        return this;
    }

    @Override
    public String toJson() {
        return Json.toJson("Id", Id, "Name",
                Name, "State", State,
                "CpuTime", CpuTime);
    }
}
