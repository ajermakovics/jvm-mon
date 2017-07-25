package jvmmon

import org.assertj.core.api.Assertions.assertThat
import org.junit.Test

class JvmStatsTest {

    private val jvmStats = JvmStats()

    @Test
    fun returnsAllVmStats_whenNoVmSelected() {
        val vms = jvmStats.getVmStats(0)

        assertThat(vms).isNotEmpty
    }

    @Test
    fun returnsVmStatsWithDetails_whenNoVmSelected() {
        val vms = jvmStats.getVmStats(0)
        val firstVm = vms[0]

        assertThat(firstVm.keys)
                .contains("Id", "DisplayName", "HeapUsed", "HeapMax", "HeapSize",
                        "NonHeapUsed", "NonHeapMax", "CpuLoad", "GcLoad", "VMVersion",
                        "OSUser", "ThreadCount")
    }

    @Test
    fun returnsVmStatsWithThreads_whenVmSelected() {
        val vms = jvmStats.getVmStats(0)
        val vmStats = vms[0]
        val vmId = vmStats.get("Id", 0)

        val stats = jvmStats.getVmStats(vmId)
                .find { vm -> vm.get<Any>("Id") == vmId }

        assertThat(stats).isNotNull.containsKey("threads")
    }

}