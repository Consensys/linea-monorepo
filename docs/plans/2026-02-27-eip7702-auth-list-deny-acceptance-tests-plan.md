# EIP-7702 Authorization List Deny Check - Acceptance Tests Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add acceptance tests that verify the EIP-7702 authorization list deny check against a real Besu node with Prague fork configuration.

**Architecture:** Two new Kotlin test classes extending `LineaPluginPoSTestBase`. One tests the 3 core scenarios (denied authority, denied address, clean passes), the other tests deny list hot-reload for auth list entries. Both construct real EIP-7702 transactions with separate authority/sender key pairs using Besu's `CodeDelegation.builder()`.

**Tech Stack:** Kotlin, JUnit 5, AssertJ, Besu acceptance test framework, Web3j, SECP256K1

---

### Task 1: Create EIP7702AuthListDenyTest with 3 core test cases

**Files:**
- Create: `besu-plugins/linea-sequencer/acceptance-tests/src/test/kotlin/linea/plugin/acc/test/EIP7702AuthListDenyTest.kt`

**Context:**
- The test extends `LineaPluginPoSTestBase()` which starts a real Besu node with Prague fork
- `denyList.txt` resource contains `0x627306090abab3a6e1400e9345bc60c78a8bef57` (GENESIS_ACCOUNT_TWO)
- `Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY` = not denied, `Accounts.GENESIS_ACCOUNT_TWO_PRIVATE_KEY` = denied
- The existing `sendRawEIP7702Transaction()` in the base class constructs a CodeDelegation signed by the sender - but we need the authority to differ from the sender for Path 1

**Key references (read these first):**
- `LineaPluginPoSTestBase.kt` lines 269-309: existing `sendRawEIP7702Transaction()` - shows how to build CodeDelegation and Transaction
- `TransactionPoolDenialTest.kt` lines 24-28: how to override CLI options with deny list
- `TestCommandLineOptionsBuilder.kt`: how CLI options are built
- `EIP7702TransactionDenialTest.kt`: existing EIP-7702 test pattern

**Step 1: Write the complete test class**

```kotlin
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

/**
 * Tests that verify the DeniedAddressValidator correctly rejects EIP-7702 DELEGATE_CODE
 * transactions when the authorization list contains denied authorities or addresses.
 */
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
```

**Step 2: Verify it compiles**

Run: `./gradlew :besu-plugins:linea-sequencer:acceptance-tests:compileTestKotlin`
(May fail due to go-corset - if so, verify syntax manually against the existing patterns)

**Step 3: Commit**

```bash
git add besu-plugins/linea-sequencer/acceptance-tests/src/test/kotlin/linea/plugin/acc/test/EIP7702AuthListDenyTest.kt
git commit -m "test: add EIP-7702 auth list deny acceptance tests

Three acceptance tests against a real Besu node:
- Denied authority (Path 1 Puppet bypass) is rejected
- Denied delegation address (Path 2 Parasite bypass) is rejected  
- Clean authorization list is accepted"
```

---

### Task 2: Create EIP7702AuthListDenyReloadTest for hot-reload

**Files:**
- Create: `besu-plugins/linea-sequencer/acceptance-tests/src/test/kotlin/linea/plugin/acc/test/EIP7702AuthListDenyReloadTest.kt`

**Context:**
- This test needs a SEPARATE class because it uses a temp deny list file (starts empty, then adds an address)
- Follow the pattern from `TransactionPoolDenyListReloadTest.kt` for: temp file setup, `reloadPluginConfig()`, and deny list manipulation
- Must also enable `--plugin-linea-delegate-code-tx-enabled=true`

**Key references:**
- `TransactionPoolDenyListReloadTest.kt` lines 30-157: temp deny list setup, reload mechanism, companion object with @TempDir

**Step 1: Write the complete test class**

```kotlin
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
import org.hyperledger.besu.tests.acceptance.dsl.transaction.NodeRequests
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.io.TempDir
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameterName
import org.web3j.protocol.core.Request
import org.web3j.protocol.core.methods.response.EthSendTransaction
import org.web3j.tx.gas.DefaultGasProvider
import java.io.IOException
import java.math.BigInteger
import java.nio.file.Files
import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.writeText

/**
 * Tests that verify the DeniedAddressValidator correctly rejects EIP-7702 authorization list
 * entries after deny list hot-reload.
 */
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

    // With empty deny list, tx with willBeDenied as authority should pass
    val responseBeforeReload = sendEIP7702WithSeparateAuth(
      senderCredentials = sender,
      authSignerCredentials = willBeDenied,
      delegationAddress = Address.fromHexStringStrict(sender.address),
    )
    assertThat(responseBeforeReload.error).isNull()
    assertThat(responseBeforeReload.transactionHash).isNotNull()

    // Add willBeDenied to deny list and reload
    tempDenyList.writeText(willBeDenied.address)
    reloadPluginConfig()

    // Now the same pattern should be rejected
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

  private fun reloadPluginConfig() {
    val reqLinea = ReloadPluginConfigRequest()
    val respLinea = reqLinea.execute(minerNode.nodeRequests())
    assertThat(respLinea).isEqualTo("Success")
  }

  class ReloadPluginConfigRequest : org.hyperledger.besu.tests.acceptance.dsl.transaction.Transaction<String> {
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

  class ReloadPluginConfigResponse : org.web3j.protocol.core.Response<String>()

  companion object {
    @JvmStatic
    @TempDir
    lateinit var tempDir: Path

    @JvmStatic
    lateinit var tempDenyList: Path
  }
}
```

**Step 2: Verify it compiles**

Run: `./gradlew :besu-plugins:linea-sequencer:acceptance-tests:compileTestKotlin`

**Step 3: Commit**

```bash
git add besu-plugins/linea-sequencer/acceptance-tests/src/test/kotlin/linea/plugin/acc/test/EIP7702AuthListDenyReloadTest.kt
git commit -m "test: add EIP-7702 auth list deny reload acceptance test

Verifies that deny list hot-reload correctly catches authorization
list authority entries that were added after initial startup."
```
