package net.consensys.zkevm.ethereum.finalization

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.setFirstByteToZero
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.createProofToFinalize
import net.consensys.zkevm.persistence.aggregation.AggregationsRepository
import org.apache.logging.log4j.LogManager
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.ArgumentMatchers.eq
import org.mockito.Mockito
import org.mockito.Mockito.spy
import org.mockito.kotlin.any
import org.mockito.kotlin.atLeast
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import org.mockito.kotlin.never
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.TimeUnit
import kotlin.random.Random
import kotlin.time.Duration.Companion.hours
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

@ExtendWith(VertxExtension::class)
class AggregationFinalizationCoordinatorTest {
  private val pollingInterval = 50.milliseconds
  private val maxAggregationsToFinalizePerIteration = 1U
  private val proofSubmissionDelay = 2.hours
  private val fixedClockInstant = Instant.parse("2023-07-10T00:00:00Z")
  private val fixedClock = mock<Clock> { on { now() } doReturn fixedClockInstant }
  private val expectedStartBlockTime = Instant.fromEpochMilliseconds(fixedClock.now().toEpochMilliseconds())
  private val aggregationsRepository = mock<AggregationsRepository>()
  private val lineaRollup = mock<LineaRollupAsyncFriendly>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
  private val asyncFriendlyTransactionManager = mock<AsyncFriendlyTransactionManager>()
  private val aggregationFinalization = mock<AggregationFinalization>()
  private var finalizationState = 0L

  @BeforeEach
  fun beforeEach() {
    // To warmup assertions otherwise first test may fail
    assertThat(true).isTrue()
    mock<LineaRollupAsyncFriendly>().currentNonce()

    whenever(asyncFriendlyTransactionManager.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))

