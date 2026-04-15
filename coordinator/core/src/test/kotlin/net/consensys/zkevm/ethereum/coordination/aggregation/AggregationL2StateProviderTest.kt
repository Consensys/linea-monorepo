package net.consensys.zkevm.ethereum.coordination.aggregation

import linea.contract.l2.FakeL2MessageService
import linea.ethapi.FakeEthApiClient
import linea.forcedtx.ForcedTransactionInclusionResult
import linea.persistence.ftx.ForcedTransactionsDao
import net.consensys.zkevm.domain.BlobAndBatchCounters
import net.consensys.zkevm.domain.ForcedTransactionRecord
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.createProofToFinalize
import net.consensys.zkevm.persistence.AggregationsRepository
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Instant

class AggregationL2StateProviderTest {
  @Test
  fun `uses previous aggregation finalized ftx state instead of latest processed ftx`() {
    val rollingHash = byteArrayOf(0x11)
    val finalizedFtxRollingHash = byteArrayOf(0x22)
    val unfinalizedFtxRollingHash = byteArrayOf(0x33)
    val fakeEthApiClient = FakeEthApiClient().apply {
      setLatestBlockTag(42uL)
    }
    val fakeMessageService = FakeL2MessageService(contractDeployBlock = 0uL).apply {
      setLastAnchoredL1Message(7uL, rollingHash)
    }
    val previousAggregation =
      createProofToFinalize(
        firstBlockNumber = 41,
        finalBlockNumber = 42,
        parentAggregationFtxNumber = 0uL,
        parentAggregationFtxRollingHash = ByteArray(32),
        finalFtxNumber = 1uL,
        finalFtxRollingHash = finalizedFtxRollingHash,
      )
    val aggregationsRepository =
      object : AggregationsRepository {
        override fun findConsecutiveProvenBlobs(fromBlockNumber: Long) =
          SafeFuture.completedFuture(emptyList<BlobAndBatchCounters>())

        override fun saveNewAggregation(aggregation: net.consensys.zkevm.domain.Aggregation) =
          SafeFuture.failedFuture<Unit>(IllegalStateException("not implemented"))

        override fun getProofsToFinalize(
          fromBlockNumber: Long,
          finalEndBlockCreatedBefore: Instant,
          maximumNumberOfProofs: Int,
        ) = SafeFuture.completedFuture(emptyList<ProofToFinalize>())

        override fun findHighestConsecutiveEndBlockNumber(fromBlockNumber: Long?) =
          SafeFuture.completedFuture<Long?>(null)

        override fun findAggregationProofByEndBlockNumber(endBlockNumber: Long) =
          SafeFuture.completedFuture<ProofToFinalize?>(if (endBlockNumber == 42L) previousAggregation else null)

        override fun findLatestProvenAggregationProofUpToEndBlockNumber(endBlockNumberInclusive: Long) =
          SafeFuture.completedFuture<ProofToFinalize?>(if (endBlockNumberInclusive >= 42L) previousAggregation else null)

        override fun deleteAggregationsUpToEndBlockNumber(endBlockNumberInclusive: Long) =
          SafeFuture.completedFuture(0)

        override fun deleteAggregationsAfterBlockNumber(startingBlockNumberInclusive: Long) =
          SafeFuture.completedFuture(0)
      }
    val forcedTransactionsDao =
      object : ForcedTransactionsDao {
        private val records =
          listOf(
            ForcedTransactionRecord(
              ftxNumber = 1uL,
              inclusionResult = ForcedTransactionInclusionResult.Included,
              simulatedExecutionBlockNumber = 41uL,
              simulatedExecutionBlockTimestamp = Instant.fromEpochSeconds(41),
              ftxBlockNumberDeadline = 50uL,
              ftxRollingHash = byteArrayOf(0x01),
              ftxRlp = byteArrayOf(0x01),
              proofStatus = ForcedTransactionRecord.ProofStatus.UNREQUESTED,
            ),
            ForcedTransactionRecord(
              ftxNumber = 2uL,
              inclusionResult = ForcedTransactionInclusionResult.Included,
              simulatedExecutionBlockNumber = 42uL,
              simulatedExecutionBlockTimestamp = Instant.fromEpochSeconds(42),
              ftxBlockNumberDeadline = 51uL,
              ftxRollingHash = unfinalizedFtxRollingHash,
              ftxRlp = byteArrayOf(0x02),
              proofStatus = ForcedTransactionRecord.ProofStatus.UNREQUESTED,
            ),
          )

        override fun save(ftx: ForcedTransactionRecord): SafeFuture<Unit> =
          SafeFuture.failedFuture(IllegalStateException("not implemented"))

        override fun findByNumber(ftxNumber: ULong): SafeFuture<ForcedTransactionRecord?> {
          return SafeFuture.completedFuture(records.find { it.ftxNumber == ftxNumber })
        }

        override fun list(): SafeFuture<List<ForcedTransactionRecord>> = SafeFuture.completedFuture(records)

        override fun deleteFtxUpToInclusive(ftxNumber: ULong): SafeFuture<Int> = SafeFuture.completedFuture(0)
      }

    val aggregationL2StateProvider =
      AggregationL2StateProviderImpl(
        ethApiClient = fakeEthApiClient,
        messageService = fakeMessageService,
        aggregationsRepository = aggregationsRepository,
        forcedTransactionsDao = forcedTransactionsDao,
      )

    val aggregationL2State = aggregationL2StateProvider.getAggregationL2State(42).get()

    assertThat(aggregationL2State.parentAggregationLastL1RollingHashMessageNumber).isEqualTo(7uL)
    assertThat(aggregationL2State.parentAggregationLastL1RollingHash).isEqualTo(rollingHash)
    assertThat(aggregationL2State.parentAggregationLastFtxNumber).isEqualTo(1uL)
    assertThat(aggregationL2State.parentAggregationLastFtxRollingHash).isEqualTo(finalizedFtxRollingHash)
  }

