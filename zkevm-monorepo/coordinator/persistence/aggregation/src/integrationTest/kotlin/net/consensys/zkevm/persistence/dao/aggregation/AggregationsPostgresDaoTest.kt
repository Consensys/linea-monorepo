package net.consensys.zkevm.persistence.dao.aggregation

import io.vertx.core.Future
import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import io.vertx.sqlclient.Row
import io.vertx.sqlclient.RowSet
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.linea.async.get
import net.consensys.trimToSecondPrecision
import net.consensys.zkevm.coordinator.clients.prover.serialization.ProofToFinalizeJsonResponse
import net.consensys.zkevm.domain.Aggregation
import net.consensys.zkevm.domain.BlobAndBatchCounters
import net.consensys.zkevm.domain.BlobCounters
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.BlobStatus
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.ExecutionProofVersions
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.VersionedExecutionProofs
import net.consensys.zkevm.domain.createAggregation
import net.consensys.zkevm.domain.createProofToFinalize
import net.consensys.zkevm.persistence.dao.blob.BlobsPostgresDao
import net.consensys.zkevm.persistence.db.DbHelper
import net.consensys.zkevm.persistence.db.DuplicatedRecordException
import net.consensys.zkevm.persistence.test.CleanDbTestSuite
import net.consensys.zkevm.persistence.test.DbQueries
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterAll
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.assertThrows
import org.junit.jupiter.api.extension.ExtendWith
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.ExecutionException
import java.util.concurrent.TimeUnit

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class AggregationsPostgresDaoTest : CleanDbTestSuite() {
  override val databaseName = DbHelper.generateUniqueDbName("coordinator-tests-aggregations-dao")

  private val BATCH_PENDING_STATUS = 5U
  private val BATCH_PROVEN_STATUS = 2U
  private var fakeClockCounter: Long = System.currentTimeMillis()

  data class Batch(
    val createdEpochMilli: Long,
    val startBlockNumber: ULong,
    val endBlockNumber: ULong,
    val proverVersion: String,
    val conflationCalculatorVersion: String,
    val status: UInt
  )

  private val maxBlobReturnLimit = 10u
  private var timeToReturn: Instant = Instant.DISTANT_PAST
  private val sampleResponse = ProofToFinalize(
    aggregatedProof = "mock_aggregatedProof".toByteArray(),
    aggregatedVerifierIndex = 1,
    aggregatedProofPublicInput = "mock_aggregatedProofPublicInput".toByteArray(),
    dataHashes = listOf("mock_dataHashes_1".toByteArray()),
    dataParentHash = "mock_dataParentHash".toByteArray(),
    parentStateRootHash = "mock_parentStateRootHash".toByteArray(),
    parentAggregationLastBlockTimestamp = Clock.System.now().trimToSecondPrecision(),
    finalTimestamp = Clock.System.now().trimToSecondPrecision(),
    firstBlockNumber = 1,
    finalBlockNumber = 23,
    l1RollingHash = "mock_l1RollingHash".toByteArray(),
    l1RollingHashMessageNumber = 4,
    l2MerkleRoots = listOf("mock_l2MerkleRoots".toByteArray()),
    l2MerkleTreesDepth = 5,
    l2MessagingBlocksOffsets = "mock_l2MessagingBlocksOffsets".toByteArray()
  )

  private val fixedClock =
    object : Clock {
      override fun now(): Instant {
        return timeToReturn
      }
    }

  private lateinit var aggregationsPostgresDaoImpl: PostgresAggregationsDao
  private lateinit var blobsPostgresDaoImpl: BlobsPostgresDao

  @BeforeEach
  fun beforeEach(vertx: Vertx) {
    super.setUp(vertx)
  }

  @BeforeAll
  override fun beforeAll(vertx: Vertx) {
    super.beforeAll(vertx)
    aggregationsPostgresDaoImpl = PostgresAggregationsDao(sqlClient, fixedClock)
    blobsPostgresDaoImpl = BlobsPostgresDao(BlobsPostgresDao.Config(maxBlobReturnLimit), sqlClient)
  }

  @AfterEach
  override fun tearDown() {
    super.tearDown()
  }

  @AfterAll
  override fun afterAll() {
    super.afterAll()
  }

  @Test
  fun saveNewAggregation_saves_aggregation_to_db() {
    val aggregation1 = Aggregation(
      startBlockNumber = 1UL,
      endBlockNumber = 10UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "a0.0.1",
      batchCount = 5UL,
      aggregationProof = sampleResponse
    )

    val dbContent1 = performInsertTest(aggregation1)
    assertThat(dbContent1).size().isEqualTo(1)

    val aggregation2 = Aggregation(
      startBlockNumber = 11UL,
      endBlockNumber = 12UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "a0.0.2",
      batchCount = 51UL,
      aggregationProof = sampleResponse
    )

    val dbContent2 = performInsertTest(aggregation2)
    assertThat(dbContent2).size().isEqualTo(2)
  }

  @Test
  fun saveNewAggregation_returns_error_when_duplicated() {
    val aggregation = Aggregation(
      startBlockNumber = 1UL,
      endBlockNumber = 10UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "a0.0.1",
      batchCount = 5UL,
      aggregationProof = sampleResponse
    )

    val dbContent1 = performInsertTest(aggregation)
    assertThat(dbContent1).size().isEqualTo(1)

    assertThrows<ExecutionException> {
      aggregationsPostgresDaoImpl.saveNewAggregation(aggregation).get()
    }.also { executionException ->
      assertThat(executionException.cause).isInstanceOf(DuplicatedRecordException::class.java)
      assertThat(executionException.cause!!.message)
        .isEqualTo(
          "Aggregation startBlockNumber=${aggregation.startBlockNumber}, " +
            "endBlockNumber=${aggregation.endBlockNumber}, " +
            "aggregationCalculatorVersion=${aggregation.aggregationCalculatorVersion} " +
            "batchCount=${aggregation.batchCount} is already persisted!"
        )
    }
  }

  @Test
  fun `updateAggregationAsProven updates proven records`() {
    val aggregation = Aggregation(
      startBlockNumber = 1UL,
      endBlockNumber = 10UL,
      status = Aggregation.Status.Proving,
      aggregationCalculatorVersion = "a0.0.1",
      batchCount = 5UL,
      aggregationProof = null
    )

    val dbContent1 = performInsertTest(aggregation)
    assertThat(dbContent1).size().isEqualTo(1)

    val updatedAggregation = Aggregation(
      startBlockNumber = aggregation.startBlockNumber,
      endBlockNumber = aggregation.endBlockNumber,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = aggregation.aggregationCalculatorVersion,
      batchCount = aggregation.batchCount,
      aggregationProof = sampleResponse
    )

    val dbContent2 = performUpdateTest(updatedAggregation)
    assertThat(dbContent2).size().isEqualTo(1)
  }

  @Test
  fun `updateAggregationAsProven doesn't update proven records`() {
    val aggregation = Aggregation(
      startBlockNumber = 1UL,
      endBlockNumber = 10UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "a0.0.1",
      batchCount = 5UL,
      aggregationProof = sampleResponse
    )
    val rowsUpdated = aggregationsPostgresDaoImpl.updateAggregationAsProven(aggregation).get()
    assertThat(rowsUpdated).isZero()
  }

  private fun buildBatch(
    startB: ULong,
    endB: ULong,
    status: UInt = BATCH_PROVEN_STATUS,
    conflationCalculatorVersion: String = "c1.0.0"
  ): Batch {
    fakeClockCounter++
    return Batch(
      createdEpochMilli = fakeClockCounter,
      startBlockNumber = startB,
      endBlockNumber = endB,
      proverVersion = "p0.0.1",
      conflationCalculatorVersion = conflationCalculatorVersion,
      status = status
    )
  }

  data class BlobAndBatches(
    val blob: BlobRecord,
    val batches: List<Batch>
  )

  private fun buildBlob(
    batches: List<Batch>,
    blobStatus: BlobStatus = BlobStatus.COMPRESSION_PROVEN,
    conflationCalculatorVersion: String = "c1.0.0"
  ): BlobAndBatches {
    val startBatch = batches.first()
    val endBatch = batches.last()
    val blob = BlobRecord(
      startBlockNumber = startBatch.startBlockNumber,
      endBlockNumber = endBatch.endBlockNumber,
      conflationCalculatorVersion = conflationCalculatorVersion,
      blobHash = ByteArray(0),
      startBlockTime = Instant.fromEpochMilliseconds(startBatch.createdEpochMilli),
      endBlockTime = Instant.fromEpochMilliseconds(endBatch.createdEpochMilli),
      batchesCount = batches.size.toUInt(),
      status = blobStatus,
      expectedShnarf = ByteArray(0),
      blobCompressionProof = null
    )
    return BlobAndBatches(blob, batches)
  }

  private fun BlobAndBatches.toBlobAndBatchCounters(): BlobAndBatchCounters {
    val blobCounters = BlobCounters(
      numberOfBatches = blob.batchesCount,
      startBlockNumber = blob.startBlockNumber,
      endBlockNumber = blob.endBlockNumber,
      startBlockTimestamp = blob.startBlockTime,
      endBlockTimestamp = blob.endBlockTime
    )
    return BlobAndBatchCounters(
      blobCounters = blobCounters,
      versionedExecutionProofs = VersionedExecutionProofs(
        executionProofs = BlockIntervals(
          startingBlockNumber = blob.startBlockNumber,
          upperBoundaries = this.batches.map { it.endBlockNumber }
        ),
        executionVersion = this.batches.map {
          ExecutionProofVersions(
            conflationCalculatorVersion = it.conflationCalculatorVersion,
            executionProverVersion = it.proverVersion
          )
        }
      )
    )
  }

  private fun List<BlobAndBatches>.toBlobAndBatchCounters(): List<BlobAndBatchCounters> {
    return this.map { it.toBlobAndBatchCounters() }
  }

  @Test
  fun getCandidatesForAggregation_from_block_number_mismatch() {
    val batch1 = buildBatch(1UL, 10UL)
    val batch2 = buildBatch(11UL, 20UL)
    val batch3 = buildBatch(21UL, 30UL)
    val batch4 = buildBatch(31UL, 40UL)
    val batches = listOf(batch1, batch2, batch3, batch4)

    val blob1 = buildBlob(listOf(batch1, batch2))
    val blob2 = buildBlob(listOf(batch3, batch4))

    insertDbRecords(batches, listOf(blob1, blob2).map { it.blob })

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(fromBlockNumber = 0).get()
      .also { blobCounters ->
        assertThat(blobCounters).isEmpty()
      }
    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(fromBlockNumber = 1).get()
      .also { blobCounters ->
        assertThat(blobCounters).hasSameElementsAs(
          listOf(blob1, blob2).toBlobAndBatchCounters()
        )
      }
  }

  @Test
  fun getCandidatesForAggregation_unproven_batch_in_the_middle() {
    val batch1 = buildBatch(1UL, 10UL)
    val batch2 = buildBatch(11UL, 20UL)
    val batch3 = buildBatch(21UL, 30UL, status = BATCH_PENDING_STATUS)
    val batch4 = buildBatch(31UL, 40UL)
    val batches = listOf(batch1, batch2, batch3, batch4)

    val blob1 = buildBlob(listOf(batch1, batch2))
    val blob2 = buildBlob(listOf(batch3, batch4))

    insertDbRecords(batches, listOf(blob1, blob2).map { it.blob })

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(fromBlockNumber = 1).get()
      .also { blobCounters ->
        assertThat(blobCounters).hasSameElementsAs(listOf(blob1).toBlobAndBatchCounters())
      }
  }

  @Test
  fun getCandidatesForAggregation_unproven_blob_in_the_middle() {
    val batch1 = buildBatch(1UL, 10UL)
    val batch2 = buildBatch(11UL, 20UL)
    val batch3 = buildBatch(21UL, 30UL)
    val batch4 = buildBatch(31UL, 40UL)
    val batch5 = buildBatch(41UL, 50UL)
    val batch6 = buildBatch(51UL, 60UL)
    val batches = listOf(batch1, batch2, batch3, batch4, batch5)

    val blob1 = buildBlob(listOf(batch1, batch2))
    val blob2 = buildBlob(listOf(batch3, batch4))
    val blob3 = buildBlob(listOf(batch5), BlobStatus.COMPRESSION_PROVING)
    val blob4 = buildBlob(listOf(batch6))
    val blobs = listOf(blob1, blob2, blob3, blob4)

    insertDbRecords(batches, blobs.map { it.blob })

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(fromBlockNumber = 1).get()
      .also { blobCounters ->
        assertThat(blobCounters).hasSameElementsAs(listOf(blob1, blob2).toBlobAndBatchCounters())
      }
  }

  @Test
  fun getCandidatesForAggregation_duplicated_blobs_with_different_versions() {
    // v1
    // blob1, {batch1 (1-10)}
    // blob2  { batch2 (11-20), batch3 (21-30) } -- error cannot prove blob2 V1
    // blob3  { batch4 (31-40)}
    // blob4  { batch5 (41-50)}
    // blob compressor version is updated, yields different batches/blobs
    // v2
    // blob2 { batch2 (11-20), batch3 (21-25) }
    // blob3 { batch4 (26-35), batch5 (36-45)}
    // blob4 { batch6 (46-50) }

    val batch1v1 = buildBatch(1UL, 10UL)
    val batch2v1 = buildBatch(11UL, 20UL)
    val batch3v1 = buildBatch(21UL, 30UL)
    val batch4v1 = buildBatch(31UL, 40UL)
    val batch5v1 = buildBatch(41UL, 50UL)
    val blob1v1 = buildBlob(listOf(batch1v1))
    val blob2v1 = buildBlob(listOf(batch2v1, batch3v1))
    val blob3v1 = buildBlob(listOf(batch4v1))
    val blob4v1 = buildBlob(listOf(batch5v1))

    val conflationCalculatorVersion2 = "c2.0.0"
    val batch2v2 = buildBatch(11UL, 20UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch3v2 = buildBatch(21UL, 25UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch4v2 = buildBatch(26UL, 35UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch5v2 = buildBatch(36UL, 45UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch6v2 = buildBatch(46UL, 50UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val blob2v2 =
      buildBlob(
        listOf(batch2v2, batch3v2),
        conflationCalculatorVersion = conflationCalculatorVersion2
      )
    val blob3v2 =
      buildBlob(
        listOf(batch4v2, batch5v2),
        conflationCalculatorVersion = conflationCalculatorVersion2
      )
    val blob4v2 = buildBlob(
      listOf(batch6v2),
      conflationCalculatorVersion = conflationCalculatorVersion2
    )

    val batches =
      listOf(batch1v1, batch2v1, batch3v1, batch4v1, batch5v1, batch2v2, batch3v2, batch4v2, batch5v2, batch6v2)
    val blobs = listOf(blob1v1, blob2v1, blob3v1, blob4v1, blob2v2, blob3v2, blob4v2)

    insertDbRecords(batches, blobs.map { it.blob })

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(1).get().also { blobCounters ->
      assertThat(blobCounters).hasSameElementsAs(listOf(blob1v1, blob2v2, blob3v2).toBlobAndBatchCounters())
    }
  }

  @Test
  fun getCandidatesForAggregation_duplicated_blobs_with_different_versions_unproven_blob_in_the_middle() {
    // v1
    // blob1, {batch1 (1-10)}
    // blob2  { batch2 (11-20), batch3 (21-30) } -- error cannot prove blob2 V1
    // blob3  { batch4 (31-40)}
    // blob4  { batch5 (41-50)}
    // blob compressor version is updated, yields different batches/blobs
    // v2
    // blob2 { batch2 (11-20), batch3 (21-25) }
    // blob3 { batch4 (26-35), batch5 (36-45)} UNPROVEN
    // blob4 { batch6 (46-50) }

    val batch1v1 = buildBatch(1UL, 10UL)
    val batch2v1 = buildBatch(11UL, 20UL)
    val batch3v1 = buildBatch(21UL, 30UL)
    val batch4v1 = buildBatch(31UL, 40UL)
    val batch5v1 = buildBatch(41UL, 50UL)
    val blob1v1 = buildBlob(listOf(batch1v1))
    val blob2v1 = buildBlob(listOf(batch2v1, batch3v1))
    val blob3v1 = buildBlob(listOf(batch4v1))
    val blob4v1 = buildBlob(listOf(batch5v1))

    val conflationCalculatorVersion2 = "c2.0.0"
    val batch2v2 = buildBatch(11UL, 20UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch3v2 = buildBatch(21UL, 25UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch4v2 = buildBatch(26UL, 35UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch5v2 = buildBatch(36UL, 45UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch6v2 = buildBatch(46UL, 50UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val blob2v2 =
      buildBlob(
        listOf(batch2v2, batch3v2),
        conflationCalculatorVersion = conflationCalculatorVersion2
      )
    val blob3v2 =
      buildBlob(
        listOf(batch4v2, batch5v2),
        blobStatus = BlobStatus.COMPRESSION_PROVING,
        conflationCalculatorVersion = conflationCalculatorVersion2
      )
    val blob4v2 = buildBlob(
      listOf(batch6v2),
      conflationCalculatorVersion = conflationCalculatorVersion2
    )

    val batches =
      listOf(batch1v1, batch2v1, batch3v1, batch4v1, batch5v1, batch2v2, batch3v2, batch4v2, batch5v2, batch6v2)
    val blobs = listOf(blob1v1, blob2v1, blob3v1, blob4v1, blob2v2, blob3v2, blob4v2)

    insertDbRecords(batches, blobs.map { it.blob })

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(1).get().also { blobCounters ->
      assertThat(blobCounters).hasSameElementsAs(listOf(blob1v1, blob2v2).toBlobAndBatchCounters())
    }
  }

  @Test
  fun getCandidatesForAggregation_duplicated_blobs_with_different_versions_unproven_batch_in_the_middle() {
    // v1
    // blob1, {batch1 (1-10)}
    // blob2  { batch2 (11-20), batch3 (21-30) } -- error cannot prove blob2 V1
    // blob3  { batch4 (31-40)}
    // blob4  { batch5 (41-50)}
    // blob compressor version is updated, yields different batches/blobs
    // v2
    // blob2 { batch2 (11-20), batch3 (21-25) }
    // blob3 { batch4 (26-35), batch5 (36-45)} UNPROVEN
    // blob4 { batch6 (46-50) }

    // [blob1 {batch1}, blob2 {batch2, batch3}, blob3 {batch4, batch5}, blob4 {batch6}]
    // [blob1 {1,[10]}, v1; blob2 {11[20,25] v2}, blob3 26 [35, 45] v2, blob4 46 [50] v2]
    val batch1v1 = buildBatch(1UL, 10UL)
    val batch2v1 = buildBatch(11UL, 20UL)
    val batch3v1 = buildBatch(21UL, 30UL)
    val batch4v1 = buildBatch(31UL, 40UL)
    val batch5v1 = buildBatch(41UL, 50UL)
    val blob1v1 = buildBlob(listOf(batch1v1))
    val blob2v1 = buildBlob(listOf(batch2v1, batch3v1))
    val blob3v1 = buildBlob(listOf(batch4v1))
    val blob4v1 = buildBlob(listOf(batch5v1))

    val conflationCalculatorVersion2 = "c2.0.0"
    val batch2v2 = buildBatch(11UL, 20UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch3v2 = buildBatch(21UL, 25UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch4v2 = buildBatch(26UL, 35UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch5v2 =
      buildBatch(36UL, 40UL, conflationCalculatorVersion = conflationCalculatorVersion2, status = BATCH_PENDING_STATUS)
    val batch5_1v2 = buildBatch(41UL, 45UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch6v2 = buildBatch(46UL, 50UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val blob2v2 =
      buildBlob(
        listOf(batch2v2, batch3v2),
        conflationCalculatorVersion = conflationCalculatorVersion2
      )
    val blob3v2 = buildBlob(
      batches = listOf(batch4v2, batch5v2, batch5_1v2),
      conflationCalculatorVersion = conflationCalculatorVersion2
    )
    val blob4v2 = buildBlob(
      listOf(batch6v2),
      conflationCalculatorVersion = conflationCalculatorVersion2
    )

    val batches =
      listOf(
        batch1v1, batch2v1, batch3v1, batch4v1, batch5v1, batch2v2, batch3v2, batch4v2, batch5v2, batch5_1v2, batch6v2
      )
    val blobs = listOf(blob1v1, blob2v1, blob3v1, blob4v1, blob2v2, blob3v2, blob4v2)

    insertDbRecords(batches, blobs.map { it.blob })

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(1).get()
      .also { blobCounters ->
        assertThat(blobCounters).hasSameElementsAs(listOf(blob1v1, blob2v2).toBlobAndBatchCounters())
      }
  }

  @Test
  fun getCandidatesForAggregation_duplicated_blobs_with_different_batch_conflation_versions() {
    // conflation v1
    // blob1, {batch1 (1-10)}
    // blob2  { batch2 (11-20), batch3 (21-30) }
    // blob3  { batch4 (31-40)}
    // blob4  { batch5 (41-50)}
    // conflation v2
    // blob2 { batch2 (11-20), batch3 (21-25) }
    // blob3 { batch4 (26-35), batch5 (36-45)}
    // blob4 { batch6 (46-50) }

    val batch1v1 = buildBatch(1UL, 10UL)
    val batch2v1 = buildBatch(11UL, 20UL)
    val batch3v1 = buildBatch(21UL, 30UL)
    val batch4v1 = buildBatch(31UL, 40UL)
    val batch5v1 = buildBatch(41UL, 50UL)
    val blob1v1 = buildBlob(listOf(batch1v1))
    val blob2v1 = buildBlob(listOf(batch2v1, batch3v1))
    val blob3v1 = buildBlob(listOf(batch4v1))
    val blob4v1 = buildBlob(listOf(batch5v1))

    val conflationCalculatorVersion2 = "c2.0.0"
    val batch2v2 = buildBatch(11UL, 20UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch3v2 = buildBatch(21UL, 25UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch4v2 = buildBatch(26UL, 35UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch5v2 =
      buildBatch(36UL, 40UL, conflationCalculatorVersion = conflationCalculatorVersion2, status = BATCH_PENDING_STATUS)
    val batch5_1v2 = buildBatch(41UL, 45UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val batch6v2 = buildBatch(46UL, 50UL, conflationCalculatorVersion = conflationCalculatorVersion2)
    val blob2v2 = buildBlob(listOf(batch2v2, batch3v2), conflationCalculatorVersion = conflationCalculatorVersion2)
    val blob3v2 =
      buildBlob(listOf(batch4v2, batch5v2, batch5_1v2), conflationCalculatorVersion = conflationCalculatorVersion2)
    val blob4v2 = buildBlob(listOf(batch6v2), conflationCalculatorVersion = conflationCalculatorVersion2)

    val batches =
      listOf(
        batch1v1, batch2v1, batch3v1, batch4v1, batch5v1, batch2v2, batch3v2, batch4v2, batch5v2, batch5_1v2, batch6v2
      )
    val blobs = listOf(blob1v1, blob2v1, blob3v1, blob4v1, blob2v2, blob3v2, blob4v2)

    insertDbRecords(batches, blobs.map { it.blob })

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(1).get().also { blobCounters ->
      assertThat(blobCounters).hasSameElementsAs(listOf(blob1v1, blob2v2).toBlobAndBatchCounters())
    }

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(11).get().also { blobCounters ->
      assertThat(blobCounters).hasSameElementsAs(listOf(blob2v2).toBlobAndBatchCounters())
    }
  }

  @Test
  fun getCandidatesForAggregation_first_blob_unproven_shall_return_no_results() {
    // conflation v1
    // blob1, {batch1 (4-10)} -- UNPROVEN
    // blob2  { batch2 (11-20), batch3 (21-30) }

    val batch1v1 = buildBatch(4UL, 10UL, status = BATCH_PENDING_STATUS)
    val batch2v1 = buildBatch(11UL, 20UL)
    val batch3v1 = buildBatch(21UL, 30UL)
    val blob1v1 = buildBlob(listOf(batch1v1))
    val blob2v1 = buildBlob(listOf(batch2v1, batch3v1))

    val batches = listOf(batch1v1, batch2v1, batch3v1)
    val blobs = listOf(blob1v1, blob2v1)

    insertDbRecords(batches, blobs.map { it.blob })

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(4).get().also { blobCounters ->
      assertThat(blobCounters).hasSameElementsAs(emptyList())
    }
  }

  @Test
  fun getProofsToFinalize_parsesAndReturnsProofToFinalize() {
    val aggregationProof1 = createProofToFinalize(
      firstBlockNumber = 0,
      finalBlockNumber = 10,
      parentAggregationLastBlockTimestamp = Instant.fromEpochSeconds(0),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:00:00Z")
    )

    val aggregation1 = Aggregation(
      startBlockNumber = 1UL,
      endBlockNumber = 10UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof1
    )
    performInsertTest(aggregation1)

    val aggregationProof2 = createProofToFinalize(
      firstBlockNumber = 11,
      finalBlockNumber = 20,
      parentAggregationLastBlockTimestamp = Instant.parse("2024-04-28T15:00:00Z"),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:00:00Z")
    )

    val aggregation2 = Aggregation(
      startBlockNumber = 11UL,
      endBlockNumber = 20UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof2
    )
    performInsertTest(aggregation2)

    aggregationsPostgresDaoImpl.getProofsToFinalize(
      1L,
      Aggregation.Status.Proven,
      aggregationProof1.finalTimestamp,
      1
    ).get().also { aggregation ->
      assertThat(aggregation).isNotNull
      assertThat(aggregation).isEqualTo(listOf(aggregationProof1))
    }
    aggregationsPostgresDaoImpl.getProofsToFinalize(
      11L,
      Aggregation.Status.Proven,
      aggregationProof2.finalTimestamp,
      1
    ).get().also { aggregation ->
      assertThat(aggregation).isNotNull
      assertThat(aggregation).isEqualTo(listOf(aggregationProof2))
    }
    aggregationsPostgresDaoImpl.getProofsToFinalize(
      1L,
      Aggregation.Status.Proven,
      aggregationProof1.finalTimestamp,
      2
    ).get().also { aggregation ->
      assertThat(aggregation).isNotNull
      assertThat(aggregation).isEqualTo(listOf(aggregationProof1, aggregationProof2))
    }
  }

  @Test
  fun getProofsToFinalize_selectsLatestAggregationCalculatorVersion() {
    val aggregationProof1 = createProofToFinalize(
      firstBlockNumber = 0,
      finalBlockNumber = 10,
      parentAggregationLastBlockTimestamp = Instant.fromEpochSeconds(0),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:00:00Z")
    )

    val aggregation1 = Aggregation(
      startBlockNumber = 1UL,
      endBlockNumber = 10UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof1
    )
    performInsertTest(aggregation1)

    val aggregationProof2 = createProofToFinalize(
      firstBlockNumber = 11,
      finalBlockNumber = 15,
      parentAggregationLastBlockTimestamp = Instant.fromEpochSeconds(0),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:00:00Z")
    )

    val aggregation2 = Aggregation(
      startBlockNumber = 1UL,
      endBlockNumber = 15UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.1.1",
      batchCount = 7UL,
      aggregationProof = aggregationProof2
    )
    performInsertTest(aggregation2)

    aggregationsPostgresDaoImpl.getProofsToFinalize(
      1L,
      Aggregation.Status.Proven,
      aggregationProof1.finalTimestamp,
      2
    ).get().also { aggregation ->
      assertThat(aggregation).isNotNull
      assertThat(aggregation).isEqualTo(listOf(aggregationProof2))
    }
  }

  @Test
  fun getProofsToFinalize_consecutiveProofsOnly() {
    val aggregationProof1 = createProofToFinalize(
      firstBlockNumber = 0,
      finalBlockNumber = 10,
      parentAggregationLastBlockTimestamp = Instant.fromEpochSeconds(0),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:00:00Z")
    )

    val aggregation1 = Aggregation(
      startBlockNumber = 1UL,
      endBlockNumber = 10UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof1
    )
    performInsertTest(aggregation1)

    val aggregationProof2 = createProofToFinalize(
      firstBlockNumber = 11,
      finalBlockNumber = 20,
      parentAggregationLastBlockTimestamp = Instant.parse("2024-04-28T15:00:00Z"),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:01:00Z")
    )

    val aggregation2 = Aggregation(
      startBlockNumber = 11UL,
      endBlockNumber = 20UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof2
    )
    performInsertTest(aggregation2)

    val aggregationProof3 = createProofToFinalize(
      firstBlockNumber = 21,
      finalBlockNumber = 39,
      parentAggregationLastBlockTimestamp = Instant.parse("2024-04-28T15:01:00Z"),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:02:00Z")
    )

    val aggregation3 = Aggregation(
      startBlockNumber = 31UL,
      endBlockNumber = 39UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof3
    )
    performInsertTest(aggregation3)

    val aggregationProof4 = createProofToFinalize(
      firstBlockNumber = 40,
      finalBlockNumber = 50,
      parentAggregationLastBlockTimestamp = Instant.parse("2024-04-28T15:01:00Z"),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:02:00Z")
    )

    val aggregation4 = Aggregation(
      startBlockNumber = 40UL,
      endBlockNumber = 50UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof4
    )
    performInsertTest(aggregation4)

    aggregationsPostgresDaoImpl.getProofsToFinalize(
      1L,
      Aggregation.Status.Proven,
      aggregationProof4.finalTimestamp,
      5
    ).get().also { aggregation ->
      assertThat(aggregation).isNotNull
      assertThat(aggregation).isEqualTo(listOf(aggregationProof1, aggregationProof2))
    }

    aggregationsPostgresDaoImpl.getProofsToFinalize(
      31L,
      Aggregation.Status.Proven,
      aggregationProof4.finalTimestamp,
      5
    ).get().also { aggregation ->
      assertThat(aggregation).isNotNull
      assertThat(aggregation).isEqualTo(listOf(aggregationProof3, aggregationProof4))
    }
  }

  @Test
  fun `findConsecutiveProvenBlobs takes smaller blobs all else being equal`() {
    val batch1 = buildBatch(0UL, 3UL)
    val batch2 = buildBatch(5UL, 5UL)
    val batch4 = buildBatch(4UL, 5UL)
    val batch3 = buildBatch(4UL, 4UL)
    val batch5 = buildBatch(6UL, 6UL)
    val batch6 = buildBatch(7UL, 7UL)
    val batches = listOf(batch1, batch2, batch3, batch4, batch5, batch6)

    val blob = buildBlob(listOf(batch1, batch3, batch2, batch5, batch6))

    insertDbRecords(batches, listOf(blob).map { it.blob })

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(fromBlockNumber = 0).get()
      .also { blobCounters ->
        assertThat(blobCounters).hasSameElementsAs(
          listOf(blob).toBlobAndBatchCounters()
        )
      }
  }

  @Test
  fun `getProofsToFinalize gets nothing if first aggregation is missing`() {
    val aggregationProof1 = createProofToFinalize(
      firstBlockNumber = 11,
      finalBlockNumber = 20,
      parentAggregationLastBlockTimestamp = Instant.parse("2024-04-28T15:00:00Z"),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:01:00Z")
    )

    val aggregation1 = Aggregation(
      startBlockNumber = 11UL,
      endBlockNumber = 20UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof1
    )
    performInsertTest(aggregation1)

    val aggregationProof2 = createProofToFinalize(
      firstBlockNumber = 21,
      finalBlockNumber = 30,
      parentAggregationLastBlockTimestamp = Instant.parse("2024-04-28T15:01:00Z"),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:02:00Z")
    )

    val aggregation2 = Aggregation(
      startBlockNumber = 21UL,
      endBlockNumber = 30UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof2
    )
    performInsertTest(aggregation2)

    aggregationsPostgresDaoImpl.getProofsToFinalize(
      1L,
      Aggregation.Status.Proven,
      aggregationProof2.finalTimestamp,
      3
    ).get().also { aggregation ->
      assertThat(aggregation).isNotNull
      assertThat(aggregation).isEqualTo(emptyList<ProofToFinalize>())
    }
  }

  @Test
  fun `getProofsToFinalize gets nothing is first aggregation is still proving`() {
    val aggregationProof1 = createProofToFinalize(
      firstBlockNumber = 1,
      finalBlockNumber = 10,
      parentAggregationLastBlockTimestamp = Instant.fromEpochSeconds(0),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:00:00Z")
    )

    val aggregation1 = Aggregation(
      startBlockNumber = 1UL,
      endBlockNumber = 10UL,
      status = Aggregation.Status.Proving,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof1
    )
    performInsertTest(aggregation1)

    val aggregationProof2 = createProofToFinalize(
      firstBlockNumber = 11,
      finalBlockNumber = 20,
      parentAggregationLastBlockTimestamp = Instant.parse("2024-04-28T15:00:00Z"),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:01:00Z")
    )

    val aggregation2 = Aggregation(
      startBlockNumber = 11UL,
      endBlockNumber = 20UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof2
    )
    performInsertTest(aggregation2)

    val aggregationProof3 = createProofToFinalize(
      firstBlockNumber = 21,
      finalBlockNumber = 30,
      parentAggregationLastBlockTimestamp = Instant.parse("2024-04-28T15:01:00Z"),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:02:00Z")
    )

    val aggregation3 = Aggregation(
      startBlockNumber = 21UL,
      endBlockNumber = 30UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof3
    )
    performInsertTest(aggregation3)

    aggregationsPostgresDaoImpl.getProofsToFinalize(
      1L,
      Aggregation.Status.Proven,
      aggregationProof3.finalTimestamp,
      3
    ).get().also { aggregation ->
      assertThat(aggregation).isNotNull
      assertThat(aggregation).isEqualTo(emptyList<ProofToFinalize>())
    }
  }

  @Test
  fun deleteAggregationsUpToEndBlockNumber_deletes_aggregations_till_block_number() {
    val aggregationProof1 = createProofToFinalize(
      firstBlockNumber = 1,
      finalBlockNumber = 10,
      parentAggregationLastBlockTimestamp = Instant.fromEpochSeconds(0L),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:00:00Z")
    )

    val aggregation1 = Aggregation(
      startBlockNumber = 1UL,
      endBlockNumber = 10UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof1
    )
    performInsertTest(aggregation1)

    val aggregationProof2 = createProofToFinalize(
      firstBlockNumber = 11,
      finalBlockNumber = 20,
      parentAggregationLastBlockTimestamp = Instant.parse("2024-04-28T15:00:00Z"),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:01:00Z")
    )

    val aggregation2 = Aggregation(
      startBlockNumber = 11UL,
      endBlockNumber = 20UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof2
    )
    performInsertTest(aggregation2)

    val aggregationProof3 = createProofToFinalize(
      firstBlockNumber = 21,
      finalBlockNumber = 30,
      parentAggregationLastBlockTimestamp = Instant.parse("2024-04-28T15:01:00Z"),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:02:00Z")
    )

    val aggregation3 = Aggregation(
      startBlockNumber = 21UL,
      endBlockNumber = 30UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof3
    )
    performInsertTest(aggregation3)

    val deletedAggregationsCount = aggregationsPostgresDaoImpl
      .deleteAggregationsUpToEndBlockNumber(aggregation2.endBlockNumber.toLong())
      .get()
    assertThat(deletedAggregationsCount).isEqualTo(2)
    val dbContent = DbQueries.getTableContent(sqlClient, DbQueries.aggregationsTable).execute().get()
    assertThat(dbContent.rowCount()).isEqualTo(1)
    val remainingAggregation = dbContent.first()!!
    assertThat(remainingAggregation.getLong("end_block_number")).isEqualTo(aggregation3.endBlockNumber.toLong())
  }

  @Test
  fun deleteAggregationsAfterBlockNumber_deletes_aggregations_after_block_number() {
    val aggregationProof1 = createProofToFinalize(
      firstBlockNumber = 1,
      finalBlockNumber = 10,
      parentAggregationLastBlockTimestamp = Instant.fromEpochSeconds(0L),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:00:00Z")
    )

    val aggregation1 = Aggregation(
      startBlockNumber = 1UL,
      endBlockNumber = 10UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof1
    )
    performInsertTest(aggregation1)

    val aggregationProof2 = createProofToFinalize(
      firstBlockNumber = 11,
      finalBlockNumber = 20,
      parentAggregationLastBlockTimestamp = Instant.parse("2024-04-28T15:00:00Z"),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:01:00Z")
    )

    val aggregation2 = Aggregation(
      startBlockNumber = 11UL,
      endBlockNumber = 20UL,
      status = Aggregation.Status.Proven,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof2
    )
    performInsertTest(aggregation2)

    val aggregationProof3 = createProofToFinalize(
      firstBlockNumber = 21,
      finalBlockNumber = 30,
      parentAggregationLastBlockTimestamp = Instant.parse("2024-04-28T15:01:00Z"),
      startBlockTime = Instant.fromEpochSeconds(0),
      finalTimestamp = Instant.parse("2024-04-28T15:02:00Z")
    )

    val aggregation3 = Aggregation(
      startBlockNumber = 21UL,
      endBlockNumber = 30UL,
      status = Aggregation.Status.Proving,
      aggregationCalculatorVersion = "0.0.1",
      batchCount = 5UL,
      aggregationProof = aggregationProof3
    )
    performInsertTest(aggregation3)

    val deletedAggregationsCount = aggregationsPostgresDaoImpl
      .deleteAggregationsAfterBlockNumber(aggregation2.startBlockNumber.toLong())
      .get()
    assertThat(deletedAggregationsCount).isEqualTo(2)
    val dbContent = DbQueries.getTableContent(sqlClient, DbQueries.aggregationsTable).execute().get()
    assertThat(dbContent.rowCount()).isEqualTo(1)
    val remainingAggregation = dbContent.first()!!
    assertThat(remainingAggregation.getLong("end_block_number")).isEqualTo(aggregation1.endBlockNumber.toLong())
  }

  @Test
  fun `findAggregationProofByEndBlockNumber when multiple aggregations found throws error`() {
    val aggregation1 = createAggregation(
      startBlockNumber = 1,
      endBlockNumber = 10
    )
    val aggregation2 = createAggregation(
      startBlockNumber = 5,
      endBlockNumber = 10
    )
    SafeFuture.collectAll(
      aggregationsPostgresDaoImpl.saveNewAggregation(aggregation1),
      aggregationsPostgresDaoImpl.saveNewAggregation(aggregation2)
    ).get()

    assertThat(aggregationsPostgresDaoImpl.findAggregationProofByEndBlockNumber(10))
      .failsWithin(1, TimeUnit.SECONDS)
      .withThrowableThat()
      .withCauseInstanceOf(IllegalStateException::class.java)
      .withMessageContaining("Multiple aggregations found for endBlockNumber=10")
  }

  @Test
  fun `findAggregationProofByEndBlockNumber when single aggregation found returns it`() {
    val aggregation1 = createAggregation(
      startBlockNumber = 1,
      endBlockNumber = 4
    )
    val aggregation2 = createAggregation(
      endBlockNumber = 10,
      parentAggregation = aggregation1
    )
    SafeFuture.collectAll(
      aggregationsPostgresDaoImpl.saveNewAggregation(aggregation1),
      aggregationsPostgresDaoImpl.saveNewAggregation(aggregation2)
    ).get()

    assertThat(aggregationsPostgresDaoImpl.findAggregationProofByEndBlockNumber(10))
      .succeedsWithin(1, TimeUnit.SECONDS)
      .isEqualTo(aggregation2.aggregationProof!!)
  }

  @Test
  fun `findAggregationProofByEndBlockNumber when no aggregation is found returns null`() {
    val aggregation1 = createAggregation(
      startBlockNumber = 1,
      endBlockNumber = 4
    )
    aggregationsPostgresDaoImpl.saveNewAggregation(aggregation1).get()

    assertThat(aggregationsPostgresDaoImpl.findAggregationProofByEndBlockNumber(11))
      .succeedsWithin(1, TimeUnit.SECONDS)
      .isEqualTo(null)
  }

  private fun performInsertTest(
    aggregation: Aggregation
  ): RowSet<Row>? {
    aggregationsPostgresDaoImpl.saveNewAggregation(aggregation).get()
    val dbContent = DbQueries.getTableContent(sqlClient, DbQueries.aggregationsTable).execute().get()
    val newlyInsertedRow =
      dbContent.find {
        it.getLong("start_block_number") == aggregation.startBlockNumber.toLong() &&
          it.getString("aggregation_calculator_version") == aggregation.aggregationCalculatorVersion
      }
    assertThat(newlyInsertedRow).isNotNull

    assertThat(newlyInsertedRow!!.getLong("start_block_number"))
      .isEqualTo(aggregation.startBlockNumber.toLong())
    assertThat(newlyInsertedRow.getLong("end_block_number"))
      .isEqualTo(aggregation.endBlockNumber.toLong())
    assertThat(newlyInsertedRow.getInteger("status"))
      .isEqualTo(PostgresAggregationsDao.aggregationStatusToDbValue(aggregation.status))
    assertThat(newlyInsertedRow.getLong("batch_count"))
      .isEqualTo(aggregation.batchCount.toLong())
    assertThat(newlyInsertedRow.getString("aggregation_calculator_version"))
      .isEqualTo(aggregation.aggregationCalculatorVersion)
    return dbContent
  }

  private fun performUpdateTest(
    aggregation: Aggregation
  ): RowSet<Row>? {
    val rowsUpdated = aggregationsPostgresDaoImpl.updateAggregationAsProven(aggregation).get()
    assertThat(rowsUpdated).isEqualTo(1)
    val dbContent = DbQueries.getTableContent(sqlClient, DbQueries.aggregationsTable).execute().get()
    val updatedRow =
      dbContent.find {
        it.getLong("start_block_number") == aggregation.startBlockNumber.toLong() &&
          it.getString("aggregation_calculator_version") == aggregation.aggregationCalculatorVersion
      }
    assertThat(updatedRow).isNotNull

    assertThat(updatedRow!!.getLong("start_block_number"))
      .isEqualTo(aggregation.startBlockNumber.toLong())
    assertThat(updatedRow.getLong("end_block_number"))
      .isEqualTo(aggregation.endBlockNumber.toLong())
    assertThat(updatedRow.getInteger("status"))
      .isEqualTo(PostgresAggregationsDao.aggregationStatusToDbValue(Aggregation.Status.Proven))
    assertThat(ProofToFinalizeJsonResponse.fromJsonString(updatedRow.getJsonObject("aggregation_proof").encode()))
      .isEqualTo(ProofToFinalizeJsonResponse.fromDomainObject(aggregation.aggregationProof!!))

    return dbContent
  }

  private fun insertNewBatch(
    batch: Batch
  ): Future<Unit> {
    val params = listOf(
      batch.createdEpochMilli,
      batch.startBlockNumber.toLong(),
      batch.endBlockNumber.toLong(),
      batch.proverVersion,
      batch.conflationCalculatorVersion,
      batch.status.toInt()
    )

    return DbQueries.insertBatch(sqlClient, DbQueries.insertBatchQueryV2, params).map { }
  }

  private fun insertBatches(batches: List<Batch>) {
    Future.all(
      batches.map { batch ->
        insertNewBatch(batch)
      }
    ).get()
  }

  private fun insertBlobs(blobs: List<BlobRecord>) {
    SafeFuture.collectAll(
      blobs.map { blob ->
        blobsPostgresDaoImpl.saveNewBlob(blob)
      }.stream()
    ).get()
  }

  private fun insertDbRecords(
    batches: List<Batch> = emptyList(),
    blobs: List<BlobRecord> = emptyList(),
    aggregations: List<Aggregation> = emptyList()
  ) {
    insertBatches(batches)
    insertBlobs(blobs)
    SafeFuture.collectAll(
      aggregations.map { aggregation ->
        aggregationsPostgresDaoImpl.saveNewAggregation(aggregation)
      }.stream()
    ).get()
  }
}
