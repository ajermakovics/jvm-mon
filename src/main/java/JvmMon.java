import com.eclipsesource.v8.*;
import com.jvmtop.monitor.VMInfo;
import com.jvmtop.monitor.VMInfoState;
import com.jvmtop.view.VMOverviewView;

import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

public class JvmMon implements JavaCallback {

    static NodeJS nodeJS = NodeJS.createNodeJS(); //nodeJS.require(new File("blessed-contrib"));
    static V8 v8 = nodeJS.getRuntime();

    public static void main(String[] args) throws Exception {
        Path script = Paths.get("jvm-mon.js");
        if(!Files.exists(script)) {
            script = Paths.get("src", "dist", "jvm-mon.js");
        }

        JvmMon jvmMon = new JvmMon();
        jvmMon.update();
        nodeJS.getRuntime().registerJavaMethod(jvmMon, "getData");
        nodeJS.getRuntime().executeScript("var refreshDelay = 1000");

        nodeJS.exec(script.toFile());

        while(nodeJS.isRunning()) {
            nodeJS.handleMessage();
        }

        nodeJS.release();
    }

    private VMOverviewView vmOverviewView = new VMOverviewView(0);

    public void update() throws Exception {
        vmOverviewView.updateVMs(vmOverviewView.scanForNewVMs());
    }

    private V8Array getVmStats() throws Exception {
        V8Array res = new V8Array(v8);

        vmOverviewView.getVMInfoList().stream()
                .filter(vm -> vm.getState() == VMInfoState.ATTACHED)
                .map(this::toJsObject)
                .forEach(res::push);

        return res;
    }

    private V8Object toJsObject(VMInfo vm) {
        V8Object res = new V8Object(v8);
        res.add("Id", vm.getId());
        res.add("DisplayName", displayName(vm));
        res.add("HeapUsed", vm.getHeapUsed());
        res.add("HeapMax", vm.getHeapMax());
        res.add("HeapSize", vm.getHeapSize());
        res.add("NonHeapUsed", vm.getNonHeapUsed());
        res.add("NonHeapMax", vm.getNonHeapMax());
        res.add("CpuLoad", vm.getCpuLoad());
        res.add("GcLoad", vm.getGcLoad());
        res.add("VMVersion", vm.getVMVersion());
        res.add("OSUser", vm.getOSUser());
        res.add("ThreadCount", vm.getThreadCount());
        res.add("hasDeadlockThreads", vm.hasDeadlockThreads());
        return res;
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
            update();
            return getVmStats();
        } catch (Exception e) {
            e.printStackTrace();
            return new V8Array(v8);
        }
    }
}
