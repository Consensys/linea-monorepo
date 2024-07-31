package net.consensys.zkevm.ethereum

import org.awaitility.Awaitility
import org.web3j.protocol.Web3j
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

fun Web3j.waitForTransactionExecution(
  transactionHash: String,
  expectedStatus: String? = null,
  timeout: Duration = 30.seconds,
  pollInterval: Duration = 500.milliseconds
) {
  Awaitility.await()
    .timeout(timeout.toJavaDuration())
    .pollInterval(pollInterval.toJavaDuration())
    .untilAsserted {
      val lastBlobTxReceipt = this.ethGetTransactionReceipt(transactionHash).send()
      if (lastBlobTxReceipt.result == null) {
        throw AssertionError("Transaction receipt not found: txHash=$transactionHash, timeout=$timeout")
      }
      expectedStatus?.also {
        if (lastBlobTxReceipt.result.status != expectedStatus) {
          throw AssertionError(
            "Transaction status does not match expected status: " +
              "txHash=$transactionHash, expected=$expectedStatus, actual=${lastBlobTxReceipt.result.status}"
          )
        }
      }
    }
}
