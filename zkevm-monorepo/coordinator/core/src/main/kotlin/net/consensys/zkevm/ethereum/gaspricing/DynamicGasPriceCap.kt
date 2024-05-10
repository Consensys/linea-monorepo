package net.consensys.zkevm.ethereum.gaspricing

import kotlinx.datetime.DayOfWeek
import net.consensys.linea.FeeHistory
import net.consensys.toGWei
import net.consensys.zkevm.persistence.feehistory.FeeHistoriesRepository
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import kotlin.time.Duration
import kotlin.time.Duration.Companion.days

const val L1_BLOCK_TIME_SECONDS = 12L
const val MAX_REWARD_PERCENTILES_SIZE = 10
const val MAX_FEE_HISTORY_BLOCK_COUNT = 1024U
val MAX_FEE_HISTORIES_STORAGE_PERIOD = 30.days

data class PercentileGasFees(
  val percentileBaseFeePerGas: BigInteger,
  val percentileBaseFeePerBlobGas: BigInteger,
  val percentileAvgReward: BigInteger
) {
  override fun toString(): String {
    return "percentileBaseFeePerGas=${percentileBaseFeePerGas.toGWei()} GWei," +
      " percentileBaseFeePerBlobGas=${percentileBaseFeePerBlobGas.toGWei()} GWei" +
      " percentileAvgReward=${percentileAvgReward.toGWei()} GWei"
  }
}

data class GasPriceCaps(
  val maxPriorityFeePerGasCap: BigInteger,
  val maxFeePerGasCap: BigInteger,
  val maxFeePerBlobGasCap: BigInteger
) {
  override fun toString(): String {
    return "maxPriorityFeePerGasCap=${maxPriorityFeePerGasCap.toGWei()} GWei," +
      " maxFeePerGasCap=${maxFeePerGasCap.toGWei()} GWei" +
      " maxFeePerBlobGasCap=${maxFeePerBlobGasCap.toGWei()} GWei"
  }
}

typealias TimeOfDayMultipliers = Map<String, Double>

fun getTimeOfDayKey(dayOfWeek: DayOfWeek, hour: Int): String {
  return "${dayOfWeek}_$hour"
}

fun getAllTimeOfDayKeys(): Set<String> {
  return DayOfWeek.values()
    .flatMap { dayOfWeek ->
      (0..23).map { hour -> getTimeOfDayKey(dayOfWeek, hour) }
    }.toSet()
}

interface GasPriceCapFeeHistoryFetcher {
  fun getEthFeeHistoryData(startBlockNumberInclusive: Long, endBlockNumberInclusive: Long): SafeFuture<FeeHistory>
}

interface GasPriceCapCalculator {
  fun calculateGasPriceCap(
    adjustmentConstant: UInt,
    finalizationTargetMaxDelay: Duration,
    baseFeePerGasAtPercentile: BigInteger,
    elapsedTimeSinceBlockTimestamp: Duration,
    avgRewardAtPercentile: BigInteger = BigInteger.ZERO,
    timeOfDayMultiplier: Double = 1.0
  ): BigInteger
}

interface GasPriceCapProvider {
  fun getDefaultGasPriceCaps(): GasPriceCaps

  fun getGasPriceCaps(targetL2BlockNumber: Long): SafeFuture<GasPriceCaps>
}

interface GasPriceCapFeeHistoryCache {
  fun getCachedNumOfFeeHistoriesFromBlockNumber(): Int

  fun cacheNumOfFeeHistoriesFromBlockNumber(
    rewardPercentile: Double,
    fromBlockNumber: Long
  ): SafeFuture<Int>

  fun getCachedPercentileGasFees(): PercentileGasFees

  fun cachePercentileGasFees(
    percentile: Double,
    fromBlockNumber: Long
  ): SafeFuture<Unit>
}

interface FeeHistoriesRepositoryWithCache : FeeHistoriesRepository, GasPriceCapFeeHistoryCache
