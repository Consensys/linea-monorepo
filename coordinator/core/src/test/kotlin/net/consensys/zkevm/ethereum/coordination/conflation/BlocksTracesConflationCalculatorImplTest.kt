package net.consensys.zkevm.ethereum.coordination.conflation

import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.linea.traces.TracesCounters
import net.consensys.linea.traces.TracingModule
import net.consensys.linea.traces.fakeTracesCounters
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockHeaderSummary
import net.consensys.zkevm.ethereum.coordination.blockcreation.SafeBlockProvider
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration.Companion.days
import kotlin.time.Duration.Companion.hours
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

class BlocksTracesConflationCalculatorImplTest {
  private val conflationDeadlineLastblockConfirmationDelay = 12.seconds
  private val conflationDeadline1Week = 7.days
  private val tracesCap = TracingModule.values().associateWith { 100u }
  private val dataLimits = DataLimits(100u, 0u, 0u)
  private val noDeadlineConflationConfig = ConflationCalculatorConfig(
    tracesCap,
    dataLimits,
    conflationDeadline1Week,
    conflationDeadlineCheckInterval = 10.minutes,
    conflationDeadlineLastBlockConfirmationDelay = conflationDeadlineLastblockConfirmationDelay,
    blocksLimit = null
  )
  private val blockTimestamp = Instant.parse("2021-01-01T00:00:00Z")
  private val latestBlockProvider: SafeBlockProvider = mock<SafeBlockProvider>() {
    on { getLatestSafeBlockHeader() } doReturn SafeFuture.completedFuture(
      BlockHeaderSummary(
        number = ULong.MAX_VALUE,
        timestamp = blockTimestamp,
        hash = Bytes32.random()
      )
    )
  }

