package net.consensys.linea.consensus.app

import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.core.LoggerContext
import org.apache.logging.log4j.core.config.Configurator
import picocli.CommandLine
import kotlin.system.exitProcess

class ConsensusLayerAppMain {
  companion object {
    private val log = LogManager.getLogger(ConsensusLayerAppMain::class)

    @JvmStatic
    fun main(args: Array<String>) {
      val cmd = CommandLine(ConsensusLayerAppCli.withAction(::startApp))
      cmd.setExecutionExceptionHandler { ex, _, _ ->
        log.error("Execution failure: ", ex)
        1
      }
      cmd.setParameterExceptionHandler { ex, _ ->
        log.error("Invalid args!: ", ex)
        1
      }
      exitProcess(cmd.execute(*args))
    }

    private fun startApp(configs: ConsensusLayerAppConfig) {
      val app = ConsensusLayerApp(configs)
      Runtime.getRuntime()
        .addShutdownHook(
          Thread {
            app.stop()
            if (LogManager.getContext() is LoggerContext) {
              // Disable log4j auto shutdown hook is not used otherwise
              // Messages in App.stop won't appear in the logs
              Configurator.shutdown(LogManager.getContext() as LoggerContext)
            }
          }
        )
      app.start()
    }
  }
}
