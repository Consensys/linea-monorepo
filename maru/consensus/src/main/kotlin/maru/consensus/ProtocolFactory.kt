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
package maru.consensus

import java.time.Clock
import maru.config.MaruConfig
import maru.consensus.delegated.ElDelegatedConsensus
import maru.consensus.dummy.DummyConsensusConfig
import maru.consensus.dummy.DummyConsensusProtocolBuilder
import maru.core.Protocol
import maru.executionlayer.client.ExecutionLayerClient
import org.web3j.protocol.Web3j

interface ProtocolFactory {
  fun create(protocolConfig: ConsensusConfig): Protocol
}

class OmniProtocolFactory(
  private val forksSchedule: ForksSchedule,
  private val clock: Clock,
  private val config: MaruConfig,
  private val executionLayerClient: ExecutionLayerClient,
  private val ethereumJsonRpcClient: Web3j,
  private val newBlockHandler: NewBlockHandler,
) : ProtocolFactory {
  override fun create(protocolConfig: ConsensusConfig): Protocol =
    when (protocolConfig) {
      is DummyConsensusConfig -> {
        require(config.dummyConsensusOptions != null) {
          "Next fork is dummy consensus one, but dummyConsensusOptions are undefined!"
        }

        DummyConsensusProtocolBuilder
          .build(
            forksSchedule = forksSchedule,
            clock = clock,
            minTimeTillNextBlock = config.executionClientConfig.minTimeBetweenGetPayloadAttempts,
            dummyConsensusOptions = config.dummyConsensusOptions!!,
            executionLayerClient = executionLayerClient,
            onNewBlockHandler = newBlockHandler,
          )
      }

      is ElDelegatedConsensus.Config -> {
        ElDelegatedConsensus(
          ethereumJsonRpcClient = ethereumJsonRpcClient,
          onNewBlock = newBlockHandler,
          config = protocolConfig,
        )
      }

      else -> {
        throw IllegalArgumentException("Fork $protocolConfig is unknown!")
      }
    }
}
