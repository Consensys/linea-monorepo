package net.consensys.zkevm.ethereum.submission

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.setFirstByteToZero
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.createBlobRecord
import net.consensys.zkevm.ethereum.settlement.BlobSubmitter
import net.consensys.zkevm.persistence.blob.BlobsRepository
import org.assertj.core.api.Assertions
import org.awaitility.Awaitility.await
import org.awaitility.Awaitility.waitAtMost
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.AdditionalMatchers.aryEq
import org.mockito.Mockito
import org.mockito.Mockito.spy
import org.mockito.kotlin.any
import org.mockito.kotlin.argThat
import org.mockito.kotlin.argumentCaptor
import org.mockito.kotlin.atLeast
import org.mockito.kotlin.atMost
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.eq
import org.mockito.kotlin.mock
import org.mockito.kotlin.never
import org.mockito.kotlin.times
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import org.web3j.protocol.core.RemoteFunctionCall
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.CompletableFuture
import java.util.concurrent.TimeUnit
import kotlin.random.Random
import kotlin.time.Duration.Companion.hours
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

@ExtendWith(VertxExtension::class)
class BlobSubmissionCoordinatorImplTest {
  private val pollingInterval = 50.milliseconds
  private val proofSubmissionDelay = 2.hours
  private val fixedClockInstant = Instant.parse("2023-07-10T00:00:00Z")
  private val fixedClock = mock<Clock> { on { now() } doReturn fixedClockInstant }

  private val expectedStartBlockTime = Instant.fromEpochMilliseconds(fixedClock.now().toEpochMilliseconds())

  private val blobsRepository = mock<BlobsRepository>()
  private val lineaRollup = mock<LineaRollupAsyncFriendly>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
  private val asyncFriendlyTransactionManager = mock<AsyncFriendlyTransactionManager>()

