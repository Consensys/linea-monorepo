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
package maru.consensus.qbft

import java.time.Clock
import java.util.concurrent.Executors
import kotlin.time.toJavaDuration
import maru.config.MaruConfig
import maru.consensus.ForkSpec
import maru.consensus.NextBlockTimestampProvider
import maru.consensus.ProtocolFactory
import maru.consensus.StaticValidatorProvider
import maru.consensus.blockimport.BlockBuildingBeaconBlockImporter
import maru.consensus.blockimport.SealedBeaconBlockImporter
import maru.consensus.blockimport.TransactionalSealedBeaconBlockImporter
import maru.consensus.qbft.adapters.ForksScheduleAdapter
import maru.consensus.qbft.adapters.P2PValidatorMulticaster
import maru.consensus.qbft.adapters.ProposerSelectorAdapter
import maru.consensus.qbft.adapters.QbftBlockCodecAdapter
import maru.consensus.qbft.adapters.QbftBlockImporterAdapter
import maru.consensus.qbft.adapters.QbftBlockInterfaceAdapter
import maru.consensus.qbft.adapters.QbftBlockchainAdapter
import maru.consensus.qbft.adapters.QbftFinalStateAdapter
import maru.consensus.qbft.adapters.QbftProtocolScheduleAdapter
import maru.consensus.qbft.adapters.QbftValidatorModeTransitionLoggerAdapter
import maru.consensus.qbft.adapters.QbftValidatorProviderAdapter
import maru.consensus.qbft.adapters.toSealedBeaconBlock
import maru.consensus.state.FinalizationState
import maru.consensus.state.StateTransition
import maru.consensus.state.StateTransitionImpl
import maru.core.BeaconBlockBody
import maru.core.BeaconState
import maru.core.Protocol
import maru.core.Validator
import maru.database.BeaconChain
import maru.executionlayer.manager.ExecutionLayerManager
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import maru.p2p.P2PNetwork
import maru.p2p.SealedBlockHandler
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.consensus.common.bft.BftEventQueue
import org.hyperledger.besu.consensus.common.bft.BftExecutors
import org.hyperledger.besu.consensus.common.bft.BlockTimer
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import org.hyperledger.besu.consensus.common.bft.MessageTracker
import org.hyperledger.besu.consensus.common.bft.RoundTimer
import org.hyperledger.besu.consensus.common.bft.statemachine.FutureMessageBuffer
import org.hyperledger.besu.consensus.qbft.core.network.QbftGossip
import org.hyperledger.besu.consensus.qbft.core.payload.MessageFactory
import org.hyperledger.besu.consensus.qbft.core.statemachine.QbftBlockHeightManagerFactory
import org.hyperledger.besu.consensus.qbft.core.statemachine.QbftController
import org.hyperledger.besu.consensus.qbft.core.statemachine.QbftRoundFactory
import org.hyperledger.besu.consensus.qbft.core.types.QbftMinedBlockObserver
import org.hyperledger.besu.consensus.qbft.core.types.QbftNewChainHead
import org.hyperledger.besu.consensus.qbft.core.validation.MessageValidatorFactory
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory
import org.hyperledger.besu.cryptoservices.KeyPairSecurityModule
import org.hyperledger.besu.cryptoservices.NodeKey
import org.hyperledger.besu.ethereum.core.Util
import org.hyperledger.besu.plugin.services.MetricsSystem
import org.hyperledger.besu.util.Subscribers

