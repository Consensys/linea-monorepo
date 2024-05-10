package net.consensys.zkevm.ethereum.coordination.aggregation

import net.consensys.zkevm.domain.BlobCounters
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import java.lang.IllegalStateException
import java.util.concurrent.ExecutionException

class AggregationTriggerCalculatorByProofLimitTest {
  @Test
  fun when_proof_limit_overflows_with_exact_proof_limit_include_current_blob_in_aggregation() {
    val mockBlobCounters = mock<BlobCounters>()
    whenever(mockBlobCounters.numberOfBatches).thenReturn(1u)
    val aggregationTrigger = AggregationTriggerCalculatorByProofLimit(maxProofsPerAggregation = 10u)
    for (i in 1..4) {
      assertThat(aggregationTrigger.checkAggregationTrigger(mockBlobCounters)).isNull()
      aggregationTrigger.newBlob(mockBlobCounters).get()
    }
    val overflowTrigger = aggregationTrigger.checkAggregationTrigger(mockBlobCounters)
    assertThat(overflowTrigger)
      .isEqualTo(
        AggregationTrigger(
          aggregationTriggerType = AggregationTriggerType.PROOF_LIMIT,
          includeCurrentBlob = true
        )
      )
  }

  @Test
  fun when_proof_limit_overflows_above_proof_limit_do_not_include_current_blob_in_aggregation() {
    val mockBlobCounters = mock<BlobCounters>()
    whenever(mockBlobCounters.numberOfBatches).thenReturn(2u)
    val aggregationTrigger = AggregationTriggerCalculatorByProofLimit(maxProofsPerAggregation = 8u)

    for (i in 1..2) {
      assertThat(aggregationTrigger.checkAggregationTrigger(mockBlobCounters)).isNull()
      aggregationTrigger.newBlob(mockBlobCounters).get()
    }
    val overflowTrigger = aggregationTrigger.checkAggregationTrigger(mockBlobCounters)
    assertThat(overflowTrigger)
      .isEqualTo(
        AggregationTrigger(
          aggregationTriggerType = AggregationTriggerType.PROOF_LIMIT,
          includeCurrentBlob = false
        )
      )
  }

  @Test
  fun when_reset_then_calculator_should_rest_proof_count() {
    val mockBlobCounters = mock<BlobCounters>()
    whenever(mockBlobCounters.numberOfBatches).thenReturn(1u)
    val aggregationTrigger = AggregationTriggerCalculatorByProofLimit(maxProofsPerAggregation = 10u)
    for (i in 1..4) {
      assertThat(aggregationTrigger.checkAggregationTrigger(mockBlobCounters)).isNull()
      aggregationTrigger.newBlob(mockBlobCounters).get()
    }

    aggregationTrigger.reset()

    for (i in 1..4) {
      assertThat(aggregationTrigger.checkAggregationTrigger(mockBlobCounters)).isNull()
      aggregationTrigger.newBlob(mockBlobCounters).get()
    }
    assertThat(aggregationTrigger.checkAggregationTrigger(mockBlobCounters))
      .isEqualTo(
        AggregationTrigger(
          aggregationTriggerType = AggregationTriggerType.PROOF_LIMIT,
          includeCurrentBlob = true
        )
      )
  }

  @Test
  fun when_after_trigger_new_blob_is_called_without_reset_then_throw_exception() {
    val mockBlobCounters = mock<BlobCounters>()
    whenever(mockBlobCounters.numberOfBatches).thenReturn(1u)
    val aggregationTrigger = AggregationTriggerCalculatorByProofLimit(maxProofsPerAggregation = 10u)
    for (i in 1..4) {
      assertThat(aggregationTrigger.checkAggregationTrigger(mockBlobCounters)).isNull()
      aggregationTrigger.newBlob(mockBlobCounters).get()
    }

    val exception = assertThrows<ExecutionException> { aggregationTrigger.newBlob(mockBlobCounters).get() }
    assertThat(exception.cause).isExactlyInstanceOf(IllegalStateException::class.java)
    assertThat(exception.cause!!.message)
      .isEqualTo("Proof count already overflowed, should have been reset before this new blob")
  }
}
