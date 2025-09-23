/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import java.io.File
import java.util.concurrent.Callable
import maru.config.MaruConfigLoader
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
  @CommandLine.Option(
    names = ["--config"],
    paramLabel = "CONFIG.toml,CONFIG.overrides.toml",
    description = ["Configuration files"],
    arity = "1..*",
    required = true,
  )
  private val configFiles: List<File>? = null

  @CommandLine.Option(
    names = ["--maru-genesis-file"],
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
    val parsedAppConfig = MaruConfigLoader.loadAppConfigs(configFiles)
    val parsedBeaconGenesisConfig = MaruConfigLoader.loadGenesisConfig(genesisFile)

    val app =
      MaruAppFactory()
        .create(
          config = parsedAppConfig.domainFriendly(),
          beaconGenesisConfig = parsedBeaconGenesisConfig.domainFriendly(),
        )
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
}