    whenever(lineaRollup.currentNonce()).thenReturn(BigInteger.ZERO)
    whenever(lineaRollup.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))
    whenever(lineaRollup.currentL2BlockNumber().sendAsync()).thenAnswer {
      SafeFuture.completedFuture(BigInteger.valueOf(finalizationState))
    }
    whenever(lineaRollup.dataShnarfHashes(any()).sendAsync()).thenAnswer { invocation ->
      SafeFuture.completedFuture(Random.nextBytes(32).setFirstByteToZero())
    }

    whenever(aggregationFinalization.finalizeAggregation(any())).thenAnswer {
      finalizationState = it.getArgument<ProofToFinalize>(0).finalBlockNumber
      SafeFuture.completedFuture(Unit)
    }
    whenever(aggregationFinalization.finalizeAggregationEthCall(any())).thenAnswer {
      SafeFuture.completedFuture(Unit)
    }
  }

  @Test
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun pollsAggregationsRepositoryForNewAggregation(vertx: Vertx, testContext: VertxTestContext) {
    val expectedNumberOfPolls = 6

    val proofsToFinalize = listOf(
      createProofToFinalize(
        firstBlockNumber = 1L,
        finalBlockNumber = 3L,
        startBlockTime = expectedStartBlockTime
      ),
      createProofToFinalize(
        firstBlockNumber = 4L,
        finalBlockNumber = 4L,
        startBlockTime = expectedStartBlockTime
      ),
      createProofToFinalize(
        firstBlockNumber = 5L,
        finalBlockNumber = 5L,
        startBlockTime = expectedStartBlockTime
      )
    )

    val aggregationFinalizationCoordinator =
      AggregationFinalizationCoordinator(
        config = AggregationFinalizationCoordinator.Config(
          pollingInterval,
          proofSubmissionDelay,
          maxAggregationsToFinalizePerIteration
        ),
        aggregationFinalization = aggregationFinalization,
        aggregationsRepository = aggregationsRepository,
        lineaRollup = lineaRollup,
        vertx = vertx,
        clock = fixedClock
      )

    whenever(aggregationsRepository.getProofsToFinalize(eq(1L), any(), any(), any())).thenAnswer {
      SafeFuture.completedFuture(listOf(proofsToFinalize[0]))
    }
    whenever(aggregationsRepository.getProofsToFinalize(eq(4L), any(), any(), any())).thenAnswer {
      SafeFuture.completedFuture(listOf(proofsToFinalize[1]))
    }
    whenever(aggregationsRepository.getProofsToFinalize(eq(5L), any(), any(), any())).thenAnswer {
      SafeFuture.completedFuture(listOf(proofsToFinalize[2]))
    }
    whenever(aggregationsRepository.getProofsToFinalize(eq(6L), any(), any(), any())).thenAnswer {
      SafeFuture.completedFuture(emptyList<ProofToFinalize>())
    }

    aggregationFinalizationCoordinator
      .start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * expectedNumberOfPolls) {
          testContext.verify {
            verify(aggregationFinalization).finalizeAggregation(proofsToFinalize[0])
          }
          testContext.completeNow()
        }
      }
      .whenException(testContext::failNow)
  }

  @Test
  @Timeout(1, timeUnit = TimeUnit.MINUTES)
  fun resistanceToRepositoryFailures(vertx: Vertx, testContext: VertxTestContext) {
    val proofsToFinalize = listOf(
      createProofToFinalize(1L, 3L, expectedStartBlockTime)
    )

    val aggregationFinalizationCoordinator =
      AggregationFinalizationCoordinator(
        AggregationFinalizationCoordinator.Config(
          pollingInterval,
          proofSubmissionDelay,
          maxAggregationsToFinalizePerIteration
        ),
        aggregationFinalization,
        aggregationsRepository,
        lineaRollup,
        vertx,
        fixedClock
      )

    whenever(aggregationsRepository.getProofsToFinalize(any(), any(), any(), any()))
      .thenThrow(RuntimeException("Something went wrong!"))
      .thenAnswer {
        SafeFuture.failedFuture<List<ProofToFinalize>>(RuntimeException("Something went better, but still wrong"))
      }
      .thenAnswer {
        SafeFuture.completedFuture(proofsToFinalize)
      }.thenAnswer {
        SafeFuture.completedFuture(emptyList<ProofToFinalize>())
      }

    aggregationFinalizationCoordinator
      .start()
      .thenApply {
        await()
          .atMost(5.seconds.toJavaDuration())
          .untilAsserted {
            verify(aggregationFinalization).finalizeAggregation(any())
          }
        verify(aggregationsRepository, atLeast(3)).getProofsToFinalize(any(), any(), any(), any())
        testContext.completeNow()
      }
      .whenException(testContext::failNow)
  }

  @Test
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun `if first submission fails, none of the finalizations are sent`(vertx: Vertx, testContext: VertxTestContext) {
    whenever(aggregationFinalization.finalizeAggregationEthCall(any())).thenAnswer {
      SafeFuture.failedFuture<String?>(RuntimeException("eth_call fails!"))
    }

    val proofsToFinalize = listOf(
      createProofToFinalize(
        firstBlockNumber = 1L,
        finalBlockNumber = 3L,
        startBlockTime = expectedStartBlockTime
      ),
      createProofToFinalize(
        firstBlockNumber = 4L,
        finalBlockNumber = 4L,
        startBlockTime = expectedStartBlockTime
      )
    )

    val aggregationFinalizationCoordinator =
      AggregationFinalizationCoordinator(
        AggregationFinalizationCoordinator.Config(
          pollingInterval,
          proofSubmissionDelay,
          maxAggregationsToFinalizePerIteration
        ),
        aggregationFinalization,
        aggregationsRepository,
        lineaRollup,
        vertx,
        fixedClock
      )

    whenever(aggregationsRepository.getProofsToFinalize(any(), any(), any(), any()))
      .thenAnswer { SafeFuture.completedFuture(proofsToFinalize) }

    aggregationFinalizationCoordinator
      .start()
      .thenApply {
        await()
          .atMost(5.seconds.toJavaDuration())
          .untilAsserted {
            verify(aggregationFinalization, atLeast(1)).finalizeAggregationEthCall(any())
          }
        verify(aggregationFinalization, never()).finalizeAggregation(any())
        testContext.completeNow()
      }
      .whenException(testContext::failNow)
  }

  @Test
  fun `if no aggregations have fully submitted data, no errors are are logged`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val proofsToFinalize = listOf(
      createProofToFinalize(
        firstBlockNumber = 1L,
        finalBlockNumber = 3L,
        startBlockTime = expectedStartBlockTime
      )
    )

    val log = spy(LogManager.getLogger(AggregationFinalizationCoordinator::class.java))
    val aggregationFinalizationCoordinator =
      AggregationFinalizationCoordinator(
        config = AggregationFinalizationCoordinator.Config(
          pollingInterval,
          proofSubmissionDelay,
          maxAggregationsToFinalizePerIteration
        ),
        aggregationFinalization = aggregationFinalization,
        aggregationsRepository = aggregationsRepository,
        lineaRollup = lineaRollup,
        vertx = vertx,
        clock = fixedClock,
        log = log
      )

    whenever(aggregationsRepository.getProofsToFinalize(any(), any(), any(), any()))
      .thenAnswer {
        SafeFuture.completedFuture(proofsToFinalize)
      }

    whenever(lineaRollup.dataShnarfHashes(any()).sendAsync()).thenAnswer { invocation ->
      SafeFuture.completedFuture(Bytes32.ZERO.toArray())
    }

    aggregationFinalizationCoordinator
      .start()
      .thenApply {
        await()
          .untilAsserted {
            verify(aggregationsRepository, atLeast(3)).getProofsToFinalize(any(), any(), any(), any())
            verify(log, never()).error(
              any<String>(),
              any<String>(),
              any()
            )
            testContext.completeNow()
          }
      }
      .whenException(testContext::failNow)
  }
}
