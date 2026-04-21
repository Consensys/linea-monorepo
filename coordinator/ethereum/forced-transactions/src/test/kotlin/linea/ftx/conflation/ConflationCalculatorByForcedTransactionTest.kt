package linea.ftx.conflation

import linea.domain.BlobCounters
import linea.domain.BlockCounters
import linea.domain.ConflationTrigger
import linea.forcedtx.ForcedTransactionInclusionResult
import linea.forcedtx.ForcedTransactionInclusionStatus
import net.consensys.linea.traces.TracesCountersV4
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCounters
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Nested
import org.junit.jupiter.api.Test
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.EnumSource
import java.util.LinkedList
import java.util.Queue
import kotlin.time.Instant

class ConflationCalculatorByForcedTransactionTest {

  private lateinit var queue: Queue<ForcedTransactionInclusionStatus>
  private lateinit var calculator: ConflationCalculatorByForcedTransaction

  private val timestamp = Instant.parse("2024-01-01T00:00:00Z")
  private val ftxOverflowTrigger = ConflationCalculator.OverflowTrigger(
    trigger = ConflationTrigger.FORCED_TRANSACTION,
    singleBlockOverSized = false,
  )

  @BeforeEach
  fun setUp() {
    queue = LinkedList()
    calculator = ConflationCalculatorByForcedTransaction(queue)
  }

  private fun ftx(
    ftx: ULong,
    blockNumber: ULong,
    inclusionResult: ForcedTransactionInclusionResult,
  ) = ForcedTransactionInclusionStatus(
    ftxNumber = ftx,
    blockNumber = blockNumber,
    blockTimestamp = timestamp,
    inclusionResult = inclusionResult,
    ftxHash = ByteArray(32),
    from = ByteArray(20),
  )

  private fun blockCounters(blockNumber: ULong) = BlockCounters(
    blockNumber = blockNumber,
    blockTimestamp = timestamp,
    tracesCounters = TracesCountersV4.EMPTY_TRACES_COUNT,
    blockRLPEncoded = ByteArray(0),
  )

  @Test
  fun `id should be FORCED_TRANSACTION`() {
    assertThat(calculator.id).isEqualTo(ConflationTrigger.FORCED_TRANSACTION.name)
  }

  @Test
  fun `consumes items from its dedicated queue`() {
    queue.add(ftx(ftx = 1UL, blockNumber = 10UL, inclusionResult = ForcedTransactionInclusionResult.BadNonce))

    calculator.checkOverflow(blockCounters(blockNumber = 9UL))

    assertThat(queue).isEmpty()
  }

  @Test
  fun `copyCountersTo is a no-op`() {
    queue.add(ftx(ftx = 1UL, blockNumber = 10UL, inclusionResult = ForcedTransactionInclusionResult.BadNonce))
    val counters = ConflationCounters(tracesCounters = TracesCountersV4.EMPTY_TRACES_COUNT)

    calculator.copyCountersTo(counters)

    assertThat(counters.blockCount).isEqualTo(0u)
    assertThat(counters.dataSize).isEqualTo(0u)
  }

