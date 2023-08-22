package net.consensys.zkevm.coordinator.app

import com.sksamuel.hoplite.ConfigLoaderBuilder
import com.sksamuel.hoplite.addFileSource
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import picocli.CommandLine
import picocli.CommandLine.Command
import picocli.CommandLine.Parameters
import java.io.File
import java.io.PrintWriter
import java.nio.charset.Charset
import java.util.concurrent.Callable

@Command(
  name = CoordinatorAppCli.COMMAND_NAME,
  showDefaultValues = true,
  abbreviateSynopsis = true,
  description = ["Runs Linea Coordinator"],
  version = ["0.0.1"],
  synopsisHeading = "%n",
  descriptionHeading = "%nDescription:%n%n",
  optionListHeading = "%nOptions:%n",
  footerHeading = "%n"
)
class CoordinatorAppCli
internal constructor(private val errorWriter: PrintWriter, private val startAction: StartAction) :
  Callable<Int> {
  @Parameters(paramLabel = "CONFIG.toml", description = ["Configuration files"])
  private val configFiles: List<File>? = null

  @CommandLine.Option(
    names = ["--traces-limits"],
    paramLabel = "<FILE>",
    description = ["Prover traces limits"],
    arity = "1"
  )
  private val tracesLimitsFile: File? = null

  override fun call(): Int {
    return try {
      if (configFiles == null) {
        errorWriter.println("Please provide a configuration file!")
        printUsage(errorWriter)
        return 1
      }
      if (tracesLimitsFile == null) {
        errorWriter.println("Please provide traces-limits file!")
        printUsage(errorWriter)
        return 1
      }
      for (configFile in configFiles) {
        if (!canReadFile(configFile)) {
          return 1
        }
      }
      val tracesLimitsConfigs =
        loadConfigs<TracesLimitsConfigFile>(listOf(tracesLimitsFile), errorWriter) ?: return 1
      val configs =
        loadConfigs<CoordinatorConfig>(configFiles, errorWriter)
          ?.let { config: CoordinatorConfig ->
            config.copy(
              conflation = config.conflation.copy(_tracesLimits = tracesLimitsConfigs.tracesLimits)
            )
          }
          ?: return 1

      startAction.start(configs)
      0
    } catch (e: Exception) {
      reportUserError(e)
      1
    }
  }

  private fun canReadFile(file: File): Boolean {
    if (!file.canRead()) {
      errorWriter.println("Cannot read configuration file '${file.absolutePath}'")
      return false
    }
    return true
  }

  fun reportUserError(ex: Throwable) {
    logger.fatal(ex.message, ex)
    errorWriter.println(ex.message)
    printUsage(errorWriter)
  }

  private fun printUsage(outputWriter: PrintWriter) {
    outputWriter.println()
    outputWriter.println("To display full help:")
    outputWriter.println(COMMAND_NAME + " --help")
  }

  /**
   * Not using a static field for this log instance because some code in this class executes prior
   * to the logging configuration being applied so it's not always safe to use the logger.
   *
   * Where this is used we also ensure the messages are printed to the error writer so they will be
   * printed even if logging is not yet configured.
   *
   * @return the logger for this class
   */
  private val logger: Logger = LogManager.getLogger()

  fun interface StartAction {
    fun start(configs: CoordinatorConfig)
  }

  companion object {
    const val COMMAND_NAME = "coordinator"
    fun withAction(startAction: StartAction): CoordinatorAppCli {
      val errorWriter = PrintWriter(System.err, true, Charset.defaultCharset())
      return CoordinatorAppCli(errorWriter, startAction)
    }

    inline fun <reified T : Any> loadConfigs(configFiles: List<File>, errorWriter: PrintWriter): T? {
      val confBuilder: ConfigLoaderBuilder = ConfigLoaderBuilder.Companion.empty().addDefaults()
      for (i in configFiles.indices.reversed()) {
        // files must be added in reverse order for overriding
        confBuilder.addFileSource(configFiles[i], false)
      }

      return confBuilder.build().loadConfig<T>(emptyList()).let { config ->
        if (config.isInvalid()) {
          errorWriter.println(config.getInvalidUnsafe().description())
          null
        } else {
          config.getUnsafe()
        }
      }
    }
  }
}
