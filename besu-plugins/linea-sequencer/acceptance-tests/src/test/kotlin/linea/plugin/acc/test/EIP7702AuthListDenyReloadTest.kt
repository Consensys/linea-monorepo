/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package linea.plugin.acc.test

import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.crypto.SECP256K1
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.TransactionType
import org.hyperledger.besu.datatypes.Wei
import org.hyperledger.besu.ethereum.core.CodeDelegation
import org.hyperledger.besu.ethereum.core.Transaction
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts
import org.hyperledger.besu.tests.acceptance.dsl.transaction.NodeRequests
import org.hyperledger.besu.tests.acceptance.dsl.transaction.Transaction as DslTransaction
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.io.TempDir
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameterName
import org.web3j.protocol.core.Request
import org.web3j.protocol.core.Response
import org.web3j.protocol.core.methods.response.EthSendTransaction
import org.web3j.tx.gas.DefaultGasProvider
import java.io.IOException
import java.math.BigInteger
import java.nio.file.Files
import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.writeText

class EIP7702AuthListDenyReloadTest : LineaPluginPoSTestBase() {
  private lateinit var web3j: Web3j
  private val secp256k1 = SECP256K1()

  override fun getTestCliOptions(): List<String> {
    tempDenyList = tempDir.resolve("denyList.txt")
    if (!tempDenyList.exists()) {
      Files.createFile(tempDenyList)
    }
    return TestCommandLineOptionsBuilder()
      .set("--plugin-linea-deny-list-path=", tempDenyList.toString())
      .set("--plugin-linea-delegate-code-tx-enabled=", "true")
      .build()
  }

  @BeforeEach
  override fun setup() {
    super.setup()
    web3j = minerNode.nodeRequests().eth()
  }

  @Test
  fun eip7702AuthListDenyCheckWorksAfterReload() {
    val sender = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY)
    val willBeDenied = Credentials.create(Accounts.GENESIS_ACCOUNT_TWO_PRIVATE_KEY)

    val responseBeforeReload = sendEIP7702WithSeparateAuth(
      senderCredentials = sender,
      authSignerCredentials = willBeDenied,
      delegationAddress = Address.fromHexStringStrict(sender.address),
    )
    assertThat(responseBeforeReload.error).isNull()
    assertThat(responseBeforeReload.transactionHash).isNotNull()

    tempDenyList.writeText(willBeDenied.address)
    reloadPluginConfig()

    val responseAfterReload = sendEIP7702WithSeparateAuth(
      senderCredentials = sender,
      authSignerCredentials = willBeDenied,
      delegationAddress = Address.fromHexStringStrict(sender.address),
    )
    assertThat(responseAfterReload.transactionHash).isNull()
    assertThat(responseAfterReload.error.message).contains(
      "authorization authority ${willBeDenied.address} is blocked",
    )
  }

  private fun sendEIP7702WithSeparateAuth(
    senderCredentials: Credentials,
    authSignerCredentials: Credentials,
    delegationAddress: Address,
  ): EthSendTransaction {
    val nonce = web3j
      .ethGetTransactionCount(senderCredentials.address, DefaultBlockParameterName.PENDING)
      .send()
      .transactionCount

    val codeDelegation = CodeDelegation.builder()
      .chainId(BigInteger.valueOf(CHAIN_ID))
      .address(delegationAddress)
      .nonce(0)
      .signAndBuild(
        secp256k1.createKeyPair(
          secp256k1.createPrivateKey(authSignerCredentials.ecKeyPair.privateKey),
        ),
      )

    val tx = Transaction.builder()
      .type(TransactionType.DELEGATE_CODE)
      .chainId(BigInteger.valueOf(CHAIN_ID))
      .nonce(nonce.toLong())
      .maxPriorityFeePerGas(Wei.of(DefaultGasProvider.GAS_PRICE))
      .maxFeePerGas(Wei.of(DefaultGasProvider.GAS_PRICE))
      .gasLimit(DefaultGasProvider.GAS_LIMIT.toLong())
      .to(Address.fromHexStringStrict(senderCredentials.address))
      .value(Wei.ZERO)
      .payload(Bytes.EMPTY)
      .accessList(emptyList())
      .codeDelegations(listOf(codeDelegation))
      .signAndBuild(
        secp256k1.createKeyPair(
          secp256k1.createPrivateKey(senderCredentials.ecKeyPair.privateKey),
        ),
      )

    return web3j.ethSendRawTransaction(tx.encoded().toHexString()).send()
  }

  private fun reloadPluginConfig() {
    val reqLinea = ReloadPluginConfigRequest()
    val respLinea = reqLinea.execute(minerNode.nodeRequests())
    assertThat(respLinea).isEqualTo("Success")
  }

  class ReloadPluginConfigRequest : DslTransaction<String> {
    override fun execute(nodeRequests: NodeRequests): String {
      return try {
        Request<Any, ReloadPluginConfigResponse>(
          "plugins_reloadPluginConfig",
          listOf("net.consensys.linea.sequencer.txpoolvalidation.LineaTransactionPoolValidatorPlugin"),
          nodeRequests.web3jService,
          ReloadPluginConfigResponse::class.java,
        )
          .send()
          .result
      } catch (e: IOException) {
        throw RuntimeException(e)
      }
    }
  }

  class ReloadPluginConfigResponse : Response<String>()

  companion object {
    @JvmStatic
    @TempDir
    lateinit var tempDir: Path

    @JvmStatic
    lateinit var tempDenyList: Path
  }
}