  @BeforeEach
  fun warmup() {
    // To warmup assertions otherwise first test may fail
    Assertions.assertThat(true).isTrue()

    whenever(lineaRollup.currentNonce()).thenReturn(BigInteger.ZERO)
    whenever(lineaRollup.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))
    whenever(lineaRollup.resetNonce(any())).thenReturn(SafeFuture.completedFuture(Unit))
    whenever(lineaRollup.updateNonceAndReferenceBlockToLastL1Block()).thenReturn(SafeFuture.completedFuture(Unit))
    whenever(asyncFriendlyTransactionManager.resetNonce()).thenReturn(SafeFuture.completedFuture(Unit))
  }

  @Timeout(15, timeUnit = TimeUnit.SECONDS)
  @Test
  fun pollsBlobsRepositoryForNewConsecutiveBlobsWithEip4844Blobs(vertx: Vertx, testContext: VertxTestContext) {
    var finalizationState = 1L
    val blobsSubmitter = mock<BlobSubmitter>()
    whenever(blobsSubmitter.submitBlob(any())).thenAnswer {
      finalizationState = it.getArgument<BlobRecord>(0).endBlockNumber.toLong()
      SafeFuture.completedFuture("txHash-blob-${it.getArgument<BlobRecord>(0).startBlockNumber}")
    }
    whenever(blobsSubmitter.submitBlobCall(any())).thenReturn(
      SafeFuture.completedFuture(Unit)
    )

    val preGeneratedResponses =
      generateConsecutiveBlobRecordResponses(
        initialBlockNumber = finalizationState,
        numberOfBlobsPerChunk = listOf(2, 3, 1), // 6 blobs total in 3 chunks
        eip4844EnabledBlobs = setOf(5, 6)
      )
    whenever(lineaRollup.currentL2BlockNumber().sendAsync()).thenAnswer {
      SafeFuture.completedFuture(finalizationState.toBigInteger())
    }
    whenever(lineaRollup.dataShnarfHashes(any()).sendAsync()).thenAnswer {
      CompletableFuture.completedFuture(BlobSubmissionCoordinatorImpl.zeroHash)
    }

    val blobSubmissionCoordinator =
      BlobSubmissionCoordinatorImpl(
        BlobSubmissionCoordinatorImpl.Config(
          pollingInterval,
          proofSubmissionDelay,
          maxBlobsToSubmitPerTick = 20u
        ),
        blobsSubmitter,
        blobsRepository,
        lineaRollup,
        vertx,
        fixedClock
      )
    whenever(blobsRepository.getConsecutiveBlobsFromBlockNumber(any(), any()))
      .thenAnswer { invocation ->
        val blobStartBlockNumber = invocation.getArgument<Long>(0).toInt()
        val result = mutableListOf<BlobRecord>()
        for (blobChunk in preGeneratedResponses) {
          val blobs = blobChunk.value
          val blobWithStartBlockNumberIndex = blobs.indexOfFirst {
            it.startBlockNumber == blobStartBlockNumber.toULong()
          }
          if (blobWithStartBlockNumberIndex < 0) {
            continue
          } else {
            result.addAll(blobs.subList(blobWithStartBlockNumberIndex, blobs.size))
            break
          }
        }
        SafeFuture.completedFuture(result.toList())
      }

    blobSubmissionCoordinator
      .start()
      .thenApply {
        waitAtMost(20.seconds.toJavaDuration())
          .pollInterval(pollingInterval.toJavaDuration())
          .untilAsserted {
            verify(blobsSubmitter, times(6)).submitBlob(any())
            verify(blobsRepository, times(1)).getConsecutiveBlobsFromBlockNumber(eq(2L), any())
            verify(blobsRepository, times(1)).getConsecutiveBlobsFromBlockNumber(eq(22L), any())
            verify(blobsRepository, times(1)).getConsecutiveBlobsFromBlockNumber(eq(42L), any())
            verify(blobsRepository, times(1)).getConsecutiveBlobsFromBlockNumber(eq(52L), any())
            verify(blobsRepository, atLeast(1)).getConsecutiveBlobsFromBlockNumber(eq(62L), any())
          }
        testContext.completeNow()
      }
      .whenException(testContext::failNow)
  }

  @Timeout(15, timeUnit = TimeUnit.SECONDS)
  @Test
  fun pollsBlobsRepositoryForNewConsecutiveBlobs(vertx: Vertx, testContext: VertxTestContext) {
    var finalizationState = 1L
    val blobsSubmitter = mock<BlobSubmitter>()
    whenever(blobsSubmitter.submitBlob(any())).thenAnswer {
      finalizationState = it.getArgument<BlobRecord>(0).endBlockNumber.toLong()
      SafeFuture.completedFuture("txHash-blob-${it.getArgument<BlobRecord>(0).startBlockNumber}")
    }
    whenever(blobsSubmitter.submitBlobCall(any())).thenReturn(
      SafeFuture.completedFuture(Unit)
    )

    val preGeneratedResponses =
      generateConsecutiveBlobRecordResponses(
        initialBlockNumber = finalizationState,
        numberOfBlobsPerChunk = listOf(2, 3, 1) // 6 blobs total in 3 chunks
      )
    whenever(lineaRollup.currentL2BlockNumber().sendAsync()).thenAnswer {
      SafeFuture.completedFuture(finalizationState.toBigInteger())
    }
    whenever(lineaRollup.dataShnarfHashes(any()).sendAsync()).thenAnswer {
      CompletableFuture.completedFuture(BlobSubmissionCoordinatorImpl.zeroHash)
    }

    val blobSubmissionCoordinator =
      BlobSubmissionCoordinatorImpl(
        BlobSubmissionCoordinatorImpl.Config(
          pollingInterval,
          proofSubmissionDelay,
          maxBlobsToSubmitPerTick = 20u
        ),
        blobsSubmitter,
        blobsRepository,
        lineaRollup,
        vertx,
        fixedClock
      )

    makeRepositoryMockReturnResponsesFrom(
      blobsRepository,
      preGeneratedResponses
    )

    blobSubmissionCoordinator
      .start()
      .thenApply {
        waitAtMost(20.seconds.toJavaDuration())
          .pollInterval(pollingInterval.toJavaDuration())
          .untilAsserted {
            verify(blobsSubmitter, times(6)).submitBlob(any())
            verify(blobsRepository, times(1)).getConsecutiveBlobsFromBlockNumber(eq(2L), any())
            verify(blobsRepository, times(1)).getConsecutiveBlobsFromBlockNumber(eq(22L), any())
            verify(blobsRepository, times(0)).getConsecutiveBlobsFromBlockNumber(eq(42L), any())
            verify(blobsRepository, times(1)).getConsecutiveBlobsFromBlockNumber(eq(52L), any())
            verify(blobsRepository, atLeast(1)).getConsecutiveBlobsFromBlockNumber(eq(62L), any())
          }
        testContext.completeNow()
      }
      .whenException(testContext::failNow)
  }

  @Timeout(15, timeUnit = TimeUnit.SECONDS)
  @Test
  fun submitsBlobUpToLimit(vertx: Vertx, testContext: VertxTestContext) {
    var finalizationState = 1L
    val blobsSubmitter = mock<BlobSubmitter>()
    whenever(blobsSubmitter.submitBlob(any())).thenAnswer {
      finalizationState = it.getArgument<BlobRecord>(0).endBlockNumber.toLong()
      SafeFuture.completedFuture("txHash-blob-${it.getArgument<BlobRecord>(0).startBlockNumber}")
    }
    whenever(blobsSubmitter.submitBlobCall(any())).thenReturn(SafeFuture.completedFuture(Unit))

    val preGeneratedResponses =
      generateConsecutiveBlobRecordResponses(
        initialBlockNumber = finalizationState,
        numberOfBlobsPerChunk = listOf(20) // 20 blobs
      )
    val blobs = preGeneratedResponses.values.flatten()
    whenever(lineaRollup.currentL2BlockNumber().sendAsync()).thenAnswer {
      SafeFuture.completedFuture(finalizationState.toBigInteger())
    }
    whenever(lineaRollup.dataShnarfHashes(aryEq(blobs[0].blobHash)).sendAsync()).thenAnswer {
      CompletableFuture.completedFuture(blobs[0].expectedShnarf)
    }

    whenever(lineaRollup.dataShnarfHashes(any())).thenAnswer {
      val dataHash = it.getArgument<ByteArray>(0)
      val response = when {
        dataHash.contentEquals(blobs[0].blobHash) -> CompletableFuture.completedFuture(blobs[0].expectedShnarf)
        dataHash.contentEquals(blobs[1].blobHash) -> CompletableFuture.completedFuture(blobs[1].expectedShnarf)
        else -> CompletableFuture.completedFuture(BlobSubmissionCoordinatorImpl.zeroHash)
      }
      val remoteCallResult = mock<RemoteFunctionCall<ByteArray>>()
      whenever(remoteCallResult.sendAsync()).thenReturn(response)
      remoteCallResult
    }

    val blobSubmissionCoordinator =
      BlobSubmissionCoordinatorImpl(
        BlobSubmissionCoordinatorImpl.Config(
          pollingInterval,
          proofSubmissionDelay,
          maxBlobsToSubmitPerTick = 4u
        ),
        blobsSubmitter,
        blobsRepository,
        lineaRollup,
        vertx,
        fixedClock
      )

    makeRepositoryMockReturnResponsesFrom(
      blobsRepository,
      preGeneratedResponses
    )

    blobSubmissionCoordinator
      .start()
      .thenApply {
        waitAtMost(20.seconds.toJavaDuration())
          .pollInterval(pollingInterval.toJavaDuration())
          .untilAsserted {
            verify(blobsRepository, atLeast(2)).getConsecutiveBlobsFromBlockNumber(any(), any())
          }
        verify(blobsSubmitter, never()).submitBlob(eq(blobs[0]))
        verify(blobsSubmitter, never()).submitBlob(eq(blobs[1]))
        verify(blobsSubmitter).submitBlob(eq(blobs[2]))
        verify(blobsSubmitter).submitBlob(eq(blobs[3]))
        verify(blobsSubmitter).submitBlob(eq(blobs[4]))
        verify(blobsSubmitter).submitBlob(eq(blobs[5]))
        verify(blobsSubmitter, never()).submitBlob(eq(blobs[6]))
        testContext.completeNow()
      }
      .whenException(testContext::failNow)
  }

  @Timeout(1, timeUnit = TimeUnit.MINUTES)
  @Test
  fun `if on contract shnarf is non zero, blob is skipped`(vertx: Vertx, testContext: VertxTestContext) {
    var finalizationState = 1L
    val blobsSubmitter = mock<BlobSubmitter>()
    whenever(blobsSubmitter.submitBlob(any())).thenAnswer {
      finalizationState = it.getArgument<BlobRecord>(0).endBlockNumber.toLong()
      SafeFuture.completedFuture("txHash-blob-${it.getArgument<BlobRecord>(0).startBlockNumber}")
    }
    whenever(blobsSubmitter.submitBlobCall(any())).thenReturn(
      SafeFuture.completedFuture(Unit)
    )

    val preGeneratedResponses =
      generateConsecutiveBlobRecordResponses(
        initialBlockNumber = finalizationState,
        numberOfBlobsPerChunk = listOf(2, 3, 1) // 6 blobs total in 3 chunks
      )
    whenever(lineaRollup.currentL2BlockNumber().sendAsync()).thenAnswer {
      SafeFuture.completedFuture(finalizationState.toBigInteger())
    }
    val filteredOutHashes = mutableSetOf<ByteArray>()
    whenever(lineaRollup.dataShnarfHashes(any())).thenAnswer {
      val requestedHash = it.getArgument<ByteArray>(0)
      val returnedResponse = preGeneratedResponses[finalizationState.toInt() + 1]!!
      val blobHashToFilterOut = if (returnedResponse.size > 3) {
        val middle = (returnedResponse.size / 2)
        returnedResponse[middle].blobHash.also {
          filteredOutHashes.add(it)
        }
      } else {
        BlobSubmissionCoordinatorImpl.zeroHash
      }
      val shnarfFuture = if (requestedHash.equals(blobHashToFilterOut)) {
        SafeFuture.completedFuture(Random.nextBytes(32).setFirstByteToZero())
      } else {
        SafeFuture.completedFuture(BlobSubmissionCoordinatorImpl.zeroHash)
      }
      val remoteCallResult = mock<RemoteFunctionCall<ByteArray>>()
      whenever(remoteCallResult.sendAsync()).thenReturn(shnarfFuture)
      remoteCallResult
    }

    val blobSubmissionCoordinator =
      BlobSubmissionCoordinatorImpl(
        BlobSubmissionCoordinatorImpl.Config(
          pollingInterval,
          proofSubmissionDelay,
          maxBlobsToSubmitPerTick = 20u
        ),
        blobsSubmitter,
        blobsRepository,
        lineaRollup,
        vertx,
        fixedClock
      )

    makeRepositoryMockReturnResponsesFrom(
      blobsRepository,
      preGeneratedResponses
    )

    blobSubmissionCoordinator
      .start()
      .thenApply {
        waitAtMost(10.seconds.toJavaDuration())
          .pollInterval(pollingInterval.toJavaDuration())
          .untilAsserted {
            // There is 1 blob hash filtered per poll

            val captor = argumentCaptor<BlobRecord>()
            // Last blob might not be submitted yet
            verify(blobsSubmitter, times(6)).submitBlob(any())
            Assertions.assertThat(
              captor.allValues.all { blobRecord ->
                !filteredOutHashes.contains(blobRecord.blobHash)
              }
            ).isTrue()
            testContext.completeNow()
          }
      }
      .whenException(testContext::failNow)
  }

  @Test
  @Timeout(15, timeUnit = TimeUnit.SECONDS)
  fun resistanceToRepositoryFailures(vertx: Vertx, testContext: VertxTestContext) {
    var finalizationState = 1L
    val blobsSubmitter = mock<BlobSubmitter>()
    val preGeneratedResponses =
      generateConsecutiveBlobRecordResponses(
        initialBlockNumber = finalizationState,
        numberOfBlobsPerChunk = listOf(2, 3, 1) // 6 blobs total in 3 chunks
      )

    whenever(blobsSubmitter.submitBlob(any())).thenAnswer {
      finalizationState = it.getArgument<BlobRecord>(0).endBlockNumber.toLong()
      SafeFuture.completedFuture("txHash-blob-${it.getArgument<BlobRecord>(0).startBlockNumber}")
    }
    whenever(blobsSubmitter.submitBlobCall(any())).thenReturn(
      SafeFuture.completedFuture("")
    )
    whenever(lineaRollup.currentL2BlockNumber().sendAsync()).thenAnswer {
      SafeFuture.completedFuture(finalizationState.toBigInteger())
    }
    whenever(lineaRollup.dataShnarfHashes(any()).sendAsync()).thenAnswer {
      CompletableFuture.completedFuture(BlobSubmissionCoordinatorImpl.zeroHash)
    }
    val blobSubmissionCoordinator =
      BlobSubmissionCoordinatorImpl(
        BlobSubmissionCoordinatorImpl.Config(
          pollingInterval,
          proofSubmissionDelay,
          maxBlobsToSubmitPerTick = 20u
        ),
        blobsSubmitter,
        blobsRepository,
        lineaRollup,
        vertx,
        fixedClock
      )

    whenever(blobsRepository.getConsecutiveBlobsFromBlockNumber(any(), any()))
      .thenThrow(RuntimeException("Something went wrong!"))
      .thenReturn(
        SafeFuture.failedFuture<List<BlobRecord>>(RuntimeException("Something went better, but still wrong"))
      )
      .thenAnswer { invocation ->
        val blobStartBlockNumber = invocation.getArgument<Long>(0).toInt()
        SafeFuture.completedFuture(preGeneratedResponses[blobStartBlockNumber] ?: emptyList())
      }

    blobSubmissionCoordinator
      .start()
      .thenApply {
        await()
          .atMost(5.seconds.toJavaDuration())
          .untilAsserted {
            verify(blobsSubmitter, times(6)).submitBlob(any())
          }
        testContext.verify {
          verify(blobsSubmitter, times(6)).submitBlob(any())
          verify(blobsRepository, atLeast(5)).getConsecutiveBlobsFromBlockNumber(any(), any())
        }
        testContext.completeNow()
      }
      .whenException(testContext::failNow)
  }

  @Test
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun `if first submission fails, none of the submissions are sent`(vertx: Vertx, testContext: VertxTestContext) {
    var finalizationState = 1L
    val blobsSubmitter = mock<BlobSubmitter>()
    whenever(blobsSubmitter.submitBlob(any())).thenAnswer {
      finalizationState = it.getArgument<BlobRecord>(0).endBlockNumber.toLong()
      SafeFuture.completedFuture("txHash-blob-${it.getArgument<BlobRecord>(0).startBlockNumber}")
    }
    whenever(blobsSubmitter.submitBlobCall(any())).thenReturn(
      SafeFuture.failedFuture<String?>(RuntimeException("eth_call fails!"))
    )
    whenever(lineaRollup.currentL2BlockNumber().sendAsync()).thenAnswer {
      SafeFuture.completedFuture(finalizationState.toBigInteger())
    }
    whenever(lineaRollup.dataShnarfHashes(any()).sendAsync()).thenAnswer {
      CompletableFuture.completedFuture(BlobSubmissionCoordinatorImpl.zeroHash)
    }

    // It's taking too long to generate on the flight
    val preGeneratedResponses =
      generateConsecutiveBlobRecordResponses(
        initialBlockNumber = finalizationState,
        numberOfBlobsPerChunk = listOf(3, 2) // 5 blobs total in 2 chunks
      )

    val blobSubmissionCoordinator =
      BlobSubmissionCoordinatorImpl(
        BlobSubmissionCoordinatorImpl.Config(
          pollingInterval,
          proofSubmissionDelay,
          maxBlobsToSubmitPerTick = 20u
        ),
        blobsSubmitter,
        blobsRepository,
        lineaRollup,
        vertx,
        fixedClock
      )

    val laggingResponseDelayMillis = pollingInterval.inWholeMilliseconds * 2
    whenever(blobsRepository.getConsecutiveBlobsFromBlockNumber(any(), any()))
      .thenAnswer { invocation ->
        val blobStartBlockNumber = invocation.getArgument<Long>(0).toInt()
        val response = preGeneratedResponses[blobStartBlockNumber] ?: emptyList()
        Thread.sleep(laggingResponseDelayMillis)
        SafeFuture.completedFuture(response)
      }

    blobSubmissionCoordinator
      .start()
      .thenApply {
        waitAtMost(10.seconds.toJavaDuration())
          .pollInterval(pollingInterval.toJavaDuration())
          .untilAsserted {
            verify(blobsSubmitter, atLeast(1)).submitBlobCall(any())
          }
        verify(blobsRepository, atLeast(1))
          .getConsecutiveBlobsFromBlockNumber(
            argThat { blockNumber: Long -> blockNumber == finalizationState + 1L },
            any()
          )
        verify(blobsSubmitter, Mockito.never()).submitBlob(any())
        testContext.completeNow()
      }
      .whenException(testContext::failNow)
  }

  @Test
  fun `nonce and reference block are updated at the start of each tick`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val finalizationState = 1L
    val blobsSubmitter = mock<BlobSubmitter>()
    whenever(lineaRollup.currentL2BlockNumber().sendAsync()).thenAnswer {
      SafeFuture.completedFuture(finalizationState.toBigInteger())
    }

    val blobSubmissionCoordinator =
      spy(
        BlobSubmissionCoordinatorImpl(
          BlobSubmissionCoordinatorImpl.Config(
            250.milliseconds,
            proofSubmissionDelay,
            maxBlobsToSubmitPerTick = 20u
          ),
          blobsSubmitter,
          blobsRepository,
          lineaRollup,
          vertx,
          fixedClock
        )
      )

    whenever(blobsRepository.getConsecutiveBlobsFromBlockNumber(any(), any())).thenAnswer {
      SafeFuture.completedFuture(emptyList<BlobRecord>())
    }

    blobSubmissionCoordinator.start()
      .thenApply {
        await().untilAsserted {
          verify(blobSubmissionCoordinator, times(5)).action()
          verify(lineaRollup, times(5)).updateNonceAndReferenceBlockToLastL1Block()
          testContext.completeNow()
        }
      }
      .whenException(testContext::failNow)
  }

  private fun makeRepositoryMockReturnResponsesFrom(
    blobsRepository: BlobsRepository,
    preGeneratedResponses: Map<Int, List<BlobRecord>>
  ) {
    whenever(blobsRepository.getConsecutiveBlobsFromBlockNumber(any(), any())).thenAnswer { invocation
      ->
      val blobStartBlockNumber = invocation.getArgument<Long>(0).toInt()
      preGeneratedResponses[blobStartBlockNumber]
        ?.let { SafeFuture.completedFuture(it) }
        ?: SafeFuture.completedFuture<List<BlobRecord>>(emptyList())
    }
  }

  private fun generateConsecutiveBlobRecordResponses(
    initialBlockNumber: Long,
    numberOfBlobsPerChunk: List<Int>, // List of number of blobs to generate per chunk
    maxBlocksPerBlob: Int = 10,
    eip4844EnabledBlobs: Set<Int> = setOf() // blobs numbered 1 to sum(blobChunks) with eip4844 enabled
  ): Map<Int, List<BlobRecord>> {
    require(eip4844EnabledBlobs.size <= numberOfBlobsPerChunk.sum())
    var blobNumber = 0
    var previousLastBlockNumber: ULong = initialBlockNumber.toULong()
    val result = mutableMapOf<Int, List<BlobRecord>>()
    numberOfBlobsPerChunk.forEach { chuckNumberOfBlobs ->
      val blobChunk = (0 until chuckNumberOfBlobs).map {
        blobNumber += 1
        val startBlockNumber = previousLastBlockNumber + 1UL
        val endBlockNumber = (startBlockNumber + maxBlocksPerBlob.toULong() - 1UL)
        val blobRecord = createBlobRecord(
          startBlockNumber = startBlockNumber,
          endBlockNumber = endBlockNumber,
          eip4844Enabled = blobNumber in eip4844EnabledBlobs,
          startBlockTime = expectedStartBlockTime
        )
        previousLastBlockNumber = endBlockNumber
        blobRecord
      }
      result[blobChunk.first().startBlockNumber.toInt()] = blobChunk
    }

    return result
  }
}
