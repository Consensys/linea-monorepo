package net.consensys.zkevm.ethereum.coordination.conflation

import kotlinx.datetime.Instant
import net.consensys.linea.traces.fakeTracesCounters
import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.domain.ConflationTrigger
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class ConflationCalculatorByBlockLimitTest {
  private lateinit var calculator: ConflationCalculatorByBlockLimit

  @BeforeEach
  fun beforeEach() {
    calculator = ConflationCalculatorByBlockLimit(3u)
  }

  @Test
  fun `should accumulate blockCounting`() {
    val counters = ConflationCounters.empty()
    calculator.copyCountersTo(counters)
    assertThat(counters.blockCount).isEqualTo(0u)
    calculator.appendBlock(blockCounters(1))
    calculator.appendBlock(blockCounters(2))
    calculator.appendBlock(blockCounters(3))
    calculator.copyCountersTo(counters)
    assertThat(counters.blockCount).isEqualTo(3u)

    calculator.reset()
    calculator.copyCountersTo(counters)
    assertThat(counters.blockCount).isEqualTo(0u)
  }

  @Test
  fun `checkOverflow should return overflow trigger`() {
    calculator.appendBlock(blockCounters(10))
    calculator.appendBlock(blockCounters(20))
    assertThat(calculator.checkOverflow(blockCounters(30))).isNull()
    calculator.appendBlock(blockCounters(30))

    assertThat(calculator.checkOverflow(blockCounters(40)))
      .isEqualTo((ConflationCalculator.OverflowTrigger(ConflationTrigger.BLOCKS_LIMIT, false)))
  }

  private fun blockCounters(blockNumber: Int): BlockCounters {
    return BlockCounters(
      blockNumber = blockNumber.toULong(),
      blockTimestamp = Instant.parse("2021-01-01T00:00:00.000Z"),
      tracesCounters = fakeTracesCounters(blockNumber.toUInt()),
      l1DataSize = blockNumber.toUInt(),
      blockRLPEncoded = ByteArray(0)
    )
  }
}
