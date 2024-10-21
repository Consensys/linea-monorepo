package net.consensys.zkevm.persistence.dao.aggregation

import io.vertx.junit5.VertxExtension
import io.vertx.sqlclient.Row
import io.vertx.sqlclient.RowSet
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.FakeFixedClock
import net.consensys.linea.async.get
import net.consensys.trimToSecondPrecision
import net.consensys.zkevm.coordinator.clients.prover.serialization.ProofToFinalizeJsonResponse
import net.consensys.zkevm.domain.Aggregation
import net.consensys.zkevm.domain.Batch
import net.consensys.zkevm.domain.BlobAndBatchCounters
import net.consensys.zkevm.domain.BlobCounters
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.BlobStatus
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.createAggregation
import net.consensys.zkevm.domain.createBatch
import net.consensys.zkevm.domain.createBlobRecordFromBatches
import net.consensys.zkevm.domain.createProofToFinalize
import net.consensys.zkevm.persistence.dao.batch.persistence.BatchesPostgresDao
import net.consensys.zkevm.persistence.dao.blob.BlobsPostgresDao
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
import java.util.concurrent.TimeUnit

@ExtendWith(VertxExtension::class)
class AggregationsPostgresDaoTest : CleanDbTestSuiteParallel() {
  init {
    target = "4"
  }

  override val databaseName = DbHelper.generateUniqueDbName("coordinator-tests-aggregations-dao")

  private val maxBlobReturnLimit = 10u
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

  private var fakeClockTime = Instant.parse("2023-12-11T00:00:00.000Z")
  private var fakeClock = FakeFixedClock(fakeClockTime)

  private lateinit var aggregationsPostgresDaoImpl: PostgresAggregationsDao
  private lateinit var blobsPostgresDaoImpl: BlobsPostgresDao
  private lateinit var batchesPostgresDaoImpl: BatchesPostgresDao

  @BeforeEach
  fun beforeEach() {
    aggregationsPostgresDaoImpl = PostgresAggregationsDao(sqlClient, fakeClock)
    blobsPostgresDaoImpl = BlobsPostgresDao(
      BlobsPostgresDao.Config(maxBlobReturnLimit),
      sqlClient
    )
    batchesPostgresDaoImpl = BatchesPostgresDao(sqlClient, fakeClock)
  }

  @AfterEach
  override fun tearDown() {
    super.tearDown()
  }

  @Test
  fun saveNewAggregation_saves_aggregation_to_db() {
    val aggregation1 = Aggregation(
      startBlockNumber = 1UL,
      endBlockNumber = 10UL,
      status = Aggregation.Status.Proven,
      batchCount = 5UL,
      aggregationProof = sampleResponse
    )

    val dbContent1 = performInsertTest(aggregation1)
    assertThat(dbContent1).size().isEqualTo(1)

    val aggregation2 = Aggregation(
      startBlockNumber = 11UL,
      endBlockNumber = 12UL,
      status = Aggregation.Status.Proven,
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
            "endBlockNumber=${aggregation.endBlockNumber} " +
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
      batchCount = 5UL,
      aggregationProof = null
    )

    val dbContent1 = performInsertTest(aggregation)
    assertThat(dbContent1).size().isEqualTo(1)

    val updatedAggregation = Aggregation(
      startBlockNumber = aggregation.startBlockNumber,
      endBlockNumber = aggregation.endBlockNumber,
      status = Aggregation.Status.Proven,
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
      batchCount = 5UL,
      aggregationProof = sampleResponse
    )
    val rowsUpdated = aggregationsPostgresDaoImpl.updateAggregationAsProven(aggregation).get()
    assertThat(rowsUpdated).isZero()
  }

  private fun getBlobAndBatchCounters(batches: List<Batch>, blob: BlobRecord): BlobAndBatchCounters {
    val blobCounters = BlobCounters(
      numberOfBatches = blob.batchesCount,
      startBlockNumber = blob.startBlockNumber,
      endBlockNumber = blob.endBlockNumber,
      startBlockTimestamp = blob.startBlockTime,
      endBlockTimestamp = blob.endBlockTime,
      expectedShnarf = blob.expectedShnarf
    )
    return BlobAndBatchCounters(
      blobCounters = blobCounters,
      executionProofs = BlockIntervals(
        startingBlockNumber = blob.startBlockNumber,
        upperBoundaries = batches.map { it.endBlockNumber }
      )
    )
  }

