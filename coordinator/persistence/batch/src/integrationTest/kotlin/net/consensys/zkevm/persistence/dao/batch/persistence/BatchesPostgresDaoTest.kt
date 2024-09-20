package net.consensys.zkevm.persistence.dao.batch.persistence

import io.vertx.junit5.VertxExtension
import io.vertx.sqlclient.PreparedQuery
import io.vertx.sqlclient.Row
import io.vertx.sqlclient.RowSet
import kotlinx.datetime.Instant
import net.consensys.FakeFixedClock
import net.consensys.linea.async.get
import net.consensys.zkevm.domain.Batch
import net.consensys.zkevm.domain.createBatch
import net.consensys.zkevm.persistence.dao.batch.persistence.BatchesDao.Companion.batchesDaoTableName
import net.consensys.zkevm.persistence.db.DbHelper
import net.consensys.zkevm.persistence.db.DuplicatedRecordException
import net.consensys.zkevm.persistence.test.CleanDbTestSuiteParallel
import net.consensys.zkevm.persistence.test.DbQueries
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.junit.jupiter.api.extension.ExtendWith
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.ExecutionException
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
class BatchesPostgresDaoTest : CleanDbTestSuiteParallel() {
  override val databaseName = DbHelper.generateUniqueDbName("coordinator-tests-batches")
  private var fakeClockTime = Instant.parse("2023-12-11T00:00:00.000Z")
  private var fakeClock = FakeFixedClock(fakeClockTime)
  private fun batchesContentQuery(): PreparedQuery<RowSet<Row>> =
    sqlClient.preparedQuery("select * from $batchesDaoTableName")

  private lateinit var batchesDao: BatchesDao

  @BeforeEach
  fun beforeEach() {
    batchesDao =
      BatchesPostgresDao(
        sqlClient,
        fakeClock
      )
  }

  @AfterEach
  override fun tearDown() {
    super.tearDown()
  }

  @Test
  fun saveNewBatch_savesTheBatch_to_db() {
    val expectedStartBlock1 = 1UL
    val expectedEndBlock1 = 1UL
    val batch1 =
      Batch(
        expectedStartBlock1,
        expectedEndBlock1,
        Batch.Status.Finalized
      )

    val dbContent1 =
      performInsertTest(batch1)
    assertThat(dbContent1).size().isEqualTo(1)

    val expectedStartBlock2 = 6UL
    val expectedEndBlock2 = 9UL
    val batch2 = Batch(
      expectedStartBlock2,
      expectedEndBlock2,
      Batch.Status.Proven
    )
    fakeClock.advanceBy(1.seconds)

    val dbContent2 =
      performInsertTest(batch2)
    assertThat(dbContent2).size().isEqualTo(2)
  }

  private fun performInsertTest(
    batch: Batch
  ): RowSet<Row>? {
    batchesDao.saveNewBatch(batch).get()
    val dbContent = batchesContentQuery().execute().get()
    val newlyInsertedRow =
      dbContent.find { it.getLong("created_epoch_milli") == fakeClock.now().toEpochMilliseconds() }
    assertThat(newlyInsertedRow).isNotNull

    assertThat(newlyInsertedRow!!.getLong("start_block_number"))
      .isEqualTo(batch.startBlockNumber.toLong())
    assertThat(newlyInsertedRow.getLong("end_block_number")).isEqualTo(batch.endBlockNumber.toLong())
    assertThat(newlyInsertedRow.getInteger("status")).isEqualTo(batchStatusToDbValue(batch.status))
    return dbContent
  }

  @Test
  fun saveNewBatch_returnsErrorWhenDuplicated() {
    val expectedStartBlock1 = 1UL
    val expectedEndBlock1 = 1UL
    val batch1 =
      Batch(
        expectedStartBlock1,
        expectedEndBlock1,
        Batch.Status.Finalized
      )

    val dbContent1 =
      performInsertTest(batch1)
    assertThat(dbContent1).size().isEqualTo(1)

    assertThrows<ExecutionException> {
      batchesDao.saveNewBatch(batch1).get()
    }.also { executionException ->
      assertThat(executionException.cause).isInstanceOf(DuplicatedRecordException::class.java)
      assertThat(executionException.cause!!.message)
        .isEqualTo(
          "Batch startBlockNumber=1, endBlockNumber=1 is already persisted!"
        )
    }
  }

  @Test
  fun `findHighestConsecutiveEndBlockNumberFromBlockNumber when empty returns null`() {
    assertThat(
      batchesDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(1L)
        .get()
    ).isEqualTo(null)

    assertThat(
      batchesDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(1L)
        .get()
    ).isEqualTo(null)
  }

