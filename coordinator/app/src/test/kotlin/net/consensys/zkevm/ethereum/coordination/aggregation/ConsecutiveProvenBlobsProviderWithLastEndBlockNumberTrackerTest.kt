package net.consensys.zkevm.ethereum.coordination.aggregation

import kotlinx.datetime.Instant
import net.consensys.zkevm.domain.BlobAndBatchCounters
import net.consensys.zkevm.domain.BlobCounters
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.VersionedExecutionProofs
import net.consensys.zkevm.persistence.aggregation.AggregationsRepository
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture

class ConsecutiveProvenBlobsProviderWithLastEndBlockNumberTrackerTest {
  @Test
  fun `findConsecutiveProvenBlobs the cache saves the last seen end block number`() {
    val repositoryMock = mock<AggregationsRepository>()
    val cache = ConsecutiveProvenBlobsProviderWithLastEndBlockNumberTracker(repositoryMock, 12UL)

    assertThat(cache.get()).isEqualTo(12L)

    val baseBlobCounters = BlobCounters(
      numberOfBatches = 1U,
      startBlockNumber = 1UL,
      endBlockNumber = 2UL,
      startBlockTimestamp = Instant.DISTANT_PAST,
      endBlockTimestamp = Instant.DISTANT_PAST
    )
    val baseBlobAndBatchCounters = BlobAndBatchCounters(
      baseBlobCounters,
      VersionedExecutionProofs(
        BlockIntervals(1UL, listOf(2UL, 3UL)),
        listOf()
      )
    )
    val expectedEndBLockNumber = 10UL
    val lastBlobAndBatchCounters = baseBlobAndBatchCounters.copy(
      blobCounters = baseBlobCounters.copy(startBlockNumber = 3UL, endBlockNumber = expectedEndBLockNumber)
    )
    whenever(repositoryMock.findConsecutiveProvenBlobs(any())).thenReturn(
      SafeFuture.completedFuture(
        listOf(
          baseBlobAndBatchCounters,
          lastBlobAndBatchCounters
        )
      )
    )

    cache.findConsecutiveProvenBlobs(expectedEndBLockNumber.toLong())
    assertThat(cache.get()).isEqualTo(expectedEndBLockNumber.toLong())
  }
}
