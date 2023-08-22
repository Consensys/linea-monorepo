package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import net.consensys.linea.FeeHistory
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.Test
import java.math.BigDecimal
import java.math.BigInteger

class GasUsageRatioWeightedAverageFeesCalculatorTest {
  val feesCalculator =
    GasUsageRatioWeightedAverageFeesCalculator(
      GasUsageRatioWeightedAverageFeesCalculator.Config(
        baseFeeCoefficient = BigDecimal("0.1"),
        priorityFeeCoefficient = BigDecimal("1.2")
      )
    )

  @Test
  fun calculateFees_singleBlockHistory() {
    val feeHistory = FeeHistory(
      oldestBlock = BigInteger.valueOf(100),
      baseFeePerGas = listOf(100, 110).map { BigInteger.valueOf(it.toLong()) },
      reward = listOf(1000).map { listOf(BigInteger.valueOf(it.toLong())) },
      gasUsedRatio = listOf(0.25.toBigDecimal())
    )
    // (100*0.1 + 1000*1.2)*0.25/0.25 = 1210
    val calculatedL2GasPrice = feesCalculator.calculateFees(feeHistory)
    Assertions.assertThat(calculatedL2GasPrice).isEqualTo(BigInteger.valueOf(1210))
  }

  @Test
  fun calculateFees_singleBlockHistory_zeroUsageRatio() {
    val feeHistory = FeeHistory(
      oldestBlock = BigInteger.valueOf(100),
      baseFeePerGas = listOf(100, 110).map { BigInteger.valueOf(it.toLong()) },
      reward = listOf(1000).map { listOf(BigInteger.valueOf(it.toLong())) },
      gasUsedRatio = listOf(0.0.toBigDecimal())
    )
    // (100*0.1 + 1000*1.2) = 1210
    val calculatedL2GasPrice = feesCalculator.calculateFees(feeHistory)
    Assertions.assertThat(calculatedL2GasPrice).isEqualTo(BigInteger.valueOf(1210))
  }

  @Test
  fun calculateFees_MultipleBlockHistory() {
    val feeHistory = FeeHistory(
      oldestBlock = BigInteger.valueOf(100),
      baseFeePerGas = listOf(100, 110, 120, 130, 140).map { BigInteger.valueOf(it.toLong()) },
      reward = listOf(1000, 1100, 1200, 1300).map { listOf(BigInteger.valueOf(it.toLong())) },
      gasUsedRatio = listOf(0.25.toBigDecimal(), 0.5.toBigDecimal(), 0.75.toBigDecimal(), 0.9.toBigDecimal())
    )

    // Weighted Average:
    // ((100*0.1 + 1000*1.2)*0.25 + (110*0.1 + 1100*1.2)*0.50 + (120*0.1 + 1200*1.2)*0.75 + (130*0.1 + 1300*1.2)*0.9)/(0.25+0.5+0.75+0.9) = 1446.95833333
    val calculatedL2GasPrice = feesCalculator.calculateFees(feeHistory)
    Assertions.assertThat(calculatedL2GasPrice).isEqualTo(BigInteger.valueOf(1447))
  }

  @Test
  fun calculateFees_equalUsageRatio_shouldEqualSimpleAverage() {
    val feeHistory = FeeHistory(
      oldestBlock = BigInteger.valueOf(100),
      baseFeePerGas = listOf(100, 110, 120, 130, 140).map { BigInteger.valueOf(it.toLong()) },
      reward = listOf(1000, 1100, 1200, 1300).map { listOf(BigInteger.valueOf(it.toLong())) },
      gasUsedRatio = listOf(0.5.toBigDecimal(), 0.5.toBigDecimal(), 0.5.toBigDecimal(), 0.5.toBigDecimal())
    )

    // Simple Average:
    // ((100*0.1 + 1000*1.2)*0.5 + (110*0.1 + 1100*1.2)*0.5 + (120*0.1 + 1200*1.2)*0.5 + (130*0.1 + 1300*1.2)*0.5)/(0.5+0.5+0.5+0.5) = 1391.5
    val calculatedL2GasPrice = feesCalculator.calculateFees(feeHistory)
    Assertions.assertThat(calculatedL2GasPrice).isEqualTo(BigInteger.valueOf(1392))
  }

  @Test
  fun calculateFees_zeroUsage_emptyL1_blocks_should_return_SimpleAverage() {
    val feeHistory = FeeHistory(
      oldestBlock = BigInteger.valueOf(100),
      baseFeePerGas = listOf(100, 110, 120, 130, 140).map { BigInteger.valueOf(it.toLong()) },
      reward = listOf(1000, 1100, 1200, 1300).map { listOf(BigInteger.valueOf(it.toLong())) },
      gasUsedRatio = listOf(0.0.toBigDecimal(), 0.0.toBigDecimal(), 0.0.toBigDecimal(), 0.0.toBigDecimal())
    )

    // Simple Average:
    // ((100*0.1 + 1000*1.2)*1 + (110*0.1 + 1100*1.2)*1 + (120*0.1 + 1200*1.2)*1 + (130*0.1 + 1300*1.2)*1)/4) = 1391.5
    val calculatedL2GasPrice = feesCalculator.calculateFees(feeHistory)
    Assertions.assertThat(calculatedL2GasPrice).isEqualTo(BigInteger.valueOf(1392))
  }

  @Test
  fun calculateFees_partialZeroUsage_emptyL1_blocks_should_return_SimpleAverage() {
    // Very unlikely scenario in L1 network with activity.
    // Only happen for local dev or test networks, so we don't need to optimize for it.
    val feeHistory = FeeHistory(
      oldestBlock = BigInteger.valueOf(100),
      baseFeePerGas = listOf(100, 110, 120, 130, 140).map { BigInteger.valueOf(it.toLong()) },
      reward = listOf(1000, 1100, 1200, 1300).map { listOf(BigInteger.valueOf(it.toLong())) },
      gasUsedRatio = listOf(0.75.toBigDecimal(), 0.0.toBigDecimal(), 0.75.toBigDecimal(), 0.0.toBigDecimal())
    )

    // Weighted Average:
    // ((100*0.1 + 1000*1.2)*0.75 + (110*0.1 + 1100*1.2)*0.0 + (120*0.1 + 1200*1.2)*0.75 + (130*0.1 + 1300*1.2)*0.0)/(0.75+0.0+0.75+0.0) = 1331
    val calculatedL2GasPrice = feesCalculator.calculateFees(feeHistory)
    Assertions.assertThat(calculatedL2GasPrice).isEqualTo(BigInteger.valueOf(1331))
  }
}
