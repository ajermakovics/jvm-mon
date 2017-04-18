import com.jvmtop.monitor.VMInfo;
import com.jvmtop.monitor.VMInfoState;
import com.jvmtop.view.VMDetailView;
import com.jvmtop.view.VMOverviewView;
import org.andrejs.json.Json;

import java.io.IOError;
import java.util.List;
import java.util.stream.Collectors;

public class JvmStats {

    private VMOverviewView vmOverviewView = new VMOverviewView(0);
    private VMDetailView vmDetailView;

    public JvmStats() {
        update(0);
    }

    List<Json> getVmStats(int vmId) throws Exception {
        update(vmId);

        List<VMInfo> vmInfos = vmOverviewView.getVMInfoList().stream()
                .filter(vm -> vm.getState() == VMInfoState.ATTACHED)
                .collect(Collectors.toList());

        List<Json> vmStats = vmInfos.stream()
                .map(JvmStats::toVmJsObject)
                .map(vmJs -> JvmStats.maybeAddThreads(vmJs, vmId, vmDetailView))
                .collect(Collectors.toList());

        return vmStats;
    }

    private void update(int vmId) {
        try {
            vmOverviewView.updateVMs(vmOverviewView.scanForNewVMs());
        } catch (Exception e) {
            throw new IOError(e);
        }

        if(vmId != 0 && (vmDetailView == null || vmId != vmDetailView.getVmId())) {
            vmDetailView = new VMDetailView(vmOverviewView.getVMInfo(vmId));
        }
    }

    private static Json maybeAddThreads(Json vmJson, int vmId, VMDetailView view) {
        if(vmJson.get("Id", -1) == vmId) {
            try {
                List<Json> threads = view.getTopThreads().stream()
                        .map(JvmStats::toThreadJs)
                        .collect(Collectors.toList());
                vmJson.put("threads", threads);
            } catch (Exception e) {
                throw new IOError(e);
            }
        }
        return vmJson;
    }

    private static Json toThreadJs(VMDetailView.ThreadStats ts) {
        Json threadJs = new Json()
            .set("TID", ts.TID)
            .set("name", ts.name)
            .set("state", ts.state.toString())
            .set("cpu", ts.cpu)
            .set("totalCpu", ts.totalCpu)
            .set("blockedBy", ts.blockedBy);
        return threadJs;
    }

    private static Json toVmJsObject(VMInfo vm) {
        Json vmJs = new Json()
            .set("Id", vm.getId())
            .set("DisplayName", displayName(vm))
            .set("HeapUsed", vm.getHeapUsed())
            .set("HeapMax", vm.getHeapMax())
            .set("HeapSize", vm.getHeapSize())
            .set("NonHeapUsed", vm.getNonHeapUsed())
            .set("NonHeapMax", vm.getNonHeapMax())
            .set("CpuLoad", vm.getCpuLoad())
            .set("GcLoad", vm.getGcLoad())
            .set("VMVersion", vm.getVMVersion())
            .set("OSUser", vm.getOSUser())
            .set("ThreadCount", vm.getThreadCount())
            .set("hasDeadlockThreads", vm.hasDeadlockThreads());
        return vmJs;
    }

    private static String displayName(VMInfo vm) {
        String name = vm.getDisplayName().trim();
        if(name.contains(" ")) {
            return name.substring(0, name.indexOf(' '));
        }
        return name;
    }
}
