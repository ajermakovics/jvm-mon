package jvmmon.core;

import jvmmon.model.JThread;
import jvmmon.model.JThreads;
import jvmmon.model.Metrics;

import javax.management.Attribute;
import javax.management.AttributeList;
import javax.management.MBeanServer;
import javax.management.ObjectName;
import java.lang.management.ManagementFactory;
import java.lang.management.ThreadInfo;
import java.lang.management.ThreadMXBean;
import java.util.*;
import java.util.stream.Collectors;

import static java.lang.management.ManagementFactory.OPERATING_SYSTEM_MXBEAN_NAME;
import static java.lang.management.ManagementFactory.getThreadMXBean;

public class JvmMon {

    final MBeanServer mbs = ManagementFactory.getPlatformMBeanServer();
    final GcMonitor gcMon = new GcMonitor(mbs);
    final Map<Long, JThread> threads = new HashMap<>();

    public String getMetricsJson() throws Exception {
        return getMetrics().toJson();
    }

    public Metrics getMetrics() throws Exception {
        Runtime rt = Runtime.getRuntime();
        long usedMem = (rt.totalMemory() - rt.freeMemory())/1024/1024;
        long maxMem = rt.maxMemory()/1024/1024;
        double load = getProcessCpuLoad();

        Metrics metrics = new Metrics();
        metrics.Used = usedMem;
        metrics.Max = maxMem;
        metrics.Load = load;
        metrics.GcUsage = gcMon.getGcUsage();
        metrics.Threads = getThreads(10);

        return metrics;
    }

    public double getProcessCpuLoad() throws Exception {
        ObjectName osObj = ObjectName.getInstance(OPERATING_SYSTEM_MXBEAN_NAME);
        AttributeList osAttrs = mbs.getAttributes(osObj, new String[]{"ProcessCpuLoad"});

        if (osAttrs.isEmpty())
            return 0;

        Attribute att = (Attribute) osAttrs.get(0);
        Double value =  (Double) att.getValue();

        // usually takes a couple of seconds before we get real values
        if (value == -1.0)
            return 0;
        // returns a percentage value with 1 decimal point precision
        return ((value * 1000) / 10.0);
    }

    public JThreads getThreads(int max) throws Exception {
        ThreadMXBean mbean = getThreadMXBean();
        List<JThread> threadList = new ArrayList<>();

        long[] ids = mbean.getAllThreadIds();
        List<ThreadInfo> threadInfos = Arrays.asList(mbean.getThreadInfo(ids));

        for(ThreadInfo ti: threadInfos) {
            long cpuTime = 0;
            long tid = ti.getThreadId();
            if(mbean.isThreadCpuTimeSupported() && mbean.isThreadCpuTimeEnabled())
                cpuTime = mbean.getThreadCpuTime(tid);

            JThread thread = threads.compute(tid, (id, tCur) -> tCur == null ? new JThread(ti) : tCur);

            threadList.add(thread.withCpuTime(cpuTime));
        }

        List<JThread> jThreads = threads.values().stream()
                .sorted(Comparator.<JThread>comparingLong(t -> t.CpuTime).reversed())
                .limit(max)
                .collect(Collectors.toList());

        return new JThreads(ids.length, jThreads);
    }

    public void stop() {
        threads.clear();
        gcMon.stop();
    }
}
