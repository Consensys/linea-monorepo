package net.consensys.zkevm.ethereum.coordination.aggregation

import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.metrics.micrometer.MicrometerMetricsFacade
import net.consensys.zkevm.domain.BlobCounters
import net.consensys.zkevm.domain.BlobsToAggregate
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockHeaderSummary
import net.consensys.zkevm.ethereum.coordination.blockcreation.SafeBlockProvider
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.Mockito
import org.mockito.kotlin.mock
import org.mockito.kotlin.reset
import org.mockito.kotlin.spy
import org.mockito.kotlin.times
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds

class GlobalAggregationCalculatorTest {
  @Test
  fun when_out_of_order_blob_then_throw_exception() {
    val aggregationTriggerByProofLimit = AggregationTriggerCalculatorByProofLimit(maxProofsPerAggregation = 15u)
    val globalAggregationCalculator = GlobalAggregationCalculator(
      0u,
      syncAggregationTrigger = listOf(aggregationTriggerByProofLimit),
      deferredAggregationTrigger = listOf(),
      metricsFacade = mock<MetricsFacade>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
    )
    val expectedErrorMessage = "Blobs to aggregate must be sequential: lastBlockNumber=0, startBlockNumber=2 for " +
      "new blob"
    Assertions.assertThatThrownBy {
      globalAggregationCalculator.newBlob(
        BlobCounters(
          numberOfBatches = 5u,
          startBlockNumber = 2u,
          endBlockNumber = 10u,
          startBlockTimestamp = Instant.fromEpochMilliseconds(100),
          endBlockTimestamp = Instant.fromEpochMilliseconds(500)
        )
      ).get()
    }
      .message()
      .isEqualTo(expectedErrorMessage)
  }

