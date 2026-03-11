/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package linea.plugin.acc.test

import linea.plugin.acc.test.tests.web3j.generated.AddressCaller
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts
import org.hyperledger.besu.tests.acceptance.dsl.transaction.NodeRequests
import org.hyperledger.besu.tests.acceptance.dsl.transaction.Transaction
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.io.TempDir
import org.web3j.crypto.Credentials
import org.web3j.protocol.core.Request
import org.web3j.tx.RawTransactionManager
import org.web3j.tx.gas.DefaultGasProvider
import org.web3j.tx.response.PollingTransactionReceiptProcessor
import java.io.IOException
import java.math.BigInteger
import java.nio.file.Files
import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.writeText

/**
 * Tests that the DenylistExecutionSelector rejects transactions which CALL a denied address
 * during EVM execution. This covers the EIP-7702 scenario where an EOA has a delegation set
 * by a prior transaction, and a subsequent transaction CALLs that EOA triggering code execution.
 */
class EIP7702RuntimeDenylistTest : LineaPluginPoSTestBase() {

  override fun getTestCliOptions(): List<String> {
    tempDenyList = tempDir.resolve("runtimeDenyList.txt")
    if (!tempDenyList.exists()) {
      Files.createFile(tempDenyList)
    }
    return TestCommandLineOptionsBuilder()
      .set("--plugin-linea-deny-list-path=", tempDenyList.toString())
      .build()
  }

  @Test
  fun transactionCallingDeniedAddressIsNotSelected() {
    val addressCaller = deployAddressCaller()
    val web3j = minerNode.nodeRequests().eth()
    val txManager = RawTransactionManager(
      web3j,
      Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY),
      CHAIN_ID,
    )

    // Deny the address of the AddressCaller contract itself so that when a tx
    // is sent TO the contract (which triggers a context enter with the contract
    // as recipient), the denylist execution selector rejects it.
    val deniedAddress = addressCaller.contractAddress
    tempDenyList.writeText(deniedAddress)
    reloadSelectorPlugin()

    // Send a tx that calls the denied address via the contract
    val txResponse = txManager.sendTransaction(
      DefaultGasProvider.GAS_PRICE,
      DefaultGasProvider.GAS_LIMIT,
      addressCaller.contractAddress,
      addressCaller.callAddress(deniedAddress).encodeFunctionCall(),
      BigInteger.ZERO,
    )

    // The tx enters the pool (no pool-level denial for caller addresses)
    assertThat(txResponse.transactionHash).isNotNull()

    // Use a canary transfer to verify a block was mined
    val canaryTxHash = accountTransactions
      .createTransfer(accounts.secondaryBenefactor, accounts.secondaryBenefactor, 1)
      .execute(minerNode.nodeRequests())
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(canaryTxHash.toHexString()))

    // The denied tx should NOT be mined (rejected by DenylistExecutionSelector)
    minerNode.verify(eth.expectNoTransactionReceipt(txResponse.transactionHash))

    // It should be discarded from the pool
    assertTransactionNotInThePool(txResponse.transactionHash)
  }

  @Test
  fun transactionCallingNonDeniedAddressIsSelected() {
    val addressCaller = deployAddressCaller()
    val web3j = minerNode.nodeRequests().eth()
    val txManager = RawTransactionManager(
      web3j,
      Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY),
      CHAIN_ID,
    )

    // Deny some unrelated address
    tempDenyList.writeText("0x0000000000000000000000000000000000099999")
    reloadSelectorPlugin()

    // Send a tx that calls a non-denied address
    val targetAddress = accounts.secondaryBenefactor.address
    val txResponse = txManager.sendTransaction(
      DefaultGasProvider.GAS_PRICE,
      DefaultGasProvider.GAS_LIMIT,
      addressCaller.contractAddress,
      addressCaller.callAddress(targetAddress).encodeFunctionCall(),
      BigInteger.ZERO,
    )

    assertThat(txResponse.transactionHash).isNotNull()
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(txResponse.transactionHash))
  }

  private fun deployAddressCaller(): AddressCaller {
    val web3j = minerNode.nodeRequests().eth()
    val credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY)
    val receiptProcessor = PollingTransactionReceiptProcessor(
      web3j,
      1000L,
      BLOCK_PERIOD_SECONDS * 6,
    )
    val txManager = RawTransactionManager(web3j, credentials, CHAIN_ID, receiptProcessor)
    return AddressCaller.deploy(web3j, txManager, DefaultGasProvider()).send()
  }

  private fun reloadSelectorPlugin() {
    val request = ReloadPluginConfigRequest()
    val result = request.execute(minerNode.nodeRequests())
    assertThat(result).isEqualTo("Success")
  }

  class ReloadPluginConfigRequest : Transaction<String> {
    override fun execute(nodeRequests: NodeRequests): String {
      return try {
        Request<Any, ReloadPluginConfigResponse>(
          "plugins_reloadPluginConfig",
          listOf("net.consensys.linea.sequencer.txselection.LineaTransactionSelectorPlugin"),
          nodeRequests.web3jService,
          ReloadPluginConfigResponse::class.java,
        )
          .send()
          .result
      } catch (e: IOException) {
        throw RuntimeException("Failed to reload plugin configuration", e)
      }
    }
  }

  class ReloadPluginConfigResponse : org.web3j.protocol.core.Response<String>()

  companion object {
    @JvmStatic
    @TempDir
    lateinit var tempDir: Path

    @JvmStatic
    lateinit var tempDenyList: Path
  }
}