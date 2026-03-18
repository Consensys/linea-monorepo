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
import linea.plugin.acc.test.tests.web3j.generated.Eip7702TestEntrypoint
import linea.plugin.acc.test.tests.web3j.generated.Eip77022Delegated
import linea.plugin.acc.test.tests.web3j.generated.Eip7702TestNested
import net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.hyperledger.besu.datatypes.Address
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
import java.util.concurrent.TimeUnit
import kotlin.io.path.exists
import kotlin.io.path.writeText

/**
 * Tests that the DenylistExecutionSelector rejects transactions which invoke a denied address
 * during EVM execution via CALL, DELEGATECALL, STATICCALL, and CALLCODE opcodes.
 * This covers the EIP-7702 scenario where an EOA has a delegation set by a prior transaction,
 * and a subsequent transaction invokes that EOA triggering code execution.
 */
class EIP7702RuntimeDenylistTest : LineaPluginPoSTestBase() {

  override fun getRequestedPlugins(): List<String> =
    DEFAULT_REQUESTED_PLUGINS + "RecordingTransactionSelectorPlugin"

  override fun getAdditionalRpcApis(): Set<String> = setOf("TEST")

  override fun getTestCliOptions(): List<String> {
    tempDenyList = tempDir.resolve("runtimeDenyList.txt")
    if (!tempDenyList.exists()) {
      Files.createFile(tempDenyList)
    }
    return TestCommandLineOptionsBuilder()
      .set("--plugin-linea-deny-list-path=", tempDenyList.toString())
      .set("--plugin-linea-delegate-code-tx-enabled=", "true")
      .build()
  }