  @Test
  fun when_new_blobs_exceed_proof_limit_verify_trigger_type_and_aggregation() {
    val aggregationTriggerByProofLimit = AggregationTriggerCalculatorByProofLimit(maxProofsPerAggregation = 15u)
    val mockClock = mock<Clock>()
    whenever(mockClock.now()).thenReturn(Instant.fromEpochMilliseconds(99.seconds.inWholeMilliseconds))
    val mockLatestSafeBlockProvider = mock<SafeBlockProvider>()
    whenever(mockLatestSafeBlockProvider.getLatestSafeBlockHeader())
      .thenReturn(
        SafeFuture.completedFuture(
          BlockHeaderSummary(
            number = 200u,
            hash = Bytes32.random(),
            timestamp = Instant.fromEpochMilliseconds(2000.seconds.inWholeMilliseconds)
          )
        )
      )

    val aggregationTriggerByDeadline = AggregationTriggerCalculatorByDeadline(
      AggregationTriggerCalculatorByDeadline.Config(100.seconds),
      mockClock,
      mockLatestSafeBlockProvider
    )
    var aggregation: BlobsToAggregate? = null
    val aggregationHandler = AggregationHandler { blobsToAggregate ->
      aggregation = blobsToAggregate
      SafeFuture.completedFuture(Unit)
    }
    val globalAggregationCalculator = spy(
      GlobalAggregationCalculator(
        0u,
        syncAggregationTrigger = listOf(aggregationTriggerByProofLimit),
        deferredAggregationTrigger = listOf(aggregationTriggerByDeadline),
        metricsFacade = mock<MetricsFacade>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
      )
    )
    globalAggregationCalculator.onAggregation(aggregationHandler)
    aggregationTriggerByDeadline.onAggregationTrigger(globalAggregationCalculator::handleAggregationTrigger)

    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 5u,
        startBlockNumber = 1u,
        endBlockNumber = 10u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(100),
        endBlockTimestamp = Instant.fromEpochMilliseconds(500)
      )
    ).get()
    aggregationTriggerByDeadline.checkAggregation().get()

    assertThat(aggregation).isNull()

    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 12u,
        startBlockNumber = 11u,
        endBlockNumber = 30u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(600),
        endBlockTimestamp = Instant.fromEpochMilliseconds(1500)
      )
    ).get()
    aggregationTriggerByDeadline.checkAggregation().get()

    assertThat(aggregation).isEqualTo(BlobsToAggregate(1u, 10u))
    verify(globalAggregationCalculator, times(1)).handleAggregationTrigger(AggregationTriggerType.PROOF_LIMIT)
    reset(globalAggregationCalculator)

    aggregation = null

    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 9u,
        startBlockNumber = 31u,
        endBlockNumber = 45u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(1600),
        endBlockTimestamp = Instant.fromEpochMilliseconds(2500)
      )
    ).get()
    aggregationTriggerByDeadline.checkAggregation().get()

    assertThat(aggregation).isEqualTo(BlobsToAggregate(11u, 30u))
    verify(globalAggregationCalculator, times(1)).handleAggregationTrigger(AggregationTriggerType.PROOF_LIMIT)
    reset(globalAggregationCalculator)

    aggregation = null

    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 4u,
        startBlockNumber = 46u,
        endBlockNumber = 61u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(2600),
        endBlockTimestamp = Instant.fromEpochMilliseconds(4500)
      )
    ).get()
    aggregationTriggerByDeadline.checkAggregation().get()

    assertThat(aggregation).isEqualTo(BlobsToAggregate(31u, 61u))
    verify(globalAggregationCalculator, times(1)).handleAggregationTrigger(AggregationTriggerType.PROOF_LIMIT)
    verify(mockLatestSafeBlockProvider, times(0)).getLatestSafeBlockHeader()
    reset(globalAggregationCalculator)
  }

  @Test
  fun when_new_blobs_exceed_time_deadline_verify_trigger_type_and_aggregation() {
    val aggregationTriggerByProofLimit = AggregationTriggerCalculatorByProofLimit(maxProofsPerAggregation = 1500u)
    val mockClock = mock<Clock>()
    val mockLatestSafeBlockProvider = mock<SafeBlockProvider>()
    val aggregationTriggerByDeadline = AggregationTriggerCalculatorByDeadline(
      AggregationTriggerCalculatorByDeadline.Config(100.milliseconds),
      mockClock,
      mockLatestSafeBlockProvider
    )

    val globalAggregationCalculator = spy(
      GlobalAggregationCalculator(
        0u,
        syncAggregationTrigger = listOf(aggregationTriggerByProofLimit),
        deferredAggregationTrigger = listOf(aggregationTriggerByDeadline),
        metricsFacade = mock<MetricsFacade>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
      )
    )
    var aggregation: BlobsToAggregate? = null
    val aggregationHandler = AggregationHandler { blobsToAggregate ->
      aggregation = blobsToAggregate
      SafeFuture.completedFuture(Unit)
    }
    globalAggregationCalculator.onAggregation(aggregationHandler)
    aggregationTriggerByDeadline.onAggregationTrigger(globalAggregationCalculator::handleAggregationTrigger)

    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 5u,
        startBlockNumber = 1u,
        endBlockNumber = 10u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(100),
        endBlockTimestamp = Instant.fromEpochMilliseconds(130)
      )
    ).get()
    whenever(mockClock.now()).thenReturn(Instant.fromEpochMilliseconds(135))
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(aggregation).isNull()

    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 12u,
        startBlockNumber = 11u,
        endBlockNumber = 30u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(140),
        endBlockTimestamp = Instant.fromEpochMilliseconds(250)
      )
    ).get()

    whenever(mockLatestSafeBlockProvider.getLatestSafeBlockHeader())
      .thenReturn(
        SafeFuture.completedFuture(
          BlockHeaderSummary(
            number = 30u,
            hash = Bytes32.random(),
            timestamp = Instant.fromEpochMilliseconds(250)
          )
        )
      )

    whenever(mockClock.now()).thenReturn(Instant.fromEpochMilliseconds(300))
    aggregationTriggerByDeadline.checkAggregation().get()

    assertThat(aggregation).isEqualTo(BlobsToAggregate(1u, 30u))
    verify(globalAggregationCalculator, times(1)).handleAggregationTrigger(AggregationTriggerType.TIME_LIMIT)
    verify(mockLatestSafeBlockProvider, times(1)).getLatestSafeBlockHeader()
    reset(globalAggregationCalculator)
    reset(mockLatestSafeBlockProvider)

    aggregation = null

    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 9u,
        startBlockNumber = 31u,
        endBlockNumber = 45u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(1600),
        endBlockTimestamp = Instant.fromEpochMilliseconds(1675)
      )
    ).get()
    whenever(mockClock.now()).thenReturn(Instant.fromEpochMilliseconds(1700))
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(aggregation).isNull()

    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 6u,
        startBlockNumber = 46u,
        endBlockNumber = 61u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(1681),
        endBlockTimestamp = Instant.fromEpochMilliseconds(1725)
      )
    ).get()
    whenever(mockClock.now()).thenReturn(Instant.fromEpochMilliseconds(1755))
    whenever(mockLatestSafeBlockProvider.getLatestSafeBlockHeader())
      .thenReturn(
        SafeFuture.completedFuture(
          BlockHeaderSummary(
            number = 61u,
            hash = Bytes32.random(),
            timestamp = Instant.fromEpochMilliseconds(1725)
          )
        )
      )

    aggregationTriggerByDeadline.checkAggregation().get()

    assertThat(aggregation).isEqualTo(BlobsToAggregate(31u, 61u))
    verify(globalAggregationCalculator, times(1)).handleAggregationTrigger(AggregationTriggerType.TIME_LIMIT)
    verify(mockLatestSafeBlockProvider, times(1)).getLatestSafeBlockHeader()
    reset(globalAggregationCalculator)
  }

  @Test
  fun `metrics are exported correctly when aggregation is triggered by proof limit`() {
    val aggregationTriggerByProofLimit = AggregationTriggerCalculatorByProofLimit(maxProofsPerAggregation = 15u)

    val testMeterRegistry = SimpleMeterRegistry()
    val globalAggregationCalculator = GlobalAggregationCalculator(
      0u,
      syncAggregationTrigger = listOf(aggregationTriggerByProofLimit),
      deferredAggregationTrigger = emptyList(),
      metricsFacade = MicrometerMetricsFacade(testMeterRegistry, "test")
    )
    val pendingProofsGauge = testMeterRegistry.get("test.aggregation.proofs.ready").gauge()
    assertThat(pendingProofsGauge.value()).isEqualTo(0.0)
    globalAggregationCalculator.onAggregation {
      SafeFuture.completedFuture(Unit)
    }

    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 5u,
        startBlockNumber = 1u,
        endBlockNumber = 10u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(100),
        endBlockTimestamp = Instant.fromEpochMilliseconds(500)
      )
    ).get()
    assertThat(pendingProofsGauge.value()).isEqualTo(6.0)

    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 7u,
        startBlockNumber = 11u,
        endBlockNumber = 30u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(600),
        endBlockTimestamp = Instant.fromEpochMilliseconds(1500)
      )
    ).get()
    assertThat(pendingProofsGauge.value()).isEqualTo(14.0)

    // Next blob should cause aggregation
    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 9u,
        startBlockNumber = 31u,
        endBlockNumber = 45u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(1600),
        endBlockTimestamp = Instant.fromEpochMilliseconds(2500)
      )
    ).get()
    assertThat(pendingProofsGauge.value()).isEqualTo(10.0)

    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 4u,
        startBlockNumber = 46u,
        endBlockNumber = 61u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(2600),
        endBlockTimestamp = Instant.fromEpochMilliseconds(4500)
      )
    ).get()

    assertThat(pendingProofsGauge.value()).isEqualTo(0.0)
  }

  @Test
  fun `metrics are exported correctly when aggregation is triggered by deadline`() {
    val mockClock = mock<Clock>()
    val mockLatestSafeBlockProvider = mock<SafeBlockProvider>()
    val aggregationTriggerByDeadline = AggregationTriggerCalculatorByDeadline(
      AggregationTriggerCalculatorByDeadline.Config(100.milliseconds),
      mockClock,
      mockLatestSafeBlockProvider
    )

    val testMeterRegistry = SimpleMeterRegistry()
    val globalAggregationCalculator = spy(
      GlobalAggregationCalculator(
        0u,
        syncAggregationTrigger = emptyList(),
        deferredAggregationTrigger = listOf(aggregationTriggerByDeadline),
        metricsFacade = MicrometerMetricsFacade(testMeterRegistry, "test")
      )
    )
    globalAggregationCalculator.onAggregation { SafeFuture.completedFuture(Unit) }
    aggregationTriggerByDeadline.onAggregationTrigger(globalAggregationCalculator::handleAggregationTrigger)

    val pendingProofsGauge = testMeterRegistry.get("test.aggregation.proofs.ready").gauge()
    assertThat(pendingProofsGauge.value()).isEqualTo(0.0)
    globalAggregationCalculator.onAggregation {
      SafeFuture.completedFuture(Unit)
    }

    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 5u,
        startBlockNumber = 1u,
        endBlockNumber = 10u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(100),
        endBlockTimestamp = Instant.fromEpochMilliseconds(130)
      )
    ).get()
    globalAggregationCalculator.newBlob(
      BlobCounters(
        numberOfBatches = 12u,
        startBlockNumber = 11u,
        endBlockNumber = 30u,
        startBlockTimestamp = Instant.fromEpochMilliseconds(140),
        endBlockTimestamp = Instant.fromEpochMilliseconds(250)
      )
    ).get()

    whenever(mockLatestSafeBlockProvider.getLatestSafeBlockHeader())
      .thenReturn(
        SafeFuture.completedFuture(
          BlockHeaderSummary(
            number = 30u,
            hash = Bytes32.random(),
            timestamp = Instant.fromEpochMilliseconds(250)
          )
        )
      )
    whenever(mockClock.now()).thenReturn(Instant.fromEpochMilliseconds(1755))

    assertThat(pendingProofsGauge.value()).isEqualTo(19.0)
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(pendingProofsGauge.value()).isEqualTo(0.0)
  }
}
