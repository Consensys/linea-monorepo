package net.consensys.zkevm.ethereum.settlement.persistence

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import io.vertx.core.Future
import io.vertx.pgclient.PgException
import io.vertx.sqlclient.PreparedQuery
import io.vertx.sqlclient.Row
import io.vertx.sqlclient.RowSet
import io.vertx.sqlclient.SqlClient
import io.vertx.sqlclient.Tuple
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import kotlinx.datetime.toKotlinInstant
import net.consensys.linea.async.toSafeFuture
import net.consensys.zkevm.coordinator.clients.response.ProverResponsesRepository
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64

/**
 * WARNING: Existing mappings should not chane. Otherwise, can break production New One can be added
 * though.
 */
public fun batchStatusToDbValue(status: Batch.Status): Int {
  // using manual mapping to catch errors at compile time instead of runtime
  return when (status) {
    Batch.Status.Finalized -> 1
    Batch.Status.Pending -> 2
  }
}

internal fun dbValueToStatus(value: Int): Batch.Status {
  return when (value) {
    1 -> Batch.Status.Finalized
    2 -> Batch.Status.Pending
    else ->
      throw IllegalStateException(
        "Value '$value' does not map to any ${Batch.Status::class.simpleName}"
      )
  }
}

private data class BatchRecord(
  val index: ProverResponsesRepository.ProverResponseIndex,
  val status: Batch.Status
)

