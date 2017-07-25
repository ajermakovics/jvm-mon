package jvmmon

import com.jvmtop.monitor.VMInfo
import com.jvmtop.monitor.VMInfoState
import com.jvmtop.view.VMDetailView
import com.jvmtop.view.VMOverviewView
import org.andrejs.json.Json

class JvmStats {

    private val vmOverviewView = VMOverviewView(0)
    private var vmDetailView: VMDetailView? = null

    init {
        update()
    }

    fun getVmStats(vmId: Int): List<Json> {
        update(vmId)

        val vmInfos = vmOverviewView.vmInfoList
                .filter { vm -> vm.state == VMInfoState.ATTACHED }

        val vmStats = vmInfos
                .map { toVmJsObject(it) }
                .map { vmJs -> maybeAddThreads(vmJs, vmId, vmDetailView) }

        return vmStats
    }

    private fun update(vmId: Int = 0) {
        vmOverviewView.updateVMs(vmOverviewView.scanForNewVMs())

        if (vmId != 0 && (vmDetailView == null || vmId != vmDetailView!!.vmId)) {
            vmDetailView = VMDetailView(vmOverviewView.getVMInfo(vmId))
        }
    }

    private fun maybeAddThreads(vmJson: Json, vmId: Int, view: VMDetailView?): Json {
        if (view != null && vmJson.get("Id", -1) == vmId) {
                val threads = view.topThreads
                        .map { toThreadJs(it) }

                vmJson.put("threads", threads)
        }
        return vmJson
    }

    private fun toThreadJs(ts: VMDetailView.ThreadStats): Json {
        val threadJs = Json()
                .set("TID", ts.TID)
                .set("name", ts.name)
                .set("state", ts.state.toString())
                .set("cpu", ts.cpu)
                .set("totalCpu", ts.totalCpu)
                .set("blockedBy", ts.blockedBy)
        return threadJs
    }

    private fun toVmJsObject(vm: VMInfo): Json {
        val vmJs = Json()
                .set("Id", vm.id)
                .set("DisplayName", displayName(vm))
                .set("HeapUsed", vm.heapUsed)
                .set("HeapMax", vm.heapMax)
                .set("HeapSize", vm.heapSize)
                .set("NonHeapUsed", vm.nonHeapUsed)
                .set("NonHeapMax", vm.nonHeapMax)
                .set("CpuLoad", vm.cpuLoad)
                .set("GcLoad", vm.gcLoad)
                .set("VMVersion", vm.vmVersion)
                .set("OSUser", vm.osUser)
                .set("ThreadCount", vm.threadCount)
                .set("hasDeadlockThreads", vm.hasDeadlockThreads())
        return vmJs
    }

    private fun displayName(vm: VMInfo): String {
        return vm.displayName.trim().substringBefore(" ")
    }
}
