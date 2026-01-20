/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package linea.plugin.acc.test.rpc.linea

import linea.plugin.acc.test.TestCommandLineOptionsBuilder
import linea.plugin.acc.test.rpc.ForcedTransactionParam
import linea.plugin.acc.test.rpc.GetForcedTransactionInclusionStatusRequest
import linea.plugin.acc.test.rpc.SendForcedRawTransactionRequest
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.genesis.GenesisConfigurationFactory
import org.junit.jupiter.api.Test
import org.web3j.crypto.Hash.sha3
import org.web3j.crypto.RawTransaction
import org.web3j.crypto.TransactionEncoder
import org.web3j.utils.Numeric
import java.math.BigInteger
import java.util.concurrent.TimeUnit

class ForcedTransactionTest : AbstractSendBundleTest() {

  companion object {
    // Deadline far in the future (block number 1 million)
    private const val DEFAULT_DEADLINE = "0xF4240"
  }

  override fun getTestCliOptions(): List<String> {
    return TestCommandLineOptionsBuilder()
      .set(
        "--plugin-linea-module-limit-file-path=",
        getResourcePath("/moduleLimitsLimitless.toml"),
      )
      .set("--plugin-linea-limitless-enabled=", "true")
      .build()
  }

  override fun getCliqueOptions(): GenesisConfigurationFactory.CliqueOptions {
    return GenesisConfigurationFactory.CliqueOptions(
      BLOCK_PERIOD_SECONDS,
      GenesisConfigurationFactory.CliqueOptions.DEFAULT.epochLength(),
      false,
    )
  }

  @Test
  fun singleForcedTransactionIsAcceptedAndMined() {
    val sender = accounts.secondaryBenefactor
    val recipient = accounts.primaryBenefactor

    val rawTx = createSignedTransfer(sender, recipient, 0)
    val txHash = sha3(rawTx)

    val sendResponse = SendForcedRawTransactionRequest(
      listOf(ForcedTransactionParam(rawTx, DEFAULT_DEADLINE)),
    ).execute(minerNode.nodeRequests())

    assertThat(sendResponse.hasError()).isFalse()
    assertThat(sendResponse.result).hasSize(1)
    assertThat(sendResponse.result[0]).isEqualTo(txHash)

    // Wait for mining
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(txHash))

    // Verify inclusion status
    val statusResponse = GetForcedTransactionInclusionStatusRequest(txHash)
      .execute(minerNode.nodeRequests())

