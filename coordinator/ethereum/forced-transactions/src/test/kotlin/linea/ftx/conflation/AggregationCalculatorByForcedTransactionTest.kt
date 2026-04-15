package linea.ftx.conflation

import linea.forcedtx.ForcedTransactionInclusionResult
import linea.forcedtx.ForcedTransactionInclusionStatus
import net.consensys.zkevm.domain.blobCounters
import net.consensys.zkevm.ethereum.coordination.aggregation.AggregationTrigger
import net.consensys.zkevm.ethereum.coordination.aggregation.AggregationTriggerType
import net.consensys.zkevm.domain.BlobsToAggregate
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.kotlin.mock
import java.util.LinkedList
import java.util.Queue
import kotlin.time.Instant

class AggregationCalculatorByForcedTransactionTest {
  @Test
  fun `creates adjacent aggregation boundaries for consecutive processed ftx blocks`() {
    val processedFtxQueue: Queue<ForcedTransactionInclusionStatus> = LinkedList()
    processedFtxQueue.add(
      forcedTransactionInclusionStatus(
        ftxNumber = 3uL,
        blockNumber = 46uL,
        inclusionResult = ForcedTransactionInclusionResult.Included,
        blockTimestamp = Instant.fromEpochSeconds(1_776_183_262),
      ),
    )
    processedFtxQueue.add(
      forcedTransactionInclusionStatus(
        ftxNumber = 4uL,
        blockNumber = 47uL,
        inclusionResult = ForcedTransactionInclusionResult.BadNonce,
        blockTimestamp = Instant.fromEpochSeconds(1_776_183_264),
      ),
    )

    val calculator = AggregationCalculatorByForcedTransaction(processedFtxQueue = processedFtxQueue, log = mock())

    assertThat(calculator.checkAggregationTrigger(blobCounters(startBlockNumber = 45uL, endBlockNumber = 45uL)))
      .isEqualTo(
        AggregationTrigger(
          aggregationTriggerType = AggregationTriggerType.FORCED_TRANSACTION,
          aggregation = BlobsToAggregate(45uL, 45uL),
        ),
      )

    calculator.reset()

    assertThat(calculator.checkAggregationTrigger(blobCounters(startBlockNumber = 46uL, endBlockNumber = 46uL)))
      .isEqualTo(
        AggregationTrigger(
          aggregationTriggerType = AggregationTriggerType.FORCED_TRANSACTION,
          aggregation = BlobsToAggregate(46uL, 46uL),
        ),
      )
  }

  private fun forcedTransactionInclusionStatus(
    ftxNumber: ULong,
    blockNumber: ULong,
    inclusionResult: ForcedTransactionInclusionResult,
    blockTimestamp: Instant,
  ) = ForcedTransactionInclusionStatus(
    ftxNumber = ftxNumber,
    blockNumber = blockNumber,
    blockTimestamp = blockTimestamp,
    inclusionResult = inclusionResult,
    ftxHash = byteArrayOf(0x01),
    from = byteArrayOf(0x02),
  )
}
