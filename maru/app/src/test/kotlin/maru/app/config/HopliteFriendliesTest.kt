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
package maru.app.config

import com.sksamuel.hoplite.ConfigLoaderBuilder
import com.sksamuel.hoplite.ExperimentalHoplite
import com.sksamuel.hoplite.Secret
import com.sksamuel.hoplite.json.JsonPropertySource
import com.sksamuel.hoplite.toml.TomlPropertySource
import java.net.URI
import maru.consensus.dummy.DummyConsensusConfig
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

@OptIn(ExperimentalHoplite::class)
class HopliteFriendliesTest {
  private inline fun <reified T : Any> parseJsonConfig(json: String): T =
    ConfigLoaderBuilder
      .default()
      .withExplicitSealedTypes()
      .addSource(JsonPropertySource(json))
      .build()
      .loadConfigOrThrow<T>()

  private inline fun <reified T : Any> parseTomlConfig(toml: String): T =
    ConfigLoaderBuilder
      .default()
      .withExplicitSealedTypes()
      .addSource(TomlPropertySource(toml))
      .build()
      .loadConfigOrThrow<T>()

  @Test
  fun genesisFileIsParseable() {
    val config =
      parseJsonConfig<DummyConsensusConfig>(
        """
        {
          "blockTimeMillis": 1000
        }
        """.trimIndent(),
      )
    assertThat(config.blockTimeMillis).isEqualTo(1000u)
  }

  @Test
  fun appConfigFileIsParseable() {
    val config =
      parseTomlConfig<MaruConfigDtoToml>(
        """
        [execution-client]
        endpoint = "https://localhost"

        [p2p-config]
        port = 3322

        [validator]
        validator-key = "0xdead"
        """.trimIndent(),
      )
    assertThat(config)
      .isEqualTo(
        MaruConfigDtoToml(
          executionClient =
            ExecutionClientConfig(endpoint = URI.create("https://localhost").toURL()),
          p2pConfig = P2P(port = 3322u),
          validator = ValidatorToml(validatorKey = Secret("0xdead")),
        ),
      )
  }
}
