package net.consensys.zkevm.coordinator.app

import net.consensys.zkevm.coordinator.app.config.CoordinatorConfig
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

  @CommandLine.Option(
    names = ["--traces-limits-v2"],
    paramLabel = "<FILE>",
    description = ["Prover traces limits for linea besu"],
    arity = "1"
  )
  private val tracesLimitsV2File: File? = null

  @CommandLine.Option(
    names = ["--smart-contract-errors"],
    paramLabel = "<FILE>",
    description = ["Smart contract error codes"],
    arity = "1"
  )

  private val smartContractErrorsFile: File? = null

  @CommandLine.Option(
    names = ["--gas-price-cap-time-of-day-multipliers"],
    paramLabel = "<FILE>",
    description = ["Time-of-day multipliers for calculation of L1 dynamic gas price caps"],
    arity = "1"
  )

  private val gasPriceCapTimeOfDayMultipliersFile: File? = null

  @CommandLine.Option(
    names = ["--check-configs-only"],
    paramLabel = "<BOOLEAN>",
    description = ["Validates configuration files only, without starting the application."],
    arity = "0..1"
  )

  private var checkConfigsOnly: Boolean = false

  override fun call(): Int {
    return try {
      if (configFiles == null) {
        errorWriter.println("Please provide a configuration file!")
        printUsage(errorWriter)
        return 1
      }
      if (tracesLimitsFile == null && tracesLimitsV2File == null) {
        errorWriter.println("Please provide traces-limits or traces-limits-v2 file!")
        printUsage(errorWriter)
        return 1
      }
      if (smartContractErrorsFile == null) {
        errorWriter.println("Please provide smart-contract-errors file!")
        printUsage(errorWriter)
        return 1
      }
      if (gasPriceCapTimeOfDayMultipliersFile == null) {
        errorWriter.println("Please provide gas-price-cap-time-of-day-multipliers file!")
        printUsage(errorWriter)
        return 1
      }

      for (configFile in configFiles) {
        if (!canReadFile(configFile)) {
          return 1
        }
      }

      val configs = linea.coordinator.config.loadConfigs(
        coordinatorConfigFiles = configFiles.map { it.toPath() },
        tracesLimitsFileV1 = tracesLimitsFile?.toPath(),
        tracesLimitsFileV2 = tracesLimitsV2File?.toPath(),
        smartContractErrorsFile = smartContractErrorsFile.toPath(),
        gasPriceCapTimeOfDayMultipliersFile = gasPriceCapTimeOfDayMultipliersFile.toPath(),
        logger = logger
      )

      if (checkConfigsOnly) {
        logger.info("All configs are valid. Final configs: {}", configs)
      } else {
        startAction.start(configs)
      }
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
    logger.fatal(ex.message)
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
  }
}
