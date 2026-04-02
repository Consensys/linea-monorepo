package linea.contract.l1

import tech.pegasys.teku.infrastructure.async.SafeFuture

data class FinalizedBlockNumberAndFtxNumber(
  val blockNumber: ULong,
  val forcedTransactionNumber: ULong?,
)

interface FinalizedBlockNumberAndFtxNumberProvider {
  fun getFinalizedBlockNumberAndFtxNumber(): SafeFuture<FinalizedBlockNumberAndFtxNumber>
}
