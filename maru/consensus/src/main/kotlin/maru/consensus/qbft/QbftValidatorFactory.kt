/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft

import java.time.Clock
import java.util.concurrent.Executors
import kotlin.time.Duration.Companion.seconds
import maru.config.QbftConfig
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.NextBlockTimestampProvider
import maru.consensus.PrevRandaoProvider
import maru.consensus.PrevRandaoProviderImpl
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
import maru.consensus.state.FinalizationProvider
import maru.consensus.state.StateTransition
import maru.consensus.state.StateTransitionImpl
import maru.consensus.validation.BeaconBlockValidatorFactoryImpl
import maru.core.BeaconState
import maru.core.Protocol
import maru.core.Validator
import maru.crypto.Hashing
import maru.crypto.Signing
import maru.database.BeaconChain
import maru.executionlayer.manager.ExecutionLayerManager
import maru.p2p.P2PNetwork
import maru.p2p.SealedBeaconBlockHandler
import maru.p2p.ValidationResult
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

class QbftValidatorFactory(
  private val beaconChain: BeaconChain,
  private val privateKeyBytes: ByteArray,
  private val qbftOptions: QbftConfig,
  private val metricsSystem: MetricsSystem,
  private val finalizationStateProvider: FinalizationProvider,
  private val nextBlockTimestampProvider: NextBlockTimestampProvider,
  private val newBlockHandler: SealedBeaconBlockHandler<*>,
  private val executionLayerManager: ExecutionLayerManager,
  private val clock: Clock,
  private val p2PNetwork: P2PNetwork,
  private val allowEmptyBlocks: Boolean,
  private val forksSchedule: ForksSchedule,
) : ProtocolFactory {
  override fun create(forkSpec: ForkSpec): Protocol {
    val protocolConfig = forkSpec.configuration as QbftConsensusConfig
    val signatureAlgorithm = SignatureAlgorithmFactory.getInstance()
    val privateKey = signatureAlgorithm.createPrivateKey(Bytes32.wrap(privateKeyBytes))
    val keyPair = signatureAlgorithm.createKeyPair(privateKey)
    val securityModule = KeyPairSecurityModule(keyPair)
    val nodeKey = NodeKey(securityModule)
    val blockChain = QbftBlockchainAdapter(beaconChain)

    val localAddress = Util.publicKeyToAddress(keyPair.publicKey)
    val qbftProposerSelector = ProposerSelectorAdapter(beaconChain, ProposerSelectorImpl)

    val validatorProvider = StaticValidatorProvider(protocolConfig.validatorSet)
    val stateTransition = StateTransitionImpl(validatorProvider)
    val proposerSelector = ProposerSelectorImpl
    val besuValidatorProvider = QbftValidatorProviderAdapter(validatorProvider)
    val localValidator = Validator(localAddress.toArray())
    val sealedBeaconBlockImporter =
      createSealedBeaconBlockImporter(
        executionLayerManager = executionLayerManager,
        beaconChain = beaconChain,
        proposerSelector = proposerSelector,
        localNodeIdentity = localValidator,
        stateTransition = stateTransition,
        finalizationStateProvider = finalizationStateProvider,
        prevRandaoProvider =
          PrevRandaoProviderImpl(
            signer = Signing.ULongSigner(nodeKey),
            hasher = Hashing::keccak,
          ),
        feeRecipient = qbftOptions.feeRecipient,
      )

    val qbftBlockCreatorFactory =
      QbftBlockCreatorFactory(
        manager = executionLayerManager,
        proposerSelector = qbftProposerSelector,
        validatorProvider = validatorProvider,
        beaconChain = beaconChain,
        finalizationStateProvider = finalizationStateProvider,
        prevRandaoProvider =
          PrevRandaoProviderImpl(
            signer = Signing.ULongSigner(nodeKey),
            hasher = Hashing::keccak,
          ),
        feeRecipient = qbftOptions.feeRecipient,
        eagerQbftBlockCreatorConfig =
          EagerQbftBlockCreator.Config(
            qbftOptions.minBlockBuildTime,
          ),
      )

    val besuForksSchedule = ForksScheduleAdapter(forkSpec, qbftOptions)

    val bftExecutors = BftExecutors.create(metricsSystem, BftExecutors.ConsensusType.QBFT)
    val bftEventQueue = BftEventQueue(qbftOptions.messageQueueLimit)
    val roundExpiry = qbftOptions.roundExpiry ?: forkSpec.blockTimeSeconds.toInt().seconds
    val roundTimer =
      RoundTimer(
        /* queue = */ bftEventQueue,
        /* roundExpiryTimeCalculator = */ ConstantRoundTimeExpiryCalculator(roundExpiry),
        /* bftExecutors = */ bftExecutors,
      )
    val blockTimer = BlockTimer(bftEventQueue, besuForksSchedule, bftExecutors, clock)
    val validatorMulticaster = P2PValidatorMulticaster(p2PNetwork)
    val finalState =
      QbftFinalStateAdapter(
        localAddress = localAddress,
        nodeKey = nodeKey,
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
    val blockInterface = QbftBlockInterfaceAdapter(stateTransition)
    val beaconBlockValidatorFactory =
      BeaconBlockValidatorFactoryImpl(
        beaconChain = beaconChain,
        proposerSelector = proposerSelector,
        stateTransition = stateTransition,
        executionLayerManager = executionLayerManager,
        allowEmptyBlocks = allowEmptyBlocks,
      )
    val protocolSchedule =
      QbftProtocolScheduleAdapter(
        blockImporter = blockImporter,
        beaconBlockValidatorFactory = beaconBlockValidatorFactory,
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
    val duplicateMessageTracker = MessageTracker(qbftOptions.duplicateMessageLimit)
    val chainHeaderNumber =
      beaconChain
        .getLatestBeaconState()
        .beaconBlockHeader
        .number
        .toLong()
    val futureMessageBuffer =
      FutureMessageBuffer(
        /* futureMessagesMaxDistance = */ qbftOptions.futureMessageMaxDistance,
        /* futureMessagesLimit = */ qbftOptions.futureMessagesLimit,
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
    val eventQueueExecutor = Executors.newSingleThreadExecutor(Thread.ofPlatform().daemon().factory())

    return QbftConsensusValidator(
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
    finalizationStateProvider: FinalizationProvider,
    prevRandaoProvider: PrevRandaoProvider<ULong>,
    feeRecipient: ByteArray,
  ): SealedBeaconBlockImporter<ValidationResult> {
    val shouldBuildNextBlock =
      { beaconState: BeaconState, roundIdentifier: ConsensusRoundIdentifier, nextBlockTimestamp: ULong ->
        // We shouldn't build next block if this fork ends
        val nextForkTimestamp =
          forksSchedule.getNextForkByTimestamp(beaconState.beaconBlockHeader.timestamp)?.timestampSeconds
            ?: ULong.MAX_VALUE
        if (nextBlockTimestamp >= nextForkTimestamp) {
          false
        } else {
          val nextProposerAddress =
            proposerSelector.getProposerForBlock(beaconState, roundIdentifier).get().address
          nextProposerAddress.contentEquals(localNodeIdentity.address)
        }
      }
    val beaconBlockImporter =
      BlockBuildingBeaconBlockImporter(
        executionLayerManager = executionLayerManager,
        finalizationStateProvider = finalizationStateProvider,
        nextBlockTimestampProvider = nextBlockTimestampProvider,
        prevRandaoProvider = prevRandaoProvider,
        shouldBuildNextBlock = shouldBuildNextBlock,
        feeRecipient = feeRecipient,
      )
    return TransactionalSealedBeaconBlockImporter(
      beaconChain = beaconChain,
      stateTransition = stateTransition,
      beaconBlockImporter = beaconBlockImporter,
    )
  }
}
