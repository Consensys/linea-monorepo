/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.e2e

import java.io.File
import java.nio.file.Files
import maru.app.MaruApp
import maru.app.MaruAppCli.Companion.loadConfig
import maru.config.MaruConfigDtoToml
import maru.consensus.config.JsonFriendlyForksSchedule

const val VALIDATOR_PRIVATE_KEY_WITH_PREFIX =
  "0x080212201dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae"

object MaruFactory {
  fun buildTestMaru(pragueTime: Long): MaruApp {
    val maruConfigResource = this::class.java.getResource("/config/maru.toml")
    val maruConfig = loadConfig<MaruConfigDtoToml>(listOf(File(maruConfigResource!!.path)))
    Files.writeString(
      maruConfig
        .getUnsafe()
        .domainFriendly()
        .persistence.privateKeyPath,
      VALIDATOR_PRIVATE_KEY_WITH_PREFIX,
    )
    val consensusGenesisTemplate =
      this::class.java
        .getResource("/config/clique-to-prague.template")!!
        .readText()
    val tmpDirFile = Files.createTempDirectory("maru-clique-to-pos").toFile()
    tmpDirFile.deleteOnExit()
    val maruGenesisFile = File(tmpDirFile, "clique-to-prague.json")
    maruGenesisFile.writeText(renderTemplate(consensusGenesisTemplate, pragueTime))

    val beaconGenesisConfig =
      loadConfig<JsonFriendlyForksSchedule>(listOf(maruGenesisFile))

    return MaruApp(maruConfig.getUnsafe().domainFriendly(), beaconGenesisConfig.getUnsafe().domainFriendly())
  }

  private fun renderTemplate(
    template: String,
    pragueTime: Long,
  ): String = template.replace("%PRAGUE_TIME%", pragueTime.toString())
}
