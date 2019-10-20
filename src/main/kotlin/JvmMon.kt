import com.eclipsesource.v8.JavaCallback
import com.eclipsesource.v8.NodeJS
import com.eclipsesource.v8.V8Array
import com.eclipsesource.v8.V8Object
import com.eclipsesource.v8.utils.V8ObjectUtils
import jvmmon.JvmStats
import java.io.File
import java.io.File.pathSeparator

class JvmMon : JavaCallback {

    private val jvmStats = JvmStats()

    /** Called from node to get new stats */
    override fun invoke(receiver: V8Object, parameters: V8Array): Any {
        val selectedVmId = parameters.getInteger(0)
        val vmStats = jvmStats.getVmStats(selectedVmId)

        return V8ObjectUtils.toV8Array(v8, vmStats)
    }

    companion object {

        private val nodeJS = NodeJS.createNodeJS()
        private val v8 = nodeJS.runtime

        @JvmStatic fun main(args: Array<String>) {
            v8.registerJavaMethod(JvmMon(), "getJvmStats")

            runNodeScript(findScript())
        }

        private fun runNodeScript(script: File) {
            nodeJS.exec(script)

            while (nodeJS.isRunning) {
                nodeJS.handleMessage()
            }

            nodeJS.release()
        }

        private fun findScript(): File {
            val script = File("jvm-mon.js")

            return if (script.exists())
                return script
            else {
                File("src${pathSeparator}dist", "jvm-mon.js")
            }

        }
    }
}
