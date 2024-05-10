package net.consensys.zkevm.ethereum.coordination.aggregation

import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.zkevm.domain.BlobCounters
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockHeaderSummary
import net.consensys.zkevm.ethereum.coordination.blockcreation.SafeBlockProvider
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration.Companion.milliseconds

class AggregationTriggerCalculatorByDeadlineTest {

  @Test
  fun when_past_deadline_latest_block_match_trigger_aggregation() {
    val mockClock = mock<Clock>()
    val mockLatestSafeBlockProvider = mock<SafeBlockProvider>()
    val deadlineTriggered = SafeFuture<Boolean>()
    val aggregationTriggerTypeHandler = AggregationTriggerHandler { it ->
      if (it == AggregationTriggerType.TIME_LIMIT) {
        deadlineTriggered.complete(true)
      } else {
        deadlineTriggered.complete(false)
      }
      SafeFuture.completedFuture(Unit)
    }

    val deadline = 200.milliseconds

    val aggregationTriggerByDeadline = AggregationTriggerCalculatorByDeadline(
      AggregationTriggerCalculatorByDeadline.Config(deadline),
      mockClock,
      mockLatestSafeBlockProvider
    )

    aggregationTriggerByDeadline.onAggregationTrigger(aggregationTriggerTypeHandler)

    whenever(mockClock.now()).thenReturn(Instant.fromEpochMilliseconds(50))
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(deadlineTriggered).isNotCompleted

    val firstBlobStartBlockTimeStamp = Instant.fromEpochMilliseconds(100)
    val firstBlobEndBlockTimeStamp = firstBlobStartBlockTimeStamp.plus(50.milliseconds)

    aggregationTriggerByDeadline.newBlob(
      BlobCounters(
        numberOfBatches = 3u,
        startBlockNumber = 11u,
        endBlockNumber = 15u,
        startBlockTimestamp = firstBlobStartBlockTimeStamp,
        endBlockTimestamp = firstBlobStartBlockTimeStamp.plus(50.milliseconds)
      )
    )

    whenever(mockClock.now()).thenReturn(
      firstBlobEndBlockTimeStamp
        .plus(1.milliseconds)
    )
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(deadlineTriggered).isNotCompleted

    whenever(mockLatestSafeBlockProvider.getLatestSafeBlockHeader())
      .thenReturn(
        SafeFuture.completedFuture(
          BlockHeaderSummary(
            number = 15u,
            hash = Bytes32.random(),
            timestamp = firstBlobEndBlockTimeStamp
          )
        )
      )
    whenever(mockClock.now())
      .thenReturn(firstBlobStartBlockTimeStamp.plus(deadline).plus(1.milliseconds))
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(deadlineTriggered).isCompleted
    assertThat(deadlineTriggered.get()).isTrue()
  }

