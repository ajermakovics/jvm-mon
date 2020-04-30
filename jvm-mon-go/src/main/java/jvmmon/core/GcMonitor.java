package jvmmon.core;

import javax.management.AttributeNotFoundException;
import javax.management.MBeanServerConnection;
import javax.management.MalformedObjectNameException;
import javax.management.ObjectName;
import java.io.IOError;
import java.lang.management.GarbageCollectorMXBean;
import java.lang.management.ManagementFactory;
import java.lang.management.OperatingSystemMXBean;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;

public class GcMonitor {

    final MBeanServerConnection con;
    final OperatingSystemMXBean os;
    final int procCount;
    final long cpuTimeMultiplier;
    List<GarbageCollectorMXBean> gcMbeans;
    long gcTime;
    long prevGcTime;
    long upTime;
    long prevUpTime;

    public GcMonitor(MBeanServerConnection con) {
        this.con = con;
        try {
            this.os = ManagementFactory.newPlatformMXBeanProxy(con,
                    ManagementFactory.OPERATING_SYSTEM_MXBEAN_NAME, OperatingSystemMXBean.class);
            procCount = os.getAvailableProcessors();
            cpuTimeMultiplier = getCpuMultiplier(con);
        } catch (Exception e) {
            throw new IOError(e);
        }
    }

    public long getGcUsage() {
        try {
            prevUpTime = upTime;
            prevGcTime = gcTime;

            gcMbeans = getGarbageCollectorMXBeans();
            gcTime = -1;
            for (GarbageCollectorMXBean gcBean : gcMbeans)
                gcTime += gcBean.getCollectionTime();

            long processGcTime = gcTime * 1000000 / procCount;
            long prevProcessGcTime = prevGcTime * 1000000 / procCount;
            long processGcTimeDiff = processGcTime - prevProcessGcTime;

            Long jmxUpTime = (Long) con.getAttribute(getRuntimeName(), "Uptime");
            upTime = jmxUpTime;
            long upTimeDiff = (upTime * 1000000) - (prevUpTime * 1000000);

            long gcUsage = upTimeDiff > 0 ? Math.min((long)
                    (1000 * (float) processGcTimeDiff / (float) upTimeDiff), 1000) : 0;

            return gcUsage;

        } catch (Exception e) {
            e.printStackTrace();
            return 0;
        }
    }

    private List<GarbageCollectorMXBean> getGarbageCollectorMXBeans() throws Exception {
        List<GarbageCollectorMXBean> gcMbeans = null;
        if (con != null) {
            ObjectName gcName = new ObjectName(ManagementFactory.GARBAGE_COLLECTOR_MXBEAN_DOMAIN_TYPE + ",*");
            Set<ObjectName> mbeans = con.queryNames(gcName, null);
            if (mbeans != null) {
                gcMbeans = new ArrayList<GarbageCollectorMXBean>();
                for (ObjectName on : mbeans) {
                    String name = ManagementFactory.GARBAGE_COLLECTOR_MXBEAN_DOMAIN_TYPE + ",name=" + on.getKeyProperty("name");
                    GarbageCollectorMXBean mbean = ManagementFactory.newPlatformMXBeanProxy(con, name, GarbageCollectorMXBean.class);
                    gcMbeans.add(mbean);
                }
            }
        }
        return gcMbeans;
    }

    public static long getCpuMultiplier(MBeanServerConnection con) throws Exception {
        Number num;
        try {
            num = (Number) con.getAttribute(getOSName(), "ProcessingCapacity");
        } catch (AttributeNotFoundException e) {
            num = 1;
        }
        return num.longValue();
    }

    private static ObjectName getOSName() {
        try {
            return new ObjectName(ManagementFactory.OPERATING_SYSTEM_MXBEAN_NAME);
        } catch (MalformedObjectNameException ex) {
            throw new RuntimeException(ex);
        }
    }

    private static ObjectName getRuntimeName() {
        try {
            return new ObjectName(ManagementFactory.RUNTIME_MXBEAN_NAME);
        } catch (MalformedObjectNameException ex) {
            throw new RuntimeException(ex);
        }
    }

    public void stop() {
        gcMbeans.clear();
    }
}
