package net.consensys.zkevm.persistence.dao.rejectedtransaction

import io.vertx.junit5.VertxExtension
import io.vertx.sqlclient.PreparedQuery
import io.vertx.sqlclient.Row
import io.vertx.sqlclient.RowSet
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.FakeFixedClock
import net.consensys.decodeHex
import net.consensys.encodeHex
import net.consensys.linea.ModuleOverflow
import net.consensys.linea.RejectedTransaction
import net.consensys.linea.TransactionInfo
import net.consensys.linea.async.get
import net.consensys.trimToMillisecondPrecision
import net.consensys.zkevm.persistence.db.DbHelper
import net.consensys.zkevm.persistence.db.DuplicatedRecordException
import net.consensys.zkevm.persistence.test.CleanDbTestSuiteParallel
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.junit.jupiter.api.extension.ExtendWith
import java.util.concurrent.ExecutionException
import kotlin.time.Duration.Companion.hours
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
class RejectedTransactionsPostgresDaoTest : CleanDbTestSuiteParallel() {
  init {
    target = "1"
    migrationLocations = "filesystem:../../../config/transaction-exclusion-api/db/migration/"
  }

  override val databaseName = DbHelper.generateUniqueDbName("coordinator-tests-rejectedtxns-dao")
  private var fakeClock = FakeFixedClock(Clock.System.now())
  private lateinit var rejectedTransactionsPostgresDao: RejectedTransactionsPostgresDao

  // Default RejectedTransaction field values
  private val defaultTxRejectionStage = RejectedTransaction.Stage.SEQUENCER
  private val defaultTimestamp = fakeClock.now().minus(10.seconds)
  private val defaultBlockNumber = 10000UL
  private val defaultTransactionRLP =
    (
      "0x02f8388204d2648203e88203e88203e8941195cf65f83b3a5768f3c4" +
        "96d3a05ad6412c64b38203e88c666d93e9cc5f73748162cea9c0017b8201c8"
      )
      .decodeHex()
  private val defaultReasonMessage = "Transaction line count for module ADD=402 is above the limit 70"
  private val defaultOverflows = listOf(
    ModuleOverflow(
      module = "ADD",
      count = 402,
      limit = 70
    ),
    ModuleOverflow(
      module = "MUL",
      count = 587,
      limit = 401
    ),
    ModuleOverflow(
      module = "EXP",
      count = 9000,
      limit = 8192
    )
  )
  private val defaultTransactionInfo = TransactionInfo(
    hash = "0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350ad7".decodeHex(),
    from = "0x4d144d7b9c96b26361d6ac74dd1d8267edca4fc2".decodeHex(),
    to = "0x1195cf65f83b3a5768f3c496d3a05ad6412c64b3".decodeHex(),
    nonce = 100UL
  )

  // Helper functions
  private fun createRejectedTransaction(
    txRejectionStage: RejectedTransaction.Stage = defaultTxRejectionStage,
    timestamp: Instant = defaultTimestamp,
    blockNumber: ULong? = defaultBlockNumber,
    transactionRLP: ByteArray = defaultTransactionRLP,
    reasonMessage: String = defaultReasonMessage,
    overflows: List<ModuleOverflow> = defaultOverflows,
    transactionInfo: TransactionInfo = defaultTransactionInfo
  ): RejectedTransaction {
    return RejectedTransaction(
      txRejectionStage = txRejectionStage,
      timestamp = timestamp.trimToMillisecondPrecision(),
      blockNumber = blockNumber,
      transactionRLP = transactionRLP,
      reasonMessage = reasonMessage,
      overflows = overflows,
      transactionInfo = transactionInfo
    )
  }

  private fun dbTableContentQuery(dbTableName: String): PreparedQuery<RowSet<Row>> =
    sqlClient.preparedQuery("select * from $dbTableName")

  private fun rejectedTransactionsTotalRows(): Int =
    dbTableContentQuery(RejectedTransactionsPostgresDao.rejectedTransactionsTable).execute().get().size()

  private fun fullTransactionsTotalRows(): Int =
    dbTableContentQuery(RejectedTransactionsPostgresDao.fullTransactionsTable).execute().get().size()

  @BeforeEach
  fun beforeEach() {
    fakeClock.setTimeTo(Clock.System.now())
    rejectedTransactionsPostgresDao =
      RejectedTransactionsPostgresDao(
        readConnection = sqlClient,
        writeConnection = sqlClient,
        config = RejectedTransactionsPostgresDao.Config(
          queryableWindowSinceRejectTimestamp = 1.hours
        ),
        clock = fakeClock
      )
  }

