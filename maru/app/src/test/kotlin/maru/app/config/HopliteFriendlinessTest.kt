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

import com.sksamuel.hoplite.Secret
import java.net.URI
import kotlin.time.Duration.Companion.milliseconds
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class HopliteFriendlinessTest {
  @Test
  fun appConfigFileIsParseable() {
    val config =
      Utils.parseTomlConfig<MaruConfigDtoToml>(
        """
        [execution-client]
        ethereum-json-rpc-endpoint = "http://localhost:8545"
        engine-api-json-rpc-endpoint = "http://localhost:8555"
        min-time-between-get-payload-attempts=800m

        [dummy-consensus-options]
        communication-time-margin=100m

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
            ExecutionClientConfig(
              ethereumJsonRpcEndpoint = URI.create("http://localhost:8545").toURL(),
              engineApiJsonRpcEndpoint = URI.create("http://localhost:8555").toURL(),
              minTimeBetweenGetPayloadAttempts = 800.milliseconds,
            ),
          dummyConsensusOptions = DummyConsensusOptionsDtoToml(100.milliseconds),
          p2pConfig = P2P(port = 3322u),
          validator = ValidatorDtoToml(validatorKey = Secret("0xdead")),
        ),
      )
  }

  @Test
  fun appConfigFileIsConvertableToDomain() {
    val config =
      Utils.parseTomlConfig<MaruConfigDtoToml>(
        """
        [execution-client]
        ethereum-json-rpc-endpoint = "http://localhost:8545"
        engine-api-json-rpc-endpoint = "http://localhost:8555"
        min-time-between-get-payload-attempts=800m

        [dummy-consensus-options]
        communication-time-margin=100m

        [p2p-config]
        port = 3322

        [validator]
        validator-key = "0xdead"
        """.trimIndent(),
      )
    assertThat(config.domainFriendly())
      .isEqualTo(
        MaruConfig(
          executionClientConfig =
            ExecutionClientConfig(
              engineApiJsonRpcEndpoint = URI.create("http://localhost:8555").toURL(),
              ethereumJsonRpcEndpoint = URI.create("http://localhost:8545").toURL(),
              minTimeBetweenGetPayloadAttempts = 800.milliseconds,
            ),
          dummyConsensusOptions = DummyConsensusOptions(100.milliseconds),
          p2pConfig = P2P(port = 3322u),
          validator = Validator(validatorKey = Bytes.fromHexString("0xdead").toArray()),
        ),
      )
  }
}
