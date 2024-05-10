package net.consensys.zkevm.persistence.dao.feehistory

import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import io.vertx.sqlclient.PreparedQuery
import io.vertx.sqlclient.Row
import io.vertx.sqlclient.RowSet
import kotlinx.datetime.Clock
import net.consensys.FakeFixedClock
import net.consensys.linea.FeeHistory
import net.consensys.linea.async.get
import net.consensys.zkevm.domain.createFeeHistory
import net.consensys.zkevm.persistence.db.DbHelper
import net.consensys.zkevm.persistence.test.CleanDbTestSuite
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.extension.ExtendWith
import java.math.BigDecimal
import java.math.BigInteger

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class FeeHistoriesPostgresDaoTest : CleanDbTestSuite() {
  override val databaseName = DbHelper.generateUniqueDbName("coordinator-tests-feehistories-dao")
  private var fakeClock = FakeFixedClock(Clock.System.now())

  private fun feeHistoriesContentQuery(): PreparedQuery<RowSet<Row>> =
    sqlClient.preparedQuery("select * from ${FeeHistoriesPostgresDao.TableName} order by block_number ASC")

  private lateinit var feeHistoriesPostgresDao: FeeHistoriesPostgresDao

  private val rewardPercentiles: List<Double> = listOf(10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0, 100.0)
  private val feeHistory = createFeeHistory(
    oldestBlockNumber = 100UL,
    initialReward = 1000UL,
    initialBaseFeePerGas = 10000UL,
    initialGasUsedRatio = 70U,
    initialBaseFeePerBlobGas = 1000UL,
    initialBlobGasUsedRatio = 60U,
    feeHistoryBlockCount = 5U,
    rewardPercentilesCount = rewardPercentiles.size
  ) // fee history of block 100, 101, 102, 103, 104

  @BeforeAll
  override fun beforeAll(vertx: Vertx) {
    super.beforeAll(vertx)
    feeHistoriesPostgresDao =
      FeeHistoriesPostgresDao(
        sqlClient,
        fakeClock
      )
  }

  @BeforeEach
  fun beforeEach() {
    fakeClock.setTimeTo(Clock.System.now())
  }

  @AfterEach
  override fun tearDown() {
    super.tearDown()
  }

  private fun performInsertTest(
    feeHistory: FeeHistory,
    rewardPercentiles: List<Double>
  ): RowSet<Row>? {
    feeHistoriesPostgresDao.saveNewFeeHistory(feeHistory, rewardPercentiles).get()
    val dbContent = feeHistoriesContentQuery().execute().get()
    val newlyInsertedRows =
      dbContent.filter { it.getLong("created_epoch_milli") == fakeClock.now().toEpochMilliseconds() }
    assertThat(newlyInsertedRows.size).isGreaterThan(0)

    newlyInsertedRows.forEachIndexed { i, row ->
      assertThat(row.getLong("block_number"))
        .isEqualTo(feeHistory.oldestBlock.toLong() + i)
      assertThat(row.getLong("base_fee_per_gas"))
        .isEqualTo(feeHistory.baseFeePerGas[i].toLong())
      assertThat(row.getLong("base_fee_per_blob_gas"))
        .isEqualTo(feeHistory.baseFeePerBlobGas.getOrElse(i) { BigInteger.ZERO }.toLong())
      assertThat(row.getFloat("gas_used_ratio"))
        .isEqualTo(feeHistory.gasUsedRatio[i].toFloat())
      assertThat(row.getFloat("blob_gas_used_ratio"))
        .isEqualTo(feeHistory.blobGasUsedRatio.getOrElse(i) { BigDecimal.ZERO }.toFloat())
      assertThat(row.getArrayOfLongs("rewards"))
        .containsAll(feeHistory.reward[i].map { it.toLong() })
      assertThat(row.getArrayOfFloats("reward_percentiles"))
        .containsAll(rewardPercentiles.map { it.toFloat() })
    }
    return dbContent
  }

  @Test
  fun `saveNewFeeHistory inserts new fee history to db`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles
      )
    assertThat(dbContent).size().isEqualTo(5)
  }

  @Test
  fun `saveNewFeeHistory upserts new fee history to db`() {
    var dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles
      )
    assertThat(dbContent).size().isEqualTo(5)

    // fee history of block 103, 104, 105, 106, 107
    val overlappedFeeHistory = createFeeHistory(
      oldestBlockNumber = 103UL,
      initialReward = 2000UL,
      initialBaseFeePerGas = 20000UL,
      initialGasUsedRatio = 20U,
      initialBaseFeePerBlobGas = 2000UL,
      initialBlobGasUsedRatio = 10U,
      feeHistoryBlockCount = 5U,
      rewardPercentilesCount = rewardPercentiles.size
    )
    fakeClock.setTimeTo(Clock.System.now())

    performInsertTest(
      overlappedFeeHistory,
      rewardPercentiles
    )

    dbContent = feeHistoriesContentQuery().execute().get()
    assertThat(dbContent).size().isEqualTo(8)
  }

  @Test
  fun `findBaseFeePerGasAtPercentile returns correct percentile values`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles
      )
    assertThat(dbContent).size().isEqualTo(5)

    val p10BaseFeePerGas = feeHistoriesPostgresDao.findBaseFeePerGasAtPercentile(
      percentile = 10.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(p10BaseFeePerGas).isEqualTo(BigInteger.valueOf(10000))

    val p50BaseFeePerGas = feeHistoriesPostgresDao.findBaseFeePerGasAtPercentile(
      percentile = 50.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(p50BaseFeePerGas).isEqualTo(BigInteger.valueOf(10002))

    val p75BaseFeePerGas = feeHistoriesPostgresDao.findBaseFeePerGasAtPercentile(
      percentile = 75.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(p75BaseFeePerGas).isEqualTo(BigInteger.valueOf(10003))

    val p100BaseFeePerGas = feeHistoriesPostgresDao.findBaseFeePerGasAtPercentile(
      percentile = 100.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(p100BaseFeePerGas).isEqualTo(BigInteger.valueOf(10004))
  }

  @Test
  fun `findBaseFeePerBlobGasAtPercentile returns correct percentile values`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles
      )
    assertThat(dbContent).size().isEqualTo(5)

    val p10BaseFeePerBlobGas = feeHistoriesPostgresDao.findBaseFeePerBlobGasAtPercentile(
      percentile = 10.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(p10BaseFeePerBlobGas).isEqualTo(BigInteger.valueOf(1000))

    val p50BaseFeePerBlobGas = feeHistoriesPostgresDao.findBaseFeePerBlobGasAtPercentile(
      percentile = 50.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(p50BaseFeePerBlobGas).isEqualTo(BigInteger.valueOf(1002))

    val p75BaseFeePerBlobGas = feeHistoriesPostgresDao.findBaseFeePerBlobGasAtPercentile(
      percentile = 75.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(p75BaseFeePerBlobGas).isEqualTo(BigInteger.valueOf(1003))

    val p100BaseFeePerBlobGas = feeHistoriesPostgresDao.findBaseFeePerBlobGasAtPercentile(
      percentile = 100.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(p100BaseFeePerBlobGas).isEqualTo(BigInteger.valueOf(1004))
  }

  @Test
  fun `findAverageRewardAtPercentile returns correct average values`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles
      )
    assertThat(dbContent).size().isEqualTo(5)

    val avgP10Reward = feeHistoriesPostgresDao.findAverageRewardAtPercentile(
      rewardPercentile = 10.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(avgP10Reward).isEqualTo(BigInteger.valueOf(1002))

    val avgP20Reward = feeHistoriesPostgresDao.findAverageRewardAtPercentile(
      rewardPercentile = 20.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(avgP20Reward).isEqualTo(BigInteger.valueOf(2004))

    val avgP70Reward = feeHistoriesPostgresDao.findAverageRewardAtPercentile(
      rewardPercentile = 70.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(avgP70Reward).isEqualTo(BigInteger.valueOf(7014))

    val avgP100Reward = feeHistoriesPostgresDao.findAverageRewardAtPercentile(
      rewardPercentile = 100.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(avgP100Reward).isEqualTo(BigInteger.valueOf(10020))
  }

  @Test
  fun `findAverageRewardAtPercentile returns null with unfound percentile`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles
      )
    assertThat(dbContent).size().isEqualTo(5)

    val avgP15Reward = feeHistoriesPostgresDao.findAverageRewardAtPercentile(
      rewardPercentile = 15.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(avgP15Reward).isNull()
  }

  @Test
  fun `findHighestBlockNumberWithPercentile returns highest block numbers of a given percentile`() {
    var dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles
      )
    assertThat(dbContent).size().isEqualTo(5)

    val rewardPercentile90 = listOf(90.0)
    val feeHistory = createFeeHistory(
      oldestBlockNumber = 105UL,
      initialReward = 1000UL,
      initialBaseFeePerGas = 10000UL,
      initialGasUsedRatio = 70U,
      initialBaseFeePerBlobGas = 1000UL,
      initialBlobGasUsedRatio = 60U,
      feeHistoryBlockCount = 5U,
      rewardPercentilesCount = rewardPercentile90.size
    )
    fakeClock.setTimeTo(Clock.System.now())

    dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentile90
      )
    assertThat(dbContent).size().isEqualTo(10)

    val p10HighestBlockNumber = feeHistoriesPostgresDao.findHighestBlockNumberWithPercentile(
      rewardPercentile = 10.0
    ).get()
    assertThat(p10HighestBlockNumber).isEqualTo(104L)

    val p20HighestBlockNumber = feeHistoriesPostgresDao.findHighestBlockNumberWithPercentile(
      rewardPercentile = 20.0
    ).get()
    assertThat(p20HighestBlockNumber).isEqualTo(104L)

    val p100HighestBlockNumber = feeHistoriesPostgresDao.findHighestBlockNumberWithPercentile(
      rewardPercentile = 100.0
    ).get()
    assertThat(p100HighestBlockNumber).isEqualTo(104L)

    val p90HighestBlockNumber = feeHistoriesPostgresDao.findHighestBlockNumberWithPercentile(
      rewardPercentile = 90.0
    ).get()
    assertThat(p90HighestBlockNumber).isEqualTo(109L)
  }

  @Test
  fun `findHighestBlockNumberWithPercentile returns null with unfound percentile`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles
      )
    assertThat(dbContent).size().isEqualTo(5)

    val p25HighestBlockNumber = feeHistoriesPostgresDao.findHighestBlockNumberWithPercentile(
      rewardPercentile = 25.0
    ).get()
    assertThat(p25HighestBlockNumber).isNull()
  }

  @Test
  fun `getNumOfFeeHistoriesFromBlockNumber returns number of records within the given percentile and from block`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles
      )
    assertThat(dbContent).size().isEqualTo(5)

    val p10NumOfRecords = feeHistoriesPostgresDao.getNumOfFeeHistoriesFromBlockNumber(
      rewardPercentile = 10.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(p10NumOfRecords).isEqualTo(5)

    val p50NumOfRecords = feeHistoriesPostgresDao.getNumOfFeeHistoriesFromBlockNumber(
      rewardPercentile = 50.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(p50NumOfRecords).isEqualTo(5)

    // unfound reward percentile
    val p75NumOfRecords = feeHistoriesPostgresDao.getNumOfFeeHistoriesFromBlockNumber(
      rewardPercentile = 75.0,
      fromBlockNumber = 100L
    ).get()
    assertThat(p75NumOfRecords).isEqualTo(0)

    // out of block range
    val p20NumOfRecordsOutOfRange = feeHistoriesPostgresDao.getNumOfFeeHistoriesFromBlockNumber(
      rewardPercentile = 20.0,
      fromBlockNumber = 110L
    ).get()
    assertThat(p20NumOfRecordsOutOfRange).isEqualTo(0)
  }

  @Test
  fun `deleteFeeHistoriesUpToBlockNumber deletes number of records below or equal to the given block number`() {
    var dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles
      )
    assertThat(dbContent).size().isEqualTo(5)

    var deletedNum = feeHistoriesPostgresDao.deleteFeeHistoriesUpToBlockNumber(
      blockNumberInclusive = 102L
    ).get()
    assertThat(deletedNum).isEqualTo(3)

    dbContent = feeHistoriesContentQuery().execute().get()
    assertThat(dbContent).size().isEqualTo(2)

    deletedNum = feeHistoriesPostgresDao.deleteFeeHistoriesUpToBlockNumber(
      blockNumberInclusive = 102L
    ).get()
    assertThat(deletedNum).isEqualTo(0)

    dbContent = feeHistoriesContentQuery().execute().get()
    assertThat(dbContent).size().isEqualTo(2)

    deletedNum = feeHistoriesPostgresDao.deleteFeeHistoriesUpToBlockNumber(
      blockNumberInclusive = 110L
    ).get()
    assertThat(deletedNum).isEqualTo(2)

    dbContent = feeHistoriesContentQuery().execute().get()
    assertThat(dbContent).size().isEqualTo(0)
  }
}
