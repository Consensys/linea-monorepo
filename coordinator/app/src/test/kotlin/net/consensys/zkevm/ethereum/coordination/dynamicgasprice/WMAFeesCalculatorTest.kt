package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import io.vertx.junit5.VertxExtension
import net.consensys.linea.FeeHistory
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import java.math.BigDecimal
import java.math.BigInteger

@ExtendWith(VertxExtension::class)
class WMAFeesCalculatorTest {
  @Test
  fun `calculateFees weighted moving average of priority fee`() {
    val wmaFeesCalculator =
      WMAFeesCalculator(
        WMAFeesCalculator.Config(
          baseFeeCoefficient = BigDecimal("0.1"),
          priorityFeeWmaCoefficient = BigDecimal("1.2")
        )
      )

    val feeHistory = FeeHistory(
      oldestBlock = BigInteger.valueOf(100),
      baseFeePerGas = listOf(100, 110, 120, 130, 140).map { BigInteger.valueOf(it.toLong()) },
      reward = listOf(1000, 1100, 1200, 1300).map { listOf(BigInteger.valueOf(it.toLong())) },
      gasUsedRatio = listOf(0.25.toBigDecimal(), 0.5.toBigDecimal(), 0.75.toBigDecimal(), 0.9.toBigDecimal())
    )

    // WMA = (140*0.1) + ((1000*0.25*1 + 1100*0.5*2 + 1200*0.75*3 + 1300*0.9*4) / (0.25*1 + 0.5*2 + 0.75*3 + 0.9*4)) * 1.2 = 1489.49295
    val calculatedL2GasPrice = wmaFeesCalculator.calculateFees(feeHistory)
    assertThat(calculatedL2GasPrice).isEqualTo(BigInteger.valueOf(1489))
  }

  @Test
  fun `calculateFees supports empty usage ratio`() {
    val wmaFeesCalculator =
      WMAFeesCalculator(
        WMAFeesCalculator.Config(
          baseFeeCoefficient = BigDecimal("0.1"),
          priorityFeeWmaCoefficient = BigDecimal("1.2")
        )
      )

    val feeHistory = FeeHistory(
      oldestBlock = BigInteger.valueOf(100),
      baseFeePerGas = listOf(100, 110, 120, 130, 140).map { BigInteger.valueOf(it.toLong()) },
      reward = listOf(1000, 1100, 1200, 1300).map { listOf(BigInteger.valueOf(it.toLong())) },
      gasUsedRatio = listOf(0.0.toBigDecimal(), 0.0.toBigDecimal(), 0.0.toBigDecimal(), 0.0.toBigDecimal())
    )

    // WMA = (140*0.1) = 14.0
    val calculatedL2GasPrice = wmaFeesCalculator.calculateFees(feeHistory)
    assertThat(calculatedL2GasPrice).isEqualTo(BigInteger.valueOf(14))
  }
}
