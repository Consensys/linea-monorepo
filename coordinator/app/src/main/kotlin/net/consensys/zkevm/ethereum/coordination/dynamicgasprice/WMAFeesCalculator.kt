package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import net.consensys.linea.FeeHistory
import net.consensys.linea.toIntervalString
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.math.BigDecimal
import java.math.BigInteger
import java.math.MathContext

/**
 * Weighted Average Fees Calculator
 * WMA Gasfee = (lastBaseFee * baseFeeCoefficient) +
 *  (sumOf(reward[i] * ratio[i] * (i+1)) / sumOf(ratio[i] * (i+1))) * priorityFeeWmaCoefficient
 */
class WMAFeesCalculator(
  private val config: Config
) : FeesCalculator {
  data class Config(
    val baseFeeCoefficient: BigDecimal,
    val priorityFeeWmaCoefficient: BigDecimal
  )

  private val log: Logger = LogManager.getLogger(this::class.java)

  override fun calculateFees(feeHistory: FeeHistory): BigInteger {
    val lastBlockBaseFeePerGas =
      BigDecimal(feeHistory.baseFeePerGas[feeHistory.baseFeePerGas.lastIndex])
    val scaledBaseFee = lastBlockBaseFeePerGas.multiply(this.config.baseFeeCoefficient)
    val weightedSumOfRewards = feeHistory.reward
      .mapIndexed { index, rewards ->
        feeHistory.gasUsedRatio[index]
          .multiply(BigDecimal(index + 1))
          .multiply(BigDecimal(rewards[0]))
      }
      .reduce { acc, reward -> acc.add(reward) }

    val weightedRatiosSum = feeHistory.gasUsedRatio
      .mapIndexed { index, gasUsedRatio -> gasUsedRatio.multiply(BigDecimal(index + 1)) }
      .reduce { acc, bigDecimal -> acc.add(bigDecimal) }

    try {
      val weightedPriorityFee = when {
        weightedSumOfRewards.compareTo(BigDecimal.ZERO) == 0 ||
          weightedRatiosSum.compareTo(BigDecimal.ZERO) == 0 -> BigDecimal.ZERO

        else ->
          weightedSumOfRewards
            .divide(weightedRatiosSum, MathContext.DECIMAL128)
            .multiply(this.config.priorityFeeWmaCoefficient)
      }

      val calculatedL2GasPrice = scaledBaseFee.add(weightedPriorityFee).toBigInteger()

      log.debug(
        "Calculated gasPrice={} wei, l1Blocks={}",
        calculatedL2GasPrice,
        feeHistory.blocksRange().toIntervalString()
      )
      return calculatedL2GasPrice
    } catch (e: Exception) {
      log.error(
        "Error: weightedSumOfRewards={}, weightedRatiosSum={}, feehistory={}, errorMessage={}",
        weightedSumOfRewards,
        weightedRatiosSum,
        feeHistory,
        e.message,
        e
      )
      throw e
    }
  }
}
