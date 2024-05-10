package net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap

import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import kotlinx.datetime.TimeZone
import kotlinx.datetime.toLocalDateTime
import net.consensys.toGWei
import net.consensys.toIntervalString
import net.consensys.toULong
import net.consensys.zkevm.coordinator.blockcreation.ExtendedWeb3J
import net.consensys.zkevm.ethereum.gaspricing.FeeHistoriesRepositoryWithCache
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapCalculator
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapProvider
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCaps
import net.consensys.zkevm.ethereum.gaspricing.TimeOfDayMultipliers
import net.consensys.zkevm.ethereum.gaspricing.getTimeOfDayKey
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.Web3j
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import kotlin.time.Duration

class GasPriceCapProviderImpl(
  private val config: Config,
  private val l1Web3jClient: Web3j,
  private val l2ExtendedWeb3JClient: ExtendedWeb3J,
  private val feeHistoriesRepository: FeeHistoriesRepositoryWithCache,
  private val gasPriceCapCalculator: GasPriceCapCalculator,
  private val clock: Clock = Clock.System
) : GasPriceCapProvider {
  data class Config(
    val enabled: Boolean,
    val maxFeePerGasCap: BigInteger,
    val maxFeePerBlobGasCap: BigInteger,
    val baseFeePerGasPercentile: Double,
    val baseFeePerGasPercentileWindowInBlocks: UInt,
    val baseFeePerGasPercentileWindowLeewayInBlocks: UInt,
    val timeOfDayMultipliers: TimeOfDayMultipliers,
    val adjustmentConstant: UInt,
    val blobAdjustmentConstant: UInt,
    val finalizationTargetMaxDelay: Duration
  )

  private val log: Logger = LogManager.getLogger(this::class.java)
  private val defaultGasPriceCaps: GasPriceCaps = GasPriceCaps(
    maxPriorityFeePerGasCap = config.maxFeePerGasCap,
    maxFeePerGasCap = config.maxFeePerGasCap,
    maxFeePerBlobGasCap = config.maxFeePerBlobGasCap
  )

  init {
    require(config.baseFeePerGasPercentile >= 0.0) {
      "baseFeePerGasPercentile must be no less than 0.0." +
        " Value=${config.baseFeePerGasPercentile}"
    }

    require(config.maxFeePerGasCap >= BigInteger.ZERO) {
      "maxFeePerGasCap must be no less than 0. Value=${config.maxFeePerGasCap}"
    }

    require(config.maxFeePerBlobGasCap >= BigInteger.ZERO) {
      "maxFeePerBlobGasCap must be no less than 0. Value=${config.maxFeePerBlobGasCap}"
    }

    require(config.finalizationTargetMaxDelay > Duration.ZERO) {
      "finalizationTargetMaxDelay duration must be longer than zero second." +
        " Value=${config.finalizationTargetMaxDelay}"
    }
  }

  private fun isEnoughDataForGasPriceCapCalculation(): Boolean {
    val minNumOfFeeHistoriesNeeded = BigInteger.valueOf(config.baseFeePerGasPercentileWindowInBlocks.toLong())
      .minus(BigInteger.valueOf(config.baseFeePerGasPercentileWindowLeewayInBlocks.toLong()))
      .coerceAtLeast(BigInteger.ZERO).toULong()

    val numOfValidFeeHistories = feeHistoriesRepository.getCachedNumOfFeeHistoriesFromBlockNumber()
    val isEnoughData = numOfValidFeeHistories.toULong() >= minNumOfFeeHistoriesNeeded
    if (!isEnoughData) {
      log.warn(
        "Not enough fee history data for gas price cap update: numOfValidFeeHistoriesInDb={}, " +
          "minNumOfFeeHistoriesNeeded={}",
        numOfValidFeeHistories,
        minNumOfFeeHistoriesNeeded
      )
    }
    return isEnoughData
  }

  private fun getElapsedTimeSinceBlockTimestamp(blockTimestamp: Instant): Duration {
    return (clock.now() - blockTimestamp).coerceAtLeast(Duration.ZERO)
  }

  private fun getTimeOfDayMultiplierForNow(timeOfDayMultipliers: TimeOfDayMultipliers): Double {
    val dateTime = clock.now().toLocalDateTime(TimeZone.UTC)
    val tdmKey = getTimeOfDayKey(dateTime.dayOfWeek, dateTime.hour)
    return timeOfDayMultipliers[tdmKey]!!
  }

  private fun calculateGasPriceCaps(
    latestL1BlockNumber: Long,
    targetL2BlockNumber: Long
  ): SafeFuture<GasPriceCaps> {
    return if (isEnoughDataForGasPriceCapCalculation()) {
      l2ExtendedWeb3JClient.ethGetExecutionPayloadByNumber(targetL2BlockNumber).thenApply {
        val targetL2BlockTimestamp = Instant.fromEpochSeconds(it.timestamp.longValue())
        val elapsedTimeSinceBlockTimestamp = getElapsedTimeSinceBlockTimestamp(targetL2BlockTimestamp)
        val percentileGasFees = feeHistoriesRepository.getCachedPercentileGasFees()
        val maxFeePerGasCap = gasPriceCapCalculator.calculateGasPriceCap(
          adjustmentConstant = config.adjustmentConstant,
          finalizationTargetMaxDelay = config.finalizationTargetMaxDelay,
          baseFeePerGasAtPercentile = percentileGasFees.percentileBaseFeePerGas,
          elapsedTimeSinceBlockTimestamp = elapsedTimeSinceBlockTimestamp,
          avgRewardAtPercentile = percentileGasFees.percentileAvgReward,
          timeOfDayMultiplier = getTimeOfDayMultiplierForNow(config.timeOfDayMultipliers)
        )
        val maxFeePerBlobGasCap = gasPriceCapCalculator.calculateGasPriceCap(
          adjustmentConstant = config.blobAdjustmentConstant,
          finalizationTargetMaxDelay = config.finalizationTargetMaxDelay,
          baseFeePerGasAtPercentile = percentileGasFees.percentileBaseFeePerBlobGas,
          elapsedTimeSinceBlockTimestamp = elapsedTimeSinceBlockTimestamp
        )
        GasPriceCaps(
          maxPriorityFeePerGasCap = maxFeePerGasCap,
          maxFeePerGasCap = maxFeePerGasCap,
          maxFeePerBlobGasCap = maxFeePerBlobGasCap
        )
      }.thenPeek { gasPriceCaps ->
        val l1Blocks =
          (latestL1BlockNumber - config.baseFeePerGasPercentileWindowInBlocks.toLong())
            .coerceAtLeast(0L).inc()..latestL1BlockNumber
        log.debug(
          "Calculated raw gas price caps: " +
            "maxFeePerGasCap={} GWei, maxPriorityFeePerGasCap={} GWei, maxFeePerBlobGasCap={} GWei, " +
            "l1Blocks={} percentile={}",
          gasPriceCaps.maxFeePerGasCap.toGWei(),
          gasPriceCaps.maxPriorityFeePerGasCap.toGWei(),
          gasPriceCaps.maxFeePerBlobGasCap.toGWei(),
          l1Blocks.toIntervalString(),
          config.baseFeePerGasPercentile
        )
      }.exceptionallyCompose { th ->
        SafeFuture.completedFuture(defaultGasPriceCaps).thenPeek {
          log.error(
            "Gas price caps reset to default due to failure occurred: " +
              "maxFeePerGasCap={} GWei, maxPriorityFeePerGasCap={} GWei, maxFeePerBlobGasCap={} GWei," +
              "errorMessage={}",
            it.maxFeePerGasCap.toGWei(),
            it.maxPriorityFeePerGasCap.toGWei(),
            it.maxFeePerBlobGasCap.toGWei(),
            th.message,
            th
          )
        }
      }
    } else {
      SafeFuture.completedFuture(defaultGasPriceCaps)
    }
  }

  private fun calculateGasPriceCaps(targetL2BlockNumber: Long): SafeFuture<GasPriceCaps> {
    return if (config.enabled) {
      SafeFuture.of(l1Web3jClient.ethBlockNumber().sendAsync())
        .thenCompose {
          calculateGasPriceCaps(
            latestL1BlockNumber = it.blockNumber.toLong(),
            targetL2BlockNumber = targetL2BlockNumber
          )
        }
    } else {
      SafeFuture.completedFuture(defaultGasPriceCaps)
    }
  }

  private fun coerceGasPriceCaps(gasPriceCaps: GasPriceCaps): GasPriceCaps {
    return GasPriceCaps(
      maxFeePerGasCap = gasPriceCaps.maxFeePerGasCap.coerceAtMost(config.maxFeePerGasCap)
        .run { if (this <= BigInteger.ZERO) config.maxFeePerGasCap else this },
      maxPriorityFeePerGasCap = gasPriceCaps.maxPriorityFeePerGasCap.coerceAtMost(config.maxFeePerGasCap)
        .run { if (this <= BigInteger.ZERO) config.maxFeePerGasCap else this },
      maxFeePerBlobGasCap = gasPriceCaps.maxFeePerBlobGasCap.coerceAtMost(config.maxFeePerBlobGasCap)
        .run { if (this <= BigInteger.ZERO) config.maxFeePerBlobGasCap else this }
    )
  }

  override fun getGasPriceCaps(targetL2BlockNumber: Long): SafeFuture<GasPriceCaps> {
    return calculateGasPriceCaps(targetL2BlockNumber)
      .thenApply(this::coerceGasPriceCaps)
  }

  override fun getDefaultGasPriceCaps(): GasPriceCaps {
    return defaultGasPriceCaps
  }
}
