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
package maru.app

import java.time.Clock
import kotlin.time.Duration.Companion.seconds
import maru.config.FollowersConfig
import maru.config.MaruConfig
import maru.consensus.ForksSchedule
import maru.consensus.NewBlockHandler
import maru.consensus.NewBlockHandlerMultiplexer
import maru.consensus.NextBlockTimestampProviderImpl
import maru.consensus.OmniProtocolFactory
import maru.consensus.ProtocolStarter
import maru.consensus.ProtocolStarterBlockHandler
import maru.consensus.delegated.ElDelegatedConsensusFactory
import maru.consensus.dummy.DummyConsensusProtocolFactory
import maru.executionlayer.client.Web3jMetadataProvider
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

class MaruApp(
  config: MaruConfig,
  beaconGenesisConfig: ForksSchedule,
  clock: Clock = Clock.systemUTC(),
) {
  private val log: Logger = LogManager.getLogger(this::class.java)

  init {
    if (config.p2pConfig == null) {
      log.warn("P2P is disabled!")
    }
    if (config.validator == null) {
      log.info("Maru is running in follower-only node")
    }
    log.info(config.toString())
  }

  private val ethereumJsonRpcClient =
    Helpers.createWeb3jClient(
      config.sotNode,
    )

  private val metadataProvider = Web3jMetadataProvider(ethereumJsonRpcClient.eth1Web3j)

  private val nextTargetBlockTimestampProvider =
    NextBlockTimestampProviderImpl(
      clock = clock,
      forksSchedule = beaconGenesisConfig,
      minTimeTillNextBlock = 0.seconds,
    )
  private val protocolStarter =
    let {
      val delegatedConsensusNewBlockHandler = NewBlockHandlerMultiplexer(emptyMap())
      val dummyConsensusNewBlockHandler = NewBlockHandlerMultiplexer(createFollowerHandlers(config.followers))
      ProtocolStarter(
        forksSchedule = beaconGenesisConfig,
        protocolFactory =
          OmniProtocolFactory(
            dummyConsensusFactory =
              DummyConsensusProtocolFactory(
                forksSchedule = beaconGenesisConfig,
                clock = clock,
                maruConfig = config,
                metadataProvider = metadataProvider,
                newBlockHandler = dummyConsensusNewBlockHandler,
              ),
            elDelegatedConsensusFactory =
              ElDelegatedConsensusFactory(
                ethereumJsonRpcClient = ethereumJsonRpcClient.eth1Web3j,
                newBlockHandler = delegatedConsensusNewBlockHandler,
              ),
          ),
        metadataProvider = metadataProvider,
        nextBlockTimestampProvider = nextTargetBlockTimestampProvider,
      ).also {
        delegatedConsensusNewBlockHandler.addHandler("protocol starter", ProtocolStarterBlockHandler(it))
        dummyConsensusNewBlockHandler.addHandler("protocol starter", ProtocolStarterBlockHandler(it))
      }
    }

  private fun createFollowerHandlers(followers: FollowersConfig): Map<String, NewBlockHandler> =
    followers.followers
      .mapValues {
        Helpers.createBlockImporter(it.value, metadataProvider)
      }

  fun start() {
    protocolStarter.start()
    log.info("Maru is up")
  }

  fun stop() {
    protocolStarter.stop()
  }
}
