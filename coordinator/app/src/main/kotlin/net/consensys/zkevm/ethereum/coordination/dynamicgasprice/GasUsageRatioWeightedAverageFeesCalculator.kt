package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import net.consensys.linea.FeeHistory
import net.consensys.linea.toIntervalString
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.math.BigDecimal
import java.math.BigInteger
import java.math.MathContext

/**
 * Gas Ratio Weighted Average Fees Calculator
 * L2GasPrice = sumOf((baseFee[i]*baseFeeCoefficient + reward[i]*priorityFeeWmaCoefficient)*ratio[i])/sumOf(ratio[i])
 */
class GasUsageRatioWeightedAverageFeesCalculator(
  val config: Config
) : FeesCalculator {
  data class Config(
    val baseFeeCoefficient: BigDecimal,
    val priorityFeeCoefficient: BigDecimal
  )

  private val log: Logger = LogManager.getLogger(this::class.java)

  override fun calculateFees(feeHistory: FeeHistory): BigInteger {
    val baseFeePerGasList = feeHistory.baseFeePerGas.map { BigDecimal(it) }
    val priorityFeesPerGasList = feeHistory.reward.map { BigDecimal(it[0]) }
    var gasUsageRatioList = feeHistory.gasUsedRatio
    var gasUsageRatiosSum = gasUsageRatioList.sumOf { it }

    if (gasUsageRatiosSum.compareTo(BigDecimal.ZERO) == 0) {
      log.warn(
        "GasUsedRatio is zero for all l1Blocks={}. Will fallback to Simple Average.",
        feeHistory.blocksRange().toIntervalString()
      )
      // Giving a weight of one will yield the simple average
      gasUsageRatioList = feeHistory.gasUsedRatio.map { BigDecimal.ONE }
      gasUsageRatiosSum = BigDecimal(feeHistory.gasUsedRatio.size)
    }

    val weightedFeesSum = gasUsageRatioList.mapIndexed { index, gasUsedRatio ->
      val baseFeeFractional = baseFeePerGasList[index].multiply(this.config.baseFeeCoefficient)
      val priorityFeeFractional = priorityFeesPerGasList[index].multiply(this.config.priorityFeeCoefficient)

      baseFeeFractional.add(priorityFeeFractional).multiply(gasUsedRatio)
    }.reduce { acc, bigDecimal -> acc.add(bigDecimal) }

    val l2GasPrice = weightedFeesSum
      .divide(gasUsageRatiosSum, MathContext.DECIMAL128)
      .setScale(0, java.math.RoundingMode.HALF_UP)
      .toBigInteger()

    log.debug(
      "Calculated gasPrice={} wei, l1Blocks={}",
      l2GasPrice,
      feeHistory.blocksRange().toIntervalString()
    )
    return l2GasPrice
  }
}
