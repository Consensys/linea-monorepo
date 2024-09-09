package net.consensys.zkevm.persistence.dao.rejectedtransaction

import io.vertx.core.Future
import io.vertx.sqlclient.Row
import io.vertx.sqlclient.SqlClient
import io.vertx.sqlclient.Tuple
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.encodeHex
import net.consensys.linea.ModuleOverflow
import net.consensys.linea.RejectedTransaction
import net.consensys.linea.TransactionInfo
import net.consensys.linea.async.toSafeFuture
import net.consensys.zkevm.persistence.db.DuplicatedRecordException
import net.consensys.zkevm.persistence.db.SQLQueryLogger
import net.consensys.zkevm.persistence.db.isDuplicateKeyException
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration

class RejectedTransactionsPostgresDao(
  private val readConnection: SqlClient,
  private val writeConnection: SqlClient,
  private val config: Config,
  private val clock: Clock = Clock.System
) : RejectedTransactionsDao {
  private val log = LogManager.getLogger(this.javaClass.name)
  private val queryLog = SQLQueryLogger(log)

  data class Config(
    val queryableWindowSinceRejectTimestamp: Duration
  )

  companion object {
    // Public instead of internal to allow usage in integrationTest source set
    fun rejectedStageToDbValue(txRejectionStage: RejectedTransaction.Stage): String {
      return when (txRejectionStage) {
        RejectedTransaction.Stage.SEQUENCER -> "SEQ"
        RejectedTransaction.Stage.RPC -> "RPC"
        RejectedTransaction.Stage.P2P -> "P2P"
      }
    }

    fun dbValueToRejectedStage(value: String): RejectedTransaction.Stage {
      return when (value) {
        "SEQ" -> RejectedTransaction.Stage.SEQUENCER
        "RPC" -> RejectedTransaction.Stage.RPC
        "P2P" -> RejectedTransaction.Stage.P2P
        else -> throw IllegalStateException()
      }
    }

    fun parseRecord(record: Row): RejectedTransaction {
      return RejectedTransaction(
        txRejectionStage = record.getString("reject_stage").run(::dbValueToRejectedStage),
        timestamp = Instant.fromEpochMilliseconds(record.getLong("timestamp")),
        blockNumber = record.getLong("block_number")?.toULong(),
        transactionRLP = record.getBuffer("tx_rlp").bytes,
        reasonMessage = record.getString("reject_reason"),
        overflows = record.getJsonArray("overflows").let { jsonArray ->
          ModuleOverflow.parseListFromJsonString(jsonArray.encode())
        },
        transactionInfo = TransactionInfo(
          hash = record.getBuffer("tx_hash").bytes,
          from = record.getBuffer("tx_from").bytes,
          to = record.getBuffer("tx_to").bytes,
          nonce = record.getLong("tx_nonce").toULong()
        )
      )
    }

    @JvmStatic
    val rejectedTransactionsTable = "rejected_transactions"

    @JvmStatic
    val fullTransactionsTable = "full_transactions"
  }

  private val insertSql =
    """
      with x as (
        insert into $rejectedTransactionsTable
        (created_epoch_milli, tx_hash, tx_from, tx_to, tx_nonce,
        reject_stage, reject_reason, timestamp, block_number, overflows)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CAST($10::text as jsonb))
        RETURNING tx_hash, reject_reason
      )
      insert into $fullTransactionsTable
      (tx_hash, reject_reason, tx_rlp)
      SELECT x.tx_hash, x.reject_reason, $11
      FROM x
    """
      .trimIndent()

  private val selectSql =
    """
      select
        $rejectedTransactionsTable.*,
        $fullTransactionsTable.tx_rlp as tx_rlp
      from $rejectedTransactionsTable
      join $fullTransactionsTable on $rejectedTransactionsTable.tx_hash = $fullTransactionsTable.tx_hash
        and $rejectedTransactionsTable.reject_reason = $fullTransactionsTable.reject_reason
      where $fullTransactionsTable.tx_hash = $1 and $rejectedTransactionsTable.timestamp >= $2
      order by $rejectedTransactionsTable.timestamp desc
      limit 1
    """
      .trimIndent()

  private val deleteSql =
    """
      delete from $rejectedTransactionsTable
      where timestamp < $1
    """
      .trimIndent()

  private val insertSqlQuery = writeConnection.preparedQuery(insertSql)
  private val selectSqlQuery = readConnection.preparedQuery(selectSql)
  private val deleteSqlQuery = writeConnection.preparedQuery(deleteSql)

  override fun saveNewRejectedTransaction(rejectedTransaction: RejectedTransaction): SafeFuture<Unit> {
    val params: List<Any?> =
      listOf(
        clock.now().toEpochMilliseconds(),
        rejectedTransaction.transactionInfo!!.hash,
        rejectedTransaction.transactionInfo!!.from,
        rejectedTransaction.transactionInfo!!.to,
        rejectedTransaction.transactionInfo!!.nonce.toLong(),
        rejectedStageToDbValue(rejectedTransaction.txRejectionStage),
        rejectedTransaction.reasonMessage,
        rejectedTransaction.timestamp.toEpochMilliseconds(),
        rejectedTransaction.blockNumber?.toLong(),
        ModuleOverflow.parseToJsonString(rejectedTransaction.overflows),
        rejectedTransaction.transactionRLP
      )
    queryLog.log(Level.TRACE, insertSql, params)

    return insertSqlQuery.execute(Tuple.tuple(params))
      .map { }
      .recover { th ->
        if (isDuplicateKeyException(th)) {
          Future.failedFuture(
            DuplicatedRecordException(
              "RejectedTransaction ${rejectedTransaction.transactionInfo!!.hash.encodeHex()} is already persisted!",
              th
            )
          )
        } else {
          Future.failedFuture(th)
        }
      }
      .toSafeFuture()
  }

  override fun findRejectedTransactionByTxHash(txHash: ByteArray): SafeFuture<RejectedTransaction?> {
    return selectSqlQuery
      .execute(
        Tuple.of(
          txHash,
          clock.now().minus(config.queryableWindowSinceRejectTimestamp).toEpochMilliseconds()
        )
      )
      .toSafeFuture()
      .thenApply { rowSet -> rowSet.map(::parseRecord) }
      .thenApply { rejectedTxRecords -> rejectedTxRecords.firstOrNull() }
  }

  override fun deleteRejectedTransactionsBeforeTimestamp(timestamp: Instant): SafeFuture<Int> {
    return deleteSqlQuery
      .execute(Tuple.of(timestamp.toEpochMilliseconds()))
      .map { rowSet -> rowSet.rowCount() }
      .toSafeFuture()
  }
}
