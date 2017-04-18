package jvmmon;

import org.andrejs.json.Json;
import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

public class JvmStatsTest {

    private JvmStats jvmStats;

    @Before
    public void setUp() {
        jvmStats = new JvmStats();
    }

    @Test
    public void returnsAllVmStats_whenNoVmSelected() throws Exception {
        List<Json> vms = jvmStats.getVmStats(0);

        Assert.assertNotEquals(0, vms.size());
    }

    @Test
    public void returnsVmStatsWithDetails_whenNoVmSelected() throws Exception {
        List<Json> vms = jvmStats.getVmStats(0);
        Json firstVm = vms.get(0);

        assertThat(firstVm.keySet())
                .contains("Id", "DisplayName", "HeapUsed", "HeapMax", "HeapSize",
                        "NonHeapUsed", "NonHeapMax", "CpuLoad", "GcLoad", "VMVersion",
                        "OSUser", "ThreadCount");
    }

    @Test
    public void returnsVmStatsWithThreads_whenVmSelected() throws Exception {
        List<Json> vms = jvmStats.getVmStats(0);
        Json vmStats = vms.get(0);
        int vmId = vmStats.get("Id", 0);

        vmStats = jvmStats.getVmStats(vmId).stream()
                .filter(vm -> vm.get("Id").equals(vmId))
                .findFirst()
                .orElseThrow(AssertionError::new);

        assertThat(vmStats).containsKey("threads");
    }

}