package net.consensys.linea.contract

import net.consensys.linea.FeeHistory
import net.consensys.zkevm.ethereum.gaspricing.FeesCalculator
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.mockito.kotlin.any
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.eq
import org.mockito.kotlin.mock
import java.math.BigInteger

class BoundableFeeCalculatorTest {
  private val feeHistory = FeeHistory(
    oldestBlock = BigInteger.valueOf(100),
    baseFeePerGas = listOf(100, 110, 120, 130, 140).map { it.toBigInteger() },
    reward = listOf(1000, 1100, 1200, 1300).map { listOf(it.toBigInteger()) },
    gasUsedRatio = listOf(0.25.toBigDecimal(), 0.5.toBigDecimal(), 0.75.toBigDecimal(), 0.9.toBigDecimal()),
    baseFeePerBlobGas = listOf(100, 110, 120, 130, 140).map { it.toBigInteger() },
    blobGasUsedRatio = listOf(0.25.toBigDecimal(), 0.5.toBigDecimal(), 0.75.toBigDecimal(), 0.9.toBigDecimal())
  )

  @Test
  fun calculateFees_fixedPriceAdded() {
    val gasPriceUpperBound = BigInteger("9000000000000")
    val gasPriceLowerBound = BigInteger("1000000")
    val gasPriceFixedCost = BigInteger("17")
    val calculatedL2GasPrice = gasPriceUpperBound.minus(BigInteger("100"))
    val expectedGasPrice = calculatedL2GasPrice + gasPriceFixedCost
    val mockFeesCalculator = mock<FeesCalculator> {
      on { calculateFees(eq(feeHistory)) } doReturn calculatedL2GasPrice
    }
    val boundableFeeCalculator = BoundableFeeCalculator(
      BoundableFeeCalculator.Config(gasPriceUpperBound, gasPriceLowerBound, gasPriceFixedCost),
      mockFeesCalculator
    )
    Assertions.assertThat(boundableFeeCalculator.calculateFees(feeHistory)).isEqualTo(expectedGasPrice)
  }

  @Test
  fun calculateFess_upperBound() {
    val gasPriceUpperBound = BigInteger("10000000000")
    val gasPriceLowerBound = BigInteger("1000000")
    val gasPriceFixedCost = BigInteger("100")
    val calculatedL2GasPrice = gasPriceUpperBound.minus(BigInteger("50"))
    val expectedGasPrice = gasPriceUpperBound

    val mockFeesCalculator = mock<FeesCalculator> {
      on { calculateFees(any()) } doReturn calculatedL2GasPrice
    }
    val boundableFeeCalculator = BoundableFeeCalculator(
      BoundableFeeCalculator.Config(gasPriceUpperBound, gasPriceLowerBound, gasPriceFixedCost),
      mockFeesCalculator
    )
    Assertions.assertThat(boundableFeeCalculator.calculateFees(feeHistory)).isEqualTo(expectedGasPrice)
  }

  @Test
  fun calculateFess_lowerBound() {
    val gasPriceUpperBound = BigInteger("10000000000")
    val gasPriceLowerBound = BigInteger("1000000")
    val gasPriceFixedCost = BigInteger("100")
    val calculatedL2GasPrice = gasPriceLowerBound.minus(BigInteger("300"))
    val expectedGasPrice = gasPriceLowerBound

    val mockFeesCalculator = mock<FeesCalculator> {
      on { calculateFees(any()) } doReturn calculatedL2GasPrice
    }
    val boundableFeeCalculator = BoundableFeeCalculator(
      BoundableFeeCalculator.Config(gasPriceUpperBound, gasPriceLowerBound, gasPriceFixedCost),
      mockFeesCalculator
    )
    Assertions.assertThat(boundableFeeCalculator.calculateFees(feeHistory)).isEqualTo(expectedGasPrice)
  }

  @Test
  fun constructor_throwsErrorIfNegativeParams() {
    val exception = assertThrows<IllegalArgumentException> {
      BoundableFeeCalculator(
        BoundableFeeCalculator.Config(
          BigInteger.valueOf(-1000),
          BigInteger.valueOf(-10000),
          BigInteger.valueOf(-100)
        ),
        mock<FeesCalculator>()
      )
    }
    Assertions.assertThat(exception)
      .isInstanceOf(IllegalArgumentException::class.java)
      .hasMessageContaining("feeUpperBound, feeLowerBound, and feeMargin must be no less than 0.")
  }

  @Test
  fun constructor_throwsErrorIfUpperSmallerThanLower() {
    val exception = assertThrows<IllegalArgumentException> {
      BoundableFeeCalculator(
        BoundableFeeCalculator.Config(
          BigInteger.valueOf(1000),
          BigInteger.valueOf(10000),
          BigInteger.valueOf(10)
        ),
        mock<FeesCalculator>()
      )
    }
    Assertions.assertThat(exception)
      .isInstanceOf(IllegalArgumentException::class.java)
      .hasMessageContaining("feeUpperBound must be no less than feeLowerBound.")
  }
}
