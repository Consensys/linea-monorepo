package net.consensys.zkevm.persistence.dao.aggregation

import linea.domain.Aggregation
import linea.domain.BlobAndBatchCounters
import linea.domain.ProofToFinalize
import linea.persistence.AggregationsDao
import net.consensys.zkevm.persistence.db.PersistenceRetryer
import net.consensys.zkevm.persistence.db.RetryingDaoBase
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Instant

class RetryingPostgresAggregationsDao(
  delegate: PostgresAggregationsDao,
  persistenceRetryer: PersistenceRetryer,
) : RetryingDaoBase<PostgresAggregationsDao>(delegate, persistenceRetryer), AggregationsDao {
  override fun findConsecutiveProvenBlobs(fromBlockNumber: Long): SafeFuture<List<BlobAndBatchCounters>> =
    retrying { delegate.findConsecutiveProvenBlobs(fromBlockNumber) }

  override fun saveNewAggregation(aggregation: Aggregation): SafeFuture<Unit> =
    delegate.saveNewAggregation(aggregation)

  override fun getProofsToFinalize(
    fromBlockNumber: Long,
    finalEndBlockCreatedBefore: Instant,
    maximumNumberOfProofs: Int,
  ): SafeFuture<List<ProofToFinalize>> =
    retrying { delegate.getProofsToFinalize(fromBlockNumber, finalEndBlockCreatedBefore, maximumNumberOfProofs) }

  override fun findHighestConsecutiveEndBlockNumber(fromBlockNumber: Long?): SafeFuture<Long?> =
    retrying { delegate.findHighestConsecutiveEndBlockNumber(fromBlockNumber) }

  override fun findAggregationProofByEndBlockNumber(endBlockNumber: Long): SafeFuture<ProofToFinalize?> =
    retrying { delegate.findAggregationProofByEndBlockNumber(endBlockNumber) }

  override fun deleteAggregationsUpToEndBlockNumber(endBlockNumberInclusive: Long): SafeFuture<Int> =
    retrying { delegate.deleteAggregationsUpToEndBlockNumber(endBlockNumberInclusive) }

  override fun deleteAggregationsAfterBlockNumber(startingBlockNumberInclusive: Long): SafeFuture<Int> =
    retrying { delegate.deleteAggregationsAfterBlockNumber(startingBlockNumberInclusive) }
}