  @Test
  fun `prefers higher dao ftx when latest proven aggregation is older than parent block`() {
    val rollingHash = byteArrayOf(0x11)
    val latestProvenFtxRollingHash = byteArrayOf(0x22)
    val processedFtxRollingHash = byteArrayOf(0x33)
    val fakeEthApiClient = FakeEthApiClient().apply {
      setLatestBlockTag(36uL)
    }
    val fakeMessageService = FakeL2MessageService(contractDeployBlock = 0uL).apply {
      setLastAnchoredL1Message(7uL, rollingHash)
    }
    val latestProvenAggregation =
      createProofToFinalize(
        firstBlockNumber = 33,
        finalBlockNumber = 34,
        parentAggregationFtxNumber = 0uL,
        parentAggregationFtxRollingHash = ByteArray(32),
        finalFtxNumber = 1uL,
        finalFtxRollingHash = latestProvenFtxRollingHash,
      )
    val aggregationsRepository =
      object : AggregationsRepository {
        override fun findConsecutiveProvenBlobs(fromBlockNumber: Long) =
          SafeFuture.completedFuture(emptyList<BlobAndBatchCounters>())

        override fun saveNewAggregation(aggregation: net.consensys.zkevm.domain.Aggregation) =
          SafeFuture.failedFuture<Unit>(IllegalStateException("not implemented"))

        override fun getProofsToFinalize(
          fromBlockNumber: Long,
          finalEndBlockCreatedBefore: Instant,
          maximumNumberOfProofs: Int,
        ) = SafeFuture.completedFuture(emptyList<ProofToFinalize>())

        override fun findHighestConsecutiveEndBlockNumber(fromBlockNumber: Long?) =
          SafeFuture.completedFuture<Long?>(null)

        override fun findAggregationProofByEndBlockNumber(endBlockNumber: Long) =
          SafeFuture.completedFuture<ProofToFinalize?>(null)

        override fun findLatestProvenAggregationProofUpToEndBlockNumber(endBlockNumberInclusive: Long) =
          SafeFuture.completedFuture<ProofToFinalize?>(
            if (endBlockNumberInclusive >= 34L) latestProvenAggregation else null,
          )

        override fun deleteAggregationsUpToEndBlockNumber(endBlockNumberInclusive: Long) =
          SafeFuture.completedFuture(0)

        override fun deleteAggregationsAfterBlockNumber(startingBlockNumberInclusive: Long) =
          SafeFuture.completedFuture(0)
      }
    val forcedTransactionsDao =
      object : ForcedTransactionsDao {
        private val records =
          listOf(
            ForcedTransactionRecord(
              ftxNumber = 1uL,
              inclusionResult = ForcedTransactionInclusionResult.Included,
              simulatedExecutionBlockNumber = 33uL,
              simulatedExecutionBlockTimestamp = Instant.fromEpochSeconds(33),
              ftxBlockNumberDeadline = 50uL,
              ftxRollingHash = byteArrayOf(0x01),
              ftxRlp = byteArrayOf(0x01),
              proofStatus = ForcedTransactionRecord.ProofStatus.PROVEN,
            ),
            ForcedTransactionRecord(
              ftxNumber = 2uL,
              inclusionResult = ForcedTransactionInclusionResult.Included,
              simulatedExecutionBlockNumber = 34uL,
              simulatedExecutionBlockTimestamp = Instant.fromEpochSeconds(34),
              ftxBlockNumberDeadline = 51uL,
              ftxRollingHash = processedFtxRollingHash,
              ftxRlp = byteArrayOf(0x02),
              proofStatus = ForcedTransactionRecord.ProofStatus.PROVEN,
            ),
          )

        override fun save(ftx: ForcedTransactionRecord): SafeFuture<Unit> =
          SafeFuture.failedFuture(IllegalStateException("not implemented"))

        override fun findByNumber(ftxNumber: ULong): SafeFuture<ForcedTransactionRecord?> {
          return SafeFuture.completedFuture(records.find { it.ftxNumber == ftxNumber })
        }

        override fun list(): SafeFuture<List<ForcedTransactionRecord>> = SafeFuture.completedFuture(records)

        override fun deleteFtxUpToInclusive(ftxNumber: ULong): SafeFuture<Int> = SafeFuture.completedFuture(0)
      }

    val aggregationL2StateProvider =
      AggregationL2StateProviderImpl(
        ethApiClient = fakeEthApiClient,
        messageService = fakeMessageService,
        aggregationsRepository = aggregationsRepository,
        forcedTransactionsDao = forcedTransactionsDao,
      )

    val aggregationL2State = aggregationL2StateProvider.getAggregationL2State(36).get()

    assertThat(aggregationL2State.parentAggregationLastL1RollingHashMessageNumber).isEqualTo(7uL)
    assertThat(aggregationL2State.parentAggregationLastL1RollingHash).isEqualTo(rollingHash)
    assertThat(aggregationL2State.parentAggregationLastFtxNumber).isEqualTo(2uL)
    assertThat(aggregationL2State.parentAggregationLastFtxRollingHash).isEqualTo(processedFtxRollingHash)
  }