  @Nested
  inner class CheckOverflow {
    @Test
    fun `returns null when the queue is empty`() {
      assertThat(calculator.checkOverflow(blockCounters(blockNumber = 10UL))).isNull()
    }

    @ParameterizedTest(name = "inclusionResult={0}")
    @EnumSource(
      value = ForcedTransactionInclusionResult::class,
    )
    fun `triggers conflation at ftxBlockNumber for all non-Included results`(
      result: ForcedTransactionInclusionResult,
    ) {
      queue.add(ftx(ftx = 1UL, blockNumber = 10UL, inclusionResult = result))

      assertThat(calculator.checkOverflow(blockCounters(blockNumber = 9UL))).isNull()
      assertThat(calculator.checkOverflow(blockCounters(blockNumber = 10UL))).isEqualTo(ftxOverflowTrigger)
      assertThat(calculator.checkOverflow(blockCounters(blockNumber = 11UL))).isNull()
    }

    @Test
    fun `triggers once per non-Included FTX at its respective boundary block`() {
      queue.add(ftx(ftx = 1UL, blockNumber = 10UL, inclusionResult = ForcedTransactionInclusionResult.BadNonce))
      queue.add(ftx(ftx = 2UL, blockNumber = 20UL, inclusionResult = ForcedTransactionInclusionResult.BadBalance))

      assertThat(calculator.checkOverflow(blockCounters(blockNumber = 10UL))).isEqualTo(ftxOverflowTrigger)
      assertThat(calculator.checkOverflow(blockCounters(blockNumber = 20UL))).isEqualTo(ftxOverflowTrigger)
    }

    @Test
    fun `picks up FTXs added to the queue between calls`() {
      assertThat(calculator.checkOverflow(blockCounters(blockNumber = 10UL))).isNull()

      queue.add(ftx(ftx = 1UL, blockNumber = 10UL, inclusionResult = ForcedTransactionInclusionResult.BadNonce))

      assertThat(calculator.checkOverflow(blockCounters(blockNumber = 10UL))).isEqualTo(ftxOverflowTrigger)
    }
  }

  @Nested
  inner class AppendBlock {

    @Test
    fun `clears pending triggers at or below the appended block number`() {
      queue.add(
        ftx(ftx = 1UL, blockNumber = 10UL, inclusionResult = ForcedTransactionInclusionResult.BadNonce),
      ) // trigger at 10
      queue.add(
        ftx(ftx = 2UL, blockNumber = 20UL, inclusionResult = ForcedTransactionInclusionResult.BadBalance),
      ) // trigger at 20

      // Populate pendingTriggerBlocks: {10, 20}
      calculator.checkOverflow(blockCounters(blockNumber = 5UL))

      // Append block 10 - clears triggers <= 10 from pendingTriggerBlocks
      calculator.appendBlock(blockCounters(blockNumber = 10UL))

      // Clear queue so readProcessedFtxs won't re-add triggers on next checkOverflow
      queue.clear()

      // Trigger at 10 was cleared from pendingTriggerBlocks
      assertThat(calculator.checkOverflow(blockCounters(blockNumber = 10UL))).isNull()
      // Trigger at 20 is still in pendingTriggerBlocks
      assertThat(calculator.checkOverflow(blockCounters(blockNumber = 20UL))).isEqualTo(ftxOverflowTrigger)
    }

    @Test
    fun `does not clear pending triggers for future blocks`() {
      queue.add(
        ftx(ftx = 1UL, blockNumber = 20UL, inclusionResult = ForcedTransactionInclusionResult.BadNonce),
      ) // trigger at 20

      calculator.checkOverflow(blockCounters(blockNumber = 5UL))
      calculator.appendBlock(blockCounters(blockNumber = 10UL))

      assertThat(calculator.checkOverflow(blockCounters(blockNumber = 20UL))).isEqualTo(ftxOverflowTrigger)
    }
  }

  @Nested
  inner class Reset {

    @Test
    fun `does not clear pending trigger blocks`() {
      queue.add(ftx(ftx = 1UL, blockNumber = 10UL, inclusionResult = ForcedTransactionInclusionResult.BadNonce))

      // Populate pendingTriggerBlocks
      calculator.checkOverflow(blockCounters(blockNumber = 5UL))
      // Remove from queue so the trigger can only come from pendingTriggerBlocks state
      queue.clear()

      calculator.reset()

      assertThat(calculator.checkOverflow(blockCounters(blockNumber = 10UL))).isEqualTo(ftxOverflowTrigger)
    }
  }

