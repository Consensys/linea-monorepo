package net.consensys.zkevm.ethereum.coordination.conflation

import kotlinx.datetime.Instant
import net.consensys.linea.traces.TracesCountersV2
import net.consensys.linea.traces.fakeTracesCountersV2
import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.domain.ConflationTrigger
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class TtdHardForkConflationCalculatorTest {
  private val totalTerminalDifficulty = 1000UL
  private val belowTtdDifficulty = totalTerminalDifficulty - 1UL
  private val aboveTtdDifficulty = totalTerminalDifficulty + 500UL
  private val higherDifficulty = totalTerminalDifficulty + 100UL
  private val lowDifficulties = listOf(100UL, 300UL, 500UL, 800UL)
  private val moderateDifficulty = 500UL
  private val maxDifficultyMinusOne = ULong.MAX_VALUE - 1UL
  private val zeroTtd = 0UL
  private val initialDifficultyBelowTtd = 500UL

  private lateinit var calculator: TtdHardForkConflationCalculator

  @BeforeEach
  fun beforeEach() {
    calculator = TtdHardForkConflationCalculator(totalTerminalDifficulty, initialDifficultyBelowTtd)
  }

  @Test
  fun `should have correct id`() {
    assertThat(calculator.id).isEqualTo("TTD_HARD_FORK")
  }

  @Test
  fun `should not trigger overflow when total difficulty is below TTD`() {
    val blockCounters = blockCounters(blockNumber = 1, totalDifficulty = belowTtdDifficulty)

    val result = calculator.checkOverflow(blockCounters)

    assertThat(result).isNull()
  }

  @Test
  fun `should trigger overflow when total difficulty equals TTD`() {
    val blockCounters = blockCounters(blockNumber = 100, totalDifficulty = totalTerminalDifficulty)

    val result = calculator.checkOverflow(blockCounters)

    assertThat(result).isEqualTo(
      ConflationCalculator.OverflowTrigger(
        trigger = ConflationTrigger.HARD_FORK,
        singleBlockOverSized = false,
      ),
    )
  }

  @Test
  fun `should trigger overflow when total difficulty exceeds TTD`() {
    val blockCounters = blockCounters(blockNumber = 101, totalDifficulty = aboveTtdDifficulty)

    val result = calculator.checkOverflow(blockCounters)

    assertThat(result).isEqualTo(
      ConflationCalculator.OverflowTrigger(
        trigger = ConflationTrigger.HARD_FORK,
        singleBlockOverSized = false,
      ),
    )
  }

  @Test
  fun `should only trigger once when TTD is reached`() {
    val firstBlock = blockCounters(blockNumber = 100, totalDifficulty = totalTerminalDifficulty)
    val secondBlock = blockCounters(blockNumber = 101, totalDifficulty = higherDifficulty)

    // First block reaches TTD
    val firstResult = calculator.checkOverflow(firstBlock)
    assertThat(firstResult).isNotNull

    // Second block should not trigger again
    val secondResult = calculator.checkOverflow(secondBlock)
    assertThat(secondResult).isNull()
  }

  @Test
  fun `should handle multiple blocks below TTD`() {
    val blocks = lowDifficulties.mapIndexed { index, difficulty ->
      blockCounters(blockNumber = index + 1, totalDifficulty = difficulty)
    }

    blocks.forEach { block ->
      calculator.appendBlock(block)
      val result = calculator.checkOverflow(block)
      assertThat(result).isNull()
    }
  }

  @Test
  fun `appendBlock should not affect state tracking`() {
    val blockCounters = blockCounters(blockNumber = 1, totalDifficulty = moderateDifficulty)

    // appendBlock should not change behavior
    calculator.appendBlock(blockCounters)

    // Should still not trigger as TTD not reached
    val result = calculator.checkOverflow(blockCounters)
    assertThat(result).isNull()
  }

  @Test
  fun `reset should not reset TTD reached state`() {
    val blockCounters = blockCounters(blockNumber = 100, totalDifficulty = totalTerminalDifficulty)

    // Trigger TTD
    calculator.checkOverflow(blockCounters)

    // Reset
    calculator.reset()

    // Should still not trigger again after reset
    val newBlock = blockCounters(blockNumber = 101, totalDifficulty = higherDifficulty)
    val result = calculator.checkOverflow(newBlock)
    assertThat(result).isNull()
  }

  @Test
  fun `copyCountersTo should not modify counters`() {
    val counters = ConflationCounters.empty(TracesCountersV2.EMPTY_TRACES_COUNT)
    val originalBlockCount = counters.blockCount

    calculator.copyCountersTo(counters)

    assertThat(counters.blockCount).isEqualTo(originalBlockCount)
  }

  @Test
  fun `should handle zero TTD`() {
    val zeroTtdCalculator = TtdHardForkConflationCalculator(zeroTtd, zeroTtd)
    val blockCounters = blockCounters(blockNumber = 1, totalDifficulty = zeroTtd)

    val result = zeroTtdCalculator.checkOverflow(blockCounters)

    assertThat(result).isNull() // Should not trigger since TTD already reached at initialization
  }

  @Test
  fun `should handle very large TTD values`() {
    val largeTtdCalculator = TtdHardForkConflationCalculator(ULong.MAX_VALUE, maxDifficultyMinusOne)
    val blockCounters = blockCounters(blockNumber = 1, totalDifficulty = maxDifficultyMinusOne)

    val result = largeTtdCalculator.checkOverflow(blockCounters)

    assertThat(result).isNull()
  }

  @Test
  fun `should not trigger when initialized with difficulty equal to TTD`() {
    val calculator = TtdHardForkConflationCalculator(totalTerminalDifficulty, totalTerminalDifficulty)
    val blockCounters = blockCounters(blockNumber = 1, totalDifficulty = totalTerminalDifficulty + 100UL)

    val result = calculator.checkOverflow(blockCounters)

    assertThat(result).isNull()
  }

  @Test
  fun `should not trigger when initialized with difficulty above TTD`() {
    val initialDifficultyAboveTtd = totalTerminalDifficulty + 200UL
    val calculator = TtdHardForkConflationCalculator(totalTerminalDifficulty, initialDifficultyAboveTtd)
    val blockCounters = blockCounters(blockNumber = 1, totalDifficulty = initialDifficultyAboveTtd + 100UL)

    val result = calculator.checkOverflow(blockCounters)

    assertThat(result).isNull()
  }

  @Test
  fun `should trigger when initialized below TTD and block reaches TTD`() {
    val initialDifficultyBelowTtd = totalTerminalDifficulty - 100UL
    val calculator = TtdHardForkConflationCalculator(totalTerminalDifficulty, initialDifficultyBelowTtd)
    val blockCounters = blockCounters(blockNumber = 1, totalDifficulty = totalTerminalDifficulty)

    val result = calculator.checkOverflow(blockCounters)

    assertThat(result).isEqualTo(
      ConflationCalculator.OverflowTrigger(
        trigger = ConflationTrigger.HARD_FORK,
        singleBlockOverSized = false,
      ),
    )
  }

  @Test
  fun `should prevent duplicate triggers on restart scenario`() {
    // Simulate coordinator restart after TTD has already been reached
    val currentDifficulty = totalTerminalDifficulty + 500UL
    val calculator = TtdHardForkConflationCalculator(totalTerminalDifficulty, currentDifficulty)

    // Process a block with even higher difficulty - should not trigger
    val blockCounters = blockCounters(blockNumber = 1000, totalDifficulty = currentDifficulty + 100UL)
    val result = calculator.checkOverflow(blockCounters)

    assertThat(result).isNull()
  }

  @Test
  fun `should handle edge case where initial difficulty equals TTD exactly`() {
    val calculator = TtdHardForkConflationCalculator(totalTerminalDifficulty, totalTerminalDifficulty)

    // Even if a new block has higher difficulty, should not trigger since TTD was already reached
    val blockCounters = blockCounters(blockNumber = 1, totalDifficulty = totalTerminalDifficulty + 1UL)
    val result = calculator.checkOverflow(blockCounters)

    assertThat(result).isNull()
  }

  @Test
  fun `should handle transition from below TTD to above TTD in single block`() {
    val initialDifficulty = totalTerminalDifficulty - 1UL
    val calculator = TtdHardForkConflationCalculator(totalTerminalDifficulty, initialDifficulty)

    // Block jumps from below TTD to well above TTD
    val blockCounters = blockCounters(blockNumber = 1, totalDifficulty = totalTerminalDifficulty + 1000UL)
    val result = calculator.checkOverflow(blockCounters)

    assertThat(result).isEqualTo(
      ConflationCalculator.OverflowTrigger(
        trigger = ConflationTrigger.HARD_FORK,
        singleBlockOverSized = false,
      ),
    )
  }

  private fun blockCounters(
    blockNumber: Int,
    totalDifficulty: ULong,
    timestamp: Instant = Instant.parse("2021-01-01T00:00:00.000Z"),
  ): BlockCounters {
    return BlockCounters(
      blockNumber = blockNumber.toULong(),
      blockTimestamp = timestamp,
      tracesCounters = fakeTracesCountersV2(blockNumber.toUInt()),
      blockRLPEncoded = ByteArray(0),
      totalDifficulty = totalDifficulty,
    )
  }
}
