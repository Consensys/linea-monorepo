package linea.contract.l1

import linea.domain.BlockParameter
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface FinalizedStateDataProvider {
  fun findFinalizedFtxNumber(blockParameter: BlockParameter): SafeFuture<ULong?>
  fun getFinalizedL2BlockNumber(blockParameter: BlockParameter): SafeFuture<ULong>
}
