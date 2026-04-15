package net.consensys.zkevm.ethereum.coordination.aggregation

import linea.forcedtx.ForcedTransactionInclusionResult
import linea.persistence.ftx.ForcedTransactionsDao
import net.consensys.zkevm.domain.ForcedTransactionRecord
import net.consensys.zkevm.domain.InvalidityProofIndex
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Instant

class InvalidityProofProviderTest {
  @Test
  fun `includes forced transaction simulated on aggregation end block`() {
    val forcedTransactionsDao =
      object : ForcedTransactionsDao {
        private val records =
          listOf(
            ForcedTransactionRecord(
              ftxNumber = 1uL,
              inclusionResult = ForcedTransactionInclusionResult.Included,
              simulatedExecutionBlockNumber = 38uL,
              simulatedExecutionBlockTimestamp = Instant.fromEpochSeconds(123),
              ftxBlockNumberDeadline = 100uL,
              ftxRollingHash = byteArrayOf(0x11),
              ftxRlp = byteArrayOf(0x22),
              proofStatus = ForcedTransactionRecord.ProofStatus.UNREQUESTED,
            ),
            ForcedTransactionRecord(
              ftxNumber = 2uL,
              inclusionResult = ForcedTransactionInclusionResult.Included,
              simulatedExecutionBlockNumber = 39uL,
              simulatedExecutionBlockTimestamp = Instant.fromEpochSeconds(124),
              ftxBlockNumberDeadline = 101uL,
              ftxRollingHash = byteArrayOf(0x33),
              ftxRlp = byteArrayOf(0x44),
              proofStatus = ForcedTransactionRecord.ProofStatus.UNREQUESTED,
            ),
          )

        override fun save(ftx: ForcedTransactionRecord): SafeFuture<Unit> =
          SafeFuture.failedFuture(IllegalStateException("not implemented"))

        override fun findByNumber(ftxNumber: ULong): SafeFuture<ForcedTransactionRecord?> =
          SafeFuture.completedFuture(records.find { it.ftxNumber == ftxNumber })

        override fun list(): SafeFuture<List<ForcedTransactionRecord>> = SafeFuture.completedFuture(records)

        override fun deleteFtxUpToInclusive(ftxNumber: ULong): SafeFuture<Int> = SafeFuture.completedFuture(0)
      }

    val invalidityProofProvider = InvalidityProofProviderImpl(forcedTransactionsDao)

    val invalidityProofs =
      invalidityProofProvider.getInvalidityProofs(
        ftxStartingNumber = 1uL,
        aggregationEndBlockNumber = 38uL,
      ).get()

    assertThat(invalidityProofs).containsExactly(
      InvalidityProofIndex(
        ftxNumber = 1uL,
        simulatedExecutionBlockNumber = 38uL,
        startBlockTimestamp = Instant.fromEpochSeconds(123),
      ),
    )
  }
}