    assertThat(statusResponse.hasError()).isFalse()
    assertThat(statusResponse.result).isNotNull
    assertThat(statusResponse.result!!.inclusionResult).isEqualTo("INCLUDED")
    assertThat(statusResponse.result!!.transactionHash).isEqualTo(txHash.lowercase())
  }

  @Test
  fun multipleForcedTransactionsAreAcceptedAndMined() {
    val sender = accounts.secondaryBenefactor
    val recipient = accounts.primaryBenefactor

    val rawTx1 = createSignedTransfer(sender, recipient, 0)
    val rawTx2 = createSignedTransfer(sender, recipient, 1)
    val rawTx3 = createSignedTransfer(sender, recipient, 2)
    val txHash1 = sha3(rawTx1)
    val txHash2 = sha3(rawTx2)
    val txHash3 = sha3(rawTx3)

    val sendResponse = SendForcedRawTransactionRequest(
      listOf(
        ForcedTransactionParam(rawTx1, DEFAULT_DEADLINE),
        ForcedTransactionParam(rawTx2, DEFAULT_DEADLINE),
        ForcedTransactionParam(rawTx3, DEFAULT_DEADLINE),
      ),
    ).execute(minerNode.nodeRequests())

    assertThat(sendResponse.hasError()).isFalse()
    assertThat(sendResponse.result).hasSize(3)

    // Wait for all to be mined
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(txHash1))
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(txHash2))
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(txHash3))

    // Verify all have INCLUDED status
    for (txHash in listOf(txHash1, txHash2, txHash3)) {
      val statusResponse = GetForcedTransactionInclusionStatusRequest(txHash)
        .execute(minerNode.nodeRequests())

      assertThat(statusResponse.result?.inclusionResult).isEqualTo("INCLUDED")
    }
  }

  @Test
  fun forcedTransactionWithBadNonceIsRejected() {
    val sender = accounts.secondaryBenefactor
    val recipient = accounts.primaryBenefactor

    // Create tx with very high nonce (will fail)
    val rawTx = createSignedTransfer(sender, recipient, 999)
    val txHash = sha3(rawTx)

    val sendResponse = SendForcedRawTransactionRequest(
      listOf(ForcedTransactionParam(rawTx, DEFAULT_DEADLINE)),
    ).execute(minerNode.nodeRequests())

    assertThat(sendResponse.hasError()).isFalse()
    assertThat(sendResponse.result).hasSize(1)

    // Wait for the status to be determined
    await()
      .atMost(30, TimeUnit.SECONDS)
      .pollInterval(1, TimeUnit.SECONDS)
      .untilAsserted {
        val statusResponse = GetForcedTransactionInclusionStatusRequest(txHash)
          .execute(minerNode.nodeRequests())

        assertThat(statusResponse.result).isNotNull
        assertThat(statusResponse.result!!.inclusionResult).isEqualTo("BAD_NONCE")
      }
  }

  @Test
  fun forcedTransactionWithInsufficientBalanceIsRejected() {
    // Create new account with minimal balance
    val poorSender = accounts.createAccount("poor")
    val recipient = accounts.primaryBenefactor

    // Try to send more ETH than the account has
    val rawTx = createSignedTransferWithValue(
      poorSender,
      recipient,
      0,
      BigInteger.valueOf(1_000_000_000_000_000_000L), // 1 ETH
    )
    val txHash = sha3(rawTx)

    val sendResponse = SendForcedRawTransactionRequest(
      listOf(ForcedTransactionParam(rawTx, DEFAULT_DEADLINE)),
    ).execute(minerNode.nodeRequests())

    assertThat(sendResponse.hasError()).isFalse()

    // Wait for the status to be determined
    await()
      .atMost(30, TimeUnit.SECONDS)
      .pollInterval(1, TimeUnit.SECONDS)
      .untilAsserted {
        val statusResponse = GetForcedTransactionInclusionStatusRequest(txHash)
          .execute(minerNode.nodeRequests())

        assertThat(statusResponse.result).isNotNull
        assertThat(statusResponse.result!!.inclusionResult).isEqualTo("BAD_BALANCE")
      }
  }

  @Test
  fun unknownTransactionReturnsNullStatus() {
    val unknownHash = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

    val statusResponse = GetForcedTransactionInclusionStatusRequest(unknownHash)
      .execute(minerNode.nodeRequests())

    assertThat(statusResponse.hasError()).isFalse()
    assertThat(statusResponse.result).isNull()
  }

  @Test
  fun forcedTransactionsHavePriorityOverRegularTransactions() {
    val forcedSender = accounts.secondaryBenefactor
    val regularSender = accounts.primaryBenefactor
    val recipient = accounts.createAccount("recipient")

    // Send regular transaction first (to tx pool)
    val regularTx = accountTransactions.createTransfer(regularSender, recipient, 1)
    val regularTxHash = regularTx.execute(minerNode.nodeRequests())

    // Send forced transaction after
    val forcedRawTx = createSignedTransfer(forcedSender, recipient, 0)
    val forcedTxHash = sha3(forcedRawTx)

    val sendResponse = SendForcedRawTransactionRequest(
      listOf(ForcedTransactionParam(forcedRawTx, DEFAULT_DEADLINE)),
    ).execute(minerNode.nodeRequests())

    assertThat(sendResponse.hasError()).isFalse()

    // Wait for both to be mined
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(forcedTxHash))
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(regularTxHash.toHexString()))

    // Check the block numbers - forced tx should be in an earlier or same block
    val forcedReceipt = ethTransactions.getTransactionReceipt(forcedTxHash)
      .execute(minerNode.nodeRequests())
      .orElseThrow()
    val regularReceipt = ethTransactions.getTransactionReceipt(regularTxHash.toHexString())
      .execute(minerNode.nodeRequests())
      .orElseThrow()

    // If in same block, forced should have lower transaction index
    if (forcedReceipt.blockNumber == regularReceipt.blockNumber) {
      assertThat(forcedReceipt.transactionIndex)
        .isLessThan(regularReceipt.transactionIndex)
    } else {
      assertThat(forcedReceipt.blockNumber)
        .isLessThanOrEqualTo(regularReceipt.blockNumber)
    }
  }

  /**
   * Verifies that when a valid FTX is followed by an invalid one, the invalid one
   * gets its rejection status recorded in a later block (not the same block).
   *
   * This ensures the invariant: only one invalidity proof can be generated per block.
   */
  @Test
  fun failingForcedTransactionNotInSameBlockAsSuccessfulOne() {
    val sender = accounts.secondaryBenefactor
    val recipient = accounts.primaryBenefactor

    // First tx is valid (nonce 0), second tx has bad nonce (nonce 999)
    val validRawTx = createSignedTransfer(sender, recipient, 0)
    val invalidRawTx = createSignedTransfer(sender, recipient, 999) // Bad nonce
    val validTxHash = sha3(validRawTx)
    val invalidTxHash = sha3(invalidRawTx)

    val sendResponse = SendForcedRawTransactionRequest(
      listOf(
        ForcedTransactionParam(validRawTx, DEFAULT_DEADLINE),
        ForcedTransactionParam(invalidRawTx, DEFAULT_DEADLINE),
      ),
    ).execute(minerNode.nodeRequests())

    assertThat(sendResponse.hasError()).isFalse()
    assertThat(sendResponse.result).hasSize(2)

    // Wait for valid tx to be mined
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(validTxHash))

    // Wait for invalid tx status to be determined (may take multiple blocks)
    await()
      .atMost(60, TimeUnit.SECONDS)
      .pollInterval(1, TimeUnit.SECONDS)
      .untilAsserted {
        val statusResponse = GetForcedTransactionInclusionStatusRequest(invalidTxHash)
          .execute(minerNode.nodeRequests())

        assertThat(statusResponse.result).isNotNull
        assertThat(statusResponse.result!!.inclusionResult).isEqualTo("BAD_NONCE")
      }

    // Get both statuses
    val validStatus = GetForcedTransactionInclusionStatusRequest(validTxHash)
      .execute(minerNode.nodeRequests())
    val invalidStatus = GetForcedTransactionInclusionStatusRequest(invalidTxHash)
      .execute(minerNode.nodeRequests())

    assertThat(validStatus.result).isNotNull
    assertThat(invalidStatus.result).isNotNull
    assertThat(validStatus.result!!.inclusionResult).isEqualTo("INCLUDED")

    // The failing tx must be in a different (later) block than the successful one
    val validBlockNumber = java.lang.Long.decode(validStatus.result!!.blockNumber)
    val invalidBlockNumber = java.lang.Long.decode(invalidStatus.result!!.blockNumber)

    assertThat(invalidBlockNumber)
      .withFailMessage(
        "Failing FTX (block %d) should not be in same block as successful FTX (block %d)",
        invalidBlockNumber,
        validBlockNumber,
      )
      .isGreaterThan(validBlockNumber)
  }

  /**
   * Verifies that multiple failing FTXs each get their rejection status in different blocks.
   *
   * This ensures the invariant: only one invalidity proof can be generated per block.
   */
  @Test
  fun multipleFailingForcedTransactionsAreInDifferentBlocks() {
    val sender = accounts.secondaryBenefactor
    val recipient = accounts.primaryBenefactor

    // Create 3 transactions with bad nonces (all will fail because expected nonce is 0)
    val rawTx1 = createSignedTransfer(sender, recipient, 100)
    val rawTx2 = createSignedTransfer(sender, recipient, 200)
    val rawTx3 = createSignedTransfer(sender, recipient, 300)
    val txHash1 = sha3(rawTx1)
    val txHash2 = sha3(rawTx2)
    val txHash3 = sha3(rawTx3)

    val sendResponse = SendForcedRawTransactionRequest(
      listOf(
        ForcedTransactionParam(rawTx1, DEFAULT_DEADLINE),
        ForcedTransactionParam(rawTx2, DEFAULT_DEADLINE),
        ForcedTransactionParam(rawTx3, DEFAULT_DEADLINE),
      ),
    ).execute(minerNode.nodeRequests())

    assertThat(sendResponse.hasError()).isFalse()
    assertThat(sendResponse.result).hasSize(3)

    // Wait for all statuses to be determined (may take multiple blocks)
    val txHashes = listOf(txHash1, txHash2, txHash3)
    for (txHash in txHashes) {
      await()
        .atMost(90, TimeUnit.SECONDS)
        .pollInterval(1, TimeUnit.SECONDS)
        .untilAsserted {
          val statusResponse = GetForcedTransactionInclusionStatusRequest(txHash)
            .execute(minerNode.nodeRequests())

          assertThat(statusResponse.result).isNotNull
          assertThat(statusResponse.result!!.inclusionResult).isEqualTo("BAD_NONCE")
        }
    }

    // Get all block numbers
    val blockNumbers = txHashes.map { txHash ->
      val status = GetForcedTransactionInclusionStatusRequest(txHash)
        .execute(minerNode.nodeRequests())
      java.lang.Long.decode(status.result!!.blockNumber)
    }

    // Each failing FTX must be in a different block (single invalidity proof per block constraint)
    assertThat(blockNumbers.toSet())
      .withFailMessage(
        "All failing FTXs should be in different blocks, but got blocks: %s",
        blockNumbers,
      )
      .hasSize(3)

    // They should also be in sequential order (first submitted = first processed)
    assertThat(blockNumbers[0])
      .withFailMessage("First tx should be in earliest block")
      .isLessThan(blockNumbers[1])
    assertThat(blockNumbers[1])
      .withFailMessage("Second tx should be in middle block")
      .isLessThan(blockNumbers[2])
  }

  private fun createSignedTransfer(
    sender: org.hyperledger.besu.tests.acceptance.dsl.account.Account,
    recipient: org.hyperledger.besu.tests.acceptance.dsl.account.Account,
    nonce: Int,
  ): String {
    return createSignedTransferWithValue(sender, recipient, nonce, BigInteger.valueOf(1000))
  }

  private fun createSignedTransferWithValue(
    sender: org.hyperledger.besu.tests.acceptance.dsl.account.Account,
    recipient: org.hyperledger.besu.tests.acceptance.dsl.account.Account,
    nonce: Int,
    value: BigInteger,
  ): String {
    val tx = RawTransaction.createTransaction(
      CHAIN_ID,
      BigInteger.valueOf(nonce.toLong()),
      TRANSFER_GAS_LIMIT,
      recipient.address,
      value,
      "",
      GAS_PRICE,
      GAS_PRICE.multiply(BigInteger.TEN).add(BigInteger.ONE),
    )

    return Numeric.toHexString(
      TransactionEncoder.signMessage(tx, sender.web3jCredentialsOrThrow()),
    )
  }
}
