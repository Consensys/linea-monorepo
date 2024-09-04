package net.consensys.zkevm.ethereum.coordinator.blob

import com.github.michaelbull.result.Ok
import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import net.consensys.FakeFixedClock
import net.consensys.linea.traces.TracesCountersV1
import net.consensys.zkevm.coordinator.clients.BlobCompressionProof
import net.consensys.zkevm.coordinator.clients.BlobCompressionProverClient
import net.consensys.zkevm.domain.Blob
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.ConflationCalculationResult
import net.consensys.zkevm.domain.ConflationTrigger
import net.consensys.zkevm.ethereum.coordination.blob.BlobCompressionProofCoordinator
import net.consensys.zkevm.ethereum.coordination.blob.BlobZkState
import net.consensys.zkevm.ethereum.coordination.blob.BlobZkStateProvider
import net.consensys.zkevm.ethereum.coordination.blob.RollingBlobShnarfCalculator
import net.consensys.zkevm.ethereum.coordination.blob.RollingBlobShnarfResult
import net.consensys.zkevm.ethereum.coordination.blob.ShnarfResult
import net.consensys.zkevm.persistence.blob.BlobsRepository
import org.awaitility.Awaitility.await
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.Mockito
import org.mockito.kotlin.any
import org.mockito.kotlin.eq
import org.mockito.kotlin.mock
import org.mockito.kotlin.times
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.lang.RuntimeException
import kotlin.random.Random
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class BlobCompressionProofCoordinatorTest {
  private lateinit var blobCompressionProofCoordinator: BlobCompressionProofCoordinator
  private val rollingBlobShnarfCalculator = mock<RollingBlobShnarfCalculator>()

  private val fixedClock = FakeFixedClock()
  private val blobHandlerPollingInterval = 50.milliseconds

  private val expectedStartBlock = 1UL
  private val expectedEndBlock = 100UL
  private val blobCompressionProverClient = mock<BlobCompressionProverClient>().also {
    val expectedBlobCompressionProofResponse = BlobCompressionProof(
      compressedData = Random.nextBytes(32),
      conflationOrder = BlockIntervals(
        startingBlockNumber =
        expectedStartBlock,
        upperBoundaries =
        listOf(expectedEndBlock)
      ),
      prevShnarf = Random.nextBytes(32),
      parentStateRootHash = Random.nextBytes(32),
      finalStateRootHash = Random.nextBytes(32),
      parentDataHash = Random.nextBytes(32),
      dataHash = Random.nextBytes(32),
      snarkHash = Random.nextBytes(32),
      expectedX = Random.nextBytes(32),
      expectedY = Random.nextBytes(32),
      expectedShnarf = Random.nextBytes(32),
      decompressionProof = Random.nextBytes(512),
      proverVersion = "mock-0.0.0",
      verifierID = 6789,
      commitment = Random.nextBytes(48),
      kzgProofContract = Random.nextBytes(48),
      kzgProofSidecar = Random.nextBytes(48)
    )

    whenever(
      it.requestBlobCompressionProof(any(), any(), any(), any(), any(), any(), any(), any(), any(), any())
    )
      .thenReturn(SafeFuture.completedFuture(Ok(expectedBlobCompressionProofResponse)))
  }
  private val blobZkStateProvider = mock<BlobZkStateProvider>()
  private val blobsRepository = mock<BlobsRepository>()

  @BeforeEach
  fun beforeEach(vertx: Vertx) {
    blobCompressionProofCoordinator = BlobCompressionProofCoordinator(
      vertx = vertx,
      blobsRepository = blobsRepository,
      blobCompressionProverClient = blobCompressionProverClient,
      rollingBlobShnarfCalculator = rollingBlobShnarfCalculator,
      blobZkStateProvider = blobZkStateProvider,
      config = BlobCompressionProofCoordinator.Config(
        pollingInterval = blobHandlerPollingInterval
      ),
      blobCompressionProofHandler = { _ -> SafeFuture.completedFuture(Unit) },
      metricsFacade = mock(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
    )
    blobCompressionProofCoordinator.start()
  }

  @Test
  fun `guarantee requestBlobCompressionProof is being passed the correct values in the correct order`() {
    val startBlockTime = fixedClock.now()
    val expectedParentDataHash = Random.nextBytes(32)
    val expectedPrevShnarf = Random.nextBytes(32)
    val parentStateRootHash = Random.nextBytes(32)
    val finalStateRootHash = Random.nextBytes(32)

    val blob = Blob(
      conflations = listOf(
        ConflationCalculationResult(
          startBlockNumber = expectedStartBlock,
          endBlockNumber = expectedEndBlock,
          conflationTrigger = ConflationTrigger.TRACES_LIMIT,
          tracesCounters = TracesCountersV1.EMPTY_TRACES_COUNT
        )
      ),
      compressedData = Random.nextBytes(128),
      startBlockTime = startBlockTime,
      endBlockTime = fixedClock.now().plus((12 * (expectedEndBlock - expectedStartBlock).toInt()).seconds)
    )

    whenever(blobZkStateProvider.getBlobZKState(any()))
      .thenReturn(
        SafeFuture.completedFuture(
          BlobZkState(
            parentStateRootHash = parentStateRootHash,
            finalStateRootHash = finalStateRootHash
          )
        )
      )

    val shnarfResult = ShnarfResult(
      dataHash = Random.nextBytes(32),
      snarkHash = Random.nextBytes(32),
      expectedX = Random.nextBytes(32),
      expectedY = Random.nextBytes(32),
      expectedShnarf = Random.nextBytes(32),
      commitment = Random.nextBytes(48),
      kzgProofContract = Random.nextBytes(48),
      kzgProofSideCar = Random.nextBytes(48)
    )

    whenever(rollingBlobShnarfCalculator.calculateShnarf(any(), any(), any(), any()))
      .thenAnswer {
        SafeFuture.completedFuture(
          RollingBlobShnarfResult(
            shnarfResult = shnarfResult,
            parentBlobHash = expectedParentDataHash,
            parentBlobShnarf = expectedPrevShnarf
          )
        )
      }

    blobCompressionProofCoordinator.handleBlob(blob).get()

    await()
      .untilAsserted {
        verify(blobCompressionProverClient)
          .requestBlobCompressionProof(
            compressedData = eq(blob.compressedData),
            conflations = eq(blob.conflations),
            parentStateRootHash = eq(parentStateRootHash),
            finalStateRootHash = eq(finalStateRootHash),
            parentDataHash = eq(expectedParentDataHash),
            prevShnarf = eq(expectedPrevShnarf),
            expectedShnarfResult = eq(shnarfResult),
            commitment = eq(shnarfResult.commitment),
            kzgProofContract = eq(shnarfResult.kzgProofContract),
            kzgProofSideCar = eq(shnarfResult.kzgProofSideCar)
          )
      }
  }

  @Test
  fun `verify failed blob is re-queued and processed in order`() {
    val startBlockTime = fixedClock.now()
    val expectedParentDataHash = Random.nextBytes(32)
    val expectedPrevShnarf = Random.nextBytes(32)
    val parentStateRootHash = Random.nextBytes(32)
    val finalStateRootHash = Random.nextBytes(32)

    val blob1 = Blob(
      conflations = listOf(
        ConflationCalculationResult(
          startBlockNumber = 1uL,
          endBlockNumber = 10uL,
          conflationTrigger = ConflationTrigger.TRACES_LIMIT,
          tracesCounters = TracesCountersV1.EMPTY_TRACES_COUNT
        )
      ),
      compressedData = Random.nextBytes(128),
      startBlockTime = startBlockTime,
      endBlockTime = fixedClock.now().plus((12 * (10 - 1)).seconds)
    )

    val blob2 = Blob(
      conflations = listOf(
        ConflationCalculationResult(
          startBlockNumber = 11uL,
          endBlockNumber = 20uL,
          conflationTrigger = ConflationTrigger.TRACES_LIMIT,
          tracesCounters = TracesCountersV1.EMPTY_TRACES_COUNT
        )
      ),
      compressedData = Random.nextBytes(128),
      startBlockTime = startBlockTime,
      endBlockTime = fixedClock.now().plus((12 * (20 - 11)).seconds)
    )

    whenever(blobZkStateProvider.getBlobZKState(any()))
      .thenReturn(
        SafeFuture.failedFuture(RuntimeException("Forced blobZkStateProvider mock failure")),
        SafeFuture.completedFuture(
          BlobZkState(
            parentStateRootHash = parentStateRootHash,
            finalStateRootHash = finalStateRootHash
          )
        ),
        SafeFuture.completedFuture(
          BlobZkState(
            parentStateRootHash = parentStateRootHash,
            finalStateRootHash = finalStateRootHash
          )
        )
      )

    val shnarfResult = ShnarfResult(
      dataHash = Random.nextBytes(32),
      snarkHash = Random.nextBytes(32),
      expectedX = Random.nextBytes(32),
      expectedY = Random.nextBytes(32),
      expectedShnarf = Random.nextBytes(32),
      commitment = Random.nextBytes(48),
      kzgProofContract = Random.nextBytes(48),
      kzgProofSideCar = Random.nextBytes(48)
    )

    whenever(rollingBlobShnarfCalculator.calculateShnarf(any(), any(), any(), any()))
      .thenAnswer {
        SafeFuture.completedFuture(
          RollingBlobShnarfResult(
            shnarfResult = shnarfResult,
            parentBlobHash = expectedParentDataHash,
            parentBlobShnarf = expectedPrevShnarf
          )
        )
      }

    blobCompressionProofCoordinator.handleBlob(blob1).get()
    blobCompressionProofCoordinator.handleBlob(blob2).get()

    await()
      .untilAsserted {
        verify(blobCompressionProverClient, times(1))
          .requestBlobCompressionProof(
            compressedData = eq(blob1.compressedData),
            conflations = eq(blob1.conflations),
            parentStateRootHash = eq(parentStateRootHash),
            finalStateRootHash = eq(finalStateRootHash),
            parentDataHash = eq(expectedParentDataHash),
            prevShnarf = eq(expectedPrevShnarf),
            expectedShnarfResult = eq(shnarfResult),
            commitment = eq(shnarfResult.commitment),
            kzgProofContract = eq(shnarfResult.kzgProofContract),
            kzgProofSideCar = eq(shnarfResult.kzgProofSideCar)
          )
        verify(blobCompressionProverClient, times(1))
          .requestBlobCompressionProof(
            compressedData = eq(blob2.compressedData),
            conflations = eq(blob2.conflations),
            parentStateRootHash = eq(parentStateRootHash),
            finalStateRootHash = eq(finalStateRootHash),
            parentDataHash = eq(expectedParentDataHash),
            prevShnarf = eq(expectedPrevShnarf),
            expectedShnarfResult = eq(shnarfResult),
            commitment = eq(shnarfResult.commitment),
            kzgProofContract = eq(shnarfResult.kzgProofContract),
            kzgProofSideCar = eq(shnarfResult.kzgProofSideCar)
          )
      }
  }
}
