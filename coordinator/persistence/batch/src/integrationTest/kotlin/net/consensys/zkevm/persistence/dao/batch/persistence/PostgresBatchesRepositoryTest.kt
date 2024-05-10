package net.consensys.zkevm.persistence.dao.batch.persistence

import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import io.vertx.sqlclient.Row
import io.vertx.sqlclient.RowSet
import kotlinx.datetime.Instant
import net.consensys.FakeFixedClock
import net.consensys.linea.async.get
import net.consensys.zkevm.coordinator.clients.prover.ProverResponsesRepository
import net.consensys.zkevm.domain.Batch
import net.consensys.zkevm.persistence.db.DbHelper
import net.consensys.zkevm.persistence.test.CleanDbTestSuite
import net.consensys.zkevm.persistence.test.DbQueries
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.assertThrows
import org.junit.jupiter.api.extension.ExtendWith
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.ExecutionException
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class PostgresBatchesRepositoryTest : CleanDbTestSuite() {
  private var fakeClockTime = Instant.parse("2023-12-11T00:00:00.000Z")
  override val databaseName = DbHelper.generateUniqueDbName("coordinator-tests-batches")

  private var fakeClock = FakeFixedClock(fakeClockTime)
  protected lateinit var postgresBatchesRepository: PostgresBatchesRepository

  private lateinit var batchesPostgresDao: BatchesDao

  @BeforeAll
  override fun beforeAll(vertx: Vertx) {
    super.beforeAll(vertx)
    batchesPostgresDao = BatchesPostgresDao(
      sqlClient,
      fakeClock
    )
    postgresBatchesRepository =
      PostgresBatchesRepository(
        batchesPostgresDao
      )
  }

  @Test
  fun saveNewBatch_savesTheBatch_to_db() {
    val expectedConflationVersion1 = "0.0.1"
    val expectedStartBlock1 = 1UL
    val expectedEndBlock1 = 1UL
    val batch1 =
      Batch(expectedStartBlock1, expectedEndBlock1, Batch.Status.Finalized, expectedConflationVersion1)

    val dbContent1 =
      performInsertTest(batch1)
    assertThat(dbContent1).size().isEqualTo(1)

    val expectedConflationVersion2 = "0.0.2"
    val expectedStartBlock2 = 6UL
    val expectedEndBlock2 = 9UL
    val batch2 = Batch(
      expectedStartBlock2,
      expectedEndBlock2,
      Batch.Status.Proven,
      expectedConflationVersion2
    )
    fakeClock.advanceBy(1.seconds)

    val dbContent2 =
      performInsertTest(batch2)
    assertThat(dbContent2).size().isEqualTo(2)
  }

  private fun performInsertTest(
    batch: Batch
  ): RowSet<Row>? {
    batchesPostgresDao.saveNewBatch(batch).get()
    val dbContent = DbQueries.getTableContent(sqlClient, DbQueries.batchesTable).execute().get()
    val newlyInsertedRow =
      dbContent.find { it.getLong("created_epoch_milli") == fakeClock.now().toEpochMilliseconds() }
    assertThat(newlyInsertedRow).isNotNull

    assertThat(newlyInsertedRow!!.getLong("start_block_number"))
      .isEqualTo(batch.startBlockNumber.toLong())
    assertThat(newlyInsertedRow.getLong("end_block_number")).isEqualTo(batch.endBlockNumber.toLong())
    assertThat(newlyInsertedRow.getString("prover_version"))
      .isEqualTo(BatchesPostgresDao.DEFAULT_VERSION)
    assertThat(newlyInsertedRow.getInteger("status")).isEqualTo(batchStatusToDbValue(batch.status))
    return dbContent
  }

  @Test
  fun saveNewBatch_returnsErrorWhenDuplicated() {
    val expectedConflationVersion = "0.0.1"
    val expectedStartBlock1 = 1UL
    val expectedEndBlock1 = 1UL
    val batch1 =
      Batch(
        startBlockNumber = expectedStartBlock1,
        endBlockNumber = expectedEndBlock1,
        status = Batch.Status.Finalized,
        conflationVersion = expectedConflationVersion
      )

    val dbContent1 =
      performInsertTest(batch1)
    assertThat(dbContent1).size().isEqualTo(1)

    assertThrows<ExecutionException> {
      batchesPostgresDao.saveNewBatch(batch1).get()
    }.also { executionException ->
      assertThat(executionException.cause).isInstanceOf(DuplicatedBatchException::class.java)
      assertThat(executionException.cause!!.message)
        .isEqualTo(
          "Batch startBlockNumber=1, endBlockNumber=1," +
            " proverVersion=${BatchesPostgresDao.DEFAULT_VERSION}," +
            " conflationVersion=0.0.1 is already persisted!"
        )
    }
  }

  @Test
  fun `findHighestConsecutiveEndBlockNumberFromBlockNumber when empty returns null`() {
    assertThat(
      batchesPostgresDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(1L)
        .get()
    ).isEqualTo(null)

    assertThat(
      batchesPostgresDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(1L)
        .get()
    ).isEqualTo(null)
  }

  @Test
  fun `findHighestConsecutiveEndBlockNumberFromBlockNumber ignores records before from block number`() {
    val expectedConflationVersion = "0.1.0"
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Proven, expectedConflationVersion),
        // Gap, query does not care about gaps
        createBatch(20, 30, Batch.Status.Proven, expectedConflationVersion),
        // Gap, query does not care about gaps
        createBatch(31, 32, Batch.Status.Proven, expectedConflationVersion),
        // Gap, query does not care about gaps
        createBatch(40, 42, Batch.Status.Proven, expectedConflationVersion)
      )

    SafeFuture.collectAll(batches.map { batchesPostgresDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesPostgresDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(20L)
        .get()
    )
      .isEqualTo(batches[2].endBlockNumber.toLong())
  }

  @Test
  fun `findHighestConsecutiveBatchByStatus returns highest batch without gaps`() {
    val expectedConflationVersion = "0.1.0"
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Proven, expectedConflationVersion),
        createBatch(4, 5, Batch.Status.Proven, expectedConflationVersion),
        createBatch(6, 7, Batch.Status.Proven, expectedConflationVersion),
        createBatch(8, 10, Batch.Status.Proven, expectedConflationVersion)
      )

    SafeFuture.collectAll(batches.map { batchesPostgresDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesPostgresDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(1L)
        .get()
    )
      .isEqualTo(batches[3].endBlockNumber.toLong())
  }

  @Test
  fun `findHighestConsecutiveEndBlockNumberFromBlockNumber returns highest batch without gaps - only one pending`() {
    val expectedConflationVersion = "0.1.0"
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized, expectedConflationVersion),
        createBatch(4, 5, Batch.Status.Proven, expectedConflationVersion)
      )

    SafeFuture.collectAll(batches.map { batchesPostgresDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesPostgresDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(1L)
        .get()
    )
      .isEqualTo(batches[1].endBlockNumber.toLong())
  }

  @Test
  fun `findHighestConsecutiveEndBlockNumberFromBlockNumber returns ranks records correctly`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized, "0.1.0"),
        createBatch(4, 5, Batch.Status.Proven, "0.1.0"),
        createBatch(4, 5, Batch.Status.Proven, "0.1.1"),
        // expected at index 3
        createBatch(6, 7, Batch.Status.Proven, "0.1.4"),
        createBatch(6, 7, Batch.Status.Proven, "0.1.2"),
        createBatch(6, 7, Batch.Status.Proven, "0.1.3"),
        // Gap in the DB
        createBatch(9, 9, Batch.Status.Proven, "0.1.0"),
        createBatch(10, 11, Batch.Status.Proven, "0.1.0")
      )

    SafeFuture.collectAll(batches.map { batchesPostgresDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesPostgresDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(1L)
        .get()
    )
      .isEqualTo(batches[3].endBlockNumber.toLong())
  }

  @Test
  fun `findHighestConsecutiveEndBlockNumberFromBlockNumber returns null when no relevant batches are found`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized, "0.1.0"),
        createBatch(4, 5, Batch.Status.Finalized, "0.1.0"),
        createBatch(4, 5, Batch.Status.Finalized, "0.1.1")
      )

    SafeFuture.collectAll(batches.map { batchesPostgresDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesPostgresDao
        .findHighestConsecutiveEndBlockNumberFromBlockNumber(5L)
        .get()
    )
      .isNull()
  }

  @Test
  fun `setStatus updates all records`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized, "0.1.0"),
        createBatch(4, 5, Batch.Status.Proven, "0.1.0"),
        createBatch(4, 5, Batch.Status.Proven, "0.1.1"),
        createBatch(6, 7, Batch.Status.Proven, "0.1.3"),
        createBatch(10, 11, Batch.Status.Proven, "0.1.0")
      )

    SafeFuture.collectAll(batches.map { batchesPostgresDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesPostgresDao
        .setBatchStatusUpToEndBlockNumber(
          endBlockNumberInclusive = 7L,
          currentStatus = Batch.Status.Proven,
          newStatus = Batch.Status.Finalized
        )
        .get()
    )
      .isEqualTo(3)
  }

  @Test
  fun `deleteBatchesUpToEndBlockNumber deletes all target records`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized, "0.1.0"),
        createBatch(4, 5, Batch.Status.Proven, "0.1.0"),
        createBatch(4, 5, Batch.Status.Proven, "0.1.1"),
        createBatch(6, 7, Batch.Status.Proven, "0.1.3"),
        createBatch(10, 11, Batch.Status.Proven, "0.1.0")
      )

    SafeFuture.collectAll(batches.map { batchesPostgresDao.saveNewBatch(it) }.stream()).get()

    val expectedRows = DbQueries.getTableContent(sqlClient, DbQueries.blobsTable).execute().get().filter {
      it.getLong("end_block_number") > 6L
    }

    assertThat(
      batchesPostgresDao
        .deleteBatchesUpToEndBlockNumber(
          endBlockNumberInclusive = 6L
        )
        .get()
    )
      .isEqualTo(3)

    val dbContent = DbQueries.getTableContent(sqlClient, DbQueries.blobsTable).execute().get()
    assertThat(dbContent.size()).isEqualTo(expectedRows.size)
    assertThat(dbContent.map(::getProverResponseIndex))
      .hasSameElementsAs(expectedRows.map(::getProverResponseIndex))
  }

  @Test
  fun `deleteBatchesUpToEndBlockNumber deletes none of the records`() {
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized, "0.1.0"),
        createBatch(4, 5, Batch.Status.Proven, "0.1.0"),
        createBatch(6, 7, Batch.Status.Proven, "0.1.0"),
        createBatch(8, 9, Batch.Status.Proven, "0.1.0"),
        createBatch(10, 11, Batch.Status.Proven, "0.1.0")
      )

    SafeFuture.collectAll(batches.map { batchesPostgresDao.saveNewBatch(it) }.stream()).get()
    assertThat(
      batchesPostgresDao
        .deleteBatchesUpToEndBlockNumber(
          endBlockNumberInclusive = 1L
        )
        .get()
    )
      .isEqualTo(0)

    val remainingBatches = DbQueries.getBatches(sqlClient)

    assertThat(remainingBatches).hasSameElementsAs(batches)
  }

  private fun getProverResponseIndex(
    record: Row
  ): ProverResponsesRepository.ProverResponseIndex {
    return ProverResponsesRepository.ProverResponseIndex(
      record.getLong("start_block_number").toULong(),
      record.getLong("end_block_number").toULong(),
      record.getString("prover_version")
    )
  }

  @Test
  fun `saveNewBatch with null conflation version`() {
    val nullConflationVersion: String? = null

    var batch1 = createBatch(
      startBlockNumber = 1L,
      endBlockNumber = 3L
    )
    batch1 = batch1.copy(conflationVersion = nullConflationVersion)

    SafeFuture.collectAll(
      postgresBatchesRepository.saveNewBatch(batch1)
    ).get()

    DbQueries.getBatches(sqlClient).also { batches ->
      assertThat(batches).hasSameElementsAs(listOf(batch1.copy(conflationVersion = "1.0.0")))
    }
  }

  private fun createBatch(
    startBlockNumber: Long,
    endBlockNumber: Long,
    status: Batch.Status = Batch.Status.Proven,
    conflationVersion: String = BatchesPostgresDao.DEFAULT_VERSION
  ): Batch {
    return Batch(
      startBlockNumber = startBlockNumber.toULong(),
      endBlockNumber = endBlockNumber.toULong(),
      status = status,
      conflationVersion = conflationVersion
    )
  }
}
