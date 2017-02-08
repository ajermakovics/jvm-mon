import com.eclipsesource.v8.*;
import com.jvmtop.monitor.VMInfo;
import com.jvmtop.monitor.VMInfoState;
import com.jvmtop.view.VMDetailView;
import com.jvmtop.view.VMOverviewView;

import java.io.IOError;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;
import java.util.stream.Collectors;

import static com.eclipsesource.v8.utils.V8ObjectUtils.toV8Array;

public class JvmMon implements JavaCallback {

    static NodeJS nodeJS = NodeJS.createNodeJS();
    static V8 v8 = nodeJS.getRuntime();

    public static void main(String[] args) throws Exception {
        Path script = Paths.get("jvm-mon.js");
        if(!Files.exists(script)) {
            script = Paths.get("src", "dist", "jvm-mon.js");
        }

        JvmMon jvmMon = new JvmMon();

        nodeJS.getRuntime().registerJavaMethod(jvmMon, "getData");
        nodeJS.getRuntime().executeScript("var refreshDelay = 1000");

        nodeJS.exec(script.toFile());

        while(nodeJS.isRunning()) {
            nodeJS.handleMessage();
        }

        nodeJS.release();
    }

    private VMOverviewView vmOverviewView = new VMOverviewView(0);
    private VMDetailView vmDetailView;

    JvmMon() throws Exception {
        update(0);
        vmDetailView = new VMDetailView(vmOverviewView.getVMInfoList().get(0));
    }

    public void update(int vmId) throws Exception {
        vmOverviewView.updateVMs(vmOverviewView.scanForNewVMs());

        if(vmId != 0 && vmId != vmDetailView.getVmId())
            vmDetailView = new VMDetailView(vmOverviewView.getVMInfo(vmId));
    }

    private V8Array getVmStats() throws Exception {
        List<V8Object> vmStats = vmOverviewView.getVMInfoList().stream()
                .filter(vm -> vm.getState() == VMInfoState.ATTACHED)
                .map(this::toVmJsObject)
                .map(vmJs -> maybeAddThreads(vmJs, vmDetailView))
                .collect(Collectors.toList());

        return toV8Array(v8, vmStats);
    }

    private V8Object maybeAddThreads(V8Object vmJs, VMDetailView view) {
        if(vmJs.getInteger("Id") == view.getVmId()) {
            try {
                List<V8Object> threads = view.getTopThreads().stream()
                        .map(this::toThreadJs)
                        .collect(Collectors.toList());
                vmJs.add("threads", toV8Array(v8, threads));
            } catch (Exception e) {
                throw new IOError(e);
            }
        }
        return vmJs;
    }

    private V8Object toThreadJs(VMDetailView.ThreadStats ts) {
        V8Object threadJs = new V8Object(v8);
        threadJs.add("TID", ts.TID);
        threadJs.add("name", ts.name);
        threadJs.add("state", ts.state.toString());
        threadJs.add("cpu", ts.cpu);
        threadJs.add("totalCpu", ts.totalCpu);
        threadJs.add("blockedBy", ts.blockedBy);
        return threadJs;
    }

    private V8Object toVmJsObject(VMInfo vm) {
        V8Object vmJs = new V8Object(v8);
        vmJs.add("Id", vm.getId());
        vmJs.add("DisplayName", displayName(vm));
        vmJs.add("HeapUsed", vm.getHeapUsed());
        vmJs.add("HeapMax", vm.getHeapMax());
        vmJs.add("HeapSize", vm.getHeapSize());
        vmJs.add("NonHeapUsed", vm.getNonHeapUsed());
        vmJs.add("NonHeapMax", vm.getNonHeapMax());
        vmJs.add("CpuLoad", vm.getCpuLoad());
        vmJs.add("GcLoad", vm.getGcLoad());
        vmJs.add("VMVersion", vm.getVMVersion());
        vmJs.add("OSUser", vm.getOSUser());
        vmJs.add("ThreadCount", vm.getThreadCount());
        vmJs.add("hasDeadlockThreads", vm.hasDeadlockThreads());
        return vmJs;
    }

    private V8Object addThreadStats(V8Object vm) {
        vm.getInteger("Id");
        return vm;
    }

    private String displayName(VMInfo vm) {
        String name = vm.getDisplayName().trim();
        if(name.contains(" "))
            return name.substring(0, name.indexOf(' '));
        return name;
    }

    @Override
    public Object invoke(V8Object receiver, V8Array parameters) {
        try {
            int vmId = parameters.getInteger(0);
            update(vmId);
            return getVmStats();
        } catch (Exception e) {
            throw new IOError(e);
        }
    }

}
