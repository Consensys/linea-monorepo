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

import java.io.File
import java.math.BigInteger
import java.nio.file.Files
import maru.consensus.ElFork
import maru.consensus.qbft.network.NoopValidatorMulticaster
import maru.consensus.qbft.toAddress
import maru.crypto.Crypto
import maru.extensions.fromHexToByteArray
import maru.testutils.MaruFactory
import maru.testutils.SpyingValidatorMulticaster
import maru.testutils.TransactionsHelper
import maru.testutils.besu.BesuFactory
import org.apache.logging.log4j.LogManager
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.kotlin.await
import org.hyperledger.besu.consensus.qbft.core.messagewrappers.Commit
import org.hyperledger.besu.consensus.qbft.core.messagewrappers.Prepare
import org.hyperledger.besu.consensus.qbft.core.messagewrappers.Proposal
import org.hyperledger.besu.consensus.qbft.core.messagewrappers.RoundChange
import org.hyperledger.besu.tests.acceptance.dsl.account.Account
import org.hyperledger.besu.tests.acceptance.dsl.blockchain.Amount
import org.hyperledger.besu.tests.acceptance.dsl.condition.net.NetConditions
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode
import org.hyperledger.besu.tests.acceptance.dsl.node.ThreadBesuNodeRunner
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.Cluster
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.ClusterConfigurationBuilder
import org.hyperledger.besu.tests.acceptance.dsl.transaction.net.NetTransactions
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.web3j.protocol.core.DefaultBlockParameter

class MaruQbftTest {
  private lateinit var cluster: Cluster
  private lateinit var besuNode: BesuNode
  private lateinit var maruNode: MaruApp
  private lateinit var transactionsHelper: TransactionsHelper
  private val log = LogManager.getLogger(this.javaClass)
  private lateinit var tmpDir: File
  private lateinit var spyingValidatorMulticaster: SpyingValidatorMulticaster

  @BeforeEach
  fun setUp() {
    val elFork = ElFork.Prague
    transactionsHelper = TransactionsHelper()
    besuNode = BesuFactory.buildTestBesu(elFork)
    cluster =
      Cluster(
        ClusterConfigurationBuilder().build(),
        NetConditions(NetTransactions()),
        ThreadBesuNodeRunner(),
      )
    cluster.start(besuNode)
    val ethereumJsonRpcBaseUrl = besuNode.jsonRpcBaseUrl().get()
    val engineRpcUrl = besuNode.engineRpcUrl().get()
    tmpDir = Files.createTempDirectory("maru").toFile()
    tmpDir.deleteOnExit()
    spyingValidatorMulticaster = SpyingValidatorMulticaster(NoopValidatorMulticaster)
    maruNode =
      MaruFactory.buildTestMaru(
        ethereumJsonRpcUrl = ethereumJsonRpcBaseUrl,
        engineApiRpc = engineRpcUrl,
        elFork = elFork,
        dataDir = tmpDir.toPath(),
        validatorMulticaster = spyingValidatorMulticaster,
      )
    maruNode.start()
  }

  @AfterEach
  fun tearDown() {
    cluster.close()
    maruNode.stop()
    tmpDir.deleteRecursively()
  }

  private fun sendTransactionAndAssertExecution(
    recipient: Account,
    amount: Amount,
  ) {
    val transfer = transactionsHelper.createTransfer(recipient, amount)
    val txHash = besuNode.execute(transfer)
    assertThat(txHash).isNotNull()
    log.info("Sending transaction {}, transaction data ", txHash)
    transactionsHelper.ethConditions.expectSuccessfulTransactionReceipt(txHash.toString()).verify(besuNode)
    log.info("Transaction {} was mined", txHash)
  }

  @Test
  fun `Maru is producing blocks with expected block time and emits messages`() {
    val blocksToProduce = 10
    repeat(blocksToProduce) {
      sendTransactionAndAssertExecution(transactionsHelper.createAccount("another account"), Amount.ether(100))
    }

    verifyBlockHeaders(fromBlockNumber = 1, blocksToProduce)

    // Need to wait because otherwise not all of the messages might be emitted at the time of a block being mined
    await.untilAsserted {
      validateRoundChange(maxBlockNumber = blocksToProduce.toLong())
    }

    for (blockNumber in 1L..blocksToProduce) {
      assertThat(
        anyPrepareWithBlockNumber(blockNumber),
      ).withFailMessage { "Didn't find any prepare messages for blockNumber=$blockNumber" }
        .isTrue
      assertThat(
        anyCommitWithBlockNumber(blockNumber),
      ).withFailMessage { "Didn't find any commit messages for blockNumber=$blockNumber" }
        .isTrue
      assertThat(
        anyProposalWithBlockNumber(blockNumber),
      ).withFailMessage { "Didn't find any proposal messages for blockNumber=$blockNumber" }
        .isTrue
    }
    allMessagesAreSignedByTheExpectedSigner()
  }

  private fun anyPrepareWithBlockNumber(blockNumber: Long): Boolean =
    spyingValidatorMulticaster.emittedMessages.any {
      it is Prepare && it.roundIdentifier.sequenceNumber == blockNumber
    }

