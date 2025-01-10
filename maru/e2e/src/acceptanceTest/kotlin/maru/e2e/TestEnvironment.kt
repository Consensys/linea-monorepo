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
import okhttp3.OkHttpClient
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.methods.response.EthSendTransaction
import org.web3j.protocol.http.HttpService
import org.web3j.tx.RawTransactionManager
import org.web3j.utils.Async

object TestEnvironment {
  val sequencerL2Client: Web3j = buildL2Client("http://localhost:8545")
  val geth1L2Client: Web3j = buildL2Client("http://localhost:8555")
  val geth2L2Client: Web3j = buildL2Client("http://localhost:8565")
  val gethSnapServerL2Client: Web3j = buildL2Client("http://localhost:8575")
  val besuFollowerL2Client: Web3j = buildL2Client("http://localhost:9545")
  val followerClients =
    mapOf(
      //        "geth1" to geth1L2Client,
      "geth2" to geth2L2Client,
      "gethSnapServer" to gethSnapServerL2Client,
      "besuFollower" to besuFollowerL2Client,
    )
  val transactionManager =
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
    await().untilAsserted {
      val lastTransaction =
        sequencerL2Client.ethGetTransactionByHash(transactionHash).send().transaction.get()
      assertThat(lastTransaction.blockNumberRaw)
        .withFailMessage("Transaction $transactionHash wasn't included!")
        .isNotNull()
    }
  }

  private fun buildL2Client(rpcUrl: String): Web3j = buildWeb3Client(rpcUrl)

  private fun buildWeb3Client(rpcUrl: String): Web3j =
    Web3j.build(
      HttpService(rpcUrl, OkHttpClient.Builder().build()),
      500,
      Async.defaultExecutorService(),
    )
}