  @Test
  private fun tracesOverflowShouldCallConflationListenerAndCarryBlockToNextConflation() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 0u,
      latestBlockProvider = latestBlockProvider,
      config = noDeadlineConflationConfig
    )
    calculator.onConflatedBatch(conflationConsumer)

    val tracesCounters: TracesCounters = fakeTracesCounters(10u)
    val tracesCounters2: TracesCounters = fakeTracesCounters(10u)
      .apply {
        this as MutableMap
        this[TracingModule.HUB] = 90u
      }

    calculator.newBlock(BlockCounters(1u, blockTimestamp, tracesCounters, 0u))
    calculator.newBlock(BlockCounters(2u, blockTimestamp, tracesCounters, 0u))
    calculator.newBlock(BlockCounters(3u, blockTimestamp, tracesCounters, 0u))
    calculator.newBlock(BlockCounters(4u, blockTimestamp, tracesCounters, 0u))
    calculator.newBlock(BlockCounters(5u, blockTimestamp, tracesCounters2, 0u))
    assertThat(calculator.getConflationInProgress()).isNotNull()

    assertThat(conflations.first())
      .isEqualTo(
        ConflationCalculationResult(
          1u,
          4u,
          fakeTracesCounters(40u),
          0u,
          ConflationTrigger.TRACES_LIMIT
        )
      )

    val ongoingConflation = calculator.getConflationInProgress()
    assertThat(ongoingConflation)
      .isEqualTo(
        ConflationCalculationResult(
          5u,
          5u,
          tracesCounters2,
          0u,
          ConflationTrigger.TIME_LIMIT
        )
      )
  }

  @Test
  fun nonSequentialBlockConflationThrowsError() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 0u,
      latestBlockProvider = latestBlockProvider,
      config = noDeadlineConflationConfig
    )
    calculator.onConflatedBatch(conflationConsumer)

    val ex =
      assertThrows<IllegalArgumentException> {
        calculator.newBlock(BlockCounters(2u, blockTimestamp, fakeTracesCounters(10u), 0u))
      }
    assertThat(ex.message)
      .contains("Blocks to conflate must be sequential: lastBlockNumber=0, new blockNumber=2")
  }

  @Test
  fun blockOversizedLogsTheOversizeValues() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val log: Logger = mock()
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 0u,
      latestBlockProvider = latestBlockProvider,
      config = noDeadlineConflationConfig.copy(
        dataConflationLimits = DataLimits(100u, 10u, 15u)
      ),
      log = log
    )
    calculator.onConflatedBatch(conflationConsumer)

    val tracesCounters: TracesCounters = fakeTracesCounters(10u)
      .toMutableMap()
      .apply {
        put(TracingModule.ADD, 110u)
        put(TracingModule.BIN, 210u)
      }

    calculator.newBlock(BlockCounters(1u, blockTimestamp, tracesCounters, 150u))
    verify(log).warn(
      "Block 1 has oversize traces TRACE(count, limit, overflow): [ADD(110, 100, 10), BIN(210, 100, 110)]"
    )
    verify(log).warn("Block 1 has oversize data (bytes): blockL1Size=150 perBlockOverhead=10 limit=100 overflow=60")
  }

  @Test
  fun overrideConsumerShouldThrowError() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 9u,
      latestBlockProvider = latestBlockProvider,
      config = noDeadlineConflationConfig
    )
    calculator.onConflatedBatch(conflationConsumer)

    val ex =
      assertThrows<IllegalStateException> { (calculator.onConflatedBatch(conflationConsumer)) }
    assertThat(ex.message).contains("Consumer is already set")
  }

  @Test
  fun shouldCallConflationListenerWhenLimitsAreReached() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 9u,
      latestBlockProvider = latestBlockProvider,
      config = noDeadlineConflationConfig
    )
    calculator.onConflatedBatch(conflationConsumer)

    val tracesCounters10: Map<TracingModule, UInt> = fakeTracesCounters(10u)
    calculator.newBlock(BlockCounters(10u, blockTimestamp, tracesCounters10, 0u))
    assertThat(conflations).isEmpty()
    assertThat(calculator.getConflationInProgress())
      .isEqualTo(
        ConflationCalculationResult(
          10u,
          10u,
          tracesCounters10,
          0u,
          ConflationTrigger.TIME_LIMIT
        )
      )

    val tracesCounters11 = tracesCounters10.toMutableMap()
    tracesCounters11.put(TracingModule.BIN, 80u)
    calculator.newBlock(BlockCounters(11u, blockTimestamp, tracesCounters11, 0u))
    assertThat(conflations).isEmpty()
    assertThat(calculator.getConflationInProgress())
      .isEqualTo(
        ConflationCalculationResult(
          10u,
          11u,
          fakeTracesCounters(20u).apply {
            this as MutableMap
            put(TracingModule.BIN, 90u)
          },
          0u,
          ConflationTrigger.TIME_LIMIT
        )
      )

    val tracesCounters12 = tracesCounters11.toMutableMap()
    tracesCounters12.put(TracingModule.BIN, 10u)
    calculator.newBlock(BlockCounters(12u, blockTimestamp, tracesCounters12, 0u))
    assertThat(calculator.getConflationInProgress()).isNull()
    assertThat(conflations).hasSize(1)
    assertThat(conflations.first())
      .isEqualTo(
        ConflationCalculationResult(
          10u,
          12u,
          fakeTracesCounters(30u).apply {
            this as MutableMap
            put(TracingModule.BIN, 100u)
          },
          0u,
          ConflationTrigger.TRACES_LIMIT
        )
      )
  }

  @Test
  fun `should trigger conflation when data limits are reached`() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val dataLimits = DataLimits(
      totalLimitBytes = 120u,
      perBlockOverheadBytes = 10u,
      minBlockL1SizeBytes = 20u
    )
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 0u,
      latestBlockProvider = latestBlockProvider,
      config = noDeadlineConflationConfig.copy(dataConflationLimits = dataLimits)
    )
    calculator.onConflatedBatch(conflationConsumer)
    val tracesCounters: TracesCounters = fakeTracesCounters(1u)
    calculator.newBlock(BlockCounters(1u, blockTimestamp, tracesCounters, 60u))
    calculator.newBlock(BlockCounters(2u, blockTimestamp, tracesCounters, 35u))
    calculator.newBlock(BlockCounters(3u, blockTimestamp, tracesCounters, 25u))
    // 100 limit reached with 2 blocks and 2*10 overhead
    assertThat(conflations.size).isEqualTo(1)
    assertThat(calculator.getConflationInProgress()).isEqualTo(
      ConflationCalculationResult(
        3u,
        3u,
        fakeTracesCounters(1u),
        35u,
        ConflationTrigger.TIME_LIMIT
      )
    )
    calculator.newBlock(BlockCounters(4u, blockTimestamp, tracesCounters, 55u))
    // 100 limit reached with 1 block 10 overhead, reaches limit 100: 120(limit) - 20(minBlockL1SizeBytes)
    assertThat(conflations.size).isEqualTo(2)
    assertThat(calculator.getConflationInProgress()).isNull()

    assertThat(conflations[0]).isEqualTo(
      ConflationCalculationResult(
        1u,
        2u,
        fakeTracesCounters(2u),
        115u,
        ConflationTrigger.DATA_LIMIT
      )
    )
    assertThat(conflations[1]).isEqualTo(
      ConflationCalculationResult(
        3u,
        4u,
        fakeTracesCounters(2u),
        100u,
        ConflationTrigger.DATA_LIMIT
      )
    )
    assertThat(conflations.size).isEqualTo(2)
  }

  @Test
  fun shouldCallConflationListenerWhenBlocksLimitIsReached() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 9u,
      latestBlockProvider = latestBlockProvider,
      config = noDeadlineConflationConfig.copy(blocksLimit = 5u)
    )
    calculator.onConflatedBatch(conflationConsumer)

    val emptyTraces = tracesCap.mapValues { 0u }
    calculator.newBlock(BlockCounters(10u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(11u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(12u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(13u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(14u, blockTimestamp, emptyTraces, 0u))
    assertThat(calculator.getConflationInProgress()).isNull()
    assertThat(conflations).hasSize(1)
    assertThat(conflations.first())
      .isEqualTo(
        ConflationCalculationResult(
          10u,
          14u,
          emptyTraces,
          0u,
          ConflationTrigger.BLOCKS_LIMIT
        )
      )
  }

  @Test
  fun blockLimitWorksCorrectlyAfterOverflow() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 9u,
      latestBlockProvider = latestBlockProvider,
      config = noDeadlineConflationConfig.copy(blocksLimit = 5u)
    )
    calculator.onConflatedBatch(conflationConsumer)

    val emptyTraces = tracesCap.mapValues { 0u }
    val moreThanAHalfTraces = tracesCap.mapValues { 51u }
    calculator.newBlock(BlockCounters(10u, blockTimestamp, moreThanAHalfTraces, 0u))
    calculator.newBlock(BlockCounters(11u, blockTimestamp, moreThanAHalfTraces, 0u))
    calculator.newBlock(BlockCounters(12u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(13u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(14u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(15u, blockTimestamp, emptyTraces, 0u))
    assertThat(calculator.getConflationInProgress()).isNull()
    assertThat(conflations).hasSameElementsAs(
      listOf(
        ConflationCalculationResult(
          10u,
          10u,
          moreThanAHalfTraces,
          0u,
          ConflationTrigger.TRACES_LIMIT
        ),
        ConflationCalculationResult(
          11u,
          15u,
          moreThanAHalfTraces,
          0u,
          ConflationTrigger.BLOCKS_LIMIT
        )

      )
    )
  }

  @Test
  fun blockLimitWorksCorrectlyWithoutOverflow() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 9u,
      latestBlockProvider = latestBlockProvider,
      config = noDeadlineConflationConfig.copy(blocksLimit = 5u)
    )
    calculator.onConflatedBatch(conflationConsumer)

    val emptyTraces = tracesCap.mapValues { 0u }
    val fullTraces = tracesCap
    calculator.newBlock(BlockCounters(10u, blockTimestamp, fullTraces, 0u))
    calculator.newBlock(BlockCounters(11u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(12u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(13u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(14u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(15u, blockTimestamp, emptyTraces, 0u))
    assertThat(calculator.getConflationInProgress()).isNull()
    assertThat(conflations).hasSameElementsAs(
      listOf(
        ConflationCalculationResult(
          10u,
          10u,
          fullTraces,
          0u,
          ConflationTrigger.TRACES_LIMIT
        ),
        ConflationCalculationResult(
          11u,
          15u,
          emptyTraces,
          0u,
          ConflationTrigger.BLOCKS_LIMIT
        )
      )
    )
  }

  @Test
  fun whenBlockLimitAndOverflowHappenAtTheSameTimeOverflowHandlingShouldPrevail() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 9u,
      latestBlockProvider = latestBlockProvider,
      config = noDeadlineConflationConfig.copy(blocksLimit = 3u)
    )
    calculator.onConflatedBatch(conflationConsumer)

    val emptyTraces = tracesCap.mapValues { 0u }
    val moreThanAHalfTraces = tracesCap.mapValues { 51u }
    calculator.newBlock(BlockCounters(10u, blockTimestamp, moreThanAHalfTraces, 0u))
    calculator.newBlock(BlockCounters(11u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(12u, blockTimestamp, moreThanAHalfTraces, 0u))
    calculator.newBlock(BlockCounters(13u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(14u, blockTimestamp, moreThanAHalfTraces, 0u))
    calculator.newBlock(BlockCounters(15u, blockTimestamp, emptyTraces, 0u))
    calculator.newBlock(BlockCounters(16u, blockTimestamp, emptyTraces, 0u))
    assertThat(calculator.getConflationInProgress()).isNull()
    assertThat(conflations).hasSameElementsAs(
      listOf(
        ConflationCalculationResult(
          10u,
          11u,
          moreThanAHalfTraces,
          0u,
          ConflationTrigger.TRACES_LIMIT
        ),
        ConflationCalculationResult(
          12u,
          13u,
          moreThanAHalfTraces,
          0u,
          ConflationTrigger.TRACES_LIMIT
        ),
        ConflationCalculationResult(
          14u,
          16u,
          moreThanAHalfTraces,
          0u,
          ConflationTrigger.BLOCKS_LIMIT
        )
      )
    )
  }

  @Test
  fun `when conflation deadline has elapsed but there are no blocks, should not trigger conflation flow`() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val lastConflationTime = Instant.parse("2023-07-01T00:00:00Z")
    val conflationDeadline2Hours = 2.hours
    val conflationDeadlineCheckInterval = 10.milliseconds
    val clock: Clock = mock()
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 0u,
      latestBlockProvider = latestBlockProvider,
      config = ConflationCalculatorConfig(
        tracesConflationLimit = tracesCap,
        dataConflationLimits = dataLimits,
        conflationDeadline = conflationDeadline2Hours,
        conflationDeadlineCheckInterval = conflationDeadlineCheckInterval,
        conflationDeadlineLastBlockConfirmationDelay = conflationDeadlineLastblockConfirmationDelay,
        blocksLimit = UInt.MAX_VALUE
      ),
      clock = clock
    )
    calculator.onConflatedBatch(conflationConsumer)
    // no blocks, deadline elapsed, should not trigger any event
    whenever(clock.now()).thenReturn(
      lastConflationTime
        .plus(conflationDeadline2Hours)
        .plus(10.minutes)
    )
    calculator.checkConflationDeadline()
    assertThat(conflations).isEmpty()
  }

  @Test
  fun `when conflation deadline has elapsed, should trigger conflation flow`() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val block1Timestamp = Instant.parse("2023-07-01T00:00:00Z")
    val conflationDeadline2Hours = 2.hours
    val conflationDeadlineCheckInterval = 10.milliseconds
    val clock: Clock = mock { on { now() } doReturn block1Timestamp }
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 0u,
      latestBlockProvider = latestBlockProvider,
      config = ConflationCalculatorConfig(
        tracesConflationLimit = tracesCap,
        dataConflationLimits = dataLimits,
        conflationDeadline = conflationDeadline2Hours,
        conflationDeadlineCheckInterval = conflationDeadlineCheckInterval,
        conflationDeadlineLastBlockConfirmationDelay = conflationDeadlineLastblockConfirmationDelay,
        blocksLimit = UInt.MAX_VALUE
      ),
      clock = clock
    )
    calculator.onConflatedBatch(conflationConsumer)

    val tracesCounters = fakeTracesCounters(1u)
    calculator.newBlock(BlockCounters(1u, block1Timestamp, tracesCounters, 0u))
    whenever(clock.now()).thenReturn(
      block1Timestamp
        .plus(conflationDeadline2Hours)
        .minus(1.seconds)
    )
    calculator.checkConflationDeadline()
    // no conflation should happen.
    assertThat(conflations).isEmpty()
    val block2Timestamp = block1Timestamp
      .plus(conflationDeadline2Hours)
      .plus(1.seconds)

    whenever(clock.now()).thenReturn(block2Timestamp.plus(2.seconds))
    whenever(latestBlockProvider.getLatestSafeBlockHeader()).thenReturn(
      SafeFuture.completedFuture(
        BlockHeaderSummary(
          number = 1UL,
          timestamp = block2Timestamp,
          hash = Bytes32.random()
        )
      )
    )
    // no conflation should happen because tick happened during next block slot which
    // may or may not be in creation process
    calculator.checkConflationDeadline()
    assertThat(conflations).hasSize(0)

    // no conflation should happen we wait confirmation delay
    whenever(clock.now()).thenReturn(
      block2Timestamp
        .plus(conflationDeadlineLastblockConfirmationDelay)
        .plus(1.seconds)
    )
    whenever(latestBlockProvider.getLatestSafeBlockHeader()).thenReturn(
      // return error 1st to make sure it can recover
      SafeFuture.failedFuture(RuntimeException("Request failed!"))
    )
    calculator.checkConflationDeadline()
    assertThat(conflations).hasSize(0)

    whenever(latestBlockProvider.getLatestSafeBlockHeader()).thenReturn(
      SafeFuture.completedFuture(
        BlockHeaderSummary(
          number = 1UL,
          timestamp = block2Timestamp,
          hash = Bytes32.random()
        )
      )
    )

    calculator.checkConflationDeadline()
    assertThat(conflations).hasSize(1)
    assertThat(calculator.getConflationInProgress()).isNull()
    assertThat(conflations).hasSameElementsAs(
      listOf(
        ConflationCalculationResult(
          1u,
          1u,
          tracesCounters,
          0u,
          ConflationTrigger.TIME_LIMIT
        )
      )
    )

    val block2time = block1Timestamp.plus(5.hours)
    calculator.newBlock(BlockCounters(2u, block2time, tracesCounters, 0u))
    whenever(clock.now()).thenReturn(
      block2time
        .plus(conflationDeadline2Hours)
        .minus(1.seconds)
    )
    whenever(latestBlockProvider.getLatestSafeBlockHeader()).thenReturn(
      SafeFuture.completedFuture(
        BlockHeaderSummary(
          number = 2UL,
          timestamp = blockTimestamp,
          hash = Bytes32.random()
        )
      )
    )
    // make sure it does not accidentally tracks block 1 time
    calculator.checkConflationDeadline()
    assertThat(conflations).hasSize(1)
    assertThat(calculator.getConflationInProgress()).isEqualTo(
      ConflationCalculationResult(
        2u,
        2u,
        tracesCounters,
        0u,
        ConflationTrigger.TIME_LIMIT
      )
    )

    whenever(clock.now()).thenReturn(
      block2time
        .plus(conflationDeadline2Hours)
        .plus(1.seconds)
    )
    calculator.checkConflationDeadline()
    assertThat(conflations).hasSize(2)
    assertThat(calculator.getConflationInProgress()).isNull()

    assertThat(conflations[1]).isEqualTo(
      ConflationCalculationResult(
        2u,
        2u,
        tracesCounters,
        0u,
        ConflationTrigger.TIME_LIMIT
      )
    )
  }

  @Test
  fun `conflation deadline should be triggered periodically`() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val block1Timestamp = Instant.parse("2023-07-01T00:00:00Z")
    val conflationDeadline2Hours = 2.hours
    val conflationDeadlineCheckInterval = 10.milliseconds
    val clock: Clock = mock { on { now() } doReturn block1Timestamp }
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 0u,
      latestBlockProvider = latestBlockProvider,
      config = ConflationCalculatorConfig(
        tracesConflationLimit = tracesCap,
        dataConflationLimits = dataLimits,
        conflationDeadline = conflationDeadline2Hours,
        conflationDeadlineCheckInterval = conflationDeadlineCheckInterval,
        conflationDeadlineLastBlockConfirmationDelay = conflationDeadlineLastblockConfirmationDelay,
        blocksLimit = UInt.MAX_VALUE
      ),
      clock = clock
    )
    calculator.onConflatedBatch(conflationConsumer)

    val tracesCounters = fakeTracesCounters(1u)
    calculator.newBlock(BlockCounters(1u, block1Timestamp, tracesCounters, 0u))
    whenever(clock.now()).thenReturn(
      block1Timestamp
        .plus(conflationDeadline2Hours)
        .minus(1.seconds)
    )
    Thread.sleep(conflationDeadlineCheckInterval.inWholeMilliseconds * 5)
    // conflation should not happen
    assertThat(conflations).isEmpty()
    val block2Timestamp = block1Timestamp
      .plus(conflationDeadline2Hours)
      .plus(1.seconds)
    whenever(latestBlockProvider.getLatestSafeBlockHeader()).thenReturn(
      SafeFuture.completedFuture(
        BlockHeaderSummary(
          number = 1UL,
          timestamp = block2Timestamp,
          hash = Bytes32.random()
        )
      )
    )
    whenever(clock.now()).thenReturn(
      block2Timestamp
        .plus(conflationDeadlineLastblockConfirmationDelay)
        .plus(1.seconds)
    )

    await()
      .pollDelay(conflationDeadlineCheckInterval.toJavaDuration())
      .atMost(conflationDeadlineCheckInterval.times(5).toJavaDuration())
      .until { conflations.isNotEmpty() }

    assertThat(conflations).hasSize(1)
    assertThat(calculator.getConflationInProgress()).isNull()
    assertThat(conflations).hasSameElementsAs(
      listOf(
        ConflationCalculationResult(
          1u,
          1u,
          tracesCounters,
          0u,
          ConflationTrigger.TIME_LIMIT
        )
      )
    )
  }

  @Test
  fun `all triggers blended together`() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val block1Timestamp = Instant.parse("2023-07-01T00:00:00Z")
    val conflationDeadline2Hours = 2.hours
    val conflationDeadlineCheckInterval = 10.milliseconds
    val clock: Clock = mock() { on { now() } doReturn block1Timestamp }

    val dataLimits = DataLimits(
      totalLimitBytes = 120u,
      perBlockOverheadBytes = 10u,
      minBlockL1SizeBytes = 20u
    )
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 0u,
      latestBlockProvider = latestBlockProvider,
      config = noDeadlineConflationConfig.copy(
        dataConflationLimits = dataLimits,
        conflationDeadline = conflationDeadline2Hours,
        conflationDeadlineCheckInterval = conflationDeadlineCheckInterval
      ),
      clock = clock
    )
    calculator.onConflatedBatch(conflationConsumer)
    whenever(clock.now()).thenReturn(Instant.parse("2023-07-01T01:10:00Z"))
    // send 1 block, conflation triggered by traces limit. Deadline reset
    calculator.newBlock(BlockCounters(1u, block1Timestamp, tracesCap, 80u))
    assertThat(conflations).hasSize(1)
    assertThat(calculator.getConflationInProgress()).isNull()

    whenever(clock.now()).thenReturn(Instant.parse("2023-07-01T02:10:00Z"))
    calculator.newBlock(BlockCounters(2u, Instant.parse("2023-07-01T02:00:00Z"), fakeTracesCounters(2u), 60u))
    // wait for conflation check tick: no conflation because deadline has not elapsed yet
    // no conflation shall be triggered, must have same event and inprogress conflation
    calculator.checkConflationDeadline()
    assertThat(conflations).hasSize(1)
    assertThat(calculator.getConflationInProgress()).isEqualTo(
      ConflationCalculationResult(
        2u,
        2u,
        fakeTracesCounters(2u),
        70u,
        ConflationTrigger.TIME_LIMIT
      )
    )

    // send 2nd block, conflation triggered data limit. Deadline reset
    whenever(clock.now()).thenReturn(Instant.parse("2023-07-01T03:10:00Z"))
    calculator.newBlock(BlockCounters(3u, Instant.parse("2023-07-01T03:00:00Z"), fakeTracesCounters(3u), 50u))
    // conflation triggered by data overflow by block 3, make sure it resets deadline check internally
    assertThat(conflations).hasSize(2)
    assertThat(calculator.getConflationInProgress()).isEqualTo(
      ConflationCalculationResult(
        3u,
        3u,
        fakeTracesCounters(3u),
        60u,
        ConflationTrigger.TIME_LIMIT
      )
    )
    // wait for conflation check tick: no conflation because deadline has not elapsed yet
    // no conflation shall be triggered, must have same event and inprogress conflation
    whenever(clock.now()).thenReturn(Instant.parse("2023-07-01T04:59:00Z"))
    calculator.checkConflationDeadline()
    assertThat(conflations).hasSize(2)
    assertThat(calculator.getConflationInProgress()).isEqualTo(
      ConflationCalculationResult(
        3u,
        3u,
        fakeTracesCounters(3u),
        60u,
        ConflationTrigger.TIME_LIMIT
      )
    )

    // move clock forward to elapse conflation deadline
    // wait for deadline to elapse, last block should be conflated
    whenever(clock.now()).thenReturn(Instant.parse("2023-07-01T05:01:00Z"))
    whenever(latestBlockProvider.getLatestSafeBlockHeader()).thenReturn(
      SafeFuture.completedFuture(
        BlockHeaderSummary(
          number = 3UL,
          timestamp = blockTimestamp,
          hash = Bytes32.random()
        )
      )
    )
    calculator.checkConflationDeadline()
    assertThat(conflations).hasSize(3)
    assertThat(calculator.getConflationInProgress()).isNull()

    assertThat(conflations).isEqualTo(
      listOf(
        ConflationCalculationResult(
          1u,
          1u,
          tracesCap,
          90u,
          ConflationTrigger.TRACES_LIMIT
        ),
        ConflationCalculationResult(
          2u,
          2u,
          fakeTracesCounters(2u),
          70u,
          ConflationTrigger.DATA_LIMIT
        ),
        ConflationCalculationResult(
          3u,
          3u,
          fakeTracesCounters(3u),
          60u,
          ConflationTrigger.TIME_LIMIT
        )
      )
    )
  }

  @Test
  fun `when a block exceeds limits by itself should trigger flush conflation`() {
    val conflations = mutableListOf<ConflationCalculationResult>()
    val conflationConsumer = { conflationResult: ConflationCalculationResult ->
      conflations.add(conflationResult)
      SafeFuture.completedFuture(Unit)
    }
    val conflationDeadline2Hours = 2.hours
    val conflationDeadlineCheckInterval = 10.milliseconds
    val block1Timestamp = Instant.parse("2023-07-01T00:00:00Z")
    val clock: Clock = mock() { on { now() } doReturn block1Timestamp }

    val dataLimits = DataLimits(
      totalLimitBytes = 120u,
      perBlockOverheadBytes = 10u,
      minBlockL1SizeBytes = 20u
    )
    val calculator = BlocksTracesConflationCalculatorImpl(
      lastBlockNumber = 0u,
      latestBlockProvider = latestBlockProvider,
      config = noDeadlineConflationConfig.copy(
        dataConflationLimits = dataLimits,
        conflationDeadline = conflationDeadline2Hours,
        conflationDeadlineCheckInterval = conflationDeadlineCheckInterval
      ),
      clock = clock
    )
    calculator.onConflatedBatch(conflationConsumer)
    val tracesOverLimit = fakeTracesCounters(2u).toMutableMap().apply {
      put(TracingModule.ADD, 300u)
      put(TracingModule.BIN, 310u)
      put(TracingModule.WCP, 320u)
    }

    // internal empty state - block over limit - trigger conflation
    calculator.newBlock(BlockCounters(1u, block1Timestamp, tracesOverLimit, 10u))
    assertThat(conflations).hasSize(1)
    assertThat(conflations[0]).isEqualTo(
      ConflationCalculationResult(
        1u,
        1u,
        tracesOverLimit,
        20u,
        ConflationTrigger.TRACES_LIMIT
      )
    )

    calculator.newBlock(BlockCounters(2u, block1Timestamp.plus(2.seconds), fakeTracesCounters(2u), 10u))
    calculator.newBlock(BlockCounters(3u, block1Timestamp.plus(4.seconds), tracesOverLimit, 10u))
    assertThat(conflations).hasSize(3)
    assertThat(conflations[1]).isEqualTo(
      ConflationCalculationResult(
        2u,
        2u,
        fakeTracesCounters(2u),
        20u,
        ConflationTrigger.TRACES_LIMIT
      )
    )
    assertThat(conflations[2]).isEqualTo(
      ConflationCalculationResult(
        3u,
        3u,
        tracesOverLimit,
        20u,
        ConflationTrigger.TRACES_LIMIT
      )
    )

    /*
     * conflation triggered by traces overflow, 2 conflations
     */
    calculator.newBlock(BlockCounters(4u, block1Timestamp.plus(8.seconds), fakeTracesCounters(4u), 10u))
    calculator.newBlock(BlockCounters(5u, block1Timestamp.plus(10.seconds), fakeTracesCounters(5u), 10u))
    calculator.newBlock(BlockCounters(6u, block1Timestamp.plus(12.seconds), tracesOverLimit, 10u))
    assertThat(calculator.getConflationInProgress()).isNull()
    assertThat(conflations).hasSize(5)
    assertThat(conflations[3]).isEqualTo(
      ConflationCalculationResult(
        4u,
        5u,
        fakeTracesCounters(9u),
        40u,
        ConflationTrigger.TRACES_LIMIT
      )
    )
    assertThat(conflations[4]).isEqualTo(
      ConflationCalculationResult(
        6u,
        6u,
        tracesOverLimit,
        20u,
        ConflationTrigger.TRACES_LIMIT
      )
    )

    /*
     * conflation triggered by data overflow, 2 conflations
     */
    calculator.newBlock(BlockCounters(7u, block1Timestamp.plus(14.seconds), fakeTracesCounters(7u), 10u))
    calculator.newBlock(BlockCounters(8u, block1Timestamp.plus(16.seconds), fakeTracesCounters(8u), 10u))
    calculator.newBlock(BlockCounters(9u, block1Timestamp.plus(18.seconds), fakeTracesCounters(9u), 300u))
    assertThat(calculator.getConflationInProgress()).isNull()
    assertThat(conflations).hasSize(7)
    assertThat(conflations[5]).isEqualTo(
      ConflationCalculationResult(
        7u,
        8u,
        fakeTracesCounters(15u),
        40u,
        ConflationTrigger.DATA_LIMIT
      )
    )
    assertThat(conflations[6]).isEqualTo(
      ConflationCalculationResult(
        9u,
        9u,
        fakeTracesCounters(9u),
        310u,
        ConflationTrigger.DATA_LIMIT
      )
    )

    /*
     * conflation triggered by data overflow, 2 conflations
     */
    calculator.newBlock(BlockCounters(10u, block1Timestamp.plus(20.seconds), fakeTracesCounters(10u), 10u))
    calculator.newBlock(BlockCounters(11u, block1Timestamp.plus(22.seconds), fakeTracesCounters(11u), 300u))
    assertThat(calculator.getConflationInProgress()).isNull()
    assertThat(conflations).hasSize(9)
    assertThat(conflations[7]).isEqualTo(
      ConflationCalculationResult(
        10u,
        10u,
        fakeTracesCounters(10u),
        20u,
        ConflationTrigger.DATA_LIMIT
      )
    )
    assertThat(conflations[8]).isEqualTo(
      ConflationCalculationResult(
        11u,
        11u,
        fakeTracesCounters(11u),
        310u,
        ConflationTrigger.DATA_LIMIT
      )
    )

    /*
     * inflight conflation is empty, send blocks over limit
     */
    calculator.newBlock(BlockCounters(12u, block1Timestamp.plus(4.seconds), tracesOverLimit, 10u))
    calculator.newBlock(
      BlockCounters(
        13u,
        block1Timestamp.plus(4.seconds),
        fakeTracesCounters(13u),
        300u
      )
    )
    // traces and data over limit
    calculator.newBlock(BlockCounters(14u, block1Timestamp.plus(4.seconds), tracesOverLimit, 300u))
    assertThat(calculator.getConflationInProgress()).isNull()
    assertThat(conflations).hasSize(12)
    assertThat(conflations[9]).isEqualTo(
      ConflationCalculationResult(
        12u,
        12u,
        tracesOverLimit,
        20u,
        ConflationTrigger.TRACES_LIMIT
      )
    )
    assertThat(conflations[10]).isEqualTo(
      ConflationCalculationResult(
        13u,
        13u,
        fakeTracesCounters(13u),
        310u,
        ConflationTrigger.DATA_LIMIT
      )
    )
    assertThat(conflations[11]).isEqualTo(
      ConflationCalculationResult(
        14u,
        14u,
        tracesOverLimit,
        310u,
        ConflationTrigger.DATA_LIMIT
      )
    )
  }
}