  private fun performInsertTest(
    rejectedTransaction: RejectedTransaction
  ) {
    rejectedTransactionsPostgresDao.saveNewRejectedTransaction(rejectedTransaction).get()

    // assert the corresponding record was inserted into the full_transactions table
    val newlyInsertedFullTxnsRows = dbTableContentQuery(RejectedTransactionsPostgresDao.fullTransactionsTable)
      .execute().get().filter { row ->
        row.getBuffer("tx_hash").bytes.contentEquals(rejectedTransaction.transactionInfo!!.hash) &&
          row.getString("reject_reason") == rejectedTransaction.reasonMessage
      }
    assertThat(newlyInsertedFullTxnsRows.size).isEqualTo(1)

    // assert the corresponding record was inserted into the rejected_transactions table
    val newlyInsertedRejectedTxnsRows = dbTableContentQuery(RejectedTransactionsPostgresDao.rejectedTransactionsTable)
      .execute().get().filter { row ->
        row.getLong("created_epoch_milli") == fakeClock.now().toEpochMilliseconds() &&
          row.getString("reject_stage") ==
          RejectedTransactionsPostgresDao.rejectedStageToDbValue(rejectedTransaction.txRejectionStage) &&
          row.getLong("timestamp") == rejectedTransaction.timestamp.toEpochMilliseconds() &&
          row.getLong("block_number") == rejectedTransaction.blockNumber?.toLong() &&
          row.getString("reject_reason") == rejectedTransaction.reasonMessage &&
          row.getJsonArray("overflows").encode() == ModuleOverflow.parseToJsonString(rejectedTransaction.overflows) &&
          row.getBuffer("tx_hash").bytes.contentEquals(rejectedTransaction.transactionInfo!!.hash) &&
          row.getBuffer("tx_from").bytes.contentEquals(rejectedTransaction.transactionInfo!!.from) &&
          row.getBuffer("tx_to").bytes.contentEquals(rejectedTransaction.transactionInfo!!.to) &&
          row.getLong("tx_nonce") == rejectedTransaction.transactionInfo!!.nonce.toLong()
      }
    assertThat(newlyInsertedRejectedTxnsRows.size).isEqualTo(1)
  }

  @Test
  fun `saveNewRejectedTransaction inserts new rejected transaction to db`() {
    // insert a new rejected transaction
    performInsertTest(createRejectedTransaction())

    // assert that the total number of rows in the two tables are correct
    assertThat(rejectedTransactionsTotalRows()).isEqualTo(1)
    assertThat(fullTransactionsTotalRows()).isEqualTo(1)
  }

  @Test
  fun `saveNewRejectedTransaction inserts new rejected transactions with same txHash but different reason to db`() {
    // insert a new rejected transaction
    performInsertTest(createRejectedTransaction())

    // insert another rejected transaction with same txHash but different reason
    performInsertTest(
      createRejectedTransaction(
        txRejectionStage = RejectedTransaction.Stage.P2P,
        blockNumber = null,
        reasonMessage = "Transaction line count for module MUL=587 is above the limit 401",
        overflows = listOf(
          ModuleOverflow(
            module = "ADD",
            count = 587,
            limit = 401
          )
        )
      )
    )

    // assert that the total number of rows in the two tables are correct
    assertThat(rejectedTransactionsTotalRows()).isEqualTo(2)
    assertThat(fullTransactionsTotalRows()).isEqualTo(2)
  }

  @Test
  fun `saveNewRejectedTransaction throws error when inserting rejected transactions with same txHash and reason`() {
    // insert a new rejected transaction
    performInsertTest(createRejectedTransaction())

    // another rejected transaction with same txHash and reason
    val duplicatedRejectedTransaction = createRejectedTransaction(
      txRejectionStage = RejectedTransaction.Stage.P2P,
      blockNumber = null,
      overflows = listOf(
        ModuleOverflow(
          module = "ADD",
          count = 587,
          limit = 401
        )
      )
    )

    // assert that the insertion of duplicatedRejectedTransaction would trigger DuplicatedRecordException error
    assertThrows<ExecutionException> {
      rejectedTransactionsPostgresDao.saveNewRejectedTransaction(duplicatedRejectedTransaction).get()
    }.also { executionException ->
      assertThat(executionException.cause).isInstanceOf(DuplicatedRecordException::class.java)
      assertThat(executionException.cause!!.message)
        .isEqualTo(
          "RejectedTransaction ${duplicatedRejectedTransaction.transactionInfo!!.hash.encodeHex()} " +
            "was already persisted!"
        )
    }

    // assert that the total number of rows in the two tables are correct
    assertThat(rejectedTransactionsTotalRows()).isEqualTo(1)
    assertThat(fullTransactionsTotalRows()).isEqualTo(1)
  }

