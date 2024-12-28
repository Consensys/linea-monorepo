package net.consensys.zkevm.persistence.dao.aggregation

import kotlinx.datetime.Instant
import net.consensys.zkevm.domain.Aggregation
import net.consensys.zkevm.domain.BlobAndBatchCounters
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.persistence.db.PersistenceRetryer
import tech.pegasys.teku.infrastructure.async.SafeFuture

class RetryingPostgresAggregationsDao(
  private val delegate: PostgresAggregationsDao,
  private val persistenceRetryer: PersistenceRetryer
) : AggregationsDao {
  override fun findConsecutiveProvenBlobs(fromBlockNumber: Long): SafeFuture<List<BlobAndBatchCounters>> {
    return persistenceRetryer.retryQuery(
      { delegate.findConsecutiveProvenBlobs(fromBlockNumber) }
    )
  }

  override fun saveNewAggregation(aggregation: Aggregation): SafeFuture<Unit> {
    return delegate.saveNewAggregation(aggregation)
  }

  override fun getProofsToFinalize(
    fromBlockNumber: Long,
    finalEndBlockCreatedBefore: Instant,
    maximumNumberOfProofs: Int
  ): SafeFuture<List<ProofToFinalize>> {
    return persistenceRetryer.retryQuery(
      {
        delegate.getProofsToFinalize(
          fromBlockNumber,
          finalEndBlockCreatedBefore,
          maximumNumberOfProofs
        )
      }
    )
  }

  override fun findAggregationProofByEndBlockNumber(endBlockNumber: Long): SafeFuture<ProofToFinalize?> {
    return persistenceRetryer.retryQuery({ delegate.findAggregationProofByEndBlockNumber(endBlockNumber) })
  }

  override fun deleteAggregationsUpToEndBlockNumber(endBlockNumberInclusive: Long): SafeFuture<Int> {
    return persistenceRetryer.retryQuery({ delegate.deleteAggregationsUpToEndBlockNumber(endBlockNumberInclusive) })
  }

  override fun deleteAggregationsAfterBlockNumber(startingBlockNumberInclusive: Long): SafeFuture<Int> {
    return persistenceRetryer.retryQuery({ delegate.deleteAggregationsAfterBlockNumber(startingBlockNumberInclusive) })
  }
}
