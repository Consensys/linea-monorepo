package net.consensys.zkevm.ethereum.coordination.conflation.upgrade

import kotlinx.datetime.Instant
import net.consensys.linea.traces.fakeTracesCounters
import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.domain.ConflationTrigger
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculator
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.Test

@Suppress("DEPRECATION")
class SwitchCutoffCalculatorTest {
  private val switchBlockNumber = 2UL
  private val calculator: SwitchCutoffCalculator = SwitchCutoffCalculator(switchBlockNumber)
  private val baseBlockCounters = BlockCounters(
    blockNumber = 0UL,
    blockTimestamp = Instant.parse("2021-01-01T00:00:00.000Z"),
    tracesCounters = fakeTracesCounters(0U),
    l1DataSize = 0U,
    blockRLPEncoded = ByteArray(0)
  )

  @Test
  fun `triggers conflation on block M`() {
    val block = baseBlockCounters.copy(blockNumber = 2UL)
    calculator.appendBlock(block)
    Assertions.assertThat(calculator.checkOverflow(block)).isEqualTo(
      ConflationCalculator.OverflowTrigger(ConflationTrigger.SWITCH_CUTOFF, false)
    )
  }

  @Test
  fun `Doesn't trigger conflation before block M`() {
    val block = baseBlockCounters.copy(blockNumber = 1UL)
    calculator.appendBlock(block)
    Assertions.assertThat(calculator.checkOverflow(block)).isNull()
  }

  @Test
  fun `Doesn't trigger conflation if block M is undefined`() {
    val calculator = SwitchCutoffCalculator(null)
    val block = baseBlockCounters.copy(blockNumber = 3UL)
    calculator.appendBlock(block)
    Assertions.assertThat(calculator.checkOverflow(block)).isNull()
  }

  @Test
  fun `throws an exception after block M`() {
    val block = baseBlockCounters.copy(blockNumber = 3UL)
    calculator.appendBlock(block)
    Assertions.assertThatThrownBy { calculator.checkOverflow(block) }.isInstanceOf(IllegalStateException::class.java)
  }
}
