package net.consensys.linea

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class FeeHistoryTest {

  @Test
  fun blocksRange() {
    val feeHistory = FeeHistory(
      oldestBlock = 100.toBigInteger(),
      baseFeePerGas = listOf(1.toBigInteger(), 2.toBigInteger(), 3.toBigInteger(), 4.toBigInteger()),
      reward = listOf(listOf(1.toBigInteger()), listOf(2.toBigInteger()), listOf(3.toBigInteger())),
      gasUsedRatio = listOf(0.1.toBigDecimal(), 0.2.toBigDecimal(), 0.3.toBigDecimal())
    )
    assertThat(feeHistory.blocksRange()).isEqualTo(100.toBigInteger()..102.toBigInteger())
  }
}
