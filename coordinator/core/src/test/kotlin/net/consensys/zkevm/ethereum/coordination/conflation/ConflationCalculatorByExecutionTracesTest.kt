package net.consensys.zkevm.ethereum.coordination.conflation

import kotlinx.datetime.Instant
import net.consensys.linea.traces.TracesCounters
import net.consensys.linea.traces.fakeTracesCounters
import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.domain.ConflationTrigger
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.kotlin.mock

class ConflationCalculatorByExecutionTracesTest {
  private val tracesLimit = fakeTracesCounters(100u)
  private val calculator = ConflationCalculatorByExecutionTraces(tracesLimit)
  private lateinit var conflationTriggerConsumer: ConflationTriggerConsumer

  @BeforeEach
  fun beforeEach() {
    conflationTriggerConsumer = mock<ConflationTriggerConsumer>()
  }

  private fun assertCountersEqualTo(expectedTracesCounters: TracesCounters) {
    val inflightCounters = ConflationCounters.empty()
    calculator.copyCountersTo(inflightCounters)
    assertThat(inflightCounters)
      .isEqualTo(ConflationCounters(tracesCounters = expectedTracesCounters))
  }

  @Test
  fun `appendBlock should accumulate counters`() {
    calculator.appendBlock(blockCounters(fakeTracesCounters(10u)))
    assertCountersEqualTo(fakeTracesCounters(10u))

    calculator.appendBlock(blockCounters(fakeTracesCounters(20u)))
    assertCountersEqualTo(fakeTracesCounters(30u))

    calculator.appendBlock(blockCounters(fakeTracesCounters(40u)))
    assertCountersEqualTo(fakeTracesCounters(70u))

    calculator.reset()
    assertCountersEqualTo(fakeTracesCounters(0u))
  }

  @Test
  fun `appendBlock should throw if counter go over limit when accumulated`() {
    calculator.appendBlock(blockCounters(fakeTracesCounters(10u)))
    assertThatThrownBy { calculator.appendBlock(blockCounters(fakeTracesCounters(91u))) }
      .isInstanceOf(IllegalStateException::class.java)

    // it should allow single oversized block
    calculator.reset()
    calculator.appendBlock(blockCounters(fakeTracesCounters(200u)))
  }

  @Test
  fun `copyCountersTo`() {
    val inflightConflationCounters = ConflationCounters.empty()
    calculator.appendBlock(blockCounters(fakeTracesCounters(10u)))
    calculator.copyCountersTo(inflightConflationCounters)
    assertThat(inflightConflationCounters)
      .isEqualTo(ConflationCounters(tracesCounters = fakeTracesCounters(10u)))

    calculator.appendBlock(blockCounters(fakeTracesCounters(20u)))
    calculator.copyCountersTo(inflightConflationCounters)
    assertThat(inflightConflationCounters)
      .isEqualTo(ConflationCounters(tracesCounters = fakeTracesCounters(30u)))

    calculator.appendBlock(blockCounters(fakeTracesCounters(30u)))
    calculator.copyCountersTo(inflightConflationCounters)
    assertThat(inflightConflationCounters)
      .isEqualTo(ConflationCounters(tracesCounters = fakeTracesCounters(60u)))

    calculator.reset()
    calculator.copyCountersTo(inflightConflationCounters)
    assertThat(inflightConflationCounters)
      .isEqualTo(ConflationCounters(tracesCounters = fakeTracesCounters(0u)))
  }

  @Test
  fun `checkOverflow should return trigger when block is oversized`() {
    assertThat(calculator.checkOverflow(blockCounters(fakeTracesCounters(100u)))).isNull()
    assertThat(calculator.checkOverflow(blockCounters(fakeTracesCounters(101u))))
      .isEqualTo(ConflationCalculator.OverflowTrigger(ConflationTrigger.TRACES_LIMIT, true))
  }

  @Test
  fun `checkOverflow should return trigger accumulated traces overflow`() {
    calculator.appendBlock(blockCounters(fakeTracesCounters(10u)))
    calculator.appendBlock(blockCounters(fakeTracesCounters(89u)))
    assertThat(calculator.checkOverflow(blockCounters(fakeTracesCounters(1u)))).isNull()
    assertThat(calculator.checkOverflow(blockCounters(fakeTracesCounters(2u))))
      .isEqualTo(ConflationCalculator.OverflowTrigger(ConflationTrigger.TRACES_LIMIT, false))
  }

  private fun blockCounters(
    tracesCounters: TracesCounters,
    blockNumber: ULong = 1uL
  ): BlockCounters {
    return BlockCounters(
      blockNumber = blockNumber,
      blockTimestamp = Instant.parse("2021-01-01T00:00:00Z"),
      tracesCounters = tracesCounters,
      l1DataSize = 0u,
      blockRLPEncoded = ByteArray(0)
    )
  }
}
