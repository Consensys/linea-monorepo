package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import net.consensys.linea.FeeHistory
import net.consensys.toIntervalString
import net.consensys.zkevm.ethereum.gaspricing.FeesCalculator
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.math.BigDecimal
import java.math.BigInteger
import java.math.MathContext
import java.math.RoundingMode

/**
 * Gas Ratio Weighted Average Fees Calculator
 * L2BaseGasPrice = (sumOf((baseFee[i]*baseFeeCoefficient + reward[i]*priorityFeeWmaCoefficient)*ratio[i]))/sumOf(ratio[i])
 * WeightedL2BlobGasPrice = (sumOf((baseBlobFee[i]*baseBlobFeeCoefficient)*blobRatio[i]))/sumOf(blobRatio[i]) * (131072 / 120_000)
 * L2FinalGasPrice = L2BaseGasPrice + WeightedL2BlobGasPrice
 */
class GasUsageRatioWeightedAverageFeesCalculator(
  val config: Config
) : FeesCalculator {
  data class Config(
    val baseFeeCoefficient: BigDecimal,
    val priorityFeeCoefficient: BigDecimal,
    val baseFeeBlobCoefficient: BigDecimal,
    val blobSubmissionExpectedExecutionGas: BigDecimal,
    val expectedBlobGas: BigDecimal
  )

  private val log: Logger = LogManager.getLogger(this::class.java)

  override fun calculateFees(feeHistory: FeeHistory): BigInteger {
    val baseFeePerGasList = feeHistory.baseFeePerGas.map { BigDecimal(it) }
    val priorityFeesPerGasList = feeHistory.reward.map { BigDecimal(it[0]) }
    val baseFeePerBlobGasList = if (feeHistory.baseFeePerBlobGas.isEmpty()) {
      List(baseFeePerGasList.size) { BigDecimal.ONE }
    } else {
      feeHistory.baseFeePerBlobGas.map { BigDecimal(it) }
    }

    val executionGasPrice = calculateExecutionGasPrice(feeHistory, baseFeePerGasList, priorityFeesPerGasList)
    val weightedL2BlobGasPrice = calculateWeightedBlobGasPrice(feeHistory, baseFeePerBlobGasList)

    val l2GasPrice = executionGasPrice + weightedL2BlobGasPrice

    log.debug(
      "Calculated gasPrice={} wei, executionGasPrice={} wei, weightedL2BlobGasPrice={} wei, l1Blocks={} feeHistory={}",
      l2GasPrice,
      executionGasPrice,
      weightedL2BlobGasPrice,
      feeHistory.blocksRange().toIntervalString(),
      feeHistory
    )
    return l2GasPrice
  }

  private fun calculateExecutionGasPrice(
    feeHistory: FeeHistory,
    baseFeePerGasList: List<BigDecimal>,
    priorityFeesPerGasList: List<BigDecimal>
  ): BigInteger {
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

    val executionGasPrice = weightedFeesSum
      .divide(gasUsageRatiosSum, MathContext.DECIMAL128)
      .setScale(0, RoundingMode.HALF_UP)
      .toBigInteger()
    return executionGasPrice
  }

  private fun calculateWeightedBlobGasPrice(
    feeHistory: FeeHistory,
    baseFeePerBlobGasList: List<BigDecimal>
  ): BigInteger {
    val blobGasUsageRatio = if (feeHistory.blobGasUsedRatio.sumOf { it }.compareTo(BigDecimal.ZERO) == 0) {
      log.warn(
        "BlobGasUsedRatio is zero for all l1Blocks={}. Will fallback to Simple Average.",
        feeHistory.blocksRange().toIntervalString()
      )
      List(feeHistory.gasUsedRatio.size) { BigDecimal.ONE }
    } else {
      feeHistory.blobGasUsedRatio
    }

    val blobGasUsageRatiosSum = blobGasUsageRatio.sumOf { it }

    val weightedBlobFeesSum = blobGasUsageRatio.zip(baseFeePerBlobGasList) { blobGasUsedRatio, baseFeePerBlobGas ->
      baseFeePerBlobGas.multiply(blobGasUsedRatio)
    }.reduce { acc, bigDecimal -> acc.add(bigDecimal) }.multiply(this.config.baseFeeBlobCoefficient)

    val weightedL2BlobGasPrice = weightedBlobFeesSum
      .divide(blobGasUsageRatiosSum, MathContext.DECIMAL128)
      .setScale(0, RoundingMode.HALF_UP)
      .multiply((config.expectedBlobGas / config.blobSubmissionExpectedExecutionGas))
      .toBigInteger()

    return weightedL2BlobGasPrice
  }
}
