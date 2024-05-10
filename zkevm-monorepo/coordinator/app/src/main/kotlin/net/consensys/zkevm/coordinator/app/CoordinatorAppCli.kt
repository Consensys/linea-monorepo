package net.consensys.zkevm.coordinator.app

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.get
import com.github.michaelbull.result.getError
import com.github.michaelbull.result.onFailure
import com.sksamuel.hoplite.ConfigLoaderBuilder
import com.sksamuel.hoplite.addFileSource
import net.consensys.linea.traces.TracingModule
import net.consensys.linea.traces.allModulesAreDefined
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
      if (tracesLimitsFile == null) {
        errorWriter.println("Please provide traces-limits file!")
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

      val configs = validateConfigs(
        tracesLimitsFile,
        smartContractErrorsFile,
        gasPriceCapTimeOfDayMultipliersFile,
        configFiles
      ) ?: return 1

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
    logger.fatal(ex.message, ex)
    errorWriter.println(ex.message)
    printUsage(errorWriter)
  }

  private fun printUsage(outputWriter: PrintWriter) {
    outputWriter.println()
    outputWriter.println("To display full help:")
    outputWriter.println(COMMAND_NAME + " --help")
  }

  private fun validateConfigs(
    tracesLimitsFile: File,
    smartContractErrorsFile: File,
    gasPriceCapTimeOfDayMultipliersFile: File,
    coordinatorConfigFiles: List<File>
  ): CoordinatorConfig? {
    var hasConfigError = false
    val tracesLimitsConfigs =
      loadConfigsOrError<TracesLimitsConfigFile>(listOf(tracesLimitsFile))

    val smartContractErrorCodes =
      loadConfigsOrError<SmartContractErrorCodesConfig>(listOf(smartContractErrorsFile))

    val gasPriceCapTimeOfDayMultipliers =
      loadConfigsOrError<GasPriceCapTimeOfDayMultipliersConfig>(listOf(gasPriceCapTimeOfDayMultipliersFile))

    val configs = loadConfigsOrError<CoordinatorConfig>(coordinatorConfigFiles)

    if (tracesLimitsConfigs is Err) {
      hasConfigError = true
      logger.error("Reading {} failed: {}", tracesLimitsFile, tracesLimitsConfigs.getError())
    } else {
      val tracesLimits = tracesLimitsConfigs.get()!!.tracesLimits
      if (!allModulesAreDefined(tracesLimits)) {
        hasConfigError = true
        logger.error(
          "Traces limits {} are incomplete. Missing modules: {}",
          tracesLimitsFile,
          TracingModule.allModules - tracesLimits.keys
        )
      }
    }

    if (smartContractErrorCodes is Err) {
      hasConfigError = true
      logger.error("Reading {} failed: {}", smartContractErrorsFile, smartContractErrorCodes.getError())
    }

    if (gasPriceCapTimeOfDayMultipliers is Err) {
      hasConfigError = true
      logger.error(
        "Reading {} failed: {}",
        gasPriceCapTimeOfDayMultipliersFile,
        gasPriceCapTimeOfDayMultipliers.getError()
      )
    }

    if (configs is Err) {
      hasConfigError = true
      logger.error("Reading {} failed: {}", configFiles, configs.getError())
    }

    return if (hasConfigError) {
      null
    } else {
      configs.get()?.let { config: CoordinatorConfig ->
        config.copy(
          conflation = config.conflation.copy(
            _tracesLimits = tracesLimitsConfigs.get()?.tracesLimits,
            smartContractErrors = smartContractErrorCodes.get()?.smartContractErrors
          ),
          l1DynamicGasPriceCapService = config.l1DynamicGasPriceCapService.copy(
            gasPriceCapCalculation = config.l1DynamicGasPriceCapService.gasPriceCapCalculation.copy(
              timeOfDayMultipliers = gasPriceCapTimeOfDayMultipliers.get()?.gasPriceCapTimeOfDayMultipliers
            )
          )
        )
      }
    }
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
      return loadConfigsOrError<T>(configFiles).onFailure { error ->
        errorWriter.println(error)
      }.get()
    }

    inline fun <reified T : Any> loadConfigsOrError(
      configFiles: List<File>
    ): Result<T, String> {
      val confBuilder: ConfigLoaderBuilder = ConfigLoaderBuilder.Companion.empty().addDefaults()
      for (configFile in configFiles.reversed()) {
        // files must be added in reverse order for overriding
        confBuilder.addFileSource(configFile, false)
      }

      return confBuilder.build().loadConfig<T>(emptyList()).let { config ->
        if (config.isInvalid()) {
          Err(config.getInvalidUnsafe().description())
        } else {
          Ok(config.getUnsafe())
        }
      }
    }
  }
}
