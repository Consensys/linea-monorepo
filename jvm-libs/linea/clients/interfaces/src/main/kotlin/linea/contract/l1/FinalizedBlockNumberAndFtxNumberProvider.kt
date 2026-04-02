package linea.contract.l1

import tech.pegasys.teku.infrastructure.async.SafeFuture

data class FinalizedBlockNumberAndFtxNumber(
  val blockNumber: ULong,
  val forcedTransactionNumber: ULong?, // null when ftx number is not available as contract is still before V8
)

interface FinalizedBlockNumberAndFtxNumberProvider {
  fun getFinalizedBlockNumberAndFtxNumber(): SafeFuture<FinalizedBlockNumberAndFtxNumber>
}
