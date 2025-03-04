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
package maru.e2e

import java.math.BigInteger
import java.util.Optional
import java.util.UUID
import java.util.concurrent.TimeUnit
import kotlin.io.path.Path
import kotlin.time.Duration.Companion.minutes
import kotlin.time.toJavaDuration
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.kotlin.await
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.methods.response.EthSendTransaction
import org.web3j.tx.RawTransactionManager
import tech.pegasys.teku.ethereum.executionclient.auth.JwtConfig
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3jClientBuilder
import tech.pegasys.teku.infrastructure.time.SystemTimeProvider

object TestEnvironment {
  val jwtConfig: Optional<JwtConfig> =
    JwtConfig.createIfNeeded(
      true,
      Optional.of("../docker/jwt"),
      Optional.of(UUID.randomUUID().toString()),
      Path("/tmp"),
    )
  val sequencerL2Client: Web3j = buildWeb3Client("http://localhost:8545")

  // The switch doesn't work for Geth 1.14 yet
  val geth1L2Client: Web3j = buildWeb3Client("http://localhost:8555")
  val geth2L2Client: Web3j = buildWeb3Client("http://localhost:8565")
  val gethSnapServerL2Client: Web3j = buildWeb3Client("http://localhost:8575")
  val besuFollowerL2Client: Web3j = buildWeb3Client("http://localhost:9545")

  // The switch doesn't work for nethermind yet
  val nethermindFollowerL2Client: Web3j = buildWeb3Client("http://localhost:10545", jwtConfig)
  val erigonFollowerL2Client: Web3j = buildWeb3Client("http://localhost:11545")
  val followerClients =
    mapOf(
      // "geth1" to geth1L2Client,
      "follower-geth-2" to geth2L2Client,
//      "follower-geth-snap-server" to gethSnapServerL2Client,
      "follower-besu" to besuFollowerL2Client,
      "follower-erigon" to erigonFollowerL2Client,
      "follower-nethermind" to nethermindFollowerL2Client,
    )
  private val transactionManager =
    let {
      val credentials =
        Credentials.create("0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae")
      RawTransactionManager(sequencerL2Client, credentials)
    }

  fun sendArbitraryTransaction(): EthSendTransaction {
    val gasPrice = sequencerL2Client.ethGasPrice().send().gasPrice
    val gasLimit = BigInteger.valueOf(21000)
    val to = transactionManager.fromAddress
    return transactionManager.sendTransaction(gasPrice, gasLimit, to, "", BigInteger.ZERO)
  }

  fun EthSendTransaction.waitForInclusion() {
    await
      .timeout(30, TimeUnit.SECONDS)
      .untilAsserted {
        val lastTransaction =
          sequencerL2Client
            .ethGetTransactionByHash(transactionHash)
            .send()
            .transaction
            .get()
        assertThat(lastTransaction.blockNumberRaw)
          .withFailMessage("Transaction $transactionHash wasn't included!")
          .isNotNull()
      }
  }

  private fun buildWeb3Client(
    rpcUrl: String,
    jwtConfig: Optional<JwtConfig> = Optional.empty(),
  ): Web3j = createWeb3jClient(rpcUrl, jwtConfig).eth1Web3j

  fun createWeb3jClient(
    endpoint: String,
    jwtConfig: Optional<JwtConfig>,
  ): Web3JClient =
    Web3jClientBuilder()
      .timeout(1.minutes.toJavaDuration())
      .endpoint(endpoint)
      .jwtConfigOpt(jwtConfig)
      .timeProvider(SystemTimeProvider.SYSTEM_TIME_PROVIDER)
      .executionClientEventsPublisher {}
      .build()
}
