package linea.coordinator.config.v2.toml

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.get
import com.github.michaelbull.result.getOrElse
import com.sksamuel.hoplite.ConfigLoaderBuilder
import com.sksamuel.hoplite.ExperimentalHoplite
import com.sksamuel.hoplite.toml.TomlPropertySource
import linea.coordinator.config.v2.CoordinatorConfig
import linea.coordinator.config.v2.toml.decoders.BlockParameterDecoder
import linea.coordinator.config.v2.toml.decoders.BlockParameterNumberDecoder
import linea.coordinator.config.v2.toml.decoders.BlockParameterTagDecoder
import linea.coordinator.config.v2.toml.decoders.TomlByteArrayHexDecoder
import linea.coordinator.config.v2.toml.decoders.TomlKotlinDurationDecoder
import linea.coordinator.config.v2.toml.decoders.TomlSignerTypeDecoder
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.nio.file.Path

fun ConfigLoaderBuilder.addCoordinatorTomlDecoders(): ConfigLoaderBuilder {
  return this
    .addDecoder(BlockParameterTagDecoder())
    .addDecoder(BlockParameterNumberDecoder())
    .addDecoder(BlockParameterDecoder())
    .addDecoder(TomlByteArrayHexDecoder())
    .addDecoder(TomlKotlinDurationDecoder())
    .addDecoder(TomlSignerTypeDecoder())
}

@OptIn(ExperimentalHoplite::class)
inline fun <reified T : Any> parseConfig(toml: String): T {
  return ConfigLoaderBuilder
    .default()
    .withExplicitSealedTypes()
    .addCoordinatorTomlDecoders()
    .addSource(TomlPropertySource(toml))
    .build()
    .loadConfigOrThrow<T>()
}

@OptIn(ExperimentalHoplite::class)
inline fun <reified T : Any> loadConfigsOrError(
  configFiles: List<Path>
): Result<T, String> {
  val confBuilder: ConfigLoaderBuilder = ConfigLoaderBuilder
    .empty()
    .addDefaults()
    .withExplicitSealedTypes()
    .addCoordinatorTomlDecoders()

  return confBuilder
    .build()
    .loadConfig<T>(configFiles.reversed().map { it.toAbsolutePath().toString() })
    .let { config ->
      if (config.isInvalid()) {
        Err(config.getInvalidUnsafe().description())
      } else {
        Ok(config.getUnsafe())
      }
    }
}

fun logErrorIfPresent(
  configName: String,
  configFiles: List<Path>,
  configLoadingResult: Result<Any?, String>,
  logger: Logger
) {
  if (configLoadingResult is Err) {
    logger.error("Failed to load $configName from files=$configFiles with error=${configLoadingResult.error}")
  }
}

inline fun <reified T : Any> loadConfigsAndLogErrors(
  configFiles: List<Path>,
  configName: String,
  logger: Logger = LogManager.getLogger("linea.coordinator.config")
): Result<T, String> {
  return loadConfigsOrError<T>(configFiles)
    .also { logErrorIfPresent(configName, configFiles, it, logger) }
}

fun loadConfigsOrError(
  coordinatorConfigFiles: List<Path>,
  tracesLimitsFileV2: Path,
  gasPriceCapTimeOfDayMultipliersFile: Path,
  smartContractErrorsFile: Path,
  logger: Logger = LogManager.getLogger("linea.coordinator.config")
): Result<CoordinatorConfigToml, String> {
  val coordinatorBaseConfigs =
    loadConfigsAndLogErrors<CoordinatorConfigFileToml>(coordinatorConfigFiles, "coordinator", logger)
  val tracesLimitsV2Configs =
    loadConfigsAndLogErrors<TracesLimitsConfigFileToml>(listOf(tracesLimitsFileV2), "traces limits v2", logger)
  val gasPriceCapTimeOfDayMultipliersConfig =
    loadConfigsAndLogErrors<GasPriceCapTimeOfDayMultipliersConfigFileToml>(
      listOf(gasPriceCapTimeOfDayMultipliersFile),
      "l1 submission gas prices caps",
      logger
    )
  val smartContractErrorsConfig = loadConfigsAndLogErrors<SmartContractErrorCodesConfigFileToml>(
    listOf(smartContractErrorsFile),
    "smart contract errors",
    logger
  )
  val configError = listOf(
    coordinatorBaseConfigs,
    tracesLimitsV2Configs,
    gasPriceCapTimeOfDayMultipliersConfig,
    smartContractErrorsConfig
  )
    .find { it is Err }

  if (configError != null) {
    @Suppress("UNCHECKED_CAST")
    return configError as Result<CoordinatorConfigToml, String>
  }

  val finalConfig = CoordinatorConfigToml(
    configs = coordinatorBaseConfigs.get()!!,
    tracesLimitsV2 = tracesLimitsV2Configs.get()!!,
    l1DynamicGasPriceCapTimeOfDayMultipliers = gasPriceCapTimeOfDayMultipliersConfig.get(),
    smartContractErrors = smartContractErrorsConfig.get()
  )
  return Ok(finalConfig)
}

fun loadConfigs(
  coordinatorConfigFiles: List<Path>,
  tracesLimitsFileV2: Path,
  gasPriceCapTimeOfDayMultipliersFile: Path,
  smartContractErrorsFile: Path,
  logger: Logger = LogManager.getLogger("linea.coordinator.config")
): CoordinatorConfig {
  loadConfigsOrError(
    coordinatorConfigFiles,
    tracesLimitsFileV2,
    gasPriceCapTimeOfDayMultipliersFile,
    smartContractErrorsFile,
    logger
  ).let {
    return it
      .getOrElse { throw RuntimeException("Invalid configurations: $it") }
      .reified()
  }
}
