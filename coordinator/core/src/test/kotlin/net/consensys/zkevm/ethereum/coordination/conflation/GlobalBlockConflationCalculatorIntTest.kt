package net.consensys.zkevm.ethereum.coordination.conflation

import kotlinx.datetime.Instant
import net.consensys.FakeFixedClock
import net.consensys.linea.traces.fakeTracesCounters
import net.consensys.linea.traces.sumTracesCounters
import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.domain.ConflationCalculationResult
import net.consensys.zkevm.domain.ConflationTrigger
import net.consensys.zkevm.ethereum.coordination.blob.FakeBlobCompressor
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockHeaderSummary
import net.consensys.zkevm.ethereum.coordination.blockcreation.SafeBlockProvider
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.kotlin.mock
import org.mockito.kotlin.spy
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration.Companion.days
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds

class GlobalBlockConflationCalculatorIntTest {
  // NOTE: this breaks the test isolation, but adds some confidence that the integration works
  private lateinit var calculatorByDealine: ConflationCalculatorByTimeDeadline
  private lateinit var calculatorByTraces: ConflationCalculator
  private lateinit var globalCalculator: GlobalBlockConflationCalculator
  private val lastBlockNumber: ULong = 0uL
  private lateinit var safeBlockProvider: SafeBlockProvider
  private lateinit var fakeClock: FakeFixedClock
  private val fakeClockTime = Instant.parse("2023-12-11T00:00:00.000Z")
  private lateinit var conflations: MutableList<ConflationCalculationResult>

  @BeforeEach
  fun beforeEach() {
    fakeClock = FakeFixedClock(fakeClockTime)
    conflations = mutableListOf()
    safeBlockProvider = mock<SafeBlockProvider>() {
      on { getLatestSafeBlock() }.thenReturn(
        SafeFuture.failedFuture(RuntimeException("getLatestSafeBlock should not be called"))
      )
      on { getLatestSafeBlockHeader() }.thenReturn(
        SafeFuture.failedFuture(RuntimeException("getLatestSafeBlockHeader not mocked yet"))
      )
    }
    calculatorByDealine = spy(
      ConflationCalculatorByTimeDeadline(
        config = ConflationCalculatorByTimeDeadline.Config(
          conflationDeadline = 2.seconds,
          conflationDeadlineLastBlockConfirmationDelay = 10.milliseconds
        ),
        lastBlockNumber = lastBlockNumber,
        latestBlockProvider = safeBlockProvider
      )
    )
    val fakeBlobCompressor = FakeBlobCompressor(1_000)
    val calculatorByData = spy(
      ConflationCalculatorByDataCompressed(
        blobCompressor = fakeBlobCompressor
      )
    )
    whenever(calculatorByData.reset()).then {
      fakeBlobCompressor.reset()
    }

    calculatorByTraces = ConflationCalculatorByExecutionTraces(
      tracesCountersLimit = fakeTracesCounters(100u)
    )
    globalCalculator = GlobalBlockConflationCalculator(
      lastBlockNumber = lastBlockNumber,
      syncCalculators = listOf(calculatorByTraces, calculatorByData),
      deferredTriggerConflationCalculators = listOf(calculatorByDealine)
    )
    globalCalculator.onConflatedBatch { trigger ->
      conflations.add(trigger)
      SafeFuture.completedFuture(Unit)
    }
  }

  @Test
  fun `conflation by traces limit - 1st block overflows`() {
    // block with traces oversize
    val block1Counters = BlockCounters(
      blockNumber = 1uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(101u),
      l1DataSize = 10u,
      blockRLPEncoded = ByteArray(10)
    )
    val block2Counters = BlockCounters(
      blockNumber = 2uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(10u),
      l1DataSize = 10u,
      blockRLPEncoded = ByteArray(10)
    )
    globalCalculator.newBlock(block1Counters)
    globalCalculator.newBlock(block2Counters)

    assertThat(conflations).hasSize(1)
    assertThat(conflations[0]).isEqualTo(
      ConflationCalculationResult(
        startBlockNumber = 1uL,
        endBlockNumber = 1uL,
        conflationTrigger = ConflationTrigger.TRACES_LIMIT,
        tracesCounters = block1Counters.tracesCounters,
        dataL1Size = block1Counters.l1DataSize
      )
    )
  }

