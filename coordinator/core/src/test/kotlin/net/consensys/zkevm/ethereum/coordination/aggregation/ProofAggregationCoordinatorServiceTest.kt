package net.consensys.zkevm.ethereum.coordination.aggregation

import build.linea.domain.BlockIntervals
import io.vertx.core.Vertx
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.trimToSecondPrecision
import net.consensys.zkevm.coordinator.clients.ProofAggregationProverClientV2
import net.consensys.zkevm.domain.Aggregation
import net.consensys.zkevm.domain.BlobAndBatchCounters
import net.consensys.zkevm.domain.BlobCounters
import net.consensys.zkevm.domain.BlobsToAggregate
import net.consensys.zkevm.domain.ProofIndex
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.ProofsToAggregate
import net.consensys.zkevm.persistence.AggregationsRepository
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.Mockito.anyLong
import org.mockito.Mockito.mock
import org.mockito.kotlin.any
import org.mockito.kotlin.argThat
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.function.Consumer
import kotlin.random.Random
import kotlin.time.Duration.Companion.milliseconds

class ProofAggregationCoordinatorServiceTest {
  private val mockVertx = mock<Vertx>()
  private val blobsToPoll = 500U

  private fun createBlob(startBlockNumber: ULong, endBLockNumber: ULong): BlobAndBatchCounters {
    val batches = BlockIntervals(
      startBlockNumber,
      listOf(startBlockNumber + 2UL, startBlockNumber + 6UL, endBLockNumber)
    )

    val blobCounters = BlobCounters(
      3u,
      startBlockNumber,
      endBLockNumber,
      Instant.fromEpochMilliseconds(100),
      Instant.fromEpochMilliseconds(5000),
      expectedShnarf = Random.nextBytes(32)
    )
    return BlobAndBatchCounters(blobCounters = blobCounters, executionProofs = batches)
  }

  private val aggregationProofResponse = ProofToFinalize(
    aggregatedProof = "mock_aggregatedProof".toByteArray(),
    aggregatedVerifierIndex = 1,
    aggregatedProofPublicInput = "mock_aggregatedProofPublicInput".toByteArray(),
    dataHashes = listOf("mock_dataHashes_1".toByteArray()),
    dataParentHash = "mock_dataParentHash".toByteArray(),
    parentStateRootHash = "mock_parentStateRootHash".toByteArray(),
    parentAggregationLastBlockTimestamp = Clock.System.now().trimToSecondPrecision(),
    finalTimestamp = Clock.System.now().trimToSecondPrecision(),
    firstBlockNumber = 1,
    finalBlockNumber = 23,
    l1RollingHash = "mock_l1RollingHash".toByteArray(),
    l1RollingHashMessageNumber = 4,
    l2MerkleRoots = listOf("mock_l2MerkleRoots".toByteArray()),
    l2MerkleTreesDepth = 5,
    l2MessagingBlocksOffsets = "mock_l2MessagingBlocksOffsets".toByteArray()
  )

