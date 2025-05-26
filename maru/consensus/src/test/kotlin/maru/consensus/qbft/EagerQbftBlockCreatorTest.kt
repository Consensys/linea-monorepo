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
import java.time.Duration
import java.time.Instant
import java.time.temporal.ChronoUnit
import kotlin.random.Random
import kotlin.time.Duration.Companion.milliseconds
import maru.consensus.ValidatorProvider
import maru.consensus.qbft.adapters.QbftBlockHeaderAdapter
import maru.consensus.qbft.adapters.toBeaconBlock
import maru.consensus.qbft.adapters.toBeaconBlockHeader
import maru.consensus.state.FinalizationState
import maru.core.BeaconBlock
import maru.core.BeaconState
import maru.core.EMPTY_HASH
import maru.core.HashUtil
import maru.core.SealedBeaconBlock
import maru.core.Validator
import maru.core.ext.DataGenerators
import maru.database.BeaconChain
import maru.executionlayer.client.PragueWeb3JJsonRpcExecutionLayerEngineApiClient
import maru.executionlayer.manager.ExecutionLayerManager
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import maru.mappers.Mappers.toDomain
import maru.serialization.rlp.bodyRoot
import maru.serialization.rlp.headerHash
import maru.serialization.rlp.stateRoot
import maru.testutils.besu.BesuFactory
import maru.testutils.besu.BesuTransactionsHelper
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import org.hyperledger.besu.consensus.common.bft.blockcreation.ProposerSelector
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockCreator
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockHeader
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.tests.acceptance.dsl.condition.net.NetConditions
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode
import org.hyperledger.besu.tests.acceptance.dsl.node.ThreadBesuNodeRunner
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.Cluster
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.ClusterConfigurationBuilder
import org.hyperledger.besu.tests.acceptance.dsl.transaction.net.NetTransactions
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.Mockito
import org.mockito.kotlin.reset
import org.mockito.kotlin.whenever
import org.web3j.protocol.core.DefaultBlockParameter
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JExecutionEngineClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3jClientBuilder
import tech.pegasys.teku.infrastructure.async.SafeFuture.completedFuture
import tech.pegasys.teku.infrastructure.time.SystemTimeProvider

class EagerQbftBlockCreatorTest {
  private lateinit var cluster: Cluster
  private lateinit var besuInstance: BesuNode
  private lateinit var ethApiClient: Web3JClient
  private val proposerSelector = Mockito.mock(ProposerSelector::class.java)
  private val validatorProvider = Mockito.mock(ValidatorProvider::class.java)
  private val beaconChain = Mockito.mock(BeaconChain::class.java)
  private val clock = Mockito.mock(Clock::class.java)
  private val validator = Validator(Random.nextBytes(20))
  private lateinit var executionLayerManager: ExecutionLayerManager
  private val validatorSet = DataGenerators.randomValidators() + validator

  @BeforeEach
  fun beforeEach() {
    cluster =
      Cluster(
        ClusterConfigurationBuilder().build(),
        NetConditions(NetTransactions()),
        ThreadBesuNodeRunner(),
      )
    besuInstance =
      BesuFactory.buildTestBesu().also {
        cluster.start(it)
      }
    ethApiClient =
      Web3jClientBuilder()
        .endpoint(besuInstance.engineRpcUrl().get())
        .timeout(Duration.ofMinutes(1))
        .timeProvider(SystemTimeProvider.SYSTEM_TIME_PROVIDER)
        .executionClientEventsPublisher { }
        .build()
    reset(
      proposerSelector,
      validatorProvider,
      beaconChain,
    )
    executionLayerManager = createExecutionLayerManager()
  }

  /*
   * Creates EagerQbftBlockCreator instance
   * Runs a block building attempt to ensure that Besu's state isn't a clean slate
   * Sets the mock up according to the generated genesis beacon block
   */
  private fun setup(
    sealedGenesisBeaconBlock: SealedBeaconBlock,
    adaptedGenesisBeaconBlock: QbftBlockHeaderAdapter,
  ): EagerQbftBlockCreator {
    whenever(
      beaconChain.getSealedBeaconBlock(sealedGenesisBeaconBlock.beaconBlock.beaconBlockHeader.hash()),
    ).thenReturn(
      sealedGenesisBeaconBlock,
    )
    whenever(proposerSelector.selectProposerForRound(ConsensusRoundIdentifier(1L, 1))).thenReturn(
      Address.wrap(
        Bytes.wrap
          (validator.address),
      ),
    )
    whenever(
      validatorProvider.getValidatorsAfterBlock(0u),
    ).thenReturn(completedFuture(validatorSet))

    val mainBlockCreator =
      DelayedQbftBlockCreator(
        manager = executionLayerManager,
        proposerSelector = proposerSelector,
        validatorProvider = validatorProvider,
        beaconChain = beaconChain,
        round = 1,
      )
    val genesisBlockHash = sealedGenesisBeaconBlock.beaconBlock.beaconBlockBody.executionPayload.blockHash
    val eagerQbftBlockCreator =
      EagerQbftBlockCreator(
        manager = executionLayerManager,
        delegate = mainBlockCreator,
        finalizationStateProvider = {
          FinalizationState(
            genesisBlockHash,
            genesisBlockHash,
          )
        },
        blockBuilderIdentity = validator,
        config =
          EagerQbftBlockCreator.Config(
            minBlockBuildTime = 500.milliseconds,
          ),
        beaconChain = beaconChain,
      )
    createProposalToBeRejected(
      clock = clock,
      genesisBlockHash = genesisBlockHash,
      mainBlockCreator = mainBlockCreator,
      parentHeader = adaptedGenesisBeaconBlock,
    )
    return eagerQbftBlockCreator
  }

