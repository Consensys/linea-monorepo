package net.consensys.zkevm.ethereum.settlement.persistence

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import io.vertx.sqlclient.PreparedQuery
import io.vertx.sqlclient.Row
import io.vertx.sqlclient.RowSet
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import kotlinx.datetime.toJavaInstant
import net.consensys.linea.async.get
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.coordinator.clients.ProverErrorType
import net.consensys.zkevm.coordinator.clients.response.ProverResponsesRepository
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.assertThrows
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.kotlin.any
import org.mockito.kotlin.atLeast
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.eq
import org.mockito.kotlin.mock
import org.mockito.kotlin.reset
import org.mockito.kotlin.times
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.util.concurrent.ExecutionException
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class PostgresBatchesRepositoryTest : CleanDbTestSuite() {
  private val maxBatchesToReturn = 10u
  private val responsesRepository = mock<ProverResponsesRepository>()
  private var timeToReturn: Instant = Instant.DISTANT_PAST
  private fun batchesContentQuery(): PreparedQuery<RowSet<Row>> =
    sqlClient.preparedQuery("select * from ${PostgresBatchesRepository.TableName}")

  private val fixedClock =
    object : Clock {
      override fun now(): Instant {
        return timeToReturn
      }
    }
  private lateinit var postgresBatchesRepository: PostgresBatchesRepository

  @BeforeAll
  override fun beforeAll(vertx: Vertx) {
    super.beforeAll(vertx)
    postgresBatchesRepository =
      PostgresBatchesRepository(
        PostgresBatchesRepository.Config(maxBatchesToReturn),
        sqlClient,
        responsesRepository,
        fixedClock
      )
  }

  @AfterEach
  override fun tearDown() {
    super.tearDown()
    reset(responsesRepository)
  }

  @Test
  fun saveNewBatch_savesTheBatch_to_db() {
    val proverResponse = mock<GetProofResponse>()
    val expectedVersion1 = "0.0.1"
    whenever(proverResponse.proverVersion).thenReturn(expectedVersion1)
    val expectedStartBlock1 = UInt64.ONE
    val expectedEndBlock1 = UInt64.ONE
    val batch1 =
      Batch(expectedStartBlock1, expectedEndBlock1, proverResponse, Batch.Status.Finalized)
    timeToReturn = Clock.System.now()

    val dbContent1 =
      performInsertTest(batch1, expectedStartBlock1, expectedEndBlock1, proverResponse)
    assertThat(dbContent1).size().isEqualTo(1)
    reset(proverResponse)

    val expectedVersion2 = "0.0.2"
    val expectedStartBlock2 = UInt64.valueOf(6)
    val expectedEndBlock2 = UInt64.valueOf(9)
    whenever(proverResponse.proverVersion).thenReturn(expectedVersion2)
    val batch2 = Batch(expectedStartBlock2, expectedEndBlock2, proverResponse, Batch.Status.Pending)
    timeToReturn = timeToReturn.plus(1.seconds)

    val dbContent2 =
      performInsertTest(batch2, expectedStartBlock2, expectedEndBlock2, proverResponse)
    assertThat(dbContent2).size().isEqualTo(2)
  }

  private fun performInsertTest(
    batch: Batch,
    expectedStartBlock: UInt64,
    expectedEndBlock: UInt64,
    proverResponse: GetProofResponse
  ): RowSet<Row>? {
    postgresBatchesRepository.saveNewBatch(batch).get()
    val dbContent = batchesContentQuery().execute().get()
    val newlyInsertedRow =
      dbContent.find { it.getLong("created_epoch_milli") == timeToReturn.toEpochMilliseconds() }
    assertThat(newlyInsertedRow).isNotNull

    assertThat(newlyInsertedRow!!.getLong("start_block_number"))
      .isEqualTo(expectedStartBlock.longValue())
    assertThat(newlyInsertedRow.getLong("end_block_number")).isEqualTo(expectedEndBlock.longValue())
    assertThat(newlyInsertedRow.getString("prover_version"))
      .isEqualTo(batch.proverResponse.proverVersion)
    assertThat(newlyInsertedRow.getInteger("status")).isEqualTo(batchStatusToDbValue(batch.status))

    verify(proverResponse, atLeast(1)).proverVersion
    return dbContent
  }

  @Test
  fun saveNewBatch_returnsErrorWhenDuplicated() {
    val proverResponse = mock<GetProofResponse>()
    val expectedVersion1 = "0.0.1"
    whenever(proverResponse.proverVersion).thenReturn(expectedVersion1)
    val expectedStartBlock1 = UInt64.ONE
    val expectedEndBlock1 = UInt64.ONE
    val batch1 =
      Batch(expectedStartBlock1, expectedEndBlock1, proverResponse, Batch.Status.Finalized)
    timeToReturn = Clock.System.now()

    val dbContent1 =
      performInsertTest(batch1, expectedStartBlock1, expectedEndBlock1, proverResponse)
    assertThat(dbContent1).size().isEqualTo(1)

    assertThrows<ExecutionException> {
      postgresBatchesRepository.saveNewBatch(batch1).get()
    }.also { executionException ->
      assertThat(executionException.cause).isInstanceOf(DuplicatedBatchException::class.java)
      assertThat(executionException.cause!!.message)
        .isEqualTo("Batch startBlockNumber=1, endBlockNumber=1, proverVersion=0.0.1 is already persisted!")
    }
  }

  @Test
  fun `getConsecutiveBatchesFromBlockNumber works correctly for 1 batch`() {
    val proverResponse = mock<GetProofResponse>()
    val expectedVersion = "0.0.1"
    whenever(proverResponse.proverVersion).thenReturn(expectedVersion)
    val expectedStartBlock = UInt64.ONE
    val expectedEndBlock = UInt64.valueOf(3)
    val expectedBatch =
      Batch(expectedStartBlock, expectedEndBlock, proverResponse, Batch.Status.Finalized)
    val expectedProverResponsesIndex =
      ProverResponsesRepository.ProverResponseIndex(
        expectedStartBlock,
        expectedEndBlock,
        expectedVersion
      )
    postgresBatchesRepository.saveNewBatch(expectedBatch).get()
    whenever(responsesRepository.find(eq(expectedProverResponsesIndex)))
      .thenReturn(SafeFuture.completedFuture(Ok(proverResponse)))

    val actualBatches =
      postgresBatchesRepository.getConsecutiveBatchesFromBlockNumber(expectedStartBlock).get()
    assertThat(actualBatches).hasSameElementsAs(listOf(expectedBatch))
  }

  @Test
  fun `getConsecutiveBatchesFromBlockNumber filters batches with lower prover version and timestamp`() {
    val outdatedProverResponse = mock<GetProofResponse>() {
      on { proverVersion } doReturn "0.0.1"
    }
    val batch1outDatedProverResponse = Batch(UInt64.ONE, UInt64.valueOf(3), outdatedProverResponse)
    val templateBlockData = GetProofResponse.BlockData(
      zkRootHash = Bytes32.ZERO,
      timestamp = Instant.parse("2023-07-10T00:00:00Z").toJavaInstant(),
      rlpEncodedTransactions = emptyList(),
      batchReceptionIndices = emptyList(),
      l2ToL1MsgHashes = emptyList(),
      fromAddresses = Bytes.EMPTY
    )
    val proverResponse1 = mock<GetProofResponse>() {
      on { proverVersion } doReturn "0.1.0"
      on { blocksData } doReturn listOf(
        templateBlockData.copy(timestamp = Instant.parse("2023-07-10T00:00:00Z").toJavaInstant()),
        templateBlockData.copy(timestamp = Instant.parse("2023-07-10T00:00:12Z").toJavaInstant()),
        templateBlockData.copy(timestamp = Instant.parse("2023-07-10T00:00:24Z").toJavaInstant())
      )
    }

    val proverResponse2 = mock<GetProofResponse>() {
      on { proverVersion } doReturn "0.1.0"
      on { blocksData } doReturn listOf(
        templateBlockData.copy(timestamp = Instant.parse("2023-07-10T04:00:00Z").toJavaInstant()),
        templateBlockData.copy(timestamp = Instant.parse("2023-07-10T04:00:24Z").toJavaInstant())
      )
    }

    val batch1 = batch1outDatedProverResponse.copy(proverResponse = proverResponse1)
    val batch2 = Batch(UInt64.valueOf(4), UInt64.valueOf(5), proverResponse2)
    SafeFuture.collectAll(
      postgresBatchesRepository.saveNewBatch(batch1outDatedProverResponse),
      postgresBatchesRepository.saveNewBatch(batch1),
      postgresBatchesRepository.saveNewBatch(batch2)
    ).get()
    whenever(
      responsesRepository.find(
        eq(
          ProverResponsesRepository.ProverResponseIndex(
            UInt64.valueOf(1),
            UInt64.valueOf(3),
            version = "0.1.0"
          )
        )
      )
    )
      .thenReturn(SafeFuture.completedFuture(Ok(proverResponse1)))

    whenever(
      responsesRepository.find(
        eq(
          ProverResponsesRepository.ProverResponseIndex(
            UInt64.valueOf(4),
            UInt64.valueOf(5),
            version = "0.1.0"
          )
        )
      )
    )
      .thenReturn(SafeFuture.completedFuture(Ok(proverResponse2)))

    postgresBatchesRepository.getConsecutiveBatchesFromBlockNumber(
      UInt64.valueOf(1),
      Instant.parse("2023-07-10T04:00:05Z")
    ).get().also { batches ->
      assertThat(batches).hasSameElementsAs(listOf(batch1))
    }

    postgresBatchesRepository.getConsecutiveBatchesFromBlockNumber(
      UInt64.valueOf(1),
      Instant.parse("2023-07-10T04:00:25Z")
    ).get().also { batches ->
      assertThat(batches).hasSameElementsAs(listOf(batch1, batch2))
    }
  }

  @Test
  fun `getConsecutiveBatchesFromBlockNumber returns a sequence of batches without gaps`() {
    val expectedProverVersion = "0.1.0"
    val batchBeforeTheRequestedInterval =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.ONE,
          UInt64.valueOf(3),
          expectedProverVersion
        )
      )
    val proverExpectedResponseIndex1 =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(4),
        UInt64.valueOf(4),
        expectedProverVersion
      )
    val proverExpectedResponseIndex2 =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(5),
        UInt64.valueOf(7),
        expectedProverVersion
      )
    val proverExpectedResponseIndex3 =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(8),
        UInt64.valueOf(11),
        expectedProverVersion
      )
    val expectedBatches =
      listOf(
        createBatch(proverExpectedResponseIndex1),
        createBatch(proverExpectedResponseIndex2),
        createBatch(proverExpectedResponseIndex3)
      )
    val batchAfterTheRequestedInterval =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.valueOf(13),
          UInt64.valueOf(16),
          expectedProverVersion
        )
      )
    val notNeededBatches = listOf(batchBeforeTheRequestedInterval, batchAfterTheRequestedInterval)
    expectedBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }
    notNeededBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }

    val actualBatches =
      postgresBatchesRepository
        .getConsecutiveBatchesFromBlockNumber(expectedBatches.first().startBlockNumber)
        .get()
    assertThat(actualBatches).hasSameElementsAs(expectedBatches)
    verify(responsesRepository, times(1)).find(eq(proverExpectedResponseIndex1))
    verify(responsesRepository, times(1)).find(eq(proverExpectedResponseIndex2))
  }

  @Test
  fun `getConsecutiveBatchesFromBlockNumber returns a sequence of batches until and exclude the error batch`() {
    val expectedProverVersion = "0.1.0"
    val batchBeforeTheRequestedInterval =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.ONE,
          UInt64.valueOf(3),
          expectedProverVersion
        )
      )
    val proverExpectedResponseIndex1 =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(4),
        UInt64.valueOf(4),
        expectedProverVersion
      )
    val proverExpectedResponseIndex2 =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(5),
        UInt64.valueOf(7),
        expectedProverVersion
      )
    val proverExpectedResponseIndex3 =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(8),
        UInt64.valueOf(12),
        expectedProverVersion
      )
    val batchWithError =
      createBatchWithError(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.valueOf(13),
          UInt64.valueOf(16),
          expectedProverVersion
        )
      )
    val expectedBatches =
      listOf(
        createBatch(proverExpectedResponseIndex1),
        createBatch(proverExpectedResponseIndex2),
        createBatch(proverExpectedResponseIndex3)
      )
    val batchAfterErrorBatch =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.valueOf(17),
          UInt64.valueOf(20),
          expectedProverVersion
        )
      )
    val notNeededBatches = listOf(batchBeforeTheRequestedInterval, batchWithError, batchAfterErrorBatch)
    expectedBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }
    notNeededBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }

    val actualBatches =
      postgresBatchesRepository
        .getConsecutiveBatchesFromBlockNumber(expectedBatches.first().startBlockNumber, Instant.DISTANT_FUTURE)
        .get()
    assertThat(actualBatches).hasSameElementsAs(expectedBatches)
    verify(responsesRepository, times(1)).find(eq(proverExpectedResponseIndex1))
    verify(responsesRepository, times(1)).find(eq(proverExpectedResponseIndex2))
    verify(responsesRepository, times(1)).find(eq(proverExpectedResponseIndex3))
  }

  @Test
  fun `getConsecutiveBatchesFromBlockNumber returns empty list of batches as the first batch triggered error`() {
    val expectedProverVersion = "0.1.0"
    val batchBeforeTheRequestedInterval =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.ONE,
          UInt64.valueOf(3),
          expectedProverVersion
        )
      )
    val batchWithError =
      createBatchWithError(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.valueOf(4),
          UInt64.valueOf(7),
          expectedProverVersion
        )
      )
    val proverResponseIndex1 =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(8),
        UInt64.valueOf(9),
        expectedProverVersion
      )
    val proverResponseIndex2 =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(9),
        UInt64.valueOf(12),
        expectedProverVersion
      )
    val unexpectedBatches =
      listOf(
        createBatch(proverResponseIndex1),
        createBatch(proverResponseIndex2)
      )
    val batchAfterTheRequestedInterval =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.valueOf(14),
          UInt64.valueOf(20),
          expectedProverVersion
        )
      )
    val notNeededBatches = listOf(batchBeforeTheRequestedInterval, batchWithError, batchAfterTheRequestedInterval)
    unexpectedBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }
    notNeededBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }

    val actualBatches =
      postgresBatchesRepository
        .getConsecutiveBatchesFromBlockNumber(batchWithError.startBlockNumber, Instant.DISTANT_FUTURE)
        .get()
    assertThat(actualBatches).isEmpty()
  }

  @Test
  fun `getConsecutiveBatchesFromBlockNumber returns a sequence of batches exclude error batch at the end`() {
    val expectedProverVersion = "0.1.0"
    val batchBeforeTheRequestedInterval =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.ONE,
          UInt64.valueOf(3),
          expectedProverVersion
        )
      )
    val proverExpectedResponseIndex1 =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(4),
        UInt64.valueOf(4),
        expectedProverVersion
      )
    val proverExpectedResponseIndex2 =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(5),
        UInt64.valueOf(7),
        expectedProverVersion
      )
    val proverExpectedResponseIndex3 =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(8),
        UInt64.valueOf(11),
        expectedProverVersion
      )
    val expectedBatches =
      listOf(
        createBatch(proverExpectedResponseIndex1),
        createBatch(proverExpectedResponseIndex2),
        createBatch(proverExpectedResponseIndex3)
      )
    val batchWithError =
      createBatchWithError(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.valueOf(12),
          UInt64.valueOf(16),
          expectedProverVersion
        )
      )
    val batchAfterTheRequestedInterval =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.valueOf(18),
          UInt64.valueOf(20),
          expectedProverVersion
        )
      )
    val notNeededBatches = listOf(batchBeforeTheRequestedInterval, batchWithError, batchAfterTheRequestedInterval)
    expectedBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }
    notNeededBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }

    val actualBatches =
      postgresBatchesRepository
        .getConsecutiveBatchesFromBlockNumber(expectedBatches.first().startBlockNumber, Instant.DISTANT_FUTURE)
        .get()
    assertThat(actualBatches).hasSameElementsAs(expectedBatches)
    verify(responsesRepository, times(1)).find(eq(proverExpectedResponseIndex1))
    verify(responsesRepository, times(1)).find(eq(proverExpectedResponseIndex2))
    verify(responsesRepository, times(1)).find(eq(proverExpectedResponseIndex3))
  }

  @Test
  fun `getConsecutiveBatchesFromBlockNumber returns a sequence of batches with maximum version without gaps`() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))
    val batchBeforeTheRequestedInterval1 =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(UInt64.ONE, UInt64.valueOf(3), "0.1.0")
      )
    val batchBeforeTheRequestedInterval2 =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(UInt64.ONE, UInt64.valueOf(3), "0.1.1")
      )
    val firstBatchWithOldVersion =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.valueOf(4),
          UInt64.valueOf(4),
          "0.1.0"
        )
      )
    val secondBatchWithOldVersion =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.valueOf(5),
          UInt64.valueOf(7),
          "0.1.0"
        )
      )
    val batchAfterTheGap1 =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.valueOf(13),
          UInt64.valueOf(15),
          "0.1.0"
        )
      )
    val batchAfterTheGap2 =
      createBatch(
        ProverResponsesRepository.ProverResponseIndex(
          UInt64.valueOf(13),
          UInt64.valueOf(15),
          "0.1.1"
        )
      )
    val notNeededBatches =
      listOf(
        firstBatchWithOldVersion,
        secondBatchWithOldVersion,
        batchBeforeTheRequestedInterval1,
        batchBeforeTheRequestedInterval2,
        batchAfterTheGap1,
        batchAfterTheGap2
      )

    val proverExpectedResponseIndex12 =
      ProverResponsesRepository.ProverResponseIndex(UInt64.valueOf(4), UInt64.valueOf(4), "0.1.1")
    val proverExpectedResponseIndex22 =
      ProverResponsesRepository.ProverResponseIndex(UInt64.valueOf(5), UInt64.valueOf(7), "0.1.2")
    val proverExpectedResponseIndex3 =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(8),
        UInt64.valueOf(11),
        "0.1.0"
      )
    val expectedBatches =
      listOf(
        createBatch(proverExpectedResponseIndex12),
        createBatch(proverExpectedResponseIndex22),
        createBatch(proverExpectedResponseIndex3)
      )

    expectedBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }
    notNeededBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }

    val actualBatches =
      postgresBatchesRepository
        .getConsecutiveBatchesFromBlockNumber(expectedBatches.first().startBlockNumber)
        .get()

    assertThat(actualBatches).hasSameElementsAs(expectedBatches)
    verify(responsesRepository, times(1)).find(eq(proverExpectedResponseIndex12))
    verify(responsesRepository, times(1)).find(eq(proverExpectedResponseIndex22))
    verify(responsesRepository, times(1)).find(eq(proverExpectedResponseIndex3))
  }

  @Test
  fun `getConsecutiveBatchesFromBlockNumber doesn't return more batches than the limit`() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))

    val createdBatches =
      (1..15).map { i ->
        createBatch(
          ProverResponsesRepository.ProverResponseIndex(
            UInt64.valueOf(i.toLong()),
            UInt64.valueOf(i.toLong()),
            "0.1.0"
          )
        )
      }
    createdBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }

    val actualBatches =
      postgresBatchesRepository.getConsecutiveBatchesFromBlockNumber(UInt64.ONE).get()
    assertThat(actualBatches).isEqualTo(createdBatches.subList(0, maxBatchesToReturn.toInt()))
  }

  @Test
  fun `getConsecutiveBatchesFromBlockNumber returned list doesn't have gaps`() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))

    val gappedBatchIndex = 7
    val createdBatches =
      (1..15).map { i ->
        createBatch(
          ProverResponsesRepository.ProverResponseIndex(
            UInt64.valueOf(i.toLong()),
            UInt64.valueOf(i.toLong()),
            "0.1.0"
          )
        )
      }.filterIndexed { i, _ ->
        i != gappedBatchIndex
      }

    createdBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }

    val actualBatches =
      postgresBatchesRepository.getConsecutiveBatchesFromBlockNumber(UInt64.ONE).get()
    assertThat(actualBatches).isEqualTo(createdBatches.subList(0, gappedBatchIndex))
  }

  @Test
  fun `getConsecutiveBatchesFromBlockNumber returned list is valid when there is a gap before the argument`() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))

    val gappedBatchIndexBeforeStart = 2
    val gappedBatchIndex = 13
    val createdBatches =
      (1..15).map { i ->
        createBatch(
          ProverResponsesRepository.ProverResponseIndex(
            UInt64.valueOf(i.toLong()),
            UInt64.valueOf(i.toLong()),
            "0.1.0"
          )
        )
      }.filterIndexed { i, _ ->
        i != gappedBatchIndexBeforeStart && i != gappedBatchIndex
      }

    createdBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }

    val actualBatches =
      postgresBatchesRepository.getConsecutiveBatchesFromBlockNumber(createdBatches[6].startBlockNumber).get()
    assertThat(actualBatches).isEqualTo(createdBatches.subList(6, 12))
  }

  @Test
  fun `getConsecutiveBatchesFromBlockNumber returns empty list if there are batches for the future`() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))

    val createdBatches =
      (5..15).map { i ->
        createBatch(
          ProverResponsesRepository.ProverResponseIndex(
            UInt64.valueOf(i.toLong()),
            UInt64.valueOf(i.toLong()),
            "0.1.0"
          )
        )
      }
    createdBatches.forEach { postgresBatchesRepository.saveNewBatch(it).get() }

    val actualBatches =
      postgresBatchesRepository.getConsecutiveBatchesFromBlockNumber(UInt64.ONE).get()
    assertThat(actualBatches).isEmpty()
  }

  @Test
  fun `getConsecutiveBatchesFromBlockNumber returns the shortest batch`() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))

    val longerBatch = createBatch(
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(1L),
        UInt64.valueOf(4L),
        "0.1.0"
      )
    )
    val shorterBatch = createBatch(
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(1L),
        UInt64.valueOf(3L),
        "0.1.0"
      )
    )
    postgresBatchesRepository.saveNewBatch(shorterBatch).get()
    postgresBatchesRepository.saveNewBatch(longerBatch).get()

    val actualBatches =
      postgresBatchesRepository.getConsecutiveBatchesFromBlockNumber(UInt64.ONE).get()
    assertThat(actualBatches).isEqualTo(listOf(shorterBatch))
  }

  @Test
  fun `findBatchWithHighestEndBlockNumberByStatus when empty returns null`() {
    assertThat(
      postgresBatchesRepository
        .findBatchWithHighestEndBlockNumberByStatus(Batch.Status.Finalized)
        .get()
    ).isEqualTo(null)

    assertThat(
      postgresBatchesRepository
        .findBatchWithHighestEndBlockNumberByStatus(Batch.Status.Pending)
        .get()
    ).isEqualTo(null)
  }

  @Test
  fun `findBatchWithHighestEndBlockNumberByStatus when no batches with matching status returns null`() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))

    val proverVersion = "0.1.0"
    val batches = listOf(createBatch(1, 3, Batch.Status.Finalized, proverVersion))

    SafeFuture.collectAll(batches.map { postgresBatchesRepository.saveNewBatch(it) }.stream()).get()
    assertThat(
      postgresBatchesRepository
        .findBatchWithHighestEndBlockNumberByStatus(Batch.Status.Pending)
        .get()
    )
      .isEqualTo(null)
  }

  @Test
  fun `findBatchWithHighestEndBlockNumberByStatus returns highest end_block number matching given status`() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))

    val proverVersion = "0.1.0"
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized, proverVersion),
        // Gap, query does not care about gaps
        createBatch(20, 30, Batch.Status.Finalized, proverVersion),
        // Gap, query does not care about gaps
        createBatch(31, 32, Batch.Status.Pending, proverVersion),
        // Gap, query does not care about gaps
        createBatch(40, 42, Batch.Status.Pending, proverVersion)
      )

    SafeFuture.collectAll(batches.map { postgresBatchesRepository.saveNewBatch(it) }.stream()).get()
    assertThat(
      postgresBatchesRepository
        .findBatchWithHighestEndBlockNumberByStatus(Batch.Status.Finalized)
        .get()
    )
      .isEqualTo(batches[1])
    assertThat(
      postgresBatchesRepository
        .findBatchWithHighestEndBlockNumberByStatus(Batch.Status.Pending)
        .get()
    )
      .isEqualTo(batches[3])
  }

  @Test
  fun `findHighestConsecutiveBatchByStatus returns highest batch without gaps`() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))

    val proverVersion = "0.1.0"
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Pending, proverVersion),
        createBatch(4, 5, Batch.Status.Pending, proverVersion),
        createBatch(6, 7, Batch.Status.Pending, proverVersion),
        createBatch(8, 10, Batch.Status.Pending, proverVersion)
      )

    SafeFuture.collectAll(batches.map { postgresBatchesRepository.saveNewBatch(it) }.stream()).get()
    assertThat(
      postgresBatchesRepository
        .findHighestConsecutiveBatchByStatus(Batch.Status.Pending)
        .get()
    )
      .isEqualTo(batches[3])
  }

  @Test
  fun `findHighestConsecutiveBatchByStatus returns highest batch without gaps - only one pending`() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))

    val proverVersion = "0.1.0"
    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized, proverVersion),
        createBatch(4, 5, Batch.Status.Pending, proverVersion)
      )

    SafeFuture.collectAll(batches.map { postgresBatchesRepository.saveNewBatch(it) }.stream()).get()
    assertThat(
      postgresBatchesRepository
        .findHighestConsecutiveBatchByStatus(Batch.Status.Pending)
        .get()
    )
      .isEqualTo(batches[1])
  }

  @Test
  fun `findHighestConsecutiveBatchByStatus returns highest batch without gaps and highest prover version `() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))

    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized, "0.1.0"),
        createBatch(4, 5, Batch.Status.Pending, "0.1.0"),
        createBatch(4, 5, Batch.Status.Pending, "0.1.1"),
        // expected at index 3
        createBatch(6, 7, Batch.Status.Pending, "0.1.4"),
        createBatch(6, 7, Batch.Status.Pending, "0.1.2"),
        createBatch(6, 7, Batch.Status.Pending, "0.1.3"),
        // Gap in the DB
        createBatch(9, 9, Batch.Status.Pending, "0.1.0"),
        createBatch(10, 11, Batch.Status.Pending, "0.1.0")
      )

    SafeFuture.collectAll(batches.map { postgresBatchesRepository.saveNewBatch(it) }.stream()).get()
    assertThat(
      postgresBatchesRepository
        .findHighestConsecutiveBatchByStatus(Batch.Status.Pending)
        .get()
    )
      .isEqualTo(batches[3])
  }

  @Test
  fun `findHighestConsecutiveBatchByStatus returns null when no pending batches are found`() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))

    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized, "0.1.0"),
        createBatch(4, 5, Batch.Status.Finalized, "0.1.0"),
        createBatch(4, 5, Batch.Status.Finalized, "0.1.1")
      )

    SafeFuture.collectAll(batches.map { postgresBatchesRepository.saveNewBatch(it) }.stream()).get()
    assertThat(
      postgresBatchesRepository
        .findHighestConsecutiveBatchByStatus(Batch.Status.Pending)
        .get()
    )
      .isNull()
  }

  @Test
  fun `setStatus updates all records`() {
    val expectedProverResponse = mock<GetProofResponse>()
    whenever(responsesRepository.find(any()))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))

    val batches =
      listOf(
        createBatch(1, 3, Batch.Status.Finalized, "0.1.0"),
        createBatch(4, 5, Batch.Status.Pending, "0.1.0"),
        createBatch(4, 5, Batch.Status.Pending, "0.1.1"),
        createBatch(6, 7, Batch.Status.Pending, "0.1.3"),
        createBatch(10, 11, Batch.Status.Pending, "0.1.0")
      )

    SafeFuture.collectAll(batches.map { postgresBatchesRepository.saveNewBatch(it) }.stream()).get()
    assertThat(
      postgresBatchesRepository
        .setBatchStatusUpToEndBlockNumber(
          endBlockNumberInclusive = UInt64.valueOf(7),
          currentStatus = Batch.Status.Pending,
          newStatus = Batch.Status.Finalized
        )
        .get()
    )
      .isEqualTo(3)
  }

  private fun createBatch(
    index: ProverResponsesRepository.ProverResponseIndex,
    status: Batch.Status = Batch.Status.Pending
  ): Batch {
    val templateBlockData = GetProofResponse.BlockData(
      zkRootHash = Bytes32.ZERO,
      timestamp = Instant.parse("2023-07-10T00:00:00Z").toJavaInstant(),
      rlpEncodedTransactions = emptyList(),
      batchReceptionIndices = emptyList(),
      l2ToL1MsgHashes = emptyList(),
      fromAddresses = Bytes.EMPTY
    )

    val expectedProverResponse = mock<GetProofResponse>() {
      on { proverVersion } doReturn index.version
      on { blocksData } doReturn listOf(templateBlockData)
    }

    whenever(responsesRepository.find(eq(index)))
      .thenReturn(SafeFuture.completedFuture(Ok(expectedProverResponse)))
    return Batch(index.startBlockNumber, index.endBlockNumber, expectedProverResponse, status)
  }

  private fun createBatchWithError(
    index: ProverResponsesRepository.ProverResponseIndex,
    status: Batch.Status = Batch.Status.Pending
  ): Batch {
    val expectedProverResponse = mock<GetProofResponse>() {
      on { proverVersion } doReturn index.version
    }

    whenever(responsesRepository.find(eq(index)))
      .thenReturn(
        SafeFuture.completedFuture(
          Err(
            ErrorResponse(
              ProverErrorType.ResponseNotFound,
              "Response file with $index wasn't found in the repo"
            )
          )
        )
      )
    return Batch(index.startBlockNumber, index.endBlockNumber, expectedProverResponse, status)
  }

  private fun createBatch(
    startBlockNumber: Long,
    endBlockNumber: Long,
    status: Batch.Status = Batch.Status.Pending,
    proverVersion: String = "0.0.0"
  ): Batch {
    val index =
      ProverResponsesRepository.ProverResponseIndex(
        UInt64.valueOf(startBlockNumber),
        UInt64.valueOf(endBlockNumber),
        proverVersion
      )
    return createBatch(index, status)
  }
}
