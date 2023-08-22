package net.consensys.zkevm.ethereum.settlement

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.ZkEvmV2AsyncFriendly
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import net.consensys.zkevm.ethereum.settlement.persistence.BatchesRepository
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.RepeatedTest
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.Mockito.RETURNS_DEEP_STUBS
import org.mockito.Mockito.atMost
import org.mockito.Mockito.never
import org.mockito.kotlin.any
import org.mockito.kotlin.atLeast
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.eq
import org.mockito.kotlin.inOrder
import org.mockito.kotlin.mock
import org.mockito.kotlin.times
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.math.BigInteger
import java.util.concurrent.TimeUnit
import kotlin.random.Random
import kotlin.time.Duration.Companion.hours
import kotlin.time.Duration.Companion.milliseconds

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class ZkEvmBatchSubmissionCoordinatorTest {
  private val pollingInterval = 50.milliseconds
  private val proofSubmissionDelay = 2.hours
  private val fixedClockInstant = Instant.parse("2023-07-10T00:00:00Z")
  private val fixedClock = mock<Clock> { on { now() } doReturn fixedClockInstant }

  @BeforeAll
  fun warmup() {
    // To warmup assertions otherwise first test may fail
    assertThat(true).isTrue()
    mock<ZkEvmV2AsyncFriendly>().currentNonce()
  }

  @RepeatedTest(10)
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun acceptNewBatch_addsABatchToTheRepository(vertx: Vertx, testContext: VertxTestContext) {
    val batchesRepository = mock<BatchesRepository>()
    val zkEvmV2 = mock<ZkEvmV2AsyncFriendly>()
    val asyncFriendlyTransactionManager = mock<AsyncFriendlyTransactionManager>()
    whenever(asyncFriendlyTransactionManager.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))
    val zkEvmBatchSubmissionCoordinator =
      ZkEvmBatchSubmissionCoordinator(
        ZkEvmBatchSubmissionCoordinator.Config(pollingInterval, proofSubmissionDelay),
        mock<BatchSubmitter>(),
        batchesRepository,
        zkEvmV2,
        vertx,
        fixedClock
      )
    val proverResponse = mock<GetProofResponse>()
    val batch = Batch(UInt64.ONE, UInt64.valueOf(2), proverResponse)
    whenever(batchesRepository.saveNewBatch(eq(batch))).thenReturn(SafeFuture.completedFuture(Unit))
    zkEvmBatchSubmissionCoordinator
      .acceptNewBatch(batch)
      .thenApply {
        testContext.verify { verify(batchesRepository, times(1)).saveNewBatch(eq(batch)) }
        testContext.completeNow()
      }
      .whenException(testContext::failNow)
  }

  // @RepeatedTest(1)
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  @Test
  fun pollsRepositoryForNewAppropriateBatches(vertx: Vertx, testContext: VertxTestContext) {
    val maxBatchesToReturn = 3
    val maxConflatedBlocksToGenerate = 3
    val expectedNumberOfPolls = 6
    val batchesRepository = mock<BatchesRepository>()
    val zkEvmV2 = mock<ZkEvmV2AsyncFriendly>(defaultAnswer = RETURNS_DEEP_STUBS)
    whenever(zkEvmV2.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))
    whenever(zkEvmV2.currentNonce()).thenReturn(BigInteger.ZERO)
    val asyncFriendlyTransactionManager = mock<AsyncFriendlyTransactionManager>()
    whenever(asyncFriendlyTransactionManager.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))

    var finalizationState = UInt64.ONE
    val batchesSubmitter = mock<BatchSubmitter>()
    whenever(batchesSubmitter.submitBatch(any())).thenAnswer {
      finalizationState = it.getArgument<Batch>(0).endBlockNumber
      SafeFuture.completedFuture(Unit)
    }
    whenever(batchesSubmitter.submitBatchCall(any())).thenReturn(SafeFuture.completedFuture(""))

    whenever(zkEvmV2.currentL2BlockNumber().sendAsync()).thenAnswer {
      SafeFuture.completedFuture(BigInteger.valueOf(finalizationState.longValue()))
    }
    val zkEvmBatchSubmissionCoordinator =
      ZkEvmBatchSubmissionCoordinator(
        ZkEvmBatchSubmissionCoordinator.Config(pollingInterval, proofSubmissionDelay),
        batchesSubmitter,
        batchesRepository,
        zkEvmV2,
        vertx,
        fixedClock
      )

    // It's taking too long to generate on the flight
    val preGeneratedResponses =
      generateChainedBatchesResponses(
        expectedNumberOfPolls,
        maxBatchesToReturn,
        maxConflatedBlocksToGenerate
      )

    makeRepositoryMockReturnResponsesFrom(
      testContext,
      finalizationState + 1,
      batchesRepository,
      preGeneratedResponses
    )

    zkEvmBatchSubmissionCoordinator
      .start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * expectedNumberOfPolls) {
          testContext.verify {
            verify(batchesRepository, atLeast(expectedNumberOfPolls - 1))
              .getConsecutiveBatchesFromBlockNumber(
                any(),
                eq(fixedClockInstant.minus(proofSubmissionDelay))
              )
            val totalBatchesReturned =
              preGeneratedResponses.values.sumOf { it.get().size } - maxBatchesToReturn
            // Last batch might not be submitted yet
            verify(batchesSubmitter, atLeast(totalBatchesReturned - maxBatchesToReturn))
              .submitBatch(any())
          }
          testContext.completeNow()
        }
      }
      .whenException(testContext::failNow)
  }

  @RepeatedTest(10)
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun resistanceToRepositoryFailures(vertx: Vertx, testContext: VertxTestContext) {
    val maxBatchesToReturn = 3
    val maxConflatedBlocksToGenerate = 3
    val expectedNumberOfPolls = 6
    val batchesRepository = mock<BatchesRepository>()
    val zkEvmV2 = mock<ZkEvmV2AsyncFriendly>(defaultAnswer = RETURNS_DEEP_STUBS)
    whenever(zkEvmV2.currentNonce()).thenReturn(BigInteger.ZERO)
    whenever(zkEvmV2.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))
    val asyncFriendlyTransactionManager = mock<AsyncFriendlyTransactionManager>()
    whenever(asyncFriendlyTransactionManager.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))

    var finalizationState = UInt64.ONE
    val batchesSubmitter = mock<BatchSubmitter>()
    whenever(batchesSubmitter.submitBatchCall(any())).thenReturn(SafeFuture.completedFuture(""))
    whenever(batchesSubmitter.submitBatch(any())).thenAnswer {
      finalizationState = it.getArgument<Batch>(0).endBlockNumber
      SafeFuture.completedFuture(Unit)
    }
    whenever(zkEvmV2.currentL2BlockNumber().sendAsync()).thenAnswer {
      SafeFuture.completedFuture(BigInteger.valueOf(finalizationState.longValue()))
    }
    val zkEvmBatchSubmissionCoordinator =
      ZkEvmBatchSubmissionCoordinator(
        ZkEvmBatchSubmissionCoordinator.Config(pollingInterval, proofSubmissionDelay),
        batchesSubmitter,
        batchesRepository,
        zkEvmV2,
        vertx,
        fixedClock
      )

    // It's taking too long to generate on the flight
    val preGeneratedResponses =
      generateChainedBatchesResponses(
        expectedNumberOfPolls,
        maxBatchesToReturn,
        maxConflatedBlocksToGenerate
      )

    makeRepositoryMockReturnResponsesFrom(
      testContext,
      finalizationState + 1,
      batchesRepository,
      preGeneratedResponses
    )

    val repositoryBrokenCheckpoint = testContext.checkpoint()
    val repositoryBrokenAgainCheckpoint = testContext.checkpoint()
    val repositoryRestoredCheckpoint = testContext.checkpoint()
    zkEvmBatchSubmissionCoordinator
      .start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * expectedNumberOfPolls / 2) {
          repositoryBrokenCheckpoint.flag()
          // Making repository fail during next calls
          whenever(batchesRepository.getConsecutiveBatchesFromBlockNumber(any(), any()))
            .thenThrow(RuntimeException("Something went wrong!"))
          testContext.verify {
            verify(batchesRepository, atLeast(expectedNumberOfPolls / 2))
              .getConsecutiveBatchesFromBlockNumber(any(), any())
          }
        }
        vertx.setTimer(pollingInterval.inWholeMilliseconds * expectedNumberOfPolls) {
          repositoryBrokenAgainCheckpoint.flag()
          whenever(batchesRepository.getConsecutiveBatchesFromBlockNumber(any(), any()))
            .thenReturn(
              SafeFuture.failedFuture(
                RuntimeException("Something went better, but still wrong")
              )
            )
          testContext.verify {
            verify(batchesRepository, atLeast(expectedNumberOfPolls - 1))
              .getConsecutiveBatchesFromBlockNumber(any(), any())
            val batchesReturnedByNow =
              preGeneratedResponses.values.take(maxBatchesToReturn + 1).sumOf {
                it.get().size
              }
            // Last batch might not be submitted yet
            verify(batchesSubmitter, atMost(batchesReturnedByNow)).submitBatch(any())
          }
        }
        vertx.setTimer(pollingInterval.inWholeMilliseconds * expectedNumberOfPolls) {
          repositoryRestoredCheckpoint.flag()
          makeRepositoryMockReturnResponsesFrom(
            testContext,
            finalizationState + 1,
            batchesRepository,
            preGeneratedResponses
          )
        }
        vertx.setTimer(pollingInterval.inWholeMilliseconds * expectedNumberOfPolls * 3 / 2) {
          testContext.verify {
            verify(batchesRepository, atLeast((expectedNumberOfPolls * 3 / 2) - 1))
              .getConsecutiveBatchesFromBlockNumber(any(), any())
            val totalBatchesReturned =
              preGeneratedResponses.values.sumOf { it.get().size } - maxBatchesToReturn
            // Last batch might not be submitted yet
            verify(batchesSubmitter, atLeast(totalBatchesReturned)).submitBatch(any())
          }
          testContext.completeNow()
        }
      }
      .whenException(testContext::failNow)
  }

  @RepeatedTest(10)
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun orderTest(vertx: Vertx, testContext: VertxTestContext) {
    val maxBatchesToReturn = 1
    val maxConflatedBlocksToGenerate = 1
    val expectedNumberOfPolls = 3
    val batchesRepository = mock<BatchesRepository>()
    val zkEvmV2 = mock<ZkEvmV2AsyncFriendly>(defaultAnswer = RETURNS_DEEP_STUBS)
    whenever(zkEvmV2.currentNonce()).thenReturn(BigInteger.ZERO)
    whenever(zkEvmV2.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))
    val asyncFriendlyTransactionManager = mock<AsyncFriendlyTransactionManager>()
    whenever(asyncFriendlyTransactionManager.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))

    val initialBlockNumber = 1
    var finalizationState = UInt64.valueOf(initialBlockNumber.toLong())
    val batchesSubmitter = mock<BatchSubmitter>()
    whenever(batchesSubmitter.submitBatch(any())).thenAnswer {
      finalizationState = it.getArgument<Batch>(0).endBlockNumber
      SafeFuture.completedFuture(Unit)
    }
    whenever(batchesSubmitter.submitBatchCall(any())).thenReturn(SafeFuture.completedFuture(""))
    whenever(zkEvmV2.currentL2BlockNumber().sendAsync()).thenAnswer {
      SafeFuture.completedFuture(BigInteger.valueOf(finalizationState.longValue()))
    }
    val zkEvmBatchSubmissionCoordinator =
      ZkEvmBatchSubmissionCoordinator(
        ZkEvmBatchSubmissionCoordinator.Config(pollingInterval, proofSubmissionDelay),
        batchesSubmitter,
        batchesRepository,
        zkEvmV2,
        vertx,
        fixedClock
      )

    // It's taking too long to generate on the flight
    val preGeneratedResponses =
      generateChainedBatchesResponses(
        expectedNumberOfPolls,
        maxBatchesToReturn,
        maxConflatedBlocksToGenerate
      )

    val laggingResponseDelayMillis = pollingInterval.inWholeMilliseconds * 3
    whenever(batchesRepository.getConsecutiveBatchesFromBlockNumber(any(), any()))
      .thenAnswer { invocation ->
        val batchStartBlockNumber = invocation.getArgument<UInt64>(0).intValue()
        val response = preGeneratedResponses[batchStartBlockNumber]!!
        response
      }
      .thenAnswer { invocation ->
        val batchStartBlockNumber = invocation.getArgument<UInt64>(0).intValue()
        val response = preGeneratedResponses[batchStartBlockNumber]!!
        response.thenApply {
          Thread.sleep(laggingResponseDelayMillis)
          it
        }
      }
      .thenAnswer { invocation ->
        val batchStartBlockNumber = invocation.getArgument<UInt64>(0).intValue()
        if (preGeneratedResponses.containsKey(batchStartBlockNumber)) {
          preGeneratedResponses[batchStartBlockNumber]
        } else {
          SafeFuture.completedFuture(emptyList())
        }
      }

    zkEvmBatchSubmissionCoordinator
      .start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * expectedNumberOfPolls + laggingResponseDelayMillis) {
          testContext.verify {
            val inOrder = inOrder(batchesRepository)
            (initialBlockNumber + 1..initialBlockNumber + expectedNumberOfPolls).forEach {
              inOrder
                .verify(batchesRepository)
                .getConsecutiveBatchesFromBlockNumber(eq(UInt64.valueOf(it.toLong())), any())
            }
          }
          testContext.completeNow()
        }
      }
      .whenException(testContext::failNow)
  }

  @RepeatedTest(10)
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun `if first submission fails, none of the submissions are sent`(vertx: Vertx, testContext: VertxTestContext) {
    val maxBatchesToReturn = 1
    val maxConflatedBlocksToGenerate = 1
    val expectedNumberOfPolls = 3
    val batchesRepository = mock<BatchesRepository>()
    val zkEvmV2 = mock<ZkEvmV2AsyncFriendly>(defaultAnswer = RETURNS_DEEP_STUBS)
    whenever(zkEvmV2.currentNonce()).thenReturn(BigInteger.ZERO)
    whenever(zkEvmV2.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))
    val asyncFriendlyTransactionManager = mock<AsyncFriendlyTransactionManager>()
    whenever(asyncFriendlyTransactionManager.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))

    val initialBlockNumber = 1
    var finalizationState = UInt64.valueOf(initialBlockNumber.toLong())
    val batchesSubmitter = mock<BatchSubmitter>()
    whenever(batchesSubmitter.submitBatch(any())).thenAnswer {
      finalizationState = it.getArgument<Batch>(0).endBlockNumber
      SafeFuture.completedFuture(Unit)
    }
    whenever(batchesSubmitter.submitBatchCall(any())).thenReturn(
      SafeFuture.failedFuture<String?>(RuntimeException("eth_call fails!"))
    )
    whenever(zkEvmV2.currentL2BlockNumber().sendAsync()).thenAnswer {
      SafeFuture.completedFuture(BigInteger.valueOf(finalizationState.longValue()))
    }
    val zkEvmBatchSubmissionCoordinator =
      ZkEvmBatchSubmissionCoordinator(
        ZkEvmBatchSubmissionCoordinator.Config(pollingInterval, proofSubmissionDelay),
        batchesSubmitter,
        batchesRepository,
        zkEvmV2,
        vertx,
        fixedClock
      )

    // It's taking too long to generate on the flight
    val preGeneratedResponses =
      generateChainedBatchesResponses(
        expectedNumberOfPolls,
        maxBatchesToReturn,
        maxConflatedBlocksToGenerate
      )

    val laggingResponseDelayMillis = pollingInterval.inWholeMilliseconds * 3
    whenever(batchesRepository.getConsecutiveBatchesFromBlockNumber(any(), any()))
      .thenAnswer { invocation ->
        val batchStartBlockNumber = invocation.getArgument<UInt64>(0).intValue()
        val response = preGeneratedResponses[batchStartBlockNumber]!!
        response
      }
      .thenAnswer { invocation ->
        val batchStartBlockNumber = invocation.getArgument<UInt64>(0).intValue()
        val response = preGeneratedResponses[batchStartBlockNumber]!!
        response.thenApply {
          Thread.sleep(laggingResponseDelayMillis)
          it
        }
      }
      .thenAnswer { invocation ->
        val batchStartBlockNumber = invocation.getArgument<UInt64>(0).intValue()
        if (preGeneratedResponses.containsKey(batchStartBlockNumber)) {
          preGeneratedResponses[batchStartBlockNumber]
        } else {
          SafeFuture.completedFuture(emptyList())
        }
      }

    zkEvmBatchSubmissionCoordinator
      .start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * expectedNumberOfPolls + laggingResponseDelayMillis) {
          testContext.verify {
            verify(batchesSubmitter, never()).submitBatch(any())
          }
          testContext.completeNow()
        }
      }
      .whenException(testContext::failNow)
  }

  private fun makeRepositoryMockReturnResponsesFrom(
    testContext: VertxTestContext,
    initialExpectedRequestedBlockNumber: UInt64,
    batchesRepository: BatchesRepository,
    preGeneratedResponses: Map<Int, SafeFuture<List<Batch>>>
  ) {
    var nextExpectedRequestedBlockNumber = initialExpectedRequestedBlockNumber
    whenever(batchesRepository.getConsecutiveBatchesFromBlockNumber(any(), any())).thenAnswer { invocation
      ->
      val batchStartBlockNumber = invocation.getArgument<UInt64>(0).intValue()
      testContext.verify {
        nextExpectedRequestedBlockNumber.intValue()

        assertThat(batchStartBlockNumber).isEqualTo(nextExpectedRequestedBlockNumber.intValue())
      }
      val response = preGeneratedResponses[batchStartBlockNumber]
      if (response == null) {
        SafeFuture.completedFuture(emptyList())
      } else {
        nextExpectedRequestedBlockNumber = response.get().last().endBlockNumber + 1
        response
      }
    }
  }

  private fun generateChainedBatchesResponses(
    expectedNumberOfPolls: Int,
    maxBatchesToReturn: Int,
    maxConflatedBlocksToGenerate: Int
  ): Map<Int, SafeFuture<List<Batch>>> {
    var previousLastBlockNumber = 1L
    return (1..expectedNumberOfPolls).associate {
      val startBlockNumber = previousLastBlockNumber + 1

      val batchesToGenerate =
        if (maxBatchesToReturn > 1) Random.nextInt(1, maxBatchesToReturn - 1) else 1
      val batches =
        (0 until batchesToGenerate).map {
          val blocksToGenerate =
            if (maxConflatedBlocksToGenerate > 1) {
              Random.nextInt(1, maxConflatedBlocksToGenerate - 1)
            } else {
              1
            }
          Batch(
            UInt64.valueOf(startBlockNumber),
            UInt64.valueOf(startBlockNumber + blocksToGenerate - 1),
            proverResponse = mock<GetProofResponse>()
          )
        }

      previousLastBlockNumber = batches.last().endBlockNumber.longValue()
      batches.first().startBlockNumber.intValue() to SafeFuture.completedFuture(batches)
    }
  }
}
