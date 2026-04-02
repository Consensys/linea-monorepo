package linea.contract.l1

import linea.domain.BlockParameter
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class FinalizedBlockNumberAndFtxNumber(
  val blockNumber: ULong,
  val forcedTransactionNumber: ULong?, // null when ftx number is not available as contract is still before V8
)

interface FinalizedBlockNumberAndFtxNumberProvider {
  fun getFinalizedBlockNumberAndFtxNumber(blockParameter: BlockParameter): SafeFuture<FinalizedBlockNumberAndFtxNumber>
}
