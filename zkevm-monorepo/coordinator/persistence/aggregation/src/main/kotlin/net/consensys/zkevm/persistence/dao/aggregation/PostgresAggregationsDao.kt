package net.consensys.zkevm.persistence.dao.aggregation

import io.vertx.core.Future
import io.vertx.pgclient.PgException
import io.vertx.sqlclient.Row
import io.vertx.sqlclient.SqlClient
import io.vertx.sqlclient.Tuple
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.linea.async.toSafeFuture
import net.consensys.zkevm.coordinator.clients.prover.serialization.ProofToFinalizeJsonResponse
import net.consensys.zkevm.domain.Aggregation
import net.consensys.zkevm.domain.BlobAndBatchCounters
import net.consensys.zkevm.domain.BlobCounters
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.ExecutionProofVersions
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.VersionedExecutionProofs
import net.consensys.zkevm.domain.toBlockIntervalsString
import net.consensys.zkevm.persistence.db.DuplicatedRecordException
import net.consensys.zkevm.persistence.db.SQLQueryLogger
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.infrastructure.async.SafeFuture

class PostgresAggregationsDao(
  connection: SqlClient,
  private val clock: Clock = Clock.System
) : AggregationsDao {
  private val log = LogManager.getLogger(this.javaClass.name)
  private val queryLog = SQLQueryLogger(log)

  companion object {
    // Public instead of internal to allow usage in integrationTest source set
    fun aggregationStatusToDbValue(status: Aggregation.Status): Int {
      return when (status) {
        Aggregation.Status.Proven -> 1
        Aggregation.Status.Proving -> 2
      }
    }

    @JvmStatic
    val aggregationsTable = "aggregations"

    @JvmStatic
    val batchesTable = "batches"

    @JvmStatic
    val blobsTable = "blobs"
  }

  private val selectBatchesAndBlobsForAggregation = connection.preparedQuery(
    """
      with ranked_blobs as (
          select
              start_block_number, end_block_number, conflation_calculator_version, batches_count, status, start_block_timestamp, end_block_timestamp,
              lag(conflation_calculator_version, 1) over (order by start_block_number asc) as previous_conflation_calculator_version,
              dense_rank() over (partition by start_block_number order by conflation_calculator_version DESC) version_rank
          from $blobsTable
          where
              start_block_number >= $1 -- EXPECTED_START_BLOCK_NUMBER
            AND status = 1
          order by start_block_number asc, conflation_calculator_version DESC
      ),
     blob_previous_ends as (
         select
             * ,
             coalesce(max(end_block_number) over (order by start_block_number asc, end_block_number asc, conflation_calculator_version desc ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING), 0) as max_blob_end
         from ranked_blobs
         where version_rank = 1
           AND (conflation_calculator_version >= previous_conflation_calculator_version OR previous_conflation_calculator_version is null)
     ),
     removed_old_blobs as (select *
                           from blob_previous_ends
                           where end_block_number > max_blob_end),
     first_gapped_blob as (select start_block_number
                           from removed_old_blobs
                           where start_block_number > $1 and start_block_number - 1 != max_blob_end
                           limit 1
     ),
     consecutive_proven_blobs as (
         select *
         from removed_old_blobs
         where EXISTS (select 1 from blobs where start_block_number = $1 and status = 1)
           and (removed_old_blobs.max_blob_end = removed_old_blobs.start_block_number - 1 or removed_old_blobs.start_block_number = $1)
           and ((select count(1) from first_gapped_blob) = 0 or removed_old_blobs.start_block_number < (select * from first_gapped_blob))
     ),
     ranked_batches as (
         select
             start_block_number, end_block_number, conflation_calculator_version, prover_version,
             lag(conflation_calculator_version, 1) over (order by start_block_number asc) as previous_conflation_calculator_version,
             dense_rank() over (partition by start_block_number order by conflation_calculator_version DESC, (end_block_number - start_block_number) asc) version_rank
         from $batchesTable
         where
             status = 2 AND
             start_block_number >= $1 -- EXPECTED_START_BLOCK_NUMBER
         order by start_block_number asc, conflation_calculator_version DESC
     ),
     batches_previous_ends as (
         select
             * ,
             coalesce(max(end_block_number) over (order by start_block_number asc, end_block_number asc, conflation_calculator_version desc ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING), 0) as max_batch_end
         from ranked_batches
         where version_rank = 1
           AND (conflation_calculator_version >= previous_conflation_calculator_version OR previous_conflation_calculator_version is null)
     ),
     removed_old_batches as (select *
                           from batches_previous_ends
                           where end_block_number > max_batch_end
                           ),
     first_gapped_batch as (select start_block_number
                           from removed_old_batches
                           where start_block_number > $1 and start_block_number - 1 != max_batch_end
                           limit 1
     ),
     consecutive_proven_batches as (
         select *
         from removed_old_batches
         where EXISTS (select 1 from batches where start_block_number = $1 and status = 2)
           and (removed_old_batches.max_batch_end = removed_old_batches.start_block_number - 1 or removed_old_batches.start_block_number = $1)
           and ((select count(1) from first_gapped_batch) = 0 or removed_old_batches.start_block_number < (select * from first_gapped_batch))
     ),
     final_blobs as (
         select
             bl.start_block_number as blob_start_block_number, bl.end_block_number as blob_end_block_number, bl.conflation_calculator_version, bl.batches_count, bl.start_block_timestamp, bl.end_block_timestamp,
             ba.start_block_number, ba.end_block_number, ba.prover_version
         from consecutive_proven_blobs as bl
                  LEFT JOIN consecutive_proven_batches as ba
                            ON ba.conflation_calculator_version = bl.conflation_calculator_version
                                AND ba.start_block_number >= bl.start_block_number
                                AND ba.end_block_number <= bl.end_block_number
         where ba.start_block_number is not null
         ORDER BY bl.start_block_number ASC
     )
select * from final_blobs
    """.trimIndent()
  )

  private val insertQuery =
    connection.preparedQuery(
      """
        insert into $aggregationsTable
        (start_block_number, end_block_number, status, aggregation_calculator_version, start_block_timestamp, batch_count, aggregation_proof)
        VALUES ($1, $2, $3, $4, $5, $6, CAST($7::text as jsonb))
      """.trimIndent()
    )

  private val setAggregationAsProvenQuery =
    connection.preparedQuery(
      """
      update $aggregationsTable set status = $1, aggregation_proof = CAST($2::text as jsonb)
      where start_block_number = $3 and end_block_number = $4 and aggregation_calculator_version = $5 and batch_count = $6
    """
        .trimIndent()
    )

  private val selectAggregationProofs =
    connection.preparedQuery(
      """
        with ranked_aggregation_versions as (select *,
          dense_rank() over (partition by start_block_number order by aggregation_calculator_version desc) version_rank
          from $aggregationsTable
          where start_block_number >= $1 and status = $2
          order by start_block_number asc, aggregation_calculator_version desc),
        previous_ends as (select *,
          lag(end_block_number, 1) over (order by end_block_number asc, aggregation_calculator_version desc) as previous_end_block_number
          from ranked_aggregation_versions
          where version_rank = 1),
        first_gapped_aggregation as (select start_block_number
          from previous_ends
          where start_block_number > $1 and start_block_number - 1 != previous_end_block_number
          limit 1)

        select
          aggregation_proof
        from previous_ends
        where EXISTS (select 1 from ranked_aggregation_versions where start_block_number = $1)
          and (previous_ends.previous_end_block_number = previous_ends.start_block_number - 1 or previous_ends.start_block_number = $1)
          and ((select count(1) from first_gapped_aggregation) = 0 or previous_ends.start_block_number < (select * from first_gapped_aggregation))
        limit $3
      """.trimIndent()
    )

  private val findAggregationByEndBlockNumber =
    connection.preparedQuery(
      """
        select * from $aggregationsTable
        where end_block_number = $1
        LIMIT 2
      """.trimIndent()
    )

  private val deleteUptoQuery =
    connection.preparedQuery(
      """
        delete from $aggregationsTable
        where end_block_number <= $1
      """.trimIndent()
    )

  private val deleteAfterQuery =
    connection.preparedQuery(
      """
        delete from $aggregationsTable
        where start_block_number >= $1
      """.trimIndent()
    )

  private fun parseBatchAnbBlobRecord(record: Row): BatchRecordWithBlobInfo {
    return BatchRecordWithBlobInfo(
      blobStartBlockNumber = record.getLong("blob_start_block_number").toULong(),
      blobEndBlockNumber = record.getLong("blob_end_block_number").toULong(),
      blobStartBlockTimestamp = Instant.fromEpochMilliseconds(record.getLong("start_block_timestamp")),
      blobEndBlockTimestamp = Instant.fromEpochMilliseconds(record.getLong("end_block_timestamp")),
      batchesCount = record.getLong("batches_count").toUInt(),
      batchStartBlockNumber = record.getLong("start_block_number").toULong(),
      batchEndBlockNumber = record.getLong("end_block_number").toULong(),
      batchProverVersion = record.getString("prover_version"),
      conflationCalculatorVersion = record.getString("conflation_calculator_version")
    )
  }

  private data class BatchRecordWithBlobInfo(
    val blobStartBlockNumber: ULong,
    val blobEndBlockNumber: ULong,
    val blobStartBlockTimestamp: Instant,
    val blobEndBlockTimestamp: Instant,
    val batchesCount: UInt,
    val batchStartBlockNumber: ULong,
    val batchEndBlockNumber: ULong,
    val batchProverVersion: String,
    val conflationCalculatorVersion: String
  )

  override fun findConsecutiveProvenBlobs(
    fromBlockNumber: Long
  ): SafeFuture<List<BlobAndBatchCounters>> {
    return selectBatchesAndBlobsForAggregation
      .execute(Tuple.of(fromBlockNumber))
      .toSafeFuture()
      .thenApply { rowSet ->
        val batches: List<BatchRecordWithBlobInfo> = rowSet.map { row -> parseBatchAnbBlobRecord(row) }
        val batchesByBlob: Map<ULong, List<BatchRecordWithBlobInfo>> = batches.groupBy { it.blobStartBlockNumber }
        val result = batchesByBlob
          .map { (_, batches) ->
            val firstBatch = batches.first()
            val blobCounters = BlobCounters(
              numberOfBatches = firstBatch.batchesCount,
              startBlockNumber = firstBatch.blobStartBlockNumber,
              endBlockNumber = firstBatch.blobEndBlockNumber,
              startBlockTimestamp = firstBatch.blobStartBlockTimestamp,
              endBlockTimestamp = firstBatch.blobEndBlockTimestamp
            )
            BlobAndBatchCounters(
              blobCounters = blobCounters,
              versionedExecutionProofs = VersionedExecutionProofs(
                executionProofs = BlockIntervals(
                  startingBlockNumber = batches.first().batchStartBlockNumber,
                  upperBoundaries = batches.map { it.batchEndBlockNumber }
                ),
                executionVersion = batches.map { batch ->
                  ExecutionProofVersions(
                    conflationCalculatorVersion = batch.conflationCalculatorVersion,
                    executionProverVersion = batch.batchProverVersion
                  )
                }
              )
            )
          }
          .filter {
            it.blobCounters.endBlockNumber == it.versionedExecutionProofs.executionProofs.upperBoundaries.last()
          }
          .sortedBy { it.blobCounters.startBlockNumber }
        if (result.isNotEmpty() && result.first().blobCounters.startBlockNumber == fromBlockNumber.toULong()) {
          result
        } else {
          emptyList()
        }
      }
  }

  override fun saveNewAggregation(aggregation: Aggregation): SafeFuture<Unit> {
    val startBlockNumber = aggregation.startBlockNumber.toLong()
    val endBlockNumber = aggregation.endBlockNumber.toLong()
    val status = aggregationStatusToDbValue(aggregation.status)
    val aggregationCalculatorVersion = aggregation.aggregationCalculatorVersion
    val batchCount = aggregation.batchCount.toLong()
    val aggregationProof = serializeAggregationProof(aggregation.aggregationProof)

    val params =
      listOf(
        startBlockNumber,
        endBlockNumber,
        status,
        aggregationCalculatorVersion,
        clock.now().toEpochMilliseconds(),
        batchCount,
        aggregationProof
      )
    queryLog.log(Level.TRACE, insertQuery.toString(), params)
    return insertQuery.execute(Tuple.tuple(params))
      .map { }
      .recover { th ->
        if (th is PgException &&
          th.errorMessage == "duplicate key value violates unique constraint \"aggregations_pkey\""
        ) {
          Future.failedFuture(
            DuplicatedRecordException(
              "Aggregation startBlockNumber=$startBlockNumber, endBlockNumber=$endBlockNumber, " +
                "aggregationCalculatorVersion=$aggregationCalculatorVersion " +
                "batchCount=$batchCount is already persisted!",
              th
            )
          )
        } else {
          Future.failedFuture(th)
        }
      }
      .toSafeFuture()
  }

  override fun getProofsToFinalize(
    fromBlockNumber: Long,
    status: Aggregation.Status,
    finalEndBlockCreatedBefore: Instant,
    maximumNumberOfProofs: Int
  ): SafeFuture<List<ProofToFinalize>> {
    return selectAggregationProofs
      .execute(
        Tuple.of(
          fromBlockNumber,
          aggregationStatusToDbValue(status),
          maximumNumberOfProofs
        )
      )
      .toSafeFuture()
      .thenApply { rowSet ->
        rowSet.map(::parseAggregationProofs).filter { proofToFinalize ->
          proofToFinalize.finalTimestamp <= finalEndBlockCreatedBefore
        }
      }
  }

  override fun findAggregationProofByEndBlockNumber(endBlockNumber: Long): SafeFuture<ProofToFinalize?> {
    return findAggregationByEndBlockNumber
      .execute(
        Tuple.of(endBlockNumber)
      )
      .toSafeFuture()
      .thenApply { rowSet ->
        val aggregationProofs = rowSet.map(::parseAggregationProofs)
        if (aggregationProofs.size > 1) {
          // coordinator now cleans up aggregations table on startup before resuming conflation
          // if this happens, a conflation invariant was broken
          throw IllegalStateException(
            "Multiple aggregations found for endBlockNumber=$endBlockNumber " +
              "aggregations=${aggregationProofs.toBlockIntervalsString()}"
          )
        } else {
          aggregationProofs.firstOrNull()
        }
      }
  }

  override fun updateAggregationAsProven(aggregation: Aggregation): SafeFuture<Int> {
    val startBlockNumber = aggregation.startBlockNumber.toLong()
    val endBlockNumber = aggregation.endBlockNumber.toLong()
    val status = aggregationStatusToDbValue(aggregation.status)
    val aggregationCalculatorVersion = aggregation.aggregationCalculatorVersion
    val batchCount = aggregation.batchCount.toLong()
    assert(aggregation.aggregationProof != null)
    val aggregationProof = serializeAggregationProof(aggregation.aggregationProof)

    val params =
      listOf(
        status,
        aggregationProof,
        startBlockNumber,
        endBlockNumber,
        aggregationCalculatorVersion,
        batchCount
      )
    queryLog.log(Level.TRACE, setAggregationAsProvenQuery.toString(), params)
    return setAggregationAsProvenQuery.execute(Tuple.tuple(params))
      .map { rowSet -> rowSet.rowCount() }
      .toSafeFuture()
  }

  override fun deleteAggregationsUpToEndBlockNumber(endBlockNumberInclusive: Long): SafeFuture<Int> {
    return deleteUptoQuery
      .execute(Tuple.of(endBlockNumberInclusive))
      .map { rowSet -> rowSet.rowCount() }
      .toSafeFuture()
  }

  override fun deleteAggregationsAfterBlockNumber(startingBlockNumberInclusive: Long): SafeFuture<Int> {
    return deleteAfterQuery
      .execute(Tuple.of(startingBlockNumberInclusive))
      .map { rowSet -> rowSet.rowCount() }
      .toSafeFuture()
  }

  private fun parseExecutionProofs(record: Row): Pair<ULong, ExecutionProofVersions> {
    return Pair(
      record.getLong("end_block_number").toULong(),
      ExecutionProofVersions(
        record.getString("conflation_calculator_version"),
        record.getString("prover_version")
      )
    )
  }

  private fun serializeAggregationProof(proofToFinalize: ProofToFinalize?): String? {
    return proofToFinalize?.let { ProofToFinalizeJsonResponse.fromDomainObject(proofToFinalize).toJsonString() }
  }

  private fun parseAggregationProofs(record: Row): ProofToFinalize {
    val aggregationProofJsonObj = record.getJsonObject("aggregation_proof")
    return ProofToFinalizeJsonResponse.fromJsonString(aggregationProofJsonObj.encode()).toDomainObject()
  }
}