  @Test
  fun `conflation by traces limit - single block`() {
    // block with traces oversize
    val block1Counters = BlockCounters(
      blockNumber = 1uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(60u),
      l1DataSize = 10u,
      blockRLPEncoded = ByteArray(10)
    )
    val block2Counters = BlockCounters(
      blockNumber = 2uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(50u),
      l1DataSize = 20u,
      blockRLPEncoded = ByteArray(20)
    )
    globalCalculator.newBlock(block1Counters)
    globalCalculator.newBlock(block2Counters)

    assertThat(conflations).hasSize(1)
    assertThat(conflations[0]).isEqualTo(
      ConflationCalculationResult(
        startBlockNumber = 1uL,
        endBlockNumber = 1uL,
        conflationTrigger = ConflationTrigger.TRACES_LIMIT,
        tracesCounters = block1Counters.tracesCounters,
        dataL1Size = 10u
      )
    )
  }

  @Test
  fun `conflation by traces limit - multiple blocks`() {
    // block with traces oversize
    val block1Counters = BlockCounters(
      blockNumber = 1uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(50u),
      l1DataSize = 10u,
      blockRLPEncoded = ByteArray(10)
    )
    val block2Counters = BlockCounters(
      blockNumber = 2uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(50u),
      l1DataSize = 20u,
      blockRLPEncoded = ByteArray(20)
    )
    val block3Counters = BlockCounters(
      blockNumber = 3uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(10u),
      l1DataSize = 20u,
      blockRLPEncoded = ByteArray(20)
    )
    globalCalculator.newBlock(block1Counters)
    globalCalculator.newBlock(block2Counters)
    globalCalculator.newBlock(block3Counters)

    assertThat(conflations).hasSize(1)
    assertThat(conflations[0]).isEqualTo(
      ConflationCalculationResult(
        startBlockNumber = 1uL,
        endBlockNumber = 2uL,
        conflationTrigger = ConflationTrigger.TRACES_LIMIT,
        tracesCounters = sumTracesCounters(block1Counters.tracesCounters, block2Counters.tracesCounters),
        dataL1Size = block1Counters.l1DataSize + block2Counters.l1DataSize
      )
    )
  }

  @Test
  fun `conflation by data limit`() {
    // block with traces oversize
    val block1Counters = BlockCounters(
      blockNumber = 1uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(10u),
      l1DataSize = 500u,
      blockRLPEncoded = ByteArray(500)
    )
    val block2Counters = BlockCounters(
      blockNumber = 2uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(20u),
      l1DataSize = 480u,
      blockRLPEncoded = ByteArray(480)
    )
    val block3Counters = BlockCounters(
      blockNumber = 3uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(20u),
      l1DataSize = 21u,
      blockRLPEncoded = ByteArray(21)
    )
    globalCalculator.newBlock(block1Counters)
    globalCalculator.newBlock(block2Counters)
    globalCalculator.newBlock(block3Counters)

    assertThat(conflations).hasSize(1)
    assertThat(conflations[0]).isEqualTo(
      ConflationCalculationResult(
        startBlockNumber = 1uL,
        endBlockNumber = 2uL,
        conflationTrigger = ConflationTrigger.DATA_LIMIT,
        tracesCounters = sumTracesCounters(block1Counters.tracesCounters, block2Counters.tracesCounters),
        dataL1Size = block1Counters.l1DataSize + block2Counters.l1DataSize
      )
    )
  }

