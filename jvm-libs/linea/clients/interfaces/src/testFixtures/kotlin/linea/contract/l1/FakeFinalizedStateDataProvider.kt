package linea.contract.l1

import linea.domain.BlockParameter
import tech.pegasys.teku.infrastructure.async.SafeFuture

class FakeFinalizedStateDataProvider(
  private var errorBlockNumbers: Set<ULong> = emptySet(),
) : FinalizedStateDataProvider {
  private var blockNumber = 0UL
  fun setErrorBlockNumbers(errorBlockNumbers: Set<ULong>) {
    this.errorBlockNumbers = errorBlockNumbers
  }

  override fun getFinalizedL2BlockNumber(blockParameter: BlockParameter): SafeFuture<ULong> {
    blockNumber = blockNumber + 1UL
    if (errorBlockNumbers.contains(blockNumber)) {
      throw Exception("Failure for the testing!")
    }
    return SafeFuture.completedFuture(blockNumber)
  }

  override fun findFinalizedFtxNumber(blockParameter: BlockParameter): SafeFuture<ULong?> {
    return SafeFuture.completedFuture(null)
  }
}
