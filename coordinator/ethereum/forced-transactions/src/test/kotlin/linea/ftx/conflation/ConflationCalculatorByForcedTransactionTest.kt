package linea.ftx.conflation

import linea.forcedtx.ForcedTransactionInclusionResult
import linea.forcedtx.ForcedTransactionInclusionStatus
import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.domain.ConflationTrigger
import net.consensys.linea.traces.TracesCountersV4
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.kotlin.mock
import java.util.LinkedList
import java.util.Queue
import kotlin.time.Instant

class ConflationCalculatorByForcedTransactionTest {
  @Test
  fun `creates adjacent conflation boundaries for consecutive processed ftx blocks`() {
    val processedFtxQueue: Queue<ForcedTransactionInclusionStatus> = LinkedList()
    processedFtxQueue.add(
      forcedTransactionInclusionStatus(
        ftxNumber = 1uL,
        blockNumber = 27uL,
        inclusionResult = ForcedTransactionInclusionResult.BadNonce,
        blockTimestamp = Instant.fromEpochSeconds(1_776_184_667),
      ),
    )
    processedFtxQueue.add(
      forcedTransactionInclusionStatus(
        ftxNumber = 2uL,
        blockNumber = 28uL,
        inclusionResult = ForcedTransactionInclusionResult.Included,
        blockTimestamp = Instant.fromEpochSeconds(1_776_184_668),
      ),
    )

    val calculator = ConflationCalculatorByForcedTransaction(processedFtxQueue = processedFtxQueue, log = mock())

    assertThat(calculator.checkOverflow(blockCounters(27uL))?.trigger)
      .isEqualTo(ConflationTrigger.FORCED_TRANSACTION)

    calculator.reset()

    assertThat(calculator.checkOverflow(blockCounters(28uL))?.trigger)
      .isEqualTo(ConflationTrigger.FORCED_TRANSACTION)
  }

  private fun blockCounters(blockNumber: ULong) = BlockCounters(
    blockNumber = blockNumber,
    blockTimestamp = Instant.fromEpochSeconds(blockNumber.toLong()),
    tracesCounters = TracesCountersV4.EMPTY_TRACES_COUNT,
    blockRLPEncoded = ByteArray(0),
  )

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