  @Test
  fun `can create a non empty block with new timestamp`() {
    val genesisExecutionPayload =
      ethApiClient.eth1Web3j
        .ethGetBlockByNumber(
          DefaultBlockParameter.valueOf("earliest"),
          true,
        ).send()
        .block
        .toDomain()
    val parentBlock =
      BeaconBlock(
        beaconBlockHeader = DataGenerators.randomBeaconBlockHeader(0U),
        beaconBlockBody =
          DataGenerators
            .randomBeaconBlockBody()
            .copy(executionPayload = genesisExecutionPayload),
      )
    val sealedGenesisBeaconBlock = SealedBeaconBlock(parentBlock, emptySet())
    val parentHeader = QbftBlockHeaderAdapter(sealedGenesisBeaconBlock.beaconBlock.beaconBlockHeader)

    val eagerQbftBlockCreator = setup(sealedGenesisBeaconBlock, parentHeader)

    setNextSecondMillis(0)
    // Try to create an new non-empty block instead of a non-empty proposal
    val acceptedBlockTimestamp = (clock.millis() / 1000)
    val acceptedBlock = eagerQbftBlockCreator.createBlock(acceptedBlockTimestamp, parentHeader)
    val acceptedBeaconBlock = acceptedBlock.toBeaconBlock()

    // block header fields
    val createdBlockHeader = acceptedBeaconBlock.beaconBlockHeader
    assertThat(createdBlockHeader.number).isEqualTo(1UL)
    assertThat(createdBlockHeader.round).isEqualTo(1U)
    assertThat(createdBlockHeader.timestamp).isEqualTo(acceptedBlockTimestamp.toULong())
    assertThat(createdBlockHeader.proposer).isEqualTo(validator)

    // block header roots
    val stateRoot =
      HashUtil.stateRoot(
        BeaconState(
          acceptedBeaconBlock.beaconBlockHeader.copy(stateRoot = EMPTY_HASH),
          validatorSet,
        ),
      )
    assertThat(
      createdBlockHeader.bodyRoot,
    ).isEqualTo(
      HashUtil.bodyRoot(acceptedBeaconBlock.beaconBlockBody),
    )
    assertThat(createdBlockHeader.stateRoot).isEqualTo(stateRoot)
    assertThat(createdBlockHeader.parentRoot).isEqualTo(parentHeader.toBeaconBlockHeader().hash())
    assertThat(
      acceptedBeaconBlock.beaconBlockHeader.hash(),
    ).isEqualTo(HashUtil.headerHash(acceptedBeaconBlock.beaconBlockHeader))

    // block body fields
    val createdBlockBody = acceptedBeaconBlock.beaconBlockBody
    assertThat(
      createdBlockBody.prevCommitSeals,
    ).isEqualTo(
      sealedGenesisBeaconBlock.commitSeals,
    )
    assertThat(createdBlockBody.executionPayload.timestamp).isEqualTo(acceptedBlockTimestamp.toULong())
    assertThat(createdBlockBody.executionPayload.transactions).isNotEmpty
  }

  private fun createProposalToBeRejected(
    clock: Clock,
    genesisBlockHash: ByteArray,
    mainBlockCreator: QbftBlockCreator,
    parentHeader: QbftBlockHeader,
  ) {
    val rejectedBlockTimestamp = (clock.millis() / 1000) + 1
    executionLayerManager
      .setHeadAndStartBlockBuilding(
        headHash = genesisBlockHash,
        safeHash = genesisBlockHash,
        finalizedHash = genesisBlockHash,
        nextBlockTimestamp = rejectedBlockTimestamp,
        feeRecipient = validator.address,
      ).get()
    val transaction = BesuTransactionsHelper().createTransfers(1u)
    besuInstance.execute(transaction)
    Thread.sleep(1000)
    val rejectedBlock = mainBlockCreator.createBlock(rejectedBlockTimestamp, parentHeader)
    val rejectedBlockTransactions =
      rejectedBlock
        .toBeaconBlock()
        .beaconBlockBody.executionPayload.transactions
    assertThat(
      rejectedBlockTransactions,
    ).isNotEmpty
  }

  private fun createExecutionLayerManager(): ExecutionLayerManager {
    val engineApiClient =
      Web3jClientBuilder()
        .endpoint(besuInstance.engineRpcUrl().get())
        .timeout(Duration.ofMinutes(1))
        .timeProvider(SystemTimeProvider.SYSTEM_TIME_PROVIDER)
        .executionClientEventsPublisher { }
        .build()
    return JsonRpcExecutionLayerManager(
      PragueWeb3JJsonRpcExecutionLayerEngineApiClient(
        Web3JExecutionEngineClient(engineApiClient),
      ),
    )
  }

  /*
   * Sets the clock to return the next second with the given milliseconds
   */
  private fun setNextSecondMillis(secondMillis: Int) {
    require(secondMillis in 0..999) { "secondMillis must be between 0 and 999" }

    val currentMillis = System.currentTimeMillis()
    val currentSecond = Instant.ofEpochMilli(currentMillis).truncatedTo(ChronoUnit.SECONDS)
    val nextSecond = currentSecond.plusSeconds(1).plusMillis(secondMillis.toLong())
    whenever(clock.millis()).thenAnswer {
      nextSecond.toEpochMilli()
    }
  }
}
