/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import com.sksamuel.hoplite.ConfigLoaderBuilder
import com.sksamuel.hoplite.ConfigResult
import com.sksamuel.hoplite.ExperimentalHoplite
import com.sksamuel.hoplite.addFileSource
import java.io.File
import java.util.concurrent.Callable
import maru.config.MaruConfigDtoToml
import maru.config.QbftOptionsDecoder
import maru.config.consensus.ForkConfigDecoder
import maru.config.consensus.JsonFriendlyForksSchedule
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.core.LoggerContext
import org.apache.logging.log4j.core.config.Configurator
import picocli.CommandLine
import picocli.CommandLine.Command

@Command(
  name = "maru",
  showDefaultValues = true,
  abbreviateSynopsis = true,
  description = ["Runs Maru consensus client"],
  version = ["0.0.1"],
  synopsisHeading = "%n",
  descriptionHeading = "%nDescription:%n%n",
  optionListHeading = "%nOptions:%n",
  footerHeading = "%n",
)
class MaruAppCli : Callable<Int> {
  companion object {
    @OptIn(ExperimentalHoplite::class)
    inline fun <reified T : Any> loadConfig(configFiles: List<File>): ConfigResult<T> {
      val confBuilder: ConfigLoaderBuilder =
        ConfigLoaderBuilder.Companion
          .default()
          .addDecoder(QbftOptionsDecoder)
          .addDecoder(ForkConfigDecoder)
          .withExplicitSealedTypes()
      for (configFile in configFiles.reversed()) {
        // files must be added in reverse order for overriding
        confBuilder.addFileSource(configFile, false)
      }

      return confBuilder.build().loadConfig<T>(emptyList())
    }
  }

  @CommandLine.Option(
    names = ["--config"],
    paramLabel = "CONFIG.toml,CONFIG.overrides.toml",
    description = ["Configuration files"],
    arity = "1..*",
    required = true,
  )
  private val configFiles: List<File>? = null

  @CommandLine.Option(
    names = ["--besu-genesis-file"],
    paramLabel = "BEACON_GENESIS.json",
    description = ["Beacon chain genesis file"],
    required = true,
  )
  private val genesisFile: File? = null

  override fun call(): Int {
    for (configFile in configFiles!!) {
      if (!validateConfigFile(configFile)) {
        System.err.println("Failed to read config file: \"${configFile.path}\"")
        return 1
      }
    }
    if (!validateConfigFile(genesisFile!!)) {
      System.err.println("Failed to read genesis file file: \"${genesisFile.path}\"")
      return 1
    }
    val appConfig = loadConfig<MaruConfigDtoToml>(configFiles)
    val beaconGenesisConfig = loadConfig<JsonFriendlyForksSchedule>(listOf(genesisFile))

    if (!validateParsedFile(appConfig, "app configuration", configFiles.map { it.absolutePath }.toString())) {
      return 1
    }

    if (!validateParsedFile(beaconGenesisConfig, "consensus genesis", genesisFile.absolutePath)) {
      return 1
    }

    val parsedAppConfig = appConfig.getUnsafe()
    val parsedBeaconGenesisConfig = beaconGenesisConfig.getUnsafe()

    val app = MaruApp(parsedAppConfig.domainFriendly(), parsedBeaconGenesisConfig.domainFriendly())
    app.start()

    Runtime
      .getRuntime()
      .addShutdownHook(
        Thread {
          app.stop()
          if (LogManager.getContext() is LoggerContext) {
            // Disable log4j auto shutdown hook is not used otherwise
            // Messages in App.stop won't appear in the logs
            Configurator.shutdown(LogManager.getContext() as LoggerContext)
          }
        },
      )

    return 0
  }

  private fun validateConfigFile(file: File): Boolean = file.canRead()

  private fun validateParsedFile(
    configResult: ConfigResult<*>,
    purpose: String,
    validatedFile: String,
  ): Boolean {
    if (configResult.isInvalid()) {
      System.err.println(
        "Invalid $purpose config file: $validatedFile, ${configResult.getInvalidUnsafe().description()}",
      )
      return false
    }
    return true
  }
}
