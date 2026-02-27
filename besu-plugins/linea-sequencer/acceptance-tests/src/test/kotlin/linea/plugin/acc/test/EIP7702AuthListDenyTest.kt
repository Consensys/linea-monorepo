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
import org.hyperledger.besu.ethereum.core.Transaction
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameterName
import org.web3j.protocol.core.methods.response.EthSendTransaction
import org.web3j.tx.gas.DefaultGasProvider
import java.math.BigInteger

class EIP7702AuthListDenyTest : LineaPluginPoSTestBase() {
  private lateinit var web3j: Web3j
  private val secp256k1 = SECP256K1()
  private val deniedAddress = "0x627306090abab3a6e1400e9345bc60c78a8bef57"

  override fun getTestCliOptions(): List<String> {
    return TestCommandLineOptionsBuilder()
      .set("--plugin-linea-deny-list-path=", getResourcePath("/denyList.txt"))
      .set("--plugin-linea-delegate-code-tx-enabled=", "true")
      .build()
  }

  @BeforeEach
  override fun setup() {
    super.setup()
    web3j = minerNode.nodeRequests().eth()
  }

  @Test
  fun eip7702TxWithDeniedAuthorityIsRejected() {
    val sender = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY)
    val deniedAuthSigner = Credentials.create(Accounts.GENESIS_ACCOUNT_TWO_PRIVATE_KEY)

    val response = sendEIP7702WithSeparateAuth(
      senderCredentials = sender,
      authSignerCredentials = deniedAuthSigner,
      delegationAddress = Address.fromHexStringStrict(sender.address),
    )

    assertThat(response.transactionHash).isNull()
    assertThat(response.error.message).isEqualTo(
      "authorization authority $deniedAddress is blocked as appearing on " +
        "the SDN or other legally prohibited list",
    )
  }

  @Test
  fun eip7702TxWithDeniedDelegationAddressIsRejected() {
    val sender = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY)

    val response = sendEIP7702WithSeparateAuth(
      senderCredentials = sender,
      authSignerCredentials = sender,
      delegationAddress = Address.fromHexStringStrict(deniedAddress),
    )

    assertThat(response.transactionHash).isNull()
    assertThat(response.error.message).isEqualTo(
      "authorization address $deniedAddress is blocked as appearing on " +
        "the SDN or other legally prohibited list",
    )
  }

  @Test
  fun eip7702TxWithCleanAuthListIsAccepted() {
    val sender = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY)

    val response = sendEIP7702WithSeparateAuth(
      senderCredentials = sender,
      authSignerCredentials = sender,
      delegationAddress = Address.fromHexStringStrict(sender.address),
    )

    assertThat(response.error).isNull()
    assertThat(response.transactionHash).isNotNull()
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

    val codeDelegation = org.hyperledger.besu.ethereum.core.CodeDelegation.builder()
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
}
