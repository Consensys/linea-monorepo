package net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap

import net.consensys.toGWei
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapCalculator
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.math.BigDecimal
import java.math.BigInteger
import kotlin.math.pow
import kotlin.time.Duration

class GasPriceCapCalculatorImpl : GasPriceCapCalculator {
  private val log: Logger = LogManager.getLogger(this::class.java)

  private fun calculateOverallMultiplier(
    adjustmentConstant: UInt,
    finalizationTargetMaxDelay: Duration,
    elapsedTimeSinceBlockTimestamp: Duration,
    timeOfDayMultiplier: Double
  ): Double {
    val finalizationDelayMultiplier = elapsedTimeSinceBlockTimestamp.inWholeSeconds.toDouble()
      .div(finalizationTargetMaxDelay.inWholeSeconds.toDouble()).pow(2)

    return 1.0 +
      adjustmentConstant.toDouble().times(timeOfDayMultiplier).times(finalizationDelayMultiplier)
  }

  override fun calculateGasPriceCap(
    adjustmentConstant: UInt,
    finalizationTargetMaxDelay: Duration,
    baseFeePerGasAtPercentile: BigInteger,
    elapsedTimeSinceBlockTimestamp: Duration,
    avgRewardAtPercentile: BigInteger,
    timeOfDayMultiplier: Double
  ): BigInteger {
    require(finalizationTargetMaxDelay > Duration.ZERO) {
      "finalizationTargetMaxDelay duration must be longer than zero second." +
        " Value=$finalizationTargetMaxDelay"
    }

    val overallMultiplier = calculateOverallMultiplier(
      adjustmentConstant,
      finalizationTargetMaxDelay,
      elapsedTimeSinceBlockTimestamp,
      timeOfDayMultiplier
    )

    val gasPriceCap = (
      baseFeePerGasAtPercentile.toBigDecimal()
        .times(BigDecimal.valueOf(overallMultiplier))
      ).toBigInteger()
      .add(avgRewardAtPercentile)

    log.debug(
      "Gas price cap calculation: gasPriceCap={} GWei," +
        " baseFeePerGasAtPercentile={} GWei, adjustmentConstant={}" +
        " finalizationTargetMaxDelay={} sec, elapsedTimeSinceBlockTimestamp={} sec," +
        " avgRewardAtPercentile={} GWei, timeOfDayMultiplier={} overallMultiplier={}",
      gasPriceCap.toGWei(),
      baseFeePerGasAtPercentile.toGWei(),
      adjustmentConstant,
      finalizationTargetMaxDelay.inWholeSeconds,
      elapsedTimeSinceBlockTimestamp.inWholeSeconds,
      avgRewardAtPercentile.toGWei(),
      timeOfDayMultiplier,
      overallMultiplier
    )

    return gasPriceCap
  }
}
