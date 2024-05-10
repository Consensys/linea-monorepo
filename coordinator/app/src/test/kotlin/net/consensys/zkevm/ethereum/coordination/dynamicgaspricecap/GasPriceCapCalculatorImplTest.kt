package net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap

import io.vertx.junit5.VertxExtension
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.junit.jupiter.api.extension.ExtendWith
import java.math.BigInteger
import kotlin.time.Duration.Companion.hours

@ExtendWith(VertxExtension::class)
class GasPriceCapCalculatorImplTest {
  @Test
  fun `calculateFeeGasPriceCap returns gas price cap correctly after 1 hour since last finalization`() {
    val gasPriceCapCalculator = GasPriceCapCalculatorImpl()

    // GasPriceCap = (baseFeePerGasAtPercentile * (1 + adjustmentConstant * timeOfDayMultiplier *
    //               (timeSincePrevFinalization/finalizationTargetMaxDelay)^2)) + avgRewardAtPercentile
    //             = (1GWei * (1 + 25 * 1.75 * (1/32)^2)) + 5GWei = 6.042724609GWei
    val calculatedGasPriceCap = gasPriceCapCalculator.calculateGasPriceCap(
      adjustmentConstant = 25U,
      finalizationTargetMaxDelay = 32.hours,
      baseFeePerGasAtPercentile = BigInteger.valueOf(1000000000), // 1GWei
      elapsedTimeSinceBlockTimestamp = 1.hours,
      avgRewardAtPercentile = BigInteger.valueOf(5000000000), // 5GWei
      timeOfDayMultiplier = 1.75 // i.e. SUNDAY_2 in config/common/gas-price-cap-time-of-day-multipliers.toml
    )
    assertThat(calculatedGasPriceCap).isEqualTo(BigInteger.valueOf(6042724609))
  }

  @Test
  fun `calculateFeeGasPriceCap returns gas price cap correctly after 16 hours since last finalization`() {
    val gasPriceCapCalculator = GasPriceCapCalculatorImpl()

    // GasPriceCap = (baseFeePerGasAtPercentile * (1 + adjustmentConstant * timeOfDayMultiplier *
    //               (timeSincePrevFinalization/finalizationTargetMaxDelay)^2)) + avgRewardAtPercentile
    //             = (1GWei * (1 + 25 * 1.75 * (16/32)^2)) + 5GWei = 16.9375GWei
    val calculatedGasPriceCap = gasPriceCapCalculator.calculateGasPriceCap(
      adjustmentConstant = 25U,
      finalizationTargetMaxDelay = 32.hours,
      baseFeePerGasAtPercentile = BigInteger.valueOf(1000000000), // 1GWei
      elapsedTimeSinceBlockTimestamp = 16.hours,
      avgRewardAtPercentile = BigInteger.valueOf(5000000000), // 5GWei
      timeOfDayMultiplier = 1.75 // i.e. SUNDAY_2 in config/common/gas-price-cap-time-of-day-multipliers.toml
    )
    assertThat(calculatedGasPriceCap).isEqualTo(BigInteger.valueOf(16937500000))
  }

  @Test
  fun `calculateFeeGasPriceCap returns gas price cap correctly after 32 hours since last finalization`() {
    val gasPriceCapCalculator = GasPriceCapCalculatorImpl()

    // GasPriceCap = (baseFeePerGasAtPercentile * (1 + adjustmentConstant * timeOfDayMultiplier *
    //               (timeSincePrevFinalization/finalizationTargetMaxDelay)^2)) + avgRewardAtPercentile
    //             = (1GWei * (1 + 25 * 1.75 * (32/32)^2)) + 5GWei = 49.75Wei
    val calculatedGasPriceCap = gasPriceCapCalculator.calculateGasPriceCap(
      adjustmentConstant = 25U,
      finalizationTargetMaxDelay = 32.hours,
      baseFeePerGasAtPercentile = BigInteger.valueOf(1000000000), // 1GWei
      elapsedTimeSinceBlockTimestamp = 32.hours,
      avgRewardAtPercentile = BigInteger.valueOf(5000000000), // 5GWei
      timeOfDayMultiplier = 1.75 // i.e. SUNDAY_2 in config/common/gas-price-cap-time-of-day-multipliers.toml
    )
    assertThat(calculatedGasPriceCap).isEqualTo(BigInteger.valueOf(49750000000))
  }

  @Test
  fun `calculateFeeGasPriceCap returns gas price cap correctly without specifying avgReward and tdm`() {
    val gasPriceCapCalculator = GasPriceCapCalculatorImpl()

    // GasPriceCap = (baseFeePerGasAtPercentile * (1 + adjustmentConstant * timeOfDayMultiplier *
    //               (timeSincePrevFinalization/finalizationTargetMaxDelay)^2)) + avgRewardAtPercentile
    //             = 1GWei * (1 + 25 * 1.0 * (32/32)^2) = 26GWei
    val calculatedGasPriceCap = gasPriceCapCalculator.calculateGasPriceCap(
      adjustmentConstant = 25U,
      finalizationTargetMaxDelay = 32.hours,
      baseFeePerGasAtPercentile = BigInteger.valueOf(1000000000), // 1GWei
      elapsedTimeSinceBlockTimestamp = 32.hours
      // avgRewardAtPercentile = BigInteger.ZERO as default
      // timeOfDayMultiplier = 1.0 as default
    )
    assertThat(calculatedGasPriceCap).isEqualTo(BigInteger.valueOf(26000000000))
  }

  @Test
  fun `calculateFeeGasPriceCap throws error when finalizationTargetMaxDelay is 0`() {
    assertThrows<IllegalArgumentException> {
      GasPriceCapCalculatorImpl().calculateGasPriceCap(
        adjustmentConstant = 25U,
        finalizationTargetMaxDelay = 0.hours,
        baseFeePerGasAtPercentile = BigInteger.valueOf(1000000000), // 1GWei
        elapsedTimeSinceBlockTimestamp = 32.hours
      )
    }.also { exception ->
      assertThat(exception.message)
        .isEqualTo(
          "finalizationTargetMaxDelay duration must be longer than zero second. Value=0s"
        )
    }
  }
}