  @Test
  fun `findHighestConsecutiveEndBlockNumberFromBlockNumber ignores records before from block number`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Proven),
        // Gap, query does not care about gaps
        createBatch(20, 30, Batch.Status.Proven),
        // Gap, query does not care about gaps
        createBatch(31, 32, Batch.Status.Proven),
        // Gap, query does not care about gaps
        createBatch(40, 42, Batch.Status.Proven)
      )

    SafeFuture.collectAll(batches.map { batchesDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(20L)
        .get()
    )
      .isEqualTo(batches[2].endBlockNumber.toLong())
  }

  @Test
  fun `findHighestConsecutiveBatchByStatus returns highest batch without gaps`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Proven),
        createBatch(4, 5, Batch.Status.Proven),
        createBatch(6, 7, Batch.Status.Proven),
        createBatch(8, 10, Batch.Status.Proven)
      )

    SafeFuture.collectAll(batches.map { batchesDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(1L)
        .get()
    )
      .isEqualTo(batches[3].endBlockNumber.toLong())
  }

  @Test
  fun `findHighestConsecutiveEndBlockNumberFromBlockNumber returns highest batch without gaps - only one pending`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized),
        createBatch(4, 5, Batch.Status.Proven)
      )

    SafeFuture.collectAll(batches.map { batchesDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(1L)
        .get()
    )
      .isEqualTo(batches[1].endBlockNumber.toLong())
  }

  @Test
  fun `findHighestConsecutiveEndBlockNumberFromBlockNumber returns null when no relevant batches are found`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized),
        createBatch(4, 5, Batch.Status.Finalized)
      )

    SafeFuture.collectAll(batches.map { batchesDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(5L)
        .get()
    )
      .isNull()
  }

  @Test
  fun `setStatus updates all records`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized),
        createBatch(4, 5, Batch.Status.Proven),
        createBatch(6, 7, Batch.Status.Proven),
        createBatch(10, 11, Batch.Status.Proven)
      )

    SafeFuture.collectAll(batches.map { batchesDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesDao
        .setBatchStatusUpToEndBlockNumber(
          endBlockNumberInclusive = 7L,
          currentStatus = Batch.Status.Proven,
          newStatus = Batch.Status.Finalized
        )
        .get()
    )
      .isEqualTo(2)
  }

  @Test
  fun `deleteBatchesUpToEndBlockNumber deletes all target records`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized),
        createBatch(4, 5, Batch.Status.Proven),
        createBatch(6, 7, Batch.Status.Proven),
        createBatch(10, 11, Batch.Status.Proven)
      )

    SafeFuture.collectAll(batches.map { batchesDao.saveNewBatch(it) }.stream()).get()

    val expectedRows = batchesContentQuery().execute().get().filter {
      it.getLong("end_block_number") > 6L
    }

    assertThat(
      batchesDao
        .deleteBatchesUpToEndBlockNumber(
          endBlockNumberInclusive = 6L
        )
        .get()
    )
      .isEqualTo(2)

    val dbContent = batchesContentQuery().execute().get()
    assertThat(dbContent.size()).isEqualTo(expectedRows.size)
  }

  @Test
  fun `deleteBatchesUpToEndBlockNumber deletes none of the records`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized),
        createBatch(4, 5, Batch.Status.Proven),
        createBatch(6, 7, Batch.Status.Proven),
        createBatch(8, 9, Batch.Status.Proven),
        createBatch(10, 11, Batch.Status.Proven)
      )

    SafeFuture.collectAll(batches.map { batchesDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesDao
        .deleteBatchesUpToEndBlockNumber(
          endBlockNumberInclusive = 1L
        )
        .get()
    )
      .isEqualTo(0)

    val remainingBatches = DbQueries.getBatches(sqlClient)
    assertThat(remainingBatches).hasSameElementsAs(batches)
  }

  @Test
  fun `deleteBatchesAfterBlockNumber deletes all target records`() {
    // The finalized status is just for test purposes this won't be the case in a real network
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized),
        createBatch(4, 5, Batch.Status.Proven),
        createBatch(6, 7, Batch.Status.Finalized),
        createBatch(10, 11, Batch.Status.Proven)
      )

    SafeFuture.collectAll(batches.map { batchesDao.saveNewBatch(it) }.stream()).get()

    val expectedRows = batchesContentQuery().execute().get().filter {
      it.getLong("end_block_number") < 6L
    }.sortedBy { it.getLong("start_block_number") }.toList()

    assertThat(
      batchesDao
        .deleteBatchesAfterBlockNumber(
          startingBlockNumberInclusive = 6L
        )
        .get()
    )
      .isEqualTo(2)

    val dbContent = batchesContentQuery().execute().get().sortedBy { it.getLong("start_block_number") }.toList()
    assertThat(dbContent.size).isEqualTo(expectedRows.size)
    dbContent.forEachIndexed { index, row ->
      assertThat(row.toJson()).isEqualTo(expectedRows[index].toJson())
    }
  }

  @Test
  fun `deleteBatchesAfterBlockNumber deletes none of the records`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized),
        createBatch(4, 5, Batch.Status.Finalized),
        createBatch(6, 7, Batch.Status.Proven),
        createBatch(8, 9, Batch.Status.Proven),
        createBatch(10, 11, Batch.Status.Proven)
      )

    SafeFuture.collectAll(batches.map { batchesDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesDao
        .deleteBatchesAfterBlockNumber(
          startingBlockNumberInclusive = 11L
        )
        .get()
    )
      .isEqualTo(0)

    val remainingBatches = DbQueries.getBatches(sqlClient)
    assertThat(remainingBatches).hasSameElementsAs(batches)
  }
}