  @Test
  fun getCandidatesForAggregation_from_block_number_mismatch() {
    val batch1 = createBatch(1L, 10L)
    val batch2 = createBatch(11L, 20L)
    val batch3 = createBatch(21L, 30L)
    val batch4 = createBatch(31L, 40L)
    val batches = listOf(batch1, batch2, batch3, batch4)

    val blob1 = createBlobRecordFromBatches(listOf(batch1, batch2))
    val blob2 = createBlobRecordFromBatches(listOf(batch3, batch4))
    val blobs = listOf(blob1, blob2)

    batches.forEach { insertBatch(it).get() }
    blobs.forEach { insertBlob(it).get() }

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(fromBlockNumber = 0).get()
      .also { blobCounters ->
        assertThat(blobCounters).isEmpty()
      }
    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(fromBlockNumber = 1).get()
      .also { blobCounters ->
        assertThat(blobCounters).hasSameElementsAs(
          listOf(
            getBlobAndBatchCounters(listOf(batch1, batch2), blob1),
            getBlobAndBatchCounters(listOf(batch3, batch4), blob2)
          )
        )
      }
  }

  @Test
  fun getCandidatesForAggregation_unproven_batch_in_the_middle() {
    val batch1 = createBatch(1L, 10L)
    val batch2 = createBatch(11L, 20L)
    val batch3 = createBatch(21L, 30L, status = Batch.Status.Finalized)
    val batch4 = createBatch(31L, 40L)
    val batches = listOf(batch1, batch2, batch3, batch4)

    val blob1 = createBlobRecordFromBatches(listOf(batch1, batch2))
    val blob2 = createBlobRecordFromBatches(listOf(batch3, batch4))
    val blobs = listOf(blob1, blob2)

    batches.forEach { insertBatch(it).get() }
    blobs.forEach { insertBlob(it).get() }

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(fromBlockNumber = 1).get()
      .also { blobCounters ->
        assertThat(blobCounters).hasSameElementsAs(listOf(getBlobAndBatchCounters(listOf(batch1, batch2), blob1)))
      }
  }

  @Test
  fun getCandidatesForAggregation_unproven_blob_in_the_middle() {
    val batch1 = createBatch(1L, 10L)
    val batch2 = createBatch(11L, 20L)
    val batch3 = createBatch(21L, 30L)
    val batch4 = createBatch(31L, 40L)
    val batch5 = createBatch(41L, 50L)
    val batch6 = createBatch(51L, 60L)
    val batches = listOf(batch1, batch2, batch3, batch4, batch5)

    val blob1 = createBlobRecordFromBatches(listOf(batch1, batch2))
    val blob2 = createBlobRecordFromBatches(listOf(batch3, batch4))
    val blob3 = createBlobRecordFromBatches(listOf(batch5), status = BlobStatus.COMPRESSION_PROVING)
    val blob4 = createBlobRecordFromBatches(listOf(batch6))
    val blobs = listOf(blob1, blob2, blob3, blob4)

    batches.forEach { insertBatch(it).get() }
    blobs.forEach { insertBlob(it).get() }

    aggregationsPostgresDaoImpl.findConsecutiveProvenBlobs(fromBlockNumber = 1).get()
      .also { blobCounters ->
        assertThat(blobCounters).hasSameElementsAs(
          listOf(
            getBlobAndBatchCounters(listOf(batch1, batch2), blob1),
            getBlobAndBatchCounters(listOf(batch3, batch4), blob2)
          )
        )
      }
  }

  @Test
  fun getCandidatesForAggregation_first_blob_unproven_shall_return_no_results() {
    // conflation v1
    // blob1, {batch1 (4-10)} -- UNPROVEN
    // blob2  { batch2 (11-20), batch3 (21-30) }

    val batch1v1 = createBatch(4L, 10L, status = Batch.Status.Finalized)
    val batch2v1 = createBatch(11L, 20L)
    val batch3v1 = createBatch(21L, 30L)
    val blob1v1 = createBlobRecordFromBatches(listOf(batch1v1))
    val blob2v1 = createBlobRecordFromBatches(listOf(batch2v1, batch3v1))

    val batches = listOf(batch1v1, batch2v1, batch3v1)
    val blobs = listOf(blob1v1, blob2v1)

    batches.forEach { insertBatch(it).get() }
    blobs.forEach { insertBlob(it).get() }

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
        it.getLong("start_block_number") == aggregation.startBlockNumber.toLong()
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
        it.getLong("start_block_number") == aggregation.startBlockNumber.toLong()
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

  private fun insertBatch(
    batch: Batch
  ): SafeFuture<Unit> {
    return batchesPostgresDaoImpl.saveNewBatch(batch)
  }

  private fun insertBlob(blob: BlobRecord): SafeFuture<Unit> {
    return blobsPostgresDaoImpl.saveNewBlob(blob)
  }
}