  @Test
  fun `falls back to latest proven aggregation when dao no longer retains finalized ftx`() {
    val rollingHash = byteArrayOf(0x11)
    val finalizedFtxRollingHash = byteArrayOf(0x22)
    val fakeEthApiClient = FakeEthApiClient().apply {
      setLatestBlockTag(58uL)
    }
    val fakeMessageService = FakeL2MessageService(contractDeployBlock = 0uL).apply {
      setLastAnchoredL1Message(7uL, rollingHash)
    }
    val latestProvenAggregation =
      createProofToFinalize(
        firstBlockNumber = 55,
        finalBlockNumber = 56,
        parentAggregationFtxNumber = 1uL,
        parentAggregationFtxRollingHash = finalizedFtxRollingHash,
        finalFtxNumber = 1uL,
        finalFtxRollingHash = finalizedFtxRollingHash,
      )
    val aggregationsRepository =
      object : AggregationsRepository {
        override fun findConsecutiveProvenBlobs(fromBlockNumber: Long) =
          SafeFuture.completedFuture(emptyList<BlobAndBatchCounters>())

        override fun saveNewAggregation(aggregation: net.consensys.zkevm.domain.Aggregation) =
          SafeFuture.failedFuture<Unit>(IllegalStateException("not implemented"))

        override fun getProofsToFinalize(
          fromBlockNumber: Long,
          finalEndBlockCreatedBefore: Instant,
          maximumNumberOfProofs: Int,
        ) = SafeFuture.completedFuture(emptyList<ProofToFinalize>())

        override fun findHighestConsecutiveEndBlockNumber(fromBlockNumber: Long?) =
          SafeFuture.completedFuture<Long?>(null)

        override fun findAggregationProofByEndBlockNumber(endBlockNumber: Long) =
          SafeFuture.completedFuture<ProofToFinalize?>(null)

        override fun findLatestProvenAggregationProofUpToEndBlockNumber(endBlockNumberInclusive: Long) =
          SafeFuture.completedFuture<ProofToFinalize?>(
            if (endBlockNumberInclusive >= 56L) latestProvenAggregation else null,
          )

        override fun deleteAggregationsUpToEndBlockNumber(endBlockNumberInclusive: Long) =
          SafeFuture.completedFuture(0)

        override fun deleteAggregationsAfterBlockNumber(startingBlockNumberInclusive: Long) =
          SafeFuture.completedFuture(0)
      }
    val forcedTransactionsDao =
      object : ForcedTransactionsDao {
        override fun save(ftx: ForcedTransactionRecord): SafeFuture<Unit> =
          SafeFuture.failedFuture(IllegalStateException("not implemented"))

        override fun findByNumber(ftxNumber: ULong): SafeFuture<ForcedTransactionRecord?> =
          SafeFuture.completedFuture(null)

        override fun list(): SafeFuture<List<ForcedTransactionRecord>> = SafeFuture.completedFuture(emptyList())

        override fun deleteFtxUpToInclusive(ftxNumber: ULong): SafeFuture<Int> = SafeFuture.completedFuture(0)
      }

    val aggregationL2StateProvider =
      AggregationL2StateProviderImpl(
        ethApiClient = fakeEthApiClient,
        messageService = fakeMessageService,
        aggregationsRepository = aggregationsRepository,
        forcedTransactionsDao = forcedTransactionsDao,
      )

    val aggregationL2State = aggregationL2StateProvider.getAggregationL2State(58).get()

    assertThat(aggregationL2State.parentAggregationLastL1RollingHashMessageNumber).isEqualTo(7uL)
    assertThat(aggregationL2State.parentAggregationLastL1RollingHash).isEqualTo(rollingHash)
    assertThat(aggregationL2State.parentAggregationLastFtxNumber).isEqualTo(1uL)
    assertThat(aggregationL2State.parentAggregationLastFtxRollingHash).isEqualTo(finalizedFtxRollingHash)
  }
}
