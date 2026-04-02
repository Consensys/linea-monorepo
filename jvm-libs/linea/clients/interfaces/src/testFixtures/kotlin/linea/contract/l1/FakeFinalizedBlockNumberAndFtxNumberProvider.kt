package linea.contract.l1

import tech.pegasys.teku.infrastructure.async.SafeFuture

class FakeFinalizedBlockNumberAndFtxNumberProvider(
  private var errorBlockNumbers: Set<ULong> = emptySet(),
) : FinalizedBlockNumberAndFtxNumberProvider {
  private var blockNumber = 0UL
  fun setErrorBlockNumbers(errorBlockNumbers: Set<ULong>) {
    this.errorBlockNumbers = errorBlockNumbers
  }

  override fun getFinalizedBlockNumberAndFtxNumber(): SafeFuture<FinalizedBlockNumberAndFtxNumber> {
    blockNumber = blockNumber + 1UL
    if (errorBlockNumbers.contains(blockNumber)) {
      throw Exception("Failure for the testing!")
    }
    return SafeFuture.completedFuture(
      FinalizedBlockNumberAndFtxNumber(
        blockNumber = blockNumber,
        forcedTransactionNumber = 1UL,
      ),
    )
  }
}
