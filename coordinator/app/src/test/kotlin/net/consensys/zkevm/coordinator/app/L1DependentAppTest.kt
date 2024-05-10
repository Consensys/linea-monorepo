package net.consensys.zkevm.coordinator.app

import net.consensys.zkevm.persistence.aggregation.AggregationsRepository
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.Mockito.mock
import org.mockito.Mockito.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture

class L1DependentAppTest {
  @Test
  fun `test resume conflation from uses lastFinalizedBlock + 1 for db queries`() {
    val aggregationsRepository = mock<AggregationsRepository>()
    val lastFinalizedBlock = 100uL

    whenever(aggregationsRepository.findConsecutiveProvenBlobs(101L))
      .thenReturn(SafeFuture.completedFuture(emptyList()))

    val lastProcessedBlock =
      L1DependentApp.resumeConflationFrom(
        aggregationsRepository,
        lastFinalizedBlock
      ).get()
    assertThat(lastProcessedBlock).isEqualTo(lastFinalizedBlock)
    verify(aggregationsRepository).findConsecutiveProvenBlobs(lastFinalizedBlock.toLong() + 1)
  }
}
