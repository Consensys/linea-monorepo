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