class PostgresBatchesRepository(
  config: Config,
  connection: SqlClient,
  private val responsesRepository: ProverResponsesRepository,
  private val clock: Clock = Clock.System
) : BatchesRepository {
  private val log = LogManager.getLogger(this.javaClass.name)
  private val queryLog = SQLQueryLogger(log)

  data class Config(val maxBatchesToReturn: UInt)

  companion object {
    @JvmStatic
    val TableName = "batches"
  }

  private val selectQuery =
    connection.preparedQuery(
      """
      with ranked_versions as (select *,
              dense_rank() over (partition by start_block_number order by end_block_number asc, prover_version desc) version_rank
              from $TableName
              where start_block_number >= $1
              order by start_block_number asc, prover_version desc),
          previous_ends as (select *,
              lag(end_block_number, 1) over (order by end_block_number asc, prover_version desc) as previous_end_block_number
              from ranked_versions
              where version_rank = 1),
           first_gapped_batch as (select start_block_number
              from previous_ends
              where start_block_number > $1 and start_block_number - 1 != previous_end_block_number
              limit 1)
      select *
      from previous_ends
      where EXISTS (select 1 from $TableName where start_block_number = $1)
        and (previous_ends.previous_end_block_number = previous_ends.start_block_number - 1 or previous_ends.start_block_number = $1)
        and ((select count(1) from first_gapped_batch) = 0 or previous_ends.start_block_number < (select * from first_gapped_batch))
      limit ${config.maxBatchesToReturn}
    """
        .trimIndent()
    )
  private val insertSql =
    """
      insert into $TableName
      (created_epoch_milli, start_block_number, end_block_number, prover_version, status) VALUES ($1, $2, $3, $4, $5)
    """
      .trimIndent()

  private val findHighestByStatusSql =
    """
      select * from $TableName
      where status = $1
      order by end_block_number desc limit 1
    """.trimIndent()

  private val findHighestConsecutiveRangeByStatusSql =
    """
      WITH pending_batches as (
        select *, rank() over (partition by end_block_number ORDER BY prover_version desc) version_rank
        from $TableName
        where status = $1
        order by start_block_number asc
      ),
      batches_highest_version as (
        select *, lead(start_block_number, 1, 0) over (order by start_block_number asc) as next_start_block_number
        from pending_batches
        where version_rank = 1
      ),
      batches_gaps as (
        select *, next_start_block_number - end_block_number as gap_size
        from batches_highest_version
      )
      select * from batches_gaps
      where gap_size != 1
      limit 1
    """
      .trimIndent()

  private val setStatusSql =
    """
        update $TableName set status = $1
        where end_block_number <= $2 and status = $3
    """
      .trimIndent()

  private val findHighestConsecutiveByStatusQuery = connection.preparedQuery(findHighestConsecutiveRangeByStatusSql)
  private val findHighestByStatusQuery = connection.preparedQuery(findHighestByStatusSql)
  private val insertQuery = connection.preparedQuery(insertSql)
  private val setStatusQuery = connection.preparedQuery(setStatusSql)

  override fun saveNewBatch(batch: Batch): SafeFuture<Unit> {
    val startBlockNumber = batch.startBlockNumber.longValue()
    val endBlockNumber = batch.endBlockNumber.longValue()
    val proverVersion = batch.proverResponse.proverVersion
    val params =
      listOf(
        clock.now().toEpochMilliseconds(),
        startBlockNumber,
        endBlockNumber,
        proverVersion,
        batchStatusToDbValue(batch.status)
      )
    queryLog.log(Level.TRACE, insertSql, params)
    return insertQuery.execute(Tuple.tuple(params))
      .map { Unit }
      .recover { th ->
        if (th is PgException && th.errorMessage == "duplicate key value violates unique constraint \"batches_pkey\"") {
          Future.failedFuture<Unit>(
            DuplicatedBatchException(
              "Batch startBlockNumber=$startBlockNumber, endBlockNumber=$endBlockNumber, " +
                "proverVersion=$proverVersion is already persisted!",
              th
            )
          )
        } else {
          Future.failedFuture<Unit>(th)
        }
      }
      .toSafeFuture()
  }

  fun getConsecutiveBatchesFromBlockNumber(
    startingBlockNumberInclusive: UInt64
  ): SafeFuture<List<Batch>> {
    return selectQuery
      .execute(Tuple.of(startingBlockNumberInclusive.longValue()))
      .toSafeFuture()
      .thenCompose { rowSet ->
        SafeFuture.collectAll(
          rowSet.map(this::parseRecord)
            .map(this::loadProverResponseFromFileSystem)
            .stream()
        ).thenApply { batches ->
          batches.takeWhile { it != null }
        }
      }
  }

  override fun getConsecutiveBatchesFromBlockNumber(
    startingBlockNumberInclusive: UInt64,
    endBlockCreatedBefore: Instant
  ): SafeFuture<List<Batch>> {
    return getConsecutiveBatchesFromBlockNumber(startingBlockNumberInclusive)
      .thenApply { batches ->
        batches.filter { batch ->
          batch.proverResponse.blocksData.last().timestamp.toKotlinInstant() < endBlockCreatedBefore
        }
      }
  }

  private fun loadProverResponseFromFileSystem(batchRecord: BatchRecord): SafeFuture<Batch> {
    return responsesRepository.find(batchRecord.index).thenApply { proverResponse ->
      when (proverResponse) {
        is Ok -> {
          Batch(
            batchRecord.index.startBlockNumber,
            batchRecord.index.endBlockNumber,
            proverResponse.value,
            batchRecord.status
          )
        }

        is Err -> {
          log.warn(
            "Trying to get batch by index, {}, but response is unsuccessful! errorMessage: {}",
            batchRecord.index,
            proverResponse.error.asException().message
          )
          null
        }
      }
    }
  }

  fun findHighestBatchByStatus(
    sql: String,
    preparedQuery: PreparedQuery<RowSet<Row>>,
    status: Batch.Status
  ): SafeFuture<Batch?> {
    val params = listOf(batchStatusToDbValue(status))
    queryLog.log(Level.TRACE, sql, params)
    return preparedQuery
      .execute(Tuple.tuple(params))
      .toSafeFuture()
      .thenCompose { rowSet ->
        if (rowSet.size() > 0) {
          this.loadProverResponseFromFileSystem(parseRecord(rowSet.first()))
        } else {
          SafeFuture.completedFuture(null)
        }
      }
  }

  override fun findBatchWithHighestEndBlockNumberByStatus(status: Batch.Status): SafeFuture<Batch?> {
    return findHighestBatchByStatus(findHighestByStatusSql, findHighestByStatusQuery, status)
  }

  override fun findHighestConsecutiveBatchByStatus(status: Batch.Status): SafeFuture<Batch?> {
    return findHighestBatchByStatus(findHighestConsecutiveRangeByStatusSql, findHighestConsecutiveByStatusQuery, status)
  }

  override fun setBatchStatusUpToEndBlockNumber(
    endBlockNumberInclusive: UInt64,
    currentStatus: Batch.Status,
    newStatus: Batch.Status
  ): SafeFuture<Int> {
    return setStatusQuery
      .execute(
        Tuple.of(
          batchStatusToDbValue(newStatus),
          endBlockNumberInclusive.longValue(),
          batchStatusToDbValue(currentStatus)
        )
      )
      .map { rowSet -> rowSet.rowCount() }
      .toSafeFuture()
  }

  private fun parseRecord(record: Row): BatchRecord {
    return BatchRecord(
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(record.getLong("start_block_number")),
        UInt64.valueOf(record.getLong("end_block_number")),
        record.getString("prover_version")
      ),
      dbValueToStatus(record.getInteger("status"))
    )
  }

  fun ULong.toUInt64(): UInt64 = UInt64.valueOf(this.toLong())
}