  @Test
  fun transactionCallingDeniedAddressIsNotSelected() {
    // Deploy all contracts upfront before any denylist rejections,
    // to avoid nonce gaps from rejected transactions affecting deployments
    val callerContract = deployAddressCaller()
    val targetContract = deployAddressCaller()
    val deniedAddress = targetContract.contractAddress

    tempDenyList.writeText(deniedAddress)
    reloadSelectorPlugin()

    // Test all call-type opcodes: CALL, DELEGATECALL, STATICCALL, CALLCODE
    val callFunctions = listOf(
      "CALL" to callerContract.callAddress(deniedAddress).encodeFunctionCall(),
      "DELEGATECALL" to callerContract.delegateCallAddress(deniedAddress).encodeFunctionCall(),
      "STATICCALL" to callerContract.staticCallAddress(deniedAddress).encodeFunctionCall(),
      "CALLCODE" to callerContract.callCodeAddress(deniedAddress).encodeFunctionCall(),
    )

    for ((opcode, encodedCall) in callFunctions) {
      val web3j = minerNode.nodeRequests().eth()
      val txManager = RawTransactionManager(
        web3j,
        Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY),
        CHAIN_ID,
      )

      val txResponse = txManager.sendTransaction(
        DefaultGasProvider.GAS_PRICE,
        DefaultGasProvider.GAS_LIMIT,
        callerContract.contractAddress,
        encodedCall,
        BigInteger.ZERO,
      )

      // The tx enters the pool (tx.to is the caller contract, not denied)
      assertThat(txResponse.transactionHash)
        .withFailMessage { "$opcode: tx hash should not be null" }
        .isNotNull()

      // Use a canary transfer to verify a block was mined
      val canaryTxHash = accountTransactions
        .createTransfer(accounts.secondaryBenefactor, accounts.secondaryBenefactor, 1)
        .execute(minerNode.nodeRequests())
      minerNode.verify(eth.expectSuccessfulTransactionReceipt(canaryTxHash.toHexString()))

      // The denied tx should NOT be mined (rejected by DenylistExecutionSelector)
      minerNode.verify(eth.expectNoTransactionReceipt(txResponse.transactionHash))

      // Verify the exact rejection reason via the recording plugin
      await()
        .atMost(4, TimeUnit.SECONDS)
        .pollInterval(50, TimeUnit.MILLISECONDS)
        .untilAsserted {
          assertThat(getRejectionReason(txResponse.transactionHash))
            .withFailMessage { "$opcode: Expected tx to be rejected with TX_FILTERED_ADDRESS_CALLED" }
            .isEqualTo(LineaTransactionSelectionResult.TX_FILTERED_ADDRESS_CALLED.toString())
        }

      // It should be discarded from the pool
      assertTransactionNotInThePool(txResponse.transactionHash)
    }
  }

  @Test
  fun nestedDelegationCallToDeniedAddressIsDetected() {
    val web3j = minerNode.nodeRequests().eth()
    val relayer = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY)

    val actor = Credentials.create(org.web3j.crypto.Keys.createEcKeyPair())

    // Deploy all 3 contracts with the relayer before any denylist changes
    val entrypoint = deployEip7702TestEntrypoint()
    val delegated = deployEip77022Delegated()
    val nested = deployEip7702TestNested()

    // Fund actor so the delegated code can send 0.001 ETH to nested contract
    val fundingManager = RawTransactionManager(web3j, relayer, CHAIN_ID)
    val fundingTx = fundingManager.sendTransaction(
      DefaultGasProvider.GAS_PRICE,
      DefaultGasProvider.GAS_LIMIT,
      actor.address,
      "",
      BigInteger("1000000000000000000"), // 1 ETH
    )
    val fundingProcessor = PollingTransactionReceiptProcessor(web3j, 1000L, BLOCK_PERIOD_SECONDS * 6)
    fundingProcessor.waitForTransactionReceipt(fundingTx.transactionHash)

    // Relayer sends EIP-7702 delegation tx on behalf of actor
    val delegationResponse = sendEIP7702WithSeparateAuth(
      web3j = web3j,
      senderCredentials = relayer,
      authSignerCredentials = actor,
      delegationAddress = Address.fromHexStringStrict(delegated.contractAddress),
    )
    assertThat(delegationResponse.error).isNull()
    assertThat(delegationResponse.transactionHash).isNotNull()

    // Wait for delegation tx to be mined
    val receiptProcessor = PollingTransactionReceiptProcessor(web3j, 1000L, BLOCK_PERIOD_SECONDS * 6)
    receiptProcessor.waitForTransactionReceipt(delegationResponse.transactionHash)

    // Add Actor's address to denylist and reload
    tempDenyList.writeText(actor.address)
    reloadSelectorPlugin()

    // Relayer calls Entrypoint.setValue(42, actorAddress, nestedAddress)
    // This triggers: Entrypoint -> CALL Actor (denied, has Delegated code) -> CALL Nested
    val txManager = RawTransactionManager(web3j, relayer, CHAIN_ID)
    val encodedCall = entrypoint.setValue(
      BigInteger.valueOf(42),
      actor.address,
      nested.contractAddress,
    ).encodeFunctionCall()

    val txResponse = txManager.sendTransaction(
      DefaultGasProvider.GAS_PRICE,
      DefaultGasProvider.GAS_LIMIT,
      entrypoint.contractAddress,
      encodedCall,
      BigInteger.ZERO,
    )

    assertThat(txResponse.transactionHash)
      .withFailMessage { "tx hash should not be null — tx should enter pool since tx.to is Entrypoint" }
      .isNotNull()

    // Use a canary transfer to verify a block was mined
    val canaryTxHash = accountTransactions
      .createTransfer(accounts.secondaryBenefactor, accounts.secondaryBenefactor, 1)
      .execute(minerNode.nodeRequests())
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(canaryTxHash.toHexString()))

    // The tx should NOT be mined (rejected by DenylistExecutionSelector)
    minerNode.verify(eth.expectNoTransactionReceipt(txResponse.transactionHash))

    // Verify the exact rejection reason via the recording plugin
    await()
      .atMost(4, TimeUnit.SECONDS)
      .pollInterval(50, TimeUnit.MILLISECONDS)
      .untilAsserted {
        assertThat(getRejectionReason(txResponse.transactionHash))
          .withFailMessage { "Expected tx to be rejected with TX_FILTERED_ADDRESS_CALLED" }
          .isEqualTo(LineaTransactionSelectionResult.TX_FILTERED_ADDRESS_CALLED.toString())
      }

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

  private fun deployEip7702TestEntrypoint(): Eip7702TestEntrypoint {
    val web3j = minerNode.nodeRequests().eth()
    val credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY)
    val receiptProcessor = PollingTransactionReceiptProcessor(
      web3j,
      1000L,
      BLOCK_PERIOD_SECONDS * 6,
    )
    val txManager = RawTransactionManager(web3j, credentials, CHAIN_ID, receiptProcessor)
    return Eip7702TestEntrypoint.deploy(web3j, txManager, DefaultGasProvider()).send()
  }

  private fun deployEip77022Delegated(): Eip77022Delegated {
    val web3j = minerNode.nodeRequests().eth()
    val credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY)
    val receiptProcessor = PollingTransactionReceiptProcessor(
      web3j,
      1000L,
      BLOCK_PERIOD_SECONDS * 6,
    )
    val txManager = RawTransactionManager(web3j, credentials, CHAIN_ID, receiptProcessor)
    return Eip77022Delegated.deploy(web3j, txManager, DefaultGasProvider()).send()
  }

  private fun deployEip7702TestNested(): Eip7702TestNested {
    val web3j = minerNode.nodeRequests().eth()
    val credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY)
    val receiptProcessor = PollingTransactionReceiptProcessor(
      web3j,
      1000L,
      BLOCK_PERIOD_SECONDS * 6,
    )
    val txManager = RawTransactionManager(web3j, credentials, CHAIN_ID, receiptProcessor)
    return Eip7702TestNested.deploy(web3j, txManager, DefaultGasProvider()).send()
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
