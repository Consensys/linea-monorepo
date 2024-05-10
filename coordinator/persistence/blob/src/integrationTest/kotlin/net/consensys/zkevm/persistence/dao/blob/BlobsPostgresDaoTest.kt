package net.consensys.zkevm.persistence.dao.blob

import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import io.vertx.sqlclient.PreparedQuery
import io.vertx.sqlclient.Row
import io.vertx.sqlclient.RowSet
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.FakeFixedClock
import net.consensys.linea.async.get
import net.consensys.linea.async.toSafeFuture
import net.consensys.setFirstByteToZero
import net.consensys.trimToMillisecondPrecision
import net.consensys.zkevm.coordinator.clients.BlobCompressionProof
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.BlobStatus
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.createBlobRecord
import net.consensys.zkevm.persistence.db.DbHelper
import net.consensys.zkevm.persistence.db.DuplicatedRecordException
import net.consensys.zkevm.persistence.test.CleanDbTestSuite
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.assertThrows
import org.junit.jupiter.api.extension.ExtendWith
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.ExecutionException
import kotlin.random.Random
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class BlobsPostgresDaoTest : CleanDbTestSuite() {
  override val databaseName = DbHelper.generateUniqueDbName("coordinator-tests-blobs-dao")
  private val maxBlobsToReturn = 6u
  private var timeToReturn: Instant = Clock.System.now()
  private fun blobsContentQuery(): PreparedQuery<RowSet<Row>> =
    sqlClient.preparedQuery("select * from ${BlobsPostgresDao.TableName}")

  private val fakeClock = FakeFixedClock()
  private lateinit var blobsPostgresDao: BlobsPostgresDao

  private val expectedStartBlock = 1UL
  private val expectedEndBlock = 100UL
  private val expectedStartBlockTime = fakeClock.now().trimToMillisecondPrecision()
  private val expectedEndBlockTime = fakeClock.now().plus(1200.seconds).trimToMillisecondPrecision()
  private val expectedShnarf = Random.nextBytes(32)

  @BeforeAll
  override fun beforeAll(vertx: Vertx) {
    super.beforeAll(vertx)
    blobsPostgresDao =
      BlobsPostgresDao(
        BlobsPostgresDao.Config(
          maxBlobsToReturn
        ),
        sqlClient,
        fakeClock
      )
  }

  @AfterEach
  override fun tearDown() {
    super.tearDown()
  }

  private fun performInsertTest(
    blobRecord: BlobRecord
  ): RowSet<Row>? {
    blobsPostgresDao.saveNewBlob(blobRecord).get()
    val dbContent = blobsContentQuery().execute().get()
    val newlyInsertedRow =
      dbContent.find { it.getLong("created_epoch_milli") == fakeClock.now().toEpochMilliseconds() }
    assertThat(newlyInsertedRow).isNotNull

    assertThat(newlyInsertedRow!!.getLong("start_block_number"))
      .isEqualTo(blobRecord.startBlockNumber.toLong())
    assertThat(newlyInsertedRow.getLong("end_block_number"))
      .isEqualTo(blobRecord.endBlockNumber.toLong())
    assertThat(newlyInsertedRow.getString("conflation_calculator_version"))
      .isEqualTo(blobRecord.conflationCalculatorVersion)
    assertThat(newlyInsertedRow.getInteger("status")).isEqualTo(BlobsPostgresDao.blobStatusToDbValue(blobRecord.status))

    return dbContent
  }

  @Test
  fun `saveNewBlob inserts new blob to db`() {
    val blobRecord1 = createBlobRecord(
      startBlockNumber = expectedStartBlock,
      endBlockNumber = expectedEndBlock,
      startBlockTime = expectedStartBlockTime,
      status = BlobStatus.COMPRESSION_PROVING
    )
    fakeClock.setTimeTo(Clock.System.now())

    val dbContent1 = performInsertTest(blobRecord1)
    assertThat(dbContent1).size().isEqualTo(1)

    val blobRecord2 = createBlobRecord(
      startBlockNumber = expectedEndBlock + 1UL,
      endBlockNumber = expectedEndBlock + 100UL,
      startBlockTime = expectedStartBlockTime,
      status = BlobStatus.COMPRESSION_PROVING
    )
    fakeClock.advanceBy(1.seconds)

    val dbContent2 = performInsertTest(blobRecord2)
    assertThat(dbContent2).size().isEqualTo(2)
  }

  @Test
  fun `saveNewBlob returns error when duplicated`() {
    val blobRecord1 = createBlobRecord(
      startBlockNumber = expectedStartBlock,
      endBlockNumber = expectedEndBlock,
      startBlockTime = expectedStartBlockTime,
      status = BlobStatus.COMPRESSION_PROVING
    )

    timeToReturn = Clock.System.now()

    val dbContent1 =
      performInsertTest(blobRecord1)
    assertThat(dbContent1).size().isEqualTo(1)

    assertThrows<ExecutionException> {
      blobsPostgresDao.saveNewBlob(blobRecord1).get()
    }.also { executionException ->
      assertThat(executionException.cause).isInstanceOf(DuplicatedRecordException::class.java)
      assertThat(executionException.cause!!.message)
        .isEqualTo(
          "Blob startBlockNumber=1, endBlockNumber=100, conflationCalculatorVersion=0.1.0 is already persisted!"
        )
    }
  }

  @Test
  fun `getConsecutiveBlobsFromBlockNumber works correctly for 1 blob`() {
    val expectedStartBlock1 = 1UL
    val expectedEndBlock1 = 90UL
    val expectedBlob = createBlobRecord(
      expectedStartBlock1,
      expectedEndBlock1,
      startBlockTime = expectedStartBlockTime
    )
    timeToReturn = Clock.System.now()

    blobsPostgresDao.saveNewBlob(expectedBlob).get()

    val actualBlobs =
      blobsPostgresDao.getConsecutiveBlobsFromBlockNumber(
        expectedStartBlock1,
        expectedEndBlockTime.plus(12.seconds)
      ).get()
    assertThat(actualBlobs).hasSameElementsAs(listOf(expectedBlob))
  }

  @Test
  fun `getConsecutiveBlobsFromBlockNumber filters blobs with lower versions and timestamp`() {
    val blobRecord1 = createBlobRecord(
      startBlockNumber = expectedStartBlock,
      endBlockNumber = expectedEndBlock,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord1WithOldVersion = createBlobRecord(
      startBlockNumber = expectedStartBlock,
      endBlockNumber = expectedEndBlock,
      conflationCalculationVersion = "0.0.1",
      startBlockTime = expectedStartBlockTime,
      blobHash = blobRecord1.blobHash
    )
    val blobRecord2 = createBlobRecord(
      startBlockNumber = expectedEndBlock + 1UL,
      endBlockNumber = expectedEndBlock + 100UL,
      startBlockTime = expectedStartBlockTime
    )
    timeToReturn = Clock.System.now()

    SafeFuture.collectAll(
      blobsPostgresDao.saveNewBlob(blobRecord1WithOldVersion),
      blobsPostgresDao.saveNewBlob(blobRecord1),
      blobsPostgresDao.saveNewBlob(blobRecord2)
    ).get()

    blobsPostgresDao.getConsecutiveBlobsFromBlockNumber(
      expectedStartBlock,
      blobRecord1.endBlockTime.plus(1.seconds)
    ).get().also { blobs ->
      assertThat(blobs).hasSameElementsAs(listOf(blobRecord1))
    }

    blobsPostgresDao.getConsecutiveBlobsFromBlockNumber(
      expectedStartBlock,
      blobRecord2.endBlockTime.plus(1.seconds)
    ).get().also { blobs ->
      assertThat(blobs).hasSameElementsAs(listOf(blobRecord1, blobRecord2))
    }
  }

  @Test
  fun `getConsecutiveBlobsFromBlockNumber returns empty list if no matched`() {
    val blobRecord1 = createBlobRecord(
      expectedStartBlock,
      expectedEndBlock,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord2 = createBlobRecord(
      startBlockNumber = expectedEndBlock + 1UL,
      endBlockNumber = expectedEndBlock + 100UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord3 = createBlobRecord(
      startBlockNumber = expectedEndBlock + 101UL,
      endBlockNumber = expectedEndBlock + 200UL,
      startBlockTime = expectedStartBlockTime
    )
    timeToReturn = Clock.System.now()

    SafeFuture.collectAll(
      blobsPostgresDao.saveNewBlob(blobRecord1),
      blobsPostgresDao.saveNewBlob(blobRecord2),
      blobsPostgresDao.saveNewBlob(blobRecord3)
    ).get()

    blobsPostgresDao.getConsecutiveBlobsFromBlockNumber(
      expectedStartBlock + 1UL,
      blobRecord3.endBlockTime.plus(1.seconds)
    ).get().also { blobs ->
      assertThat(blobs).isEmpty()
    }
  }

  @Test
  fun `getConsecutiveBlobsFromBlockNumber returns a sequence of blobs without gaps`() {
    val blobRecord1 = createBlobRecord(
      startBlockNumber = 1UL,
      endBlockNumber = 40UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord2 = createBlobRecord(
      startBlockNumber = 41UL,
      endBlockNumber = 60UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord3 = createBlobRecord(
      startBlockNumber = 61UL,
      endBlockNumber = 100UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord4 = createBlobRecord(
      startBlockNumber = 101UL,
      endBlockNumber = 111UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord5 = createBlobRecord(
      startBlockNumber = 112UL,
      endBlockNumber = 132UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord6 = createBlobRecord(
      startBlockNumber = 134UL,
      endBlockNumber = 156UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord7 = createBlobRecord(
      startBlockNumber = 157UL,
      endBlockNumber = 189UL,
      startBlockTime = expectedStartBlockTime
    )
    val expectedBlobs = listOf(
      blobRecord3,
      blobRecord4,
      blobRecord5
    )
    val otherBlobs = listOf(
      blobRecord1,
      blobRecord2,
      blobRecord6,
      blobRecord7
    )

    timeToReturn = Clock.System.now()
    expectedBlobs.forEach { blobsPostgresDao.saveNewBlob(it).get() }
    otherBlobs.forEach { blobsPostgresDao.saveNewBlob(it).get() }

    val actualBlobs =
      blobsPostgresDao
        .getConsecutiveBlobsFromBlockNumber(
          expectedBlobs.first().startBlockNumber,
          otherBlobs.last().endBlockTime
        ).get()
    assertThat(actualBlobs).hasSameElementsAs(expectedBlobs)
  }

  @Test
  fun `getConsecutiveBlobsFromBlockNumber returns a sequence of consecutive blobs with priority on versions`() {
    val blobRecord1 = createBlobRecord(
      startBlockNumber = 1UL,
      endBlockNumber = 40UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord2 = createBlobRecord(
      startBlockNumber = 41UL,
      endBlockNumber = 60UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord3 = createBlobRecord(
      61UL,
      100UL,
      startBlockTime = expectedStartBlockTime
    )
    // This will not be accepted as it has older version
    val blobRecord4 = createBlobRecord(
      startBlockNumber = 101UL,
      endBlockNumber = 111UL,
      conflationCalculationVersion = "0.0.1",
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord5 = createBlobRecord(
      startBlockNumber = 101UL,
      endBlockNumber = 109UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord6 = createBlobRecord(
      startBlockNumber = 110UL,
      endBlockNumber = 122UL,
      startBlockTime = expectedStartBlockTime
    )
    // This will break the consecutive search
    val blobRecord7 = createBlobRecord(
      startBlockNumber = 112UL,
      endBlockNumber = 132UL,
      conflationCalculationVersion = "0.0.1",
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord8 = createBlobRecord(
      startBlockNumber = 123UL,
      endBlockNumber = 132UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord9 = createBlobRecord(
      startBlockNumber = 133UL,
      endBlockNumber = 156UL,
      startBlockTime = expectedStartBlockTime
    )
    val expectedBlobs = listOf(
      blobRecord1,
      blobRecord2,
      blobRecord3,
      blobRecord5,
      blobRecord6
    )
    val olderVersionBlobs = listOf(
      blobRecord4,
      blobRecord7,
      blobRecord8,
      blobRecord9
    )

    timeToReturn = Clock.System.now()
    expectedBlobs.forEach { blobsPostgresDao.saveNewBlob(it).get() }
    olderVersionBlobs.forEach { blobsPostgresDao.saveNewBlob(it).get() }

    val actualBlobs =
      blobsPostgresDao
        .getConsecutiveBlobsFromBlockNumber(
          expectedBlobs.first().startBlockNumber,
          blobRecord9.endBlockTime.plus(1.seconds)
        ).get()
    assertThat(actualBlobs).hasSameElementsAs(expectedBlobs)
  }

  @Test
  fun `findBlobByXBlockNumber works correctly for 1 blob`() {
    val expectedBlob = createBlobRecord(
      startBlockNumber = 1UL,
      endBlockNumber = 90UL,
      startBlockTime = expectedStartBlockTime
    )

    blobsPostgresDao.saveNewBlob(expectedBlob).get()

    assertThat(blobsPostgresDao.findBlobByEndBlockNumber(90UL))
      .succeedsWithin(1.seconds.toJavaDuration())
      .isEqualTo(expectedBlob)

    assertThat(blobsPostgresDao.findBlobByEndBlockNumber(91UL))
      .succeedsWithin(1.seconds.toJavaDuration())
      .isNull()

    assertThat(blobsPostgresDao.findBlobByStartBlockNumber(1UL))
      .succeedsWithin(1.seconds.toJavaDuration())
      .isEqualTo(expectedBlob)

    assertThat(blobsPostgresDao.findBlobByStartBlockNumber(2UL))
      .succeedsWithin(1.seconds.toJavaDuration())
      .isNull()
  }

  @Test
  fun `findBlobByBlockNumber returs blob record with higher versions`() {
    val blobRecord1 = createBlobRecord(
      startBlockNumber = 1UL,
      endBlockNumber = 100UL,
      conflationCalculationVersion = "9.3.1"
    )

    val blobRecord1WithOldVersion = createBlobRecord(
      startBlockNumber = 1UL,
      endBlockNumber = 100UL,
      conflationCalculationVersion = "9.2.6"
    )

    SafeFuture.collectAll(
      blobsPostgresDao.saveNewBlob(blobRecord1WithOldVersion),
      blobsPostgresDao.saveNewBlob(blobRecord1)
    ).get()

    assertThat(blobsPostgresDao.findBlobByStartBlockNumber(1UL))
      .succeedsWithin(1.seconds.toJavaDuration())
      .isEqualTo(blobRecord1)

    assertThat(blobsPostgresDao.findBlobByEndBlockNumber(100UL))
      .succeedsWithin(1.seconds.toJavaDuration())
      .isEqualTo(blobRecord1)
  }

  @Test
  fun `updateBlobAsProven updates the target record correctly`() {
    val blobRecord1 = createBlobRecord(
      startBlockNumber = 1UL,
      endBlockNumber = 40UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord2 = createBlobRecord(
      startBlockNumber = 41UL,
      endBlockNumber = 60UL,
      blobHash = ByteArray(32),
      status = BlobStatus.COMPRESSION_PROVING,
      shnarf = expectedShnarf,
      blobCompressionProof = null,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord3 = createBlobRecord(
      startBlockNumber = 61UL,
      endBlockNumber = 100UL,
      startBlockTime = expectedStartBlockTime
    )
    val expectedBlobHash = Random.nextBytes(32).setFirstByteToZero()
    val expectedCompressionProof = BlobCompressionProof(
      compressedData = Random.nextBytes(32).setFirstByteToZero(),
      conflationOrder = BlockIntervals(41U, listOf(60U)),
      prevShnarf = Random.nextBytes(32),
      parentStateRootHash = Random.nextBytes(32).setFirstByteToZero(),
      finalStateRootHash = Random.nextBytes(32).setFirstByteToZero(),
      parentDataHash = Random.nextBytes(32).setFirstByteToZero(),
      dataHash = expectedBlobHash,
      snarkHash = Random.nextBytes(32),
      expectedX = Random.nextBytes(32),
      expectedY = Random.nextBytes(32),
      expectedShnarf = expectedShnarf,
      decompressionProof = Random.nextBytes(512),
      proverVersion = "mock-0.0.0",
      verifierID = 6789,
      eip4844Enabled = false,
      commitment = ByteArray(0),
      kzgProofContract = ByteArray(0),
      kzgProofSidecar = ByteArray(0)
    )

    val blobRecord2Expected = createBlobRecord(
      41UL,
      60UL,
      status = BlobStatus.COMPRESSION_PROVEN,
      shnarf = expectedShnarf,
      startBlockTime = expectedStartBlockTime,
      blobHash = expectedBlobHash,
      blobCompressionProof = expectedCompressionProof
    )
    val insertedBlobs = listOf(
      blobRecord1,
      blobRecord2,
      blobRecord3
    )
    val expectedBlobs = listOf(
      blobRecord1,
      blobRecord2Expected,
      blobRecord3
    )
    timeToReturn = Clock.System.now()
    insertedBlobs.forEach { blobsPostgresDao.saveNewBlob(it).get() }
    blobsPostgresDao.updateBlob(
      blobRecord2.startBlockNumber,
      blobRecord2.endBlockNumber,
      blobRecord2.conflationCalculatorVersion,
      BlobStatus.COMPRESSION_PROVEN,
      blobRecord2Expected.blobCompressionProof!!
    ).get()

    val actualBlobs =
      blobsPostgresDao
        .getConsecutiveBlobsFromBlockNumber(
          expectedBlobs.first().startBlockNumber,
          blobRecord3.endBlockTime.plus(1.seconds)
        ).get()
    assertThat(actualBlobs).hasSameElementsAs(expectedBlobs)
  }

  @Test
  fun `deleteBlobsUpToEndBlockNumber deletes the target record correctly`() {
    val blobRecord1 = createBlobRecord(
      startBlockNumber = 1UL,
      endBlockNumber = 40UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord2 = createBlobRecord(
      startBlockNumber = 41UL,
      endBlockNumber = 60UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord3 = createBlobRecord(
      startBlockNumber = 61UL,
      endBlockNumber = 100UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord4 = createBlobRecord(
      startBlockNumber = 101UL,
      endBlockNumber = 111UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord5 = createBlobRecord(
      startBlockNumber = 112UL,
      endBlockNumber = 132UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord6 = createBlobRecord(
      startBlockNumber = 133UL,
      endBlockNumber = 156UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord7 = createBlobRecord(
      startBlockNumber = 157UL,
      endBlockNumber = 189UL,
      startBlockTime = expectedStartBlockTime
    )
    val expectedBlobs = listOf(
      blobRecord4,
      blobRecord5,
      blobRecord6,
      blobRecord7
    )
    val deletedBlobs = listOf(
      blobRecord1,
      blobRecord2,
      blobRecord3
    )

    timeToReturn = Clock.System.now()
    expectedBlobs.forEach { blobsPostgresDao.saveNewBlob(it).get() }
    deletedBlobs.forEach { blobsPostgresDao.saveNewBlob(it).get() }

    blobsPostgresDao.deleteBlobsUpToEndBlockNumber(
      blobRecord3.endBlockNumber
    ).get()

    val existedBlobRecords = blobsContentQuery().execute()
      .toSafeFuture()
      .thenApply { rowSet ->
        rowSet.map(BlobsPostgresDao::parseRecord)
      }.get()

    assertThat(existedBlobRecords).hasSameElementsAs(expectedBlobs)
  }

  @Test
  fun `deleteBlobsAfterBlockNumber deletes the target record correctly`() {
    val blobRecord1 = createBlobRecord(
      startBlockNumber = 1UL,
      endBlockNumber = 40UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord2 = createBlobRecord(
      startBlockNumber = 41UL,
      endBlockNumber = 60UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord3 = createBlobRecord(
      startBlockNumber = 61UL,
      endBlockNumber = 100UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord4 = createBlobRecord(
      startBlockNumber = 101UL,
      endBlockNumber = 111UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord5 = createBlobRecord(
      startBlockNumber = 112UL,
      endBlockNumber = 132UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord6 = createBlobRecord(
      startBlockNumber = 133UL,
      endBlockNumber = 156UL,
      startBlockTime = expectedStartBlockTime
    )
    val blobRecord7 = createBlobRecord(
      startBlockNumber = 157UL,
      endBlockNumber = 189UL,
      startBlockTime = expectedStartBlockTime,
      status = BlobStatus.COMPRESSION_PROVING
    )
    val deletedBlobs = listOf(
      blobRecord4,
      blobRecord5,
      blobRecord6,
      blobRecord7
    )
    val expectedBlobs = listOf(
      blobRecord1,
      blobRecord2,
      blobRecord3
    )

    timeToReturn = Clock.System.now()
    expectedBlobs.forEach { blobsPostgresDao.saveNewBlob(it).get() }
    deletedBlobs.forEach { blobsPostgresDao.saveNewBlob(it).get() }

    blobsPostgresDao.deleteBlobsAfterBlockNumber(
      blobRecord3.endBlockNumber
    ).get()

    val existedBlobRecords = blobsContentQuery().execute()
      .toSafeFuture()
      .thenApply { rowSet ->
        rowSet.map(BlobsPostgresDao::parseRecord)
      }.get()

    assertThat(existedBlobRecords).hasSameElementsAs(expectedBlobs)
  }
}
