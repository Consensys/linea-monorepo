package net.consensys.zkevm.ethereum.coordination.aggregation

import linea.forcedtx.ForcedTransactionInclusionResult
import linea.persistence.ftx.FakeForcedTransactionsDao
import linea.persistence.ftx.ForcedTransactionRecordFactory.createForcedTransactionRecord
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import kotlin.time.Instant

class InvalidityProofProviderImplTest {
  private lateinit var dao: FakeForcedTransactionsDao
  private lateinit var provider: InvalidityProofProviderImpl

  @BeforeEach
  fun setUp() {
    dao = FakeForcedTransactionsDao()
    provider = InvalidityProofProviderImpl(dao)
  }

  private fun saveFtx(
    ftxNumber: ULong,
    simulatedExecutionBlockNumber: ULong,
    inclusionResult: ForcedTransactionInclusionResult = ForcedTransactionInclusionResult.BadNonce,
  ) {
    dao.save(
      createForcedTransactionRecord(
        ftxNumber = ftxNumber,
        simulatedExecutionBlockNumber = simulatedExecutionBlockNumber,
        simulatedExecutionBlockTimestamp = Instant.fromEpochSeconds(simulatedExecutionBlockNumber.toLong()),
        inclusionResult = inclusionResult,
      ),
    ).get()
  }

  @Test
  fun `returns FTX whose simulated execution block equals the aggregation end block`() {
    // Aggregation [20..21] containing an FTX at block 21 (its end block).
    saveFtx(ftxNumber = 4UL, simulatedExecutionBlockNumber = 21UL)

    val proofs = provider.getInvalidityProofs(
      ftxStartingNumber = 4UL,
      aggregationEndBlockNumber = 21UL,
    ).get()

    assertThat(proofs).hasSize(1)
    assertThat(proofs[0].ftxNumber).isEqualTo(4UL)
    assertThat(proofs[0].simulatedExecutionBlockNumber).isEqualTo(21UL)
  }

  @Test
  fun `returns every FTX whose simulated execution block falls inside the aggregation`() {
    // Aggregation [20..23] covering FTXs #3 (block 20), #4 (block 21), #5 (block 23).
    saveFtx(ftxNumber = 3UL, simulatedExecutionBlockNumber = 20UL)
    saveFtx(ftxNumber = 4UL, simulatedExecutionBlockNumber = 21UL)
    saveFtx(ftxNumber = 5UL, simulatedExecutionBlockNumber = 23UL)

    val proofs = provider.getInvalidityProofs(
      ftxStartingNumber = 3UL,
      aggregationEndBlockNumber = 23UL,
    ).get()

    assertThat(proofs.map { it.ftxNumber }).containsExactlyInAnyOrder(3UL, 4UL, 5UL)
  }

  @Test
  fun `excludes FTXs whose simulated execution block is past the aggregation end block`() {
    // Aggregation [20..21] — FTX#5 at block 23 belongs to the next aggregation.
    saveFtx(ftxNumber = 4UL, simulatedExecutionBlockNumber = 21UL)
    saveFtx(ftxNumber = 5UL, simulatedExecutionBlockNumber = 23UL)

    val proofs = provider.getInvalidityProofs(
      ftxStartingNumber = 4UL,
      aggregationEndBlockNumber = 21UL,
    ).get()

    assertThat(proofs.map { it.ftxNumber }).containsExactly(4UL)
  }

  @Test
  fun `excludes FTXs already accounted for by the parent aggregation`() {
    // Parent aggregation finalised FTX#3 (block 20). The current aggregation [21..21]
    // should only ask about FTX#4.
    saveFtx(ftxNumber = 3UL, simulatedExecutionBlockNumber = 20UL)
    saveFtx(ftxNumber = 4UL, simulatedExecutionBlockNumber = 21UL)

    val proofs = provider.getInvalidityProofs(
      ftxStartingNumber = 4UL,
      aggregationEndBlockNumber = 21UL,
    ).get()

    assertThat(proofs.map { it.ftxNumber }).containsExactly(4UL)
  }

  @Test
  fun `returns empty when no FTX falls inside the aggregation`() {
    saveFtx(ftxNumber = 5UL, simulatedExecutionBlockNumber = 23UL)

    val proofs = provider.getInvalidityProofs(
      ftxStartingNumber = 4UL,
      aggregationEndBlockNumber = 22UL,
    ).get()

    assertThat(proofs).isEmpty()
  }
}