class QbftProtocolFactory(
  private val beaconChain: BeaconChain,
  private val privateKeyBytes: ByteArray,
  private val maruConfig: MaruConfig,
  private val metricsSystem: MetricsSystem,
  private val finalizationStateProvider: (BeaconBlockBody) -> FinalizationState,
  private val nextBlockTimestampProvider: NextBlockTimestampProvider,
  private val newBlockHandler: SealedBlockHandler,
  private val executionLayerManager: JsonRpcExecutionLayerManager,
  private val clock: Clock,
  private val p2PNetwork: P2PNetwork,
) : ProtocolFactory {
  override fun create(forkSpec: ForkSpec): Protocol {
    require(maruConfig.validator != null) {
      "Validator configuration is not specified!"
    }

    val signatureAlgorithm = SignatureAlgorithmFactory.getInstance()
    val privateKey = signatureAlgorithm.createPrivateKey(Bytes32.wrap(privateKeyBytes))
    val keyPair = signatureAlgorithm.createKeyPair(privateKey)
    val securityModule = KeyPairSecurityModule(keyPair)
    val nodeKey = NodeKey(securityModule)
    val blockChain = QbftBlockchainAdapter(beaconChain)

    val localAddress = Util.publicKeyToAddress(keyPair.publicKey)
    val qbftProposerSelector = ProposerSelectorAdapter(beaconChain, ProposerSelectorImpl)

    val localValidator = Validator(localAddress.toArray())
    val validatorProvider = StaticValidatorProvider(setOf(localValidator))
    val stateTransition = StateTransitionImpl(validatorProvider)
    val proposerSelector = ProposerSelectorImpl
    val besuValidatorProvider = QbftValidatorProviderAdapter(validatorProvider)
    val sealedBeaconBlockImporter =
      createSealedBeaconBlockImporter(
        executionLayerManager = executionLayerManager,
        beaconChain = beaconChain,
        proposerSelector = proposerSelector,
        localNodeIdentity = localValidator,
        stateTransition = stateTransition,
      )

    val qbftBlockCreatorFactory =
      QbftBlockCreatorFactory(
        manager = executionLayerManager,
        proposerSelector = qbftProposerSelector,
        validatorProvider = validatorProvider,
        beaconChain = beaconChain,
        finalizationStateProvider = finalizationStateProvider,
        blockBuilderIdentity = Validator(localAddress.toArray()),
        eagerQbftBlockCreatorConfig =
          EagerQbftBlockCreator.Config(
            maruConfig.qbftOptions.minBlockBuildTime,
          ),
      )

    val besuForksSchedule = ForksScheduleAdapter(forkSpec, maruConfig.qbftOptions)

    val bftExecutors = BftExecutors.create(metricsSystem, BftExecutors.ConsensusType.QBFT)
    val bftEventQueue = BftEventQueue(maruConfig.qbftOptions.messageQueueLimit)
    val roundTimer =
      RoundTimer(
        /* queue = */ bftEventQueue,
        /* baseExpiryPeriod = */ maruConfig.qbftOptions.roundExpiry.toJavaDuration(),
        /* bftExecutors = */ bftExecutors,
      )
    val blockTimer = BlockTimer(bftEventQueue, besuForksSchedule, bftExecutors, clock)
    val validatorMulticaster = P2PValidatorMulticaster(p2PNetwork)
    val finalState =
      QbftFinalStateAdapter(
        localAddress = localAddress,
        nodeKey = nodeKey,
        validatorProvider = besuValidatorProvider,
        proposerSelector = qbftProposerSelector,
        validatorMulticaster = validatorMulticaster,
        roundTimer = roundTimer,
        blockTimer = blockTimer,
        blockCreatorFactory = qbftBlockCreatorFactory,
        clock = clock,
        beaconChain = beaconChain,
      )

    val minedBlockObservers = Subscribers.create<QbftMinedBlockObserver>()
    minedBlockObservers.subscribe { qbftBlock ->
      newBlockHandler.handleSealedBlock(qbftBlock.toSealedBeaconBlock())
      bftEventQueue.add(QbftNewChainHead(qbftBlock.header))
    }

    val blockImporter = QbftBlockImporterAdapter(sealedBeaconBlockImporter)

    val blockCodec = QbftBlockCodecAdapter
    val blockInterface = QbftBlockInterfaceAdapter()
    val protocolSchedule =
      QbftProtocolScheduleAdapter(
        blockImporter = blockImporter,
        beaconChain = beaconChain,
        proposerSelector = proposerSelector,
        stateTransition = stateTransition,
        executionLayerManager = executionLayerManager,
      )
    val messageValidatorFactory =
      MessageValidatorFactory(
        /* proposerSelector = */ qbftProposerSelector,
        /* protocolSchedule = */ protocolSchedule,
        /* validatorProvider = */ besuValidatorProvider,
        /* blockInterface = */ blockInterface,
      )
    val messageFactory = MessageFactory(nodeKey, blockCodec)
    val qbftRoundFactory =
      QbftRoundFactory(
        /* finalState = */ finalState,
        /* blockInterface = */ blockInterface,
        /* protocolSchedule = */ protocolSchedule,
        /* minedBlockObservers = */ minedBlockObservers,
        /* messageValidatorFactory = */ messageValidatorFactory,
        /* messageFactory = */ messageFactory,
      )

    val transitionLogger = QbftValidatorModeTransitionLoggerAdapter()
    val qbftBlockHeightManagerFactory =
      QbftBlockHeightManagerFactory(
        /* finalState = */ finalState,
        /* roundFactory = */ qbftRoundFactory,
        /* messageValidatorFactory = */ messageValidatorFactory,
        /* messageFactory = */ messageFactory,
        /* validatorProvider = */ besuValidatorProvider,
        /* validatorModeTransitionLogger = */ transitionLogger,
      )
    val duplicateMessageTracker = MessageTracker(maruConfig.qbftOptions.duplicateMessageLimit)
    val chainHeaderNumber =
      beaconChain
        .getLatestBeaconState()
        .latestBeaconBlockHeader
        .number
        .toLong()
    val futureMessageBuffer =
      FutureMessageBuffer(
        /* futureMessagesMaxDistance = */ maruConfig.qbftOptions.futureMessageMaxDistance,
        /* futureMessagesLimit = */ maruConfig.qbftOptions.futureMessagesLimit,
        /* chainHeight = */ chainHeaderNumber,
      )
    val gossiper = QbftGossip(validatorMulticaster, blockCodec)
    val qbftController =
      QbftController(
        /* blockchain = */ blockChain,
        /* finalState = */ finalState,
        /* qbftBlockHeightManagerFactory = */ qbftBlockHeightManagerFactory,
        /* gossiper = */ gossiper,
        /* duplicateMessageTracker = */ duplicateMessageTracker,
        /* futureMessageBuffer = */ futureMessageBuffer,
        /* blockEncoder = */ blockCodec,
      )

    val eventMultiplexer = QbftEventMultiplexer(qbftController)
    val eventProcessor = QbftEventProcessor(bftEventQueue, eventMultiplexer)
    val eventQueueExecutor = Executors.newSingleThreadExecutor()

    return QbftConsensus(
      qbftController = qbftController,
      eventProcessor = eventProcessor,
      bftExecutors = bftExecutors,
      eventQueueExecutor = eventQueueExecutor,
    )
  }

  private fun createSealedBeaconBlockImporter(
    executionLayerManager: ExecutionLayerManager,
    proposerSelector: ProposerSelector,
    localNodeIdentity: Validator,
    beaconChain: BeaconChain,
    stateTransition: StateTransition,
  ): SealedBeaconBlockImporter {
    val finalizationStateProvider = { beaconBlockBody: BeaconBlockBody ->
      val hash = beaconBlockBody.executionPayload.blockHash
      FinalizationState(hash, hash)
    }
    val shouldBuildNextBlock =
      { beaconState: BeaconState, roundIdentifier: ConsensusRoundIdentifier ->
        val nextProposerAddress =
          proposerSelector.getProposerForBlock(beaconState, roundIdentifier).get().address
        nextProposerAddress.contentEquals(localNodeIdentity.address)
      }
    val beaconBlockImporter =
      BlockBuildingBeaconBlockImporter(
        executionLayerManager = executionLayerManager,
        finalizationStateProvider = finalizationStateProvider,
        nextBlockTimestampProvider = nextBlockTimestampProvider,
        shouldBuildNextBlock = shouldBuildNextBlock,
        blockBuilderIdentity = localNodeIdentity,
      )
    return TransactionalSealedBeaconBlockImporter(beaconChain, stateTransition, beaconBlockImporter)
  }
}