  private fun anyProposalWithBlockNumber(blockNumber: Long): Boolean =
    spyingValidatorMulticaster.emittedMessages.any {
      it is Proposal && it.roundIdentifier.sequenceNumber == blockNumber
    }

  private fun anyCommitWithBlockNumber(blockNumber: Long): Boolean =
    spyingValidatorMulticaster.emittedMessages.any {
      it is Commit && it.roundIdentifier.sequenceNumber == blockNumber
    }

  // All blocks except the first must be produced within 0th round. Absence of transactions will trigger RoundChange
  // events post test
  private fun validateRoundChange(maxBlockNumber: Long) {
    val roundChangeMessages =
      spyingValidatorMulticaster.emittedMessages.filter {
        it is RoundChange
      }
    assertThat(roundChangeMessages).isNotEmpty()
    roundChangeMessages.forEach { roundChange ->
      assertThat(
        roundChange.roundIdentifier.sequenceNumber == 1L ||
          roundChange.roundIdentifier.sequenceNumber > maxBlockNumber,
      ).withFailMessage { "Unexpected RoundChange! $roundChange" }
        .isTrue
    }
  }

  private fun allMessagesAreSignedByTheExpectedSigner() {
    val validatorAddress =
      Crypto
        .privateKeyToValidator(MaruFactory.VALIDATOR_PRIVATE_KEY.fromHexToByteArray())
        .toAddress()
    spyingValidatorMulticaster.emittedMessages.forEach {
      assertThat(it.author)
        .withFailMessage { "Unexpected signer address for message=$it author=${it.author}" }
        .isEqualTo(validatorAddress)
    }
  }

  @Test
  fun `Maru works if Besu stops mid flight`() {
    val blocksToProduce = 5
    repeat(blocksToProduce) {
      sendTransactionAndAssertExecution(transactionsHelper.createAccount("another account"), Amount.ether(100))
    }
    cluster.stop()
    Thread.sleep(3000)
    cluster.start(besuNode)

    repeat(blocksToProduce) {
      sendTransactionAndAssertExecution(transactionsHelper.createAccount("another account"), Amount.ether(100))
    }

    verifyBlockHeaders(fromBlockNumber = 1, blocksToProduce)
    verifyBlockHeaders(fromBlockNumber = 6, blocksToProduce)
  }

  @Test
  fun `Maru works after restart`() {
    val blocksToProduce = 5
    repeat(blocksToProduce) {
      sendTransactionAndAssertExecution(transactionsHelper.createAccount("another account"), Amount.ether(100))
    }
    maruNode.stop()
    Thread.sleep(3)
    maruNode.start()
    repeat(blocksToProduce) {
      sendTransactionAndAssertExecution(transactionsHelper.createAccount("another account"), Amount.ether(100))
    }

    verifyBlockHeaders(fromBlockNumber = 1, blocksToProduce)
    verifyBlockHeaders(fromBlockNumber = 6, blocksToProduce)
  }

  @Test
  fun `Maru works after recreation`() {
    val blocksToProduce = 5
    repeat(blocksToProduce) {
      sendTransactionAndAssertExecution(transactionsHelper.createAccount("another account"), Amount.ether(100))
    }
    maruNode.stop()
    maruNode.close()

    Thread.sleep(3)
    maruNode =
      MaruFactory.buildTestMaru(
        ethereumJsonRpcUrl = besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = besuNode.engineRpcUrl().get(),
        elFork = ElFork.Prague,
        dataDir = tmpDir.toPath(),
      )
    // The difference from the previous test is that BeaconChain is instantiated with the Maru instance and it's not
    // affected by start and stop
    maruNode.start()
    repeat(blocksToProduce) {
      sendTransactionAndAssertExecution(transactionsHelper.createAccount("another account"), Amount.ether(100))
    }

    verifyBlockHeaders(fromBlockNumber = 1, blocksToProduce)
    verifyBlockHeaders(fromBlockNumber = 6, blocksToProduce)
  }

  private fun verifyBlockHeaders(
    fromBlockNumber: Int,
    blocksProduced: Int,
  ) {
    val blocks =
      (fromBlockNumber until fromBlockNumber + blocksProduced)
        .map {
          besuNode
            .nodeRequests()
            .eth()
            .ethGetBlockByNumber(
              DefaultBlockParameter.valueOf(BigInteger.valueOf(it.toLong())),
              false,
            ).sendAsync()
        }.map { it.get().block }

    val blockTimeSeconds = 1L
    val timestampsSeconds = blocks.map { it.timestamp.toLong() }
    (2.until(blocks.size)).forEach {
      assertThat(timestampsSeconds[it - 1]).isLessThan(timestampsSeconds[it])
      val actualBlockTime = timestampsSeconds[it] - timestampsSeconds[it - 1]
      assertThat(actualBlockTime)
        .withFailMessage("Timestamps: $timestampsSeconds")
        .isGreaterThanOrEqualTo(blockTimeSeconds)
      assertThat(actualBlockTime)
        .withFailMessage("Timestamps: $timestampsSeconds")
        .isLessThanOrEqualTo(blockTimeSeconds)
    }
  }
}