  @Test
  fun when_reset_verify_deadline_state_is_reset() {
    val mockClock = mock<Clock>()
    val mockLatestSafeBlockProvider = mock<SafeBlockProvider>()

    val deadlineTriggered = SafeFuture<Boolean>()
    val aggregationTriggerTypeHandler = AggregationTriggerHandler { it ->
      if (it == AggregationTriggerType.TIME_LIMIT) {
        deadlineTriggered.complete(true)
      } else {
        deadlineTriggered.complete(false)
      }
      SafeFuture.completedFuture(Unit)
    }

    val deadline = 200.milliseconds

    val aggregationTriggerByDeadline = AggregationTriggerCalculatorByDeadline(
      AggregationTriggerCalculatorByDeadline.Config(deadline),
      mockClock,
      mockLatestSafeBlockProvider
    )
    aggregationTriggerByDeadline.onAggregationTrigger(aggregationTriggerTypeHandler)

    whenever(mockClock.now()).thenReturn(Instant.fromEpochMilliseconds(50))
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(deadlineTriggered).isNotCompleted

    val firstBloblStartBlockTimeStamp = Instant.fromEpochMilliseconds(100)
    val firstBlobEndBlockTimeStamp = firstBloblStartBlockTimeStamp.plus(50.milliseconds)

    aggregationTriggerByDeadline.newBlob(
      BlobCounters(
        numberOfBatches = 3u,
        startBlockNumber = 11u,
        endBlockNumber = 15u,
        startBlockTimestamp = firstBloblStartBlockTimeStamp,
        endBlockTimestamp = firstBlobEndBlockTimeStamp
      )
    )

    whenever(mockClock.now()).thenReturn(
      firstBlobEndBlockTimeStamp
        .plus(1.milliseconds)
    )
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(deadlineTriggered).isNotCompleted

    aggregationTriggerByDeadline.reset()

    whenever(mockClock.now())
      .thenReturn(firstBloblStartBlockTimeStamp.plus(deadline).plus(1.milliseconds))
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(deadlineTriggered).isNotCompleted

    val secondBlobStartBlockTimestamp = Instant.fromEpochMilliseconds(500)
    val secondBlobEndBlockTimestamp = secondBlobStartBlockTimestamp.plus(50.milliseconds)

    aggregationTriggerByDeadline.newBlob(
      BlobCounters(
        numberOfBatches = 3u,
        startBlockNumber = 11u,
        endBlockNumber = 15u,
        startBlockTimestamp = secondBlobStartBlockTimestamp,
        endBlockTimestamp = secondBlobEndBlockTimestamp
      )
    )
    whenever(mockLatestSafeBlockProvider.getLatestSafeBlockHeader())
      .thenReturn(
        SafeFuture.completedFuture(
          BlockHeaderSummary(
            number = 15u,
            hash = Bytes32.random(),
            timestamp = firstBlobEndBlockTimeStamp
          )
        )
      )
    whenever(mockClock.now()).thenReturn(secondBlobStartBlockTimestamp.plus(deadline).plus(1.milliseconds))
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(deadlineTriggered).isCompleted
    assertThat(deadlineTriggered.get()).isTrue()
  }

  @Test
  fun when_latest_block_does_not_match_aggregation_not_triggered() {
    val mockClock = mock<Clock>()
    val mockLatestSafeBlockProvider = mock<SafeBlockProvider>()
    val deadlineTriggered = SafeFuture<Boolean>()
    val aggregationTriggerTypeHandler = AggregationTriggerHandler { it ->
      if (it == AggregationTriggerType.TIME_LIMIT) {
        deadlineTriggered.complete(true)
      } else {
        deadlineTriggered.complete(false)
      }
      SafeFuture.completedFuture(Unit)
    }

    val deadline = 200.milliseconds

    val aggregationTriggerByDeadline = AggregationTriggerCalculatorByDeadline(
      AggregationTriggerCalculatorByDeadline.Config(deadline),
      mockClock,
      mockLatestSafeBlockProvider
    )

    aggregationTriggerByDeadline.onAggregationTrigger(aggregationTriggerTypeHandler)

    whenever(mockClock.now()).thenReturn(Instant.fromEpochMilliseconds(50))
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(deadlineTriggered).isNotCompleted

    val firstBlobStartBlockTimeStamp = Instant.fromEpochMilliseconds(100)
    val firstBlobEndBlockTimeStamp = firstBlobStartBlockTimeStamp.plus(50.milliseconds)

    aggregationTriggerByDeadline.newBlob(
      BlobCounters(
        numberOfBatches = 3u,
        startBlockNumber = 11u,
        endBlockNumber = 15u,
        startBlockTimestamp = firstBlobStartBlockTimeStamp,
        endBlockTimestamp = firstBlobStartBlockTimeStamp.plus(50.milliseconds)
      )
    )

    whenever(mockClock.now()).thenReturn(
      firstBlobEndBlockTimeStamp
        .plus(1.milliseconds)
    )
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(deadlineTriggered).isNotCompleted

    whenever(mockLatestSafeBlockProvider.getLatestSafeBlockHeader())
      .thenReturn(
        SafeFuture.completedFuture(
          BlockHeaderSummary(
            number = 16u,
            hash = Bytes32.random(),
            timestamp = firstBlobEndBlockTimeStamp
          )
        )
      )
    whenever(mockClock.now())
      .thenReturn(firstBlobStartBlockTimeStamp.plus(deadline).plus(1.milliseconds))
    aggregationTriggerByDeadline.checkAggregation().get()
    assertThat(deadlineTriggered).isNotCompleted
  }
}
