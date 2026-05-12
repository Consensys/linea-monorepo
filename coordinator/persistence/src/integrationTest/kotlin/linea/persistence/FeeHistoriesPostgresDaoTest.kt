package linea.persistence

import io.vertx.junit5.VertxExtension
import io.vertx.sqlclient.PreparedQuery
import io.vertx.sqlclient.Row
import io.vertx.sqlclient.RowSet
import linea.domain.FeeHistory
import net.consensys.FakeFixedClock
import net.consensys.linea.async.get
import net.consensys.zkevm.persistence.db.DbHelper
import net.consensys.zkevm.persistence.db.test.CleanDbTestSuiteParallel
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import kotlin.time.Clock

@ExtendWith(VertxExtension::class)
class FeeHistoriesPostgresDaoTest : CleanDbTestSuiteParallel() {
  init {
    target = "4"
  }

  fun createFeeHistory(
    oldestBlockNumber: ULong,
    initialReward: ULong,
    initialBaseFeePerGas: ULong,
    initialGasUsedRatio: UInt,
    initialBaseFeePerBlobGas: ULong,
    initialBlobGasUsedRatio: UInt,
    feeHistoryBlockCount: UInt,
    rewardPercentilesCount: Int,
  ): FeeHistory {
    return FeeHistory(
      oldestBlock = oldestBlockNumber,
      baseFeePerGas = (initialBaseFeePerGas until initialBaseFeePerGas + feeHistoryBlockCount + 1u).toList(),
      reward = (initialReward until initialReward + feeHistoryBlockCount)
        .map { reward -> (1..rewardPercentilesCount).map { reward.times(it.toUInt()) } },
      gasUsedRatio = (initialGasUsedRatio until initialGasUsedRatio + feeHistoryBlockCount)
        .map { (it.toDouble() / 100.0) },
      baseFeePerBlobGas = (initialBaseFeePerBlobGas until initialBaseFeePerBlobGas + feeHistoryBlockCount + 1u)
        .toList(),
      blobGasUsedRatio = (initialBlobGasUsedRatio until initialBlobGasUsedRatio + feeHistoryBlockCount)
        .map { (it.toDouble() / 100.0) },
    )
  }

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
    rewardPercentilesCount = rewardPercentiles.size,
  ) // fee history of block 100, 101, 102, 103, 104

  @BeforeEach
  fun beforeEach() {
    fakeClock.setTimeTo(Clock.System.now())
    feeHistoriesPostgresDao =
      FeeHistoriesPostgresDao(
        sqlClient,
        fakeClock,
      )
  }

  private fun performInsertTest(feeHistory: FeeHistory, rewardPercentiles: List<Double>): RowSet<Row>? {
    feeHistoriesPostgresDao.saveNewFeeHistory(feeHistory, rewardPercentiles).get()
    val dbContent = feeHistoriesContentQuery().execute().get()
    val newlyInsertedRows =
      dbContent.filter { it.getLong("created_epoch_milli") == fakeClock.now().toEpochMilliseconds() }
    Assertions.assertThat(newlyInsertedRows.size).isGreaterThan(0)

    newlyInsertedRows.forEachIndexed { i, row ->
      Assertions.assertThat(row.getLong("block_number"))
        .isEqualTo(feeHistory.oldestBlock.toLong() + i)
      Assertions.assertThat(row.getLong("base_fee_per_gas"))
        .isEqualTo(feeHistory.baseFeePerGas[i].toLong())
      Assertions.assertThat(row.getLong("base_fee_per_blob_gas"))
        .isEqualTo(feeHistory.baseFeePerBlobGas.getOrElse(i) { 0uL }.toLong())
      Assertions.assertThat(row.getFloat("gas_used_ratio"))
        .isEqualTo(feeHistory.gasUsedRatio[i].toFloat())
      Assertions.assertThat(row.getFloat("blob_gas_used_ratio"))
        .isEqualTo(feeHistory.blobGasUsedRatio.getOrElse(i) { 0.0 }.toFloat())
      Assertions.assertThat(row.getArrayOfLongs("rewards"))
        .containsAll(feeHistory.reward[i].map { it.toLong() })
      Assertions.assertThat(row.getArrayOfFloats("reward_percentiles"))
        .containsAll(rewardPercentiles.map { it.toFloat() })
    }
    return dbContent
  }

  @Test
  fun `saveNewFeeHistory inserts new fee history to db`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles,
      )
    Assertions.assertThat(dbContent).size().isEqualTo(5)
  }

  @Test
  fun `saveNewFeeHistory upserts new fee history to db`() {
    var dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles,
      )
    Assertions.assertThat(dbContent).size().isEqualTo(5)

    // fee history of block 103, 104, 105, 106, 107
    val overlappedFeeHistory = createFeeHistory(
      oldestBlockNumber = 103UL,
      initialReward = 2000UL,
      initialBaseFeePerGas = 20000UL,
      initialGasUsedRatio = 20U,
      initialBaseFeePerBlobGas = 2000UL,
      initialBlobGasUsedRatio = 10U,
      feeHistoryBlockCount = 5U,
      rewardPercentilesCount = rewardPercentiles.size,
    )
    fakeClock.setTimeTo(Clock.System.now())

    performInsertTest(
      overlappedFeeHistory,
      rewardPercentiles,
    )

    dbContent = feeHistoriesContentQuery().execute().get()
    Assertions.assertThat(dbContent).size().isEqualTo(8)
  }

  @Test
  fun `findBaseFeePerGasAtPercentile returns correct percentile values`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles,
      )
    Assertions.assertThat(dbContent).size().isEqualTo(5)

    val p10BaseFeePerGas = feeHistoriesPostgresDao.findBaseFeePerGasAtPercentile(
      percentile = 10.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(p10BaseFeePerGas).isEqualTo(10000uL)

    val p50BaseFeePerGas = feeHistoriesPostgresDao.findBaseFeePerGasAtPercentile(
      percentile = 50.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(p50BaseFeePerGas).isEqualTo(10002uL)

    val p75BaseFeePerGas = feeHistoriesPostgresDao.findBaseFeePerGasAtPercentile(
      percentile = 75.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(p75BaseFeePerGas).isEqualTo(10003uL)

    val p100BaseFeePerGas = feeHistoriesPostgresDao.findBaseFeePerGasAtPercentile(
      percentile = 100.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(p100BaseFeePerGas).isEqualTo(10004uL)
  }

  @Test
  fun `findBaseFeePerBlobGasAtPercentile returns correct percentile values`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles,
      )
    Assertions.assertThat(dbContent).size().isEqualTo(5)

    val p10BaseFeePerBlobGas = feeHistoriesPostgresDao.findBaseFeePerBlobGasAtPercentile(
      percentile = 10.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(p10BaseFeePerBlobGas).isEqualTo(1000uL)

    val p50BaseFeePerBlobGas = feeHistoriesPostgresDao.findBaseFeePerBlobGasAtPercentile(
      percentile = 50.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(p50BaseFeePerBlobGas).isEqualTo(1002uL)

    val p75BaseFeePerBlobGas = feeHistoriesPostgresDao.findBaseFeePerBlobGasAtPercentile(
      percentile = 75.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(p75BaseFeePerBlobGas).isEqualTo(1003uL)

    val p100BaseFeePerBlobGas = feeHistoriesPostgresDao.findBaseFeePerBlobGasAtPercentile(
      percentile = 100.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(p100BaseFeePerBlobGas).isEqualTo(1004uL)
  }

  @Test
  fun `findAverageRewardAtPercentile returns correct average values`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles,
      )
    Assertions.assertThat(dbContent).size().isEqualTo(5)

    val avgP10Reward = feeHistoriesPostgresDao.findAverageRewardAtPercentile(
      rewardPercentile = 10.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(avgP10Reward).isEqualTo(1002uL)

    val avgP20Reward = feeHistoriesPostgresDao.findAverageRewardAtPercentile(
      rewardPercentile = 20.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(avgP20Reward).isEqualTo(2004uL)

    val avgP70Reward = feeHistoriesPostgresDao.findAverageRewardAtPercentile(
      rewardPercentile = 70.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(avgP70Reward).isEqualTo(7014uL)

    val avgP100Reward = feeHistoriesPostgresDao.findAverageRewardAtPercentile(
      rewardPercentile = 100.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(avgP100Reward).isEqualTo(10020uL)
  }

  @Test
  fun `findAverageRewardAtPercentile returns null with unfound percentile`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles,
      )
    Assertions.assertThat(dbContent).size().isEqualTo(5)

    val avgP15Reward = feeHistoriesPostgresDao.findAverageRewardAtPercentile(
      rewardPercentile = 15.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(avgP15Reward).isNull()
  }

  @Test
  fun `findHighestBlockNumberWithPercentile returns highest block numbers of a given percentile`() {
    var dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles,
      )
    Assertions.assertThat(dbContent).size().isEqualTo(5)

    val rewardPercentile90 = listOf(90.0)
    val feeHistory = createFeeHistory(
      oldestBlockNumber = 105UL,
      initialReward = 1000UL,
      initialBaseFeePerGas = 10000UL,
      initialGasUsedRatio = 70U,
      initialBaseFeePerBlobGas = 1000UL,
      initialBlobGasUsedRatio = 60U,
      feeHistoryBlockCount = 5U,
      rewardPercentilesCount = rewardPercentile90.size,
    )
    fakeClock.setTimeTo(Clock.System.now())

    dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentile90,
      )
    Assertions.assertThat(dbContent).size().isEqualTo(10)

    val p10HighestBlockNumber = feeHistoriesPostgresDao.findHighestBlockNumberWithPercentile(
      rewardPercentile = 10.0,
    ).get()
    Assertions.assertThat(p10HighestBlockNumber).isEqualTo(104L)

    val p20HighestBlockNumber = feeHistoriesPostgresDao.findHighestBlockNumberWithPercentile(
      rewardPercentile = 20.0,
    ).get()
    Assertions.assertThat(p20HighestBlockNumber).isEqualTo(104L)

    val p100HighestBlockNumber = feeHistoriesPostgresDao.findHighestBlockNumberWithPercentile(
      rewardPercentile = 100.0,
    ).get()
    Assertions.assertThat(p100HighestBlockNumber).isEqualTo(104L)

    val p90HighestBlockNumber = feeHistoriesPostgresDao.findHighestBlockNumberWithPercentile(
      rewardPercentile = 90.0,
    ).get()
    Assertions.assertThat(p90HighestBlockNumber).isEqualTo(109L)
  }

  @Test
  fun `findHighestBlockNumberWithPercentile returns null with unfound percentile`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles,
      )
    Assertions.assertThat(dbContent).size().isEqualTo(5)

    val p25HighestBlockNumber = feeHistoriesPostgresDao.findHighestBlockNumberWithPercentile(
      rewardPercentile = 25.0,
    ).get()
    Assertions.assertThat(p25HighestBlockNumber).isNull()
  }

  @Test
  fun `getNumOfFeeHistoriesFromBlockNumber returns number of records within the given percentile and from block`() {
    val dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles,
      )
    Assertions.assertThat(dbContent).size().isEqualTo(5)

    val p10NumOfRecords = feeHistoriesPostgresDao.getNumOfFeeHistoriesFromBlockNumber(
      rewardPercentile = 10.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(p10NumOfRecords).isEqualTo(5)

    val p50NumOfRecords = feeHistoriesPostgresDao.getNumOfFeeHistoriesFromBlockNumber(
      rewardPercentile = 50.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(p50NumOfRecords).isEqualTo(5)

    // unfound reward percentile
    val p75NumOfRecords = feeHistoriesPostgresDao.getNumOfFeeHistoriesFromBlockNumber(
      rewardPercentile = 75.0,
      fromBlockNumber = 100L,
    ).get()
    Assertions.assertThat(p75NumOfRecords).isEqualTo(0)

    // out of block range
    val p20NumOfRecordsOutOfRange = feeHistoriesPostgresDao.getNumOfFeeHistoriesFromBlockNumber(
      rewardPercentile = 20.0,
      fromBlockNumber = 110L,
    ).get()
    Assertions.assertThat(p20NumOfRecordsOutOfRange).isEqualTo(0)
  }

  @Test
  fun `deleteFeeHistoriesUpToBlockNumber deletes number of records below or equal to the given block number`() {
    var dbContent =
      performInsertTest(
        feeHistory,
        rewardPercentiles,
      )
    Assertions.assertThat(dbContent).size().isEqualTo(5)

    var deletedNum = feeHistoriesPostgresDao.deleteFeeHistoriesUpToBlockNumber(
      blockNumberInclusive = 102L,
    ).get()
    Assertions.assertThat(deletedNum).isEqualTo(3)

    dbContent = feeHistoriesContentQuery().execute().get()
    Assertions.assertThat(dbContent).size().isEqualTo(2)

    deletedNum = feeHistoriesPostgresDao.deleteFeeHistoriesUpToBlockNumber(
      blockNumberInclusive = 102L,
    ).get()
    Assertions.assertThat(deletedNum).isEqualTo(0)

    dbContent = feeHistoriesContentQuery().execute().get()
    Assertions.assertThat(dbContent).size().isEqualTo(2)

    deletedNum = feeHistoriesPostgresDao.deleteFeeHistoriesUpToBlockNumber(
      blockNumberInclusive = 110L,
    ).get()
    Assertions.assertThat(deletedNum).isEqualTo(2)

    dbContent = feeHistoriesContentQuery().execute().get()
    Assertions.assertThat(dbContent).size().isEqualTo(0)
  }
}
