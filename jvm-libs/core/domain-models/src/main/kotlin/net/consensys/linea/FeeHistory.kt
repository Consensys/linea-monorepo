package net.consensys.linea

import java.math.BigDecimal
import java.math.BigInteger

data class FeeHistory(
  val oldestBlock: BigInteger,
  val baseFeePerGas: List<BigInteger>,
  val reward: List<List<BigInteger>>,
  val gasUsedRatio: List<BigDecimal>,
  val baseFeePerBlobGas: List<BigInteger>,
  val blobGasUsedRatio: List<BigDecimal>
) {
  fun blocksRange(): ClosedRange<BigInteger> {
    return oldestBlock..oldestBlock.add((reward.size - 1).toBigInteger())
  }
}
