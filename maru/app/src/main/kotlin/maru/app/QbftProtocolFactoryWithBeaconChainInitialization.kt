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
import maru.config.MaruConfig
import maru.consensus.ForkSpec
import maru.consensus.NewBlockHandler
import maru.consensus.NextBlockTimestampProvider
import maru.consensus.ProtocolFactory
import maru.consensus.qbft.QbftConsensusConfig
import maru.consensus.qbft.QbftProtocolFactory
import maru.consensus.state.FinalizationState
import maru.core.BeaconBlock
import maru.core.BeaconBlockBody
import maru.core.BeaconBlockHeader
import maru.core.BeaconState
import maru.core.HashUtil
import maru.core.Protocol
import maru.core.SealedBeaconBlock
import maru.core.Validator
import maru.crypto.Crypto
import maru.database.BeaconChain
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import maru.mappers.Mappers.toDomain
import maru.serialization.rlp.RLPSerializers
import maru.serialization.rlp.stateRoot
import org.hyperledger.besu.consensus.common.bft.Gossiper
import org.hyperledger.besu.consensus.common.bft.network.ValidatorMulticaster
import org.hyperledger.besu.plugin.services.MetricsSystem
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter

class QbftProtocolFactoryWithBeaconChainInitialization(
  private val maruConfig: MaruConfig,
  private val metricsSystem: MetricsSystem,
  private val finalizationStateProvider: (BeaconBlockBody) -> FinalizationState,
  private val executionLayerClient: Web3j,
  private val beaconChain: BeaconChain,
  private val nextTargetBlockTimestampProvider: NextBlockTimestampProvider,
  private val newBlockHandler: NewBlockHandler<Unit>,
  private val clock: Clock,
  private val gossiper: Gossiper,
  private val validatorMulticaster: ValidatorMulticaster,
) : ProtocolFactory {
  init {
    require(maruConfig.validator != null) { "The validator is required when QBFT protocol is instantiated!" }
  }

  private fun initializeDb(updater: BeaconChain.Updater) {
    val genesisExecutionPayload =
      executionLayerClient
        .ethGetBlockByNumber(DefaultBlockParameter.valueOf("latest"), true)
        .send()
        .block
        .toDomain()

    val beaconBlockBody = BeaconBlockBody(prevCommitSeals = emptyList(), executionPayload = genesisExecutionPayload)

    val beaconBlockHeader =
      BeaconBlockHeader(
        number = 0u,
        round = 0u,
        timestamp = genesisExecutionPayload.timestamp,
        proposer = Validator(genesisExecutionPayload.feeRecipient),
        parentRoot = ByteArray(32),
        stateRoot = ByteArray(32),
        bodyRoot = ByteArray(32),
        headerHashFunction = RLPSerializers.DefaultHeaderHashFunction,
      )

    val initialValidators = setOf(Crypto.privateKeyToValidator(maruConfig.validator!!.privateKey))
    val tmpGenesisStateRoot =
      BeaconState(
        latestBeaconBlockHeader = beaconBlockHeader,
        validators = initialValidators,
      )
    val stateRootHash = HashUtil.stateRoot(tmpGenesisStateRoot)

    val genesisBlockHeader = beaconBlockHeader.copy(stateRoot = stateRootHash)
    val genesisBlock = BeaconBlock(genesisBlockHeader, beaconBlockBody)
    val genesisStateRoot = BeaconState(genesisBlockHeader, initialValidators)
    updater.putBeaconState(genesisStateRoot)
    updater.putSealedBeaconBlock(SealedBeaconBlock(genesisBlock, emptyList()))
    updater.commit()
  }

  override fun create(forkSpec: ForkSpec): Protocol {
    require(forkSpec.configuration is QbftConsensusConfig) {
      "Unexpected fork specification! ${
        forkSpec
          .configuration
      } instead of ${QbftConsensusConfig::class.simpleName}"
    }
    val qbftConsensusConfig = forkSpec.configuration as QbftConsensusConfig

    val engineApiExecutionLayerClient =
      Helpers.buildExecutionEngineClient(
        maruConfig.validator!!.engineApiClient,
        qbftConsensusConfig.elFork,
      )
    val executionLayerManager =
      JsonRpcExecutionLayerManager(
        executionLayerEngineApiClient = engineApiExecutionLayerClient,
      )

    if (!beaconChain.isInitialized()) {
      initializeDb(beaconChain.newUpdater())
    }

    val qbftProtocolFactory =
      QbftProtocolFactory(
        beaconChain = beaconChain,
        maruConfig = maruConfig,
        metricsSystem = metricsSystem,
        finalizationStateProvider = finalizationStateProvider,
        nextBlockTimestampProvider = nextTargetBlockTimestampProvider,
        newBlockHandler = newBlockHandler,
        executionLayerManager = executionLayerManager,
        clock = clock,
        gossiper = gossiper,
        validatorMulticaster = validatorMulticaster,
      )
    return qbftProtocolFactory.create(forkSpec)
  }
}