  @Test
  fun `test aggregation flow`() {
    // FIXME this it's only happy path, with should cover other scenarios
    val mockAggregationCalculator = mock<AggregationCalculator>()
    val mockAggregationsRepository = mock<AggregationsRepository>()
    val mockProofAggregationClient = mock<ProofAggregationProverClientV2>()
    val aggregationL2StateProvider = mock<AggregationL2StateProvider>()

    val config = ProofAggregationCoordinatorService.Config(
      pollingInterval = 10.milliseconds,
      proofsLimit = blobsToPoll
    )

    var provenAggregation = 0UL

    val provenAggregationEndBlockNumberConsumer = Consumer<ULong> { provenAggregation = it }
    val proofAggregationCoordinatorService = ProofAggregationCoordinatorService(
      vertx = mockVertx,
      config = config,
      nextBlockNumberToPoll = 10L,
      aggregationCalculator = mockAggregationCalculator,
      aggregationsRepository = mockAggregationsRepository,
      consecutiveProvenBlobsProvider = mockAggregationsRepository::findConsecutiveProvenBlobs,
      proofAggregationClient = mockProofAggregationClient,
      aggregationL2StateProvider = aggregationL2StateProvider,
      provenAggregationEndBlockNumberConsumer = provenAggregationEndBlockNumberConsumer
    )
    verify(mockAggregationCalculator).onAggregation(proofAggregationCoordinatorService)

    val blob1 = listOf(createBlob(11u, 19u), createBlob(20u, 33u), createBlob(34u, 41u))
    val blob2 = listOf(createBlob(42u, 60u))
    whenever(mockAggregationsRepository.findConsecutiveProvenBlobs(anyLong()))
      .thenAnswer {
        if (it.getArgument<Long>(0) == 10L) {
          SafeFuture.completedFuture(blob1)
        } else if (it.getArgument<Long>(0) == 42L) {
          SafeFuture.completedFuture(blob2)
        } else {
          throw IllegalStateException()
        }
      }

    val blobsToAggregate1 = BlobsToAggregate(11u, 33u)
    val blobsToAggregate2 = BlobsToAggregate(34u, 60u)

    whenever(mockAggregationCalculator.newBlob(any<BlobCounters>())).thenAnswer {
      if (it.getArgument<BlobCounters>(0) == blob1[1].blobCounters) {
        proofAggregationCoordinatorService.onAggregation(blobsToAggregate1)
      } else if (it.getArgument<BlobCounters>(0) == blob2[0].blobCounters) {
        proofAggregationCoordinatorService.onAggregation(blobsToAggregate2)
      } else {
        SafeFuture.completedFuture(Unit)
      }
    }

    val compressionBlobs1 = listOf(blob1[0], blob1[1])
    val compressionBlobs2 = listOf(blob1[2], blob2[0])

    val executionProofs1 = BlockIntervals(
      blob1[0].executionProofs.startingBlockNumber,
      blob1[0].executionProofs.upperBoundaries + blob1[1].executionProofs.upperBoundaries
    )

    val executionProofs2 = BlockIntervals(
      blob1[2].executionProofs.startingBlockNumber,
      blob1[2].executionProofs.upperBoundaries + blob2[0].executionProofs.upperBoundaries
    )

    val rollingInfo1 = AggregationL2State(
      parentAggregationLastBlockTimestamp = Instant.fromEpochSeconds(123456),
      parentAggregationLastL1RollingHashMessageNumber = 12UL,
      parentAggregationLastL1RollingHash = ByteArray(32)
    )
    val rollingInfo2 = AggregationL2State(
      parentAggregationLastBlockTimestamp = Instant.fromEpochSeconds(123458),
      parentAggregationLastL1RollingHashMessageNumber = 14UL,
      parentAggregationLastL1RollingHash = ByteArray(32)
    )

    whenever(aggregationL2StateProvider.getAggregationL2State(anyLong()))
      .thenAnswer { SafeFuture.completedFuture(rollingInfo1) }
      .thenAnswer { SafeFuture.completedFuture(rollingInfo2) }

    val proofsToAggregate1 = ProofsToAggregate(
      compressionProofIndexes = compressionBlobs1.map {
        ProofIndex(
          it.blobCounters.startBlockNumber,
          it.blobCounters.endBlockNumber,
          it.blobCounters.expectedShnarf
        )
      },
      executionProofs = executionProofs1,
      parentAggregationLastBlockTimestamp = rollingInfo1.parentAggregationLastBlockTimestamp,
      parentAggregationLastL1RollingHashMessageNumber = rollingInfo1.parentAggregationLastL1RollingHashMessageNumber,
      parentAggregationLastL1RollingHash = rollingInfo1.parentAggregationLastL1RollingHash
    )

    val proofsToAggregate2 = ProofsToAggregate(
      compressionProofIndexes = compressionBlobs2.map {
        ProofIndex(
          it.blobCounters.startBlockNumber,
          it.blobCounters.endBlockNumber,
          it.blobCounters.expectedShnarf
        )
      },
      executionProofs = executionProofs2,
      parentAggregationLastBlockTimestamp = rollingInfo2.parentAggregationLastBlockTimestamp,
      parentAggregationLastL1RollingHashMessageNumber = rollingInfo2.parentAggregationLastL1RollingHashMessageNumber,
      parentAggregationLastL1RollingHash = rollingInfo2.parentAggregationLastL1RollingHash
    )

    val aggregationProof1 = aggregationProofResponse.copy(finalBlockNumber = 23)
    val aggregationProof2 = aggregationProofResponse.copy(finalBlockNumber = 50)

    val aggregation1 = Aggregation(
      startBlockNumber = blobsToAggregate1.startBlockNumber,
      endBlockNumber = blobsToAggregate1.endBlockNumber,
      batchCount = compressionBlobs1.sumOf { it.blobCounters.numberOfBatches }.toULong(),
      aggregationProof = aggregationProof1
    )

    val aggregation2 = Aggregation(
      startBlockNumber = blobsToAggregate2.startBlockNumber,
      endBlockNumber = blobsToAggregate2.endBlockNumber,
      batchCount = compressionBlobs2.sumOf { it.blobCounters.numberOfBatches }.toULong(),
      aggregationProof = aggregationProof2
    )

    whenever(mockProofAggregationClient.requestProof(any()))
      .thenAnswer {
        if (it.getArgument<ProofsToAggregate>(0) == proofsToAggregate1) {
          SafeFuture.completedFuture(aggregationProof1)
        } else if (it.getArgument<ProofsToAggregate>(0) == proofsToAggregate2) {
          SafeFuture.completedFuture(aggregationProof2)
        } else {
          throw IllegalStateException()
        }
      }

    whenever(
      mockAggregationsRepository.saveNewAggregation(
        argThat<Aggregation> {
          this == aggregation1 || this == aggregation2
        }
      )
    )
      .thenReturn(SafeFuture.completedFuture(Unit))

    // First aggregation should Trigger
    proofAggregationCoordinatorService.action().get()

    verify(mockProofAggregationClient).requestProof(proofsToAggregate1)
    verify(mockAggregationsRepository).saveNewAggregation(aggregation1)
    assertThat(provenAggregation).isEqualTo(aggregation1.endBlockNumber)

    // Second aggregation should Trigger
    proofAggregationCoordinatorService.action().get()

    verify(mockProofAggregationClient).requestProof(proofsToAggregate2)
    verify(mockAggregationsRepository).saveNewAggregation(aggregation2)

    assertThat(provenAggregation).isEqualTo(aggregation2.endBlockNumber)
  }
}