  @Test
  fun `findRejectedTransactionByTxHash returns rejected transaction with most recent timestamp from db`() {
    // insert a new rejected transaction
    val oldestRejectedTransaction = createRejectedTransaction(
      timestamp = fakeClock.now().minus(10.seconds)
    )
    performInsertTest(oldestRejectedTransaction)

    // insert another rejected transaction with same txHash but different reason and
    // with a more recent timestamp
    performInsertTest(
      createRejectedTransaction(
        reasonMessage = "Transaction line count for module MUL=587 is above the limit 401",
        timestamp = fakeClock.now().minus(9.seconds)
      )
    )

    // insert another rejected transaction with same txHash but different reason
    // and with the most recent timestamp
    val newestRejectedTransaction = createRejectedTransaction(
      reasonMessage = "Transaction line count for module EXP=9000 is above the limit 8192",
      timestamp = fakeClock.now().minus(8.seconds)
    )
    performInsertTest(newestRejectedTransaction)

    // find the rejected transaction with the txHash
    val foundRejectedTransaction = rejectedTransactionsPostgresDao.findRejectedTransactionByTxHash(
      oldestRejectedTransaction.transactionInfo!!.hash
    ).get()

    // assert that the found rejected transaction is the same as the one with most recent timestamp
    assertThat(foundRejectedTransaction).isEqualTo(newestRejectedTransaction)

    // assert that the total number of rows in the two tables are correct
    assertThat(rejectedTransactionsTotalRows()).isEqualTo(3)
    assertThat(fullTransactionsTotalRows()).isEqualTo(3)
  }

  @Test
  fun `findRejectedTransactionByTxHash returns null as timestamp exceeds queryable window`() {
    // insert a new rejected transaction with timestamp exceeds the 1-hour queryable window
    val rejectedTransaction = createRejectedTransaction(
      timestamp = fakeClock.now().minus(1.hours).minus(1.seconds)
    )
    performInsertTest(rejectedTransaction)

    // find the rejected transaction with the txHash
    val foundRejectedTransaction = rejectedTransactionsPostgresDao.findRejectedTransactionByTxHash(
      rejectedTransaction.transactionInfo!!.hash
    ).get()

    // assert that null is returned from the find method
    assertThat(foundRejectedTransaction).isNull()

    // assert that the total number of rows in the two tables are both one which
    // implies the rejected transaction is still present in db
    assertThat(rejectedTransactionsTotalRows()).isEqualTo(1)
    assertThat(fullTransactionsTotalRows()).isEqualTo(1)
  }

  @Test
  fun `deleteRejectedTransactionsAfterTimestamp returns 2 row deleted as timestamp exceeds storage window`() {
    // insert a new rejected transaction with timestamp within the 10-hours storage window
    performInsertTest(
      createRejectedTransaction(
        timestamp = fakeClock.now().minus(9.hours)
      )
    )

    // insert another rejected transaction with same txHash but different reason and
    // with timestamp just exceeds the 10-hours storage window
    performInsertTest(
      createRejectedTransaction(
        reasonMessage = "Transaction line count for module EXP=9000 is above the limit 8192",
        timestamp = fakeClock.now().minus(10.hours)
      )
    )

    // insert another rejected transaction with different txHash and reason and
    // with timestamp exceeds the 10-hours storage window
    performInsertTest(
      createRejectedTransaction(
        transactionInfo = TransactionInfo(
          hash = "0x078ecd6f00bff4beca9116ca85c65ddd265971e415d7df7a96b3c10424b031e2".decodeHex(),
          from = "0x4d144d7b9c96b26361d6ac74dd1d8267edca4fc2".decodeHex(),
          to = "0x1195cf65f83b3a5768f3c496d3a05ad6412c64b3".decodeHex(),
          nonce = 101UL
        ),
        reasonMessage = "Transaction line count for module EXP=10000 is above the limit 8192",
        timestamp = fakeClock.now().minus(11.hours)
      )
    )

    // assert that the total number of rows in the two tables are both three which
    // implies all the rejected transactions above are present in db
    assertThat(rejectedTransactionsTotalRows()).isEqualTo(3)
    assertThat(fullTransactionsTotalRows()).isEqualTo(3)

    // delete the rejected transactions with storage window as 10 hours from now
    val deletedRows = rejectedTransactionsPostgresDao.deleteRejectedTransactionsAfterTimestamp(
      fakeClock.now().minus(10.hours)
    ).get()

    // assert that number of total deleted rows is two
    assertThat(deletedRows).isEqualTo(2)

    // assert that the total number of rows in the two tables are both one which
    // implies the two rejected transactions with timestamp exceeds the storage
    // window were deleted
    assertThat(rejectedTransactionsTotalRows()).isEqualTo(1)
    assertThat(fullTransactionsTotalRows()).isEqualTo(1)
  }
}
