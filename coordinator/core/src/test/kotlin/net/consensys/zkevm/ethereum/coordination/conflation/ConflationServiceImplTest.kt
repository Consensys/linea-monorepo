package net.consensys.zkevm.ethereum.coordination.conflation

import kotlinx.datetime.Instant
import net.consensys.linea.traces.fakeTracesCounters
import net.consensys.zkevm.ethereum.coordination.blockcreation.SafeBlockProvider
import net.consensys.zkevm.toULong
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.RepeatedTest
import org.junit.jupiter.api.Test
import org.mockito.kotlin.mock
import tech.pegasys.teku.ethereum.executionclient.schema.executionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.time.Duration
import java.util.concurrent.Executors
import kotlin.time.Duration.Companion.days
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds

class ConflationServiceImplTest {
  private val conflationBlockLimit = 2u
  private lateinit var conflationCalculator: TracesConflationCalculator
  private lateinit var conflationService: ConflationServiceImpl

  @BeforeEach
  fun beforeEach() {
    val tracesLimits = fakeTracesCounters(100u)
    val dataLimits = DataLimits(100u, 0u, 0u)
    conflationCalculator = BlocksTracesConflationCalculatorImpl(
      0u,
      latestBlockProvider = mock<SafeBlockProvider>(),
      ConflationCalculatorConfig(
        tracesLimits,
        dataLimits,
        conflationDeadline = 2.days,
        conflationDeadlineCheckInterval = 10.minutes,
        conflationDeadlineLastBlockConfirmationDelay = 12.seconds,
        blocksLimit = conflationBlockLimit
      )
    )
    conflationService = ConflationServiceImpl(conflationCalculator)
  }

  @Test
  fun `emits event with blocks when calculator emits conflation`() {
    val payload1 = executionPayloadV1(blockNumber = 1)
    val payload2 = executionPayloadV1(blockNumber = 2)
    val payload3 = executionPayloadV1(blockNumber = 3)
    val payload4 = executionPayloadV1(blockNumber = 4)
    val payload1Time = Instant.parse("2021-01-01T00:00:00Z")
    val payloadCounters1 = BlockCounters(
      blockNumber = 1UL,
      payload1Time.plus(0.seconds),
      tracesCounters = fakeTracesCounters(40u),
      l1DataSize = 10u
    )
    val payloadCounters2 = BlockCounters(
      blockNumber = 2UL,
      payload1Time.plus(2.seconds),
      tracesCounters = fakeTracesCounters(40u),
      l1DataSize = 10u
    )
    val payloadCounters3 = BlockCounters(
      blockNumber = 3UL,
      payload1Time.plus(4.seconds),
      tracesCounters = fakeTracesCounters(100u),
      l1DataSize = 10u
    )
    val payloadCounters4 = BlockCounters(
      blockNumber = 4UL,
      payload1Time.plus(6.seconds),
      tracesCounters = fakeTracesCounters(50u),
      l1DataSize = 10u
    )

    val conflationEvents = mutableListOf<BlocksConflation>()
    conflationService.onConflatedBatch { conflationEvent: BlocksConflation ->
      conflationEvents.add(conflationEvent)
      SafeFuture.completedFuture(Unit)
    }

    // 1st conflation
    conflationService.newBlock(payload1, payloadCounters1)
    conflationService.newBlock(payload2, payloadCounters2)
    conflationService.newBlock(payload3, payloadCounters3)
    conflationService.newBlock(payload4, payloadCounters4)

    assertThat(conflationEvents).isEqualTo(
      listOf(
        BlocksConflation(
          listOf(payload1, payload2),
          ConflationCalculationResult(
            1u,
            2u,
            fakeTracesCounters(80u),
            20u,
            ConflationTrigger.BLOCKS_LIMIT
          )
        ),
        BlocksConflation(
          listOf(payload3),
          ConflationCalculationResult(
            3u,
            3u,
            fakeTracesCounters(100u),
            10u,
            ConflationTrigger.TRACES_LIMIT
          )
        )
      )
    )
  }

  @RepeatedTest(10)
  fun `sends blocks in correct order to calculator`() {
    val numberOfThreads = 10
    val numberOfBlocks = 2000
    val moduleTracesCounter = 10u
    assertThat(numberOfBlocks % numberOfThreads).isEqualTo(0)
    val expectedConflations = numberOfBlocks / conflationBlockLimit.toInt()
    val blocks = (1..numberOfBlocks).map { executionPayloadV1(blockNumber = it.toLong()) }
    val fixedTracesCounters = fakeTracesCounters(moduleTracesCounter)
    val blockTime = Instant.parse("2021-01-01T00:00:00Z")
    val conflationEvents = mutableListOf<BlocksConflation>()
    conflationService.onConflatedBatch { conflationEvent: BlocksConflation ->
      conflationEvents.add(conflationEvent)
      SafeFuture.completedFuture(Unit)
    }
    val blockChunks = blocks.shuffled().chunked(numberOfBlocks / numberOfThreads)
    assertThat(blockChunks.size).isEqualTo(numberOfThreads)

    val executor = Executors.newFixedThreadPool(numberOfThreads)
    blockChunks.forEach { chunck ->
      executor.submit {
        chunck.forEach {
          conflationService.newBlock(
            it,
            BlockCounters(it.blockNumber.toULong(), blockTime, fixedTracesCounters, 1u)
          )
        }
      }
    }

    Awaitility.waitAtMost(Duration.ofSeconds(10)).until {
      conflationEvents.size >= expectedConflations
    }
    executor.shutdown()

    var expectedNexStartBlockNumber: ULong = 1u
    assertThat(conflationEvents.size).isEqualTo(numberOfBlocks / conflationBlockLimit.toInt())
    conflationEvents.forEachIndexed { index, event ->
      assertThat(event.conflationResult.startBlockNumber).isEqualTo(expectedNexStartBlockNumber)
      expectedNexStartBlockNumber = event.conflationResult.endBlockNumber + 1u
    }
    assertThat(conflationService.blocksToConflate).isEmpty()
  }
}
