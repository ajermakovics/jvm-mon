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
        final Runtime rt = Runtime.getRuntime();
        long usedMem = (rt.totalMemory() - rt.freeMemory())/1024/1024;
        long maxMem = rt.maxMemory()/1024/1024;
        double load = getProcessCpuLoad();

        final Metrics metrics = new Metrics();
        metrics.Used = usedMem;
        metrics.Max = maxMem;
        metrics.Load = load;
        metrics.GcUsage = gcMon.getGcUsage();
        metrics.Threads = getThreads(10);

        return metrics;
    }

    public double getProcessCpuLoad() throws Exception {
        final ObjectName osObj = ObjectName.getInstance(OPERATING_SYSTEM_MXBEAN_NAME);
        final AttributeList osAttrs = mbs.getAttributes(osObj, new String[]{"ProcessCpuLoad"});

        if (osAttrs.isEmpty())
            return 0;

        final Attribute att = (Attribute) osAttrs.get(0);
        final Double value =  (Double) att.getValue();

        // usually takes a couple of seconds before we get real values
        if (value == -1.0)
            return 0;
        // returns a percentage value with 1 decimal point precision
        return ((value * 1000) / 10.0);
    }

    public JThreads getThreads(int max) throws Exception {
        final ThreadMXBean mbean = getThreadMXBean();
        final List<JThread> threadList = new ArrayList<>();
        final Set<Long> alive = new HashSet<>();
        boolean cpuTimeOn = mbean.isThreadCpuTimeSupported() && mbean.isThreadCpuTimeEnabled();

        long[] ids = mbean.getAllThreadIds();
        final ThreadInfo[] threadInfos = mbean.getThreadInfo(ids);

        for (ThreadInfo ti : threadInfos) {
            if (ti == null) // thread exited between calls
                continue;
            long tid = ti.getThreadId();
            long cpuTime = cpuTimeOn ? mbean.getThreadCpuTime(tid) : 0;

            JThread thread = threads.computeIfAbsent(tid, id -> new JThread(ti)).update(ti);
            alive.add(tid);
            threadList.add(thread.withCpuTime(cpuTime));
        }

        threads.keySet().retainAll(alive); // drop dead threads

        final List<JThread> jThreads = threadList.stream()
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
