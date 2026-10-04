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

        if (Name == null)
            Name = "";
        if (Name.length() > 25)
            Name = Name.substring(0, 25);
    }

    public JThread update(ThreadInfo ti) {
        State = ti.getThreadState();
        return this;
    }

    /** Stores CPU time used since previous sample (0 on first sample or if unsupported) */
    public JThread withCpuTime(long cpuTime) {
        CpuTime = (prevCpuTime > 0 && cpuTime >= prevCpuTime) ? cpuTime - prevCpuTime : 0;
        prevCpuTime = Math.max(cpuTime, 0);
        return this;
    }

    @Override
    public String toJson() {
        return Json.toJson("Id", Id, "Name",
                Name, "State", State,
                "CpuTime", CpuTime);
    }
}
