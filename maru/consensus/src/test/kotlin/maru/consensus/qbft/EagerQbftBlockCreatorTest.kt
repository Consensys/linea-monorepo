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
import kotlin.random.Random
import kotlin.time.Duration.Companion.milliseconds
import maru.consensus.ConsensusConfig
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.NextBlockTimestampProviderImpl
import maru.consensus.ValidatorProvider
import maru.consensus.qbft.adapters.QbftBlockHeaderAdapter
import maru.consensus.qbft.adapters.toBeaconBlock
import maru.consensus.qbft.adapters.toBeaconBlockHeader
import maru.consensus.state.FinalizationState
import maru.core.BeaconBlock
import maru.core.BeaconBlockHeader
import maru.core.BeaconState
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
import maru.testutils.TransactionsHelper
import maru.testutils.besu.BesuFactory
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import org.hyperledger.besu.consensus.common.bft.blockcreation.ProposerSelector
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.tests.acceptance.dsl.condition.net.NetConditions
import org.hyperledger.besu.tests.acceptance.dsl.node.ThreadBesuNodeRunner
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.Cluster
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.ClusterConfigurationBuilder
import org.hyperledger.besu.tests.acceptance.dsl.transaction.net.NetTransactions
import org.junit.jupiter.api.Test
import org.mockito.Mockito
import org.mockito.kotlin.whenever
import org.web3j.protocol.core.DefaultBlockParameter
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JExecutionEngineClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3jClientBuilder
import tech.pegasys.teku.infrastructure.async.SafeFuture.completedFuture
import tech.pegasys.teku.infrastructure.time.SystemTimeProvider

class EagerQbftBlockCreatorTest {
  private var cluster =
    Cluster(
      ClusterConfigurationBuilder().build(),
      NetConditions(NetTransactions()),
      ThreadBesuNodeRunner(),
    )
  private val besuInstance =
    BesuFactory.buildTestBesu().also {
      cluster.start(it)
    }
  private val ethApiClient =
    Web3jClientBuilder()
      .endpoint(besuInstance.engineRpcUrl().get())
      .timeout(Duration.ofMinutes(1))
      .timeProvider(SystemTimeProvider.SYSTEM_TIME_PROVIDER)
      .executionClientEventsPublisher { }
      .build()
  private val proposerSelector = Mockito.mock(ProposerSelector::class.java)
  private val validatorProvider = Mockito.mock(ValidatorProvider::class.java)
  private val beaconChain = Mockito.mock(BeaconChain::class.java)
  private val validator = Validator(Random.nextBytes(20))
  private val executionLayerManager = createExecutionLayerManager()
  private val clock = Clock.systemUTC()
  private val validatorSet = DataGenerators.randomValidators() + validator
  private val forksSchedule =
    ForksSchedule(
      setOf(
        ForkSpec(
          timestampSeconds = 0,
          blockTimeSeconds = 1,
          configuration = object : ConsensusConfig {},
        ),
      ),
    )
  private val nextBlockTimestampProvider = NextBlockTimestampProviderImpl(clock, forksSchedule)

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
    val sealedGenesisBeaconBlock = SealedBeaconBlock(parentBlock, emptyList())

    val parentHeader = QbftBlockHeaderAdapter(sealedGenesisBeaconBlock.beaconBlock.beaconBlockHeader)
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
    val genesisBlockHash = genesisExecutionPayload.blockHash
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
            communicationMargin = 100.milliseconds,
          ),
        beaconChain = beaconChain,
        nextBlockTimestampProvider = nextBlockTimestampProvider,
        clock = clock,
      )
    // Create a non-empty proposal
    val rejectedBlockTimestamp = clock.millis() / 1000
    executionLayerManager
      .setHeadAndStartBlockBuilding(
        headHash = genesisBlockHash,
        safeHash = genesisBlockHash,
        finalizedHash = genesisBlockHash,
        nextBlockTimestamp = rejectedBlockTimestamp,
        feeRecipient = validator.address,
      ).get()
    val transaction = TransactionsHelper().createTransfers(1u)
    besuInstance.execute(transaction)
    Thread.sleep(1000)
    val rejectedBlock = mainBlockCreator.createBlock(rejectedBlockTimestamp, parentHeader)
    val proposedTransactions =
      rejectedBlock
        .toBeaconBlock()
        .beaconBlockBody.executionPayload.transactions
    assertThat(
      proposedTransactions,
    ).hasSize(1)

    // Try to create an empty block instead of a non-empty proposal
    val blockTimestamp = clock.millis() / 1000
    val createdBlock = eagerQbftBlockCreator.createBlock(blockTimestamp, parentHeader)
    val createdBeaconBlock = createdBlock.toBeaconBlock()

    // block header fields
    val createdBlockHeader = createdBeaconBlock.beaconBlockHeader
    assertThat(createdBlockHeader.number).isEqualTo(1UL)
    assertThat(createdBlockHeader.round).isEqualTo(1U)
    assertThat(createdBlockHeader.timestamp).isEqualTo(blockTimestamp.toULong())
    assertThat(createdBlockHeader.proposer).isEqualTo(validator)

    // block header roots
    val stateRoot =
      HashUtil.stateRoot(
        BeaconState(
          createdBeaconBlock.beaconBlockHeader.copy(stateRoot = BeaconBlockHeader.EMPTY_HASH),
          validatorSet,
        ),
      )
    assertThat(
      createdBlockHeader.bodyRoot,
    ).isEqualTo(
      HashUtil.bodyRoot(createdBeaconBlock.beaconBlockBody),
    )
    assertThat(createdBlockHeader.stateRoot).isEqualTo(stateRoot)
    assertThat(createdBlockHeader.parentRoot).isEqualTo(parentHeader.toBeaconBlockHeader().hash())
    assertThat(
      createdBeaconBlock.beaconBlockHeader.hash(),
    ).isEqualTo(HashUtil.headerHash(createdBeaconBlock.beaconBlockHeader))

    // block body fields
    val createdBlockBody = createdBeaconBlock.beaconBlockBody
    assertThat(
      createdBlockBody.prevCommitSeals,
    ).isEqualTo(
      sealedGenesisBeaconBlock.commitSeals,
    )
    assertThat(createdBlockBody.executionPayload.timestamp).isEqualTo(blockTimestamp.toULong())
    assertThat(createdBlockBody.executionPayload.transactions).isNotEmpty
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
}
