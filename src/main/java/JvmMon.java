import com.eclipsesource.v8.*;
import com.eclipsesource.v8.utils.V8ObjectUtils;
import org.andrejs.json.Json;

import java.io.IOError;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;

public class JvmMon implements JavaCallback {

    private static NodeJS nodeJS = NodeJS.createNodeJS();
    private static V8 v8 = nodeJS.getRuntime();

    private JvmStats jvmStats = new JvmStats();

    public static void main(String[] args) throws Exception {

        Path script = Paths.get("jvm-mon.js");
        if(!Files.exists(script)) {
            script = Paths.get("src", "dist", "jvm-mon.js");
        }

        JvmMon jvmMon = new JvmMon();

        v8.registerJavaMethod(jvmMon, "getJvmStats");
        nodeJS.exec(script.toFile());

        while(nodeJS.isRunning()) {
            nodeJS.handleMessage();
        }

        nodeJS.release();
    }

    public JvmMon() {
    }

    /** Called from javascript to get new stats **/
    @Override
    public Object invoke(V8Object receiver, V8Array parameters) {
        try {
            int selectedVmId = parameters.getInteger(0);
            List<Json> vmStats = jvmStats.getVmStats(selectedVmId);
            return V8ObjectUtils.toV8Array(v8, vmStats);

        } catch (Exception e) {
            throw new IOError(e);
        }
    }
}
