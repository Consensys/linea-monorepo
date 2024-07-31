package net.consensys.zkevm.load.model

import org.web3j.protocol.core.Request
import org.web3j.protocol.core.methods.response.EthSendTransaction
import java.math.BigInteger

class TransactionDetail(
  val walletId: Int,
  val nonce: BigInteger,
  @Transient val ethSendTransactionRequest: Request<*, EthSendTransaction>,
  val expectedOutcome: EXPECTED_OUTCOME = EXPECTED_OUTCOME.SUCCESS
) {
  var hash: String? = null
}

enum class EXPECTED_OUTCOME {
  // TODO: NOT_EXECUTED should be used for non profitable or underpriced transactions.
  FAILED, SUCCESS, NOT_EXECUTED
}