  @Test
  fun `consecutive FTXs arriving one at a time each produce conflation and aggregation triggers`() {
    val conflationQueue = LinkedList<ForcedTransactionInclusionStatus>()
    val aggregationQueue = LinkedList<ForcedTransactionInclusionStatus>()
    val conflationCalc = ConflationCalculatorByForcedTransaction(conflationQueue)
    val aggregationCalc = AggregationCalculatorByForcedTransaction(aggregationQueue)

    val conflationTriggers = mutableListOf<ULong>()
    val aggregationTriggers = mutableListOf<ULong>()

    // 4 consecutive FTXs, blocks 21-24 — mirrors the e2e forced-transactions test
    for (i in 0 until 4) {
      val block = 21UL + i.toULong()
      val result = ftx(
        ftx = (i + 1).toULong(),
        blockNumber = block,
        inclusionResult = ForcedTransactionInclusionResult.BadNonce,
      )

      // FTX result arrives — each calculator gets its own copy (production wiring)
      conflationQueue.add(result)
      aggregationQueue.add(result)

      // Conflation processes the FTX block
      if (conflationCalc.checkOverflow(blockCounters(block)) != null) {
        conflationTriggers.add(block)
        conflationCalc.reset()
        conflationCalc.appendBlock(blockCounters(block))
      }

      // Aggregation processes the blob that just sealed ([block-1 .. block-1])
      val blobEnd = block - 1UL
      val blob = BlobCounters(
        startBlockNumber = blobEnd,
        endBlockNumber = blobEnd,
        numberOfBatches = 1u,
        startBlockTimestamp = timestamp,
        endBlockTimestamp = timestamp,
        expectedShnarf = ByteArray(0),
      )
      if (aggregationCalc.checkAggregationTrigger(blob) != null) {
        aggregationTriggers.add(blobEnd)
        aggregationCalc.reset()
      }
    }

    // Each FTX block fires a conflation trigger (FTX is first block of new blob)
    assertThat(conflationTriggers).isEqualTo(listOf(21UL, 22UL, 23UL, 24UL))
    // Each aggregation seals at (ftxBlock - 1), so FTX starts the next aggregation
    assertThat(aggregationTriggers).isEqualTo(listOf(20UL, 21UL, 22UL, 23UL))
  }

  @Test
  fun `aggregation draining its queue does not affect conflation triggers`() {
    val conflationQueue = LinkedList<ForcedTransactionInclusionStatus>()
    val aggregationQueue = LinkedList<ForcedTransactionInclusionStatus>()
    val conflationCalc = ConflationCalculatorByForcedTransaction(conflationQueue)
    val aggregationCalc = AggregationCalculatorByForcedTransaction(aggregationQueue)

    // FTX#1@21 arrives
    val ftx1 = ftx(ftx = 1UL, blockNumber = 21UL, inclusionResult = ForcedTransactionInclusionResult.BadNonce)
    conflationQueue.add(ftx1)
    aggregationQueue.add(ftx1)

    // Conflation processes block 21 → trigger fires
    assertThat(conflationCalc.checkOverflow(blockCounters(21UL))).isEqualTo(ftxOverflowTrigger)
    conflationCalc.reset()
    conflationCalc.appendBlock(blockCounters(21UL))

    // FTX#2@22 arrives
    val ftx2 = ftx(ftx = 2UL, blockNumber = 22UL, inclusionResult = ForcedTransactionInclusionResult.BadNonce)
    conflationQueue.add(ftx2)
    aggregationQueue.add(ftx2)

    // Aggregation processes blob [20..20] — drains BOTH FTX#1 and FTX#2 from its queue
    val blob20 = BlobCounters(
      startBlockNumber = 20UL,
      endBlockNumber = 20UL,
      numberOfBatches = 1u,
      startBlockTimestamp = timestamp,
      endBlockTimestamp = timestamp,
      expectedShnarf = ByteArray(0),
    )
    val aggTrigger = aggregationCalc.checkAggregationTrigger(blob20)
    assertThat(aggTrigger).isNotNull // fires at 20 (= 21-1)
    assertThat(aggregationQueue).isEmpty() // aggregation drained everything

    // Conflation processes block 22 — still fires because it has its own queue
    assertThat(conflationCalc.checkOverflow(blockCounters(22UL))).isEqualTo(ftxOverflowTrigger)
  }
}