  @Test
  fun `integrated flow with multiple scenarios`() {
    // block with traces oversize
    val block1Counters = BlockCounters(
      blockNumber = 1uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(101u),
      l1DataSize = 100u,
      blockRLPEncoded = ByteArray(100)
    )
    // block with data in size limit
    val block2Counters = BlockCounters(
      blockNumber = 2uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(20u),
      l1DataSize = 1_000u,
      blockRLPEncoded = ByteArray(1_000)
    )

    val block3Counters = BlockCounters(
      blockNumber = 3uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(30u),
      l1DataSize = 300u,
      blockRLPEncoded = ByteArray(300)
    )
    val block4Counters = BlockCounters(
      blockNumber = 4uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(70u),
      l1DataSize = 400u,
      blockRLPEncoded = ByteArray(400)
    )
    // will trigger traces overflow
    val block5Counters = BlockCounters(
      blockNumber = 5uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(10u),
      l1DataSize = 100u,
      blockRLPEncoded = ByteArray(100)
    )
    val block6Counters = BlockCounters(
      blockNumber = 6uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(10u),
      l1DataSize = 100u,
      blockRLPEncoded = ByteArray(100)
    )
    val block7Counters = BlockCounters(
      blockNumber = 7uL,
      blockTimestamp = fakeClock.now(),
      tracesCounters = fakeTracesCounters(10u),
      l1DataSize = 100u,
      blockRLPEncoded = ByteArray(100)
    )

    globalCalculator.newBlock(block1Counters)
    globalCalculator.newBlock(block2Counters)
    globalCalculator.newBlock(block3Counters)
    globalCalculator.newBlock(block4Counters)
    globalCalculator.newBlock(block5Counters)
    globalCalculator.newBlock(block6Counters)
    globalCalculator.newBlock(block7Counters)
    // will trigger deadline overflow
    fakeClock.advanceBy(2.days)

    whenever(safeBlockProvider.getLatestSafeBlockHeader()).thenReturn(
      SafeFuture.completedFuture(
        BlockHeaderSummary(
          number = 7uL,
          hash = Bytes32.random(),
          timestamp = block7Counters.blockTimestamp
        )
      )
    )

    calculatorByDealine.checkConflationDeadline()

    assertThat(conflations[0]).isEqualTo(
      ConflationCalculationResult(
        startBlockNumber = 1uL,
        endBlockNumber = 1uL,
        conflationTrigger = ConflationTrigger.DATA_LIMIT,
        tracesCounters = block1Counters.tracesCounters,
        dataL1Size = block1Counters.l1DataSize
      )
    )
    assertThat(conflations[1]).isEqualTo(
      ConflationCalculationResult(
        startBlockNumber = 2uL,
        endBlockNumber = 2uL,
        conflationTrigger = ConflationTrigger.DATA_LIMIT,
        tracesCounters = block2Counters.tracesCounters,
        dataL1Size = block2Counters.l1DataSize
      )
    )
    assertThat(conflations[2]).isEqualTo(
      ConflationCalculationResult(
        startBlockNumber = 3uL,
        endBlockNumber = 4uL,
        conflationTrigger = ConflationTrigger.TRACES_LIMIT,
        tracesCounters = sumTracesCounters(block3Counters.tracesCounters, block4Counters.tracesCounters),
        dataL1Size = block3Counters.l1DataSize + block4Counters.l1DataSize
      )
    )
    assertThat(conflations[3]).isEqualTo(
      ConflationCalculationResult(
        startBlockNumber = 5uL,
        endBlockNumber = 7uL,
        conflationTrigger = ConflationTrigger.TIME_LIMIT,
        tracesCounters = sumTracesCounters(
          block5Counters.tracesCounters,
          block6Counters.tracesCounters,
          block7Counters.tracesCounters
        ),
        dataL1Size = block5Counters.l1DataSize + block6Counters.l1DataSize + block7Counters.l1DataSize
      )
    )
    assertThat(conflations).hasSize(4)
  }
}
