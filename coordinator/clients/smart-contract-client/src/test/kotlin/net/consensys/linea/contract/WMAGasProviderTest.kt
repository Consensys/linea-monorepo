package net.consensys.linea.contract

import net.consensys.linea.FeeHistory
import net.consensys.zkevm.ethereum.gaspricing.FeesCalculator
import net.consensys.zkevm.ethereum.gaspricing.FeesFetcher
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigDecimal
import java.math.BigInteger

class WMAGasProviderTest {
  private val feeHistory = FeeHistory(
    oldestBlock = BigInteger.valueOf(100),
    baseFeePerGas = listOf(100, 110, 120, 130, 140).map { it.toBigInteger() },
    reward = listOf(1000, 1100, 1200, 1300).map { listOf(it.toBigInteger()) },
    gasUsedRatio = listOf(0.25.toBigDecimal(), 0.5.toBigDecimal(), 0.75.toBigDecimal(), 0.9.toBigDecimal()),
    baseFeePerBlobGas = listOf(100, 110, 120, 130, 140).map { it.toBigInteger() },
    blobGasUsedRatio = listOf(0.25.toBigDecimal(), 0.5.toBigDecimal(), 0.75.toBigDecimal(), 0.9.toBigDecimal())
  )
  private val chainId = 999
  private val gasLimit = BigInteger.valueOf(100000)
  private val mockFeesFetcher = mock<FeesFetcher>() {
    on { getL1EthGasPriceData() } doReturn SafeFuture.completedFuture(feeHistory)
  }
  private val l1PriorityFeeCalculator: FeesCalculator = WMAFeesCalculator(
    WMAFeesCalculator.Config(
      baseFeeCoefficient = BigDecimal("0.0"),
      priorityFeeWmaCoefficient = BigDecimal.ONE
    )
  )

  @Test
  fun `getMaxFeePerGas should return correct value based on baseFeePerGas and WMA of priority fee`() {
    val maxFeePerGasCap = BigInteger.valueOf(1000000000)
    val maxFeePerBlobGasCap = BigInteger.valueOf(1000000000)
    val wmaGasProvider =
      WMAGasProvider(
        chainId.toLong(),
        mockFeesFetcher,
        l1PriorityFeeCalculator,
        WMAGasProvider.Config(
          gasLimit = gasLimit,
          maxFeePerGasCap = maxFeePerGasCap,
          maxFeePerBlobGasCap = maxFeePerBlobGasCap
        )
      )

    // WMA = (140*0.0) + ((1000*0.25*1 + 1100*0.5*2 + 1200*0.75*3 + 1300*0.9*4) / (0.25*1 + 0.5*2 + 0.75*3 + 0.9*4)) * 1.0 = 1229.57746
    // MaxFeePerGas = min(140*2 + 1229, 100000000) = 1509
    val calculatedMaxFeePerGas = wmaGasProvider.getMaxFeePerGas(null)
    assertThat(calculatedMaxFeePerGas).isEqualTo(BigInteger.valueOf(1509))
  }

  @Test
  fun `getMaxPriorityFeePerGas should return correct WMA of priority fee`() {
    val maxFeePerGasCap = BigInteger.valueOf(1000000000)
    val maxFeePerBlobGasCap = BigInteger.valueOf(1000000000)
    val wmaGasProvider =
      WMAGasProvider(
        chainId.toLong(),
        mockFeesFetcher,
        l1PriorityFeeCalculator,
        WMAGasProvider.Config(
          gasLimit = gasLimit,
          maxFeePerGasCap = maxFeePerGasCap,
          maxFeePerBlobGasCap = maxFeePerBlobGasCap
        )
      )

    val calculatedMaxPriorityFeePerGas = wmaGasProvider.getMaxPriorityFeePerGas(null)
    assertThat(calculatedMaxPriorityFeePerGas).isEqualTo(BigInteger.valueOf(1229))
  }

  @Test
  fun `getMaxFeePerGas should return 1000 if calculated max fee per gas is larger than 1000`() {
    val maxFeePerGasCap = BigInteger.valueOf(1000)
    val maxFeePerBlobGasCap = BigInteger.valueOf(100)
    val wmaGasProvider =
      WMAGasProvider(
        chainId.toLong(),
        mockFeesFetcher,
        l1PriorityFeeCalculator,
        WMAGasProvider.Config(
          gasLimit = gasLimit,
          maxFeePerGasCap = maxFeePerGasCap,
          maxFeePerBlobGasCap = maxFeePerBlobGasCap
        )
      )

    val calculatedMaxFeePerGas = wmaGasProvider.getMaxFeePerGas(null)
    assertThat(calculatedMaxFeePerGas).isEqualTo(maxFeePerGasCap)
  }

  @Test
  fun `getMaxPriorityFeePerGas should return 1000 if WMA of priority fee is larger than 1000`() {
    val maxFeePerGasCap = BigInteger.valueOf(1000)
    val maxFeePerBlobGasCap = BigInteger.valueOf(1000000000)
    val wmaGasProvider =
      WMAGasProvider(
        chainId.toLong(),
        mockFeesFetcher,
        l1PriorityFeeCalculator,
        WMAGasProvider.Config(
          gasLimit = gasLimit,
          maxFeePerGasCap = maxFeePerGasCap,
          maxFeePerBlobGasCap = maxFeePerBlobGasCap
        )
      )

    val calculatedMaxPriorityFeePerGas = wmaGasProvider.getMaxPriorityFeePerGas(null)
    assertThat(calculatedMaxPriorityFeePerGas).isEqualTo(maxFeePerGasCap)
  }

  @Test
  fun `getEIP1559GasFees should return correct maxPriorityFeePerGas and maxFeePerGas`() {
    val maxFeePerGasCap = BigInteger.valueOf(1000000000)
    val maxFeePerBlobGasCap = BigInteger.valueOf(1000000000)
    val wmaGasProvider =
      WMAGasProvider(
        chainId.toLong(),
        mockFeesFetcher,
        l1PriorityFeeCalculator,
        WMAGasProvider.Config(
          gasLimit = gasLimit,
          maxFeePerGasCap = maxFeePerGasCap,
          maxFeePerBlobGasCap = maxFeePerBlobGasCap
        )
      )

    val calculatedFees = wmaGasProvider.getEIP1559GasFees()
    assertThat(calculatedFees.maxPriorityFeePerGas).isEqualTo(1229)
    assertThat(calculatedFees.maxFeePerGas).isEqualTo(1509)
  }

  @Test
  fun `getEIP1559GasFees should return both maxPriorityFeePerGas and maxFeePerGas as 1000 if exceeds max cap`() {
    val maxFeePerGasCap = BigInteger.valueOf(1000)
    val maxFeePerBlobGasCap = BigInteger.valueOf(1000000000)
    val wmaGasProvider =
      WMAGasProvider(
        chainId.toLong(),
        mockFeesFetcher,
        l1PriorityFeeCalculator,
        WMAGasProvider.Config(
          gasLimit = gasLimit,
          maxFeePerGasCap = maxFeePerGasCap,
          maxFeePerBlobGasCap = maxFeePerBlobGasCap
        )
      )

    val calculatedFees = wmaGasProvider.getEIP1559GasFees()
    assertThat(calculatedFees.maxPriorityFeePerGas).isEqualTo(maxFeePerGasCap)
    assertThat(calculatedFees.maxFeePerGas).isEqualTo(maxFeePerGasCap)
  }

  @Test
  fun `getEIP4844GasFees should return correct maxPriorityFeePerGas and maxFeePerGas and maxFeePerBlobGas`() {
    val maxFeePerGasCap = BigInteger.valueOf(1000000000)
    val maxFeePerBlobGasCap = BigInteger.valueOf(1000000000)
    val wmaGasProvider =
      WMAGasProvider(
        chainId.toLong(),
        mockFeesFetcher,
        l1PriorityFeeCalculator,
        WMAGasProvider.Config(
          gasLimit = gasLimit,
          maxFeePerGasCap = maxFeePerGasCap,
          maxFeePerBlobGasCap = maxFeePerBlobGasCap
        )
      )

    val calculatedFees = wmaGasProvider.getEIP4844GasFees()
    assertThat(calculatedFees.eip1559GasFees.maxPriorityFeePerGas).isEqualTo(1229)
    assertThat(calculatedFees.eip1559GasFees.maxFeePerGas).isEqualTo(1509)
    assertThat(calculatedFees.maxFeePerBlobGas).isEqualTo(maxFeePerBlobGasCap)
  }

  @Test
  fun `getEIP4844GasFees should return both calculatedFees as 100 if exceeds max cap`() {
    val maxFeePerGasCap = BigInteger.valueOf(1000)
    val maxFeePerGasCapForEIP4844 = BigInteger.valueOf(100)
    val maxFeePerBlobGasCap = BigInteger.valueOf(100)
    val wmaGasProvider =
      WMAGasProvider(
        chainId.toLong(),
        mockFeesFetcher,
        l1PriorityFeeCalculator,
        WMAGasProvider.Config(
          gasLimit = gasLimit,
          maxFeePerGasCap = maxFeePerGasCap,
          maxFeePerBlobGasCap = maxFeePerBlobGasCap,
          maxFeePerGasCapForEIP4844 = maxFeePerGasCapForEIP4844
        )
      )

    val calculatedFees = wmaGasProvider.getEIP4844GasFees()
    assertThat(calculatedFees.eip1559GasFees.maxPriorityFeePerGas).isEqualTo(maxFeePerGasCapForEIP4844)
    assertThat(calculatedFees.eip1559GasFees.maxFeePerGas).isEqualTo(maxFeePerGasCapForEIP4844)
    assertThat(calculatedFees.maxFeePerBlobGas).isEqualTo(maxFeePerBlobGasCap)
  }

  @Test
  fun
  `getEIP4844GasFees should return both calculatedFees as 100 if maxFeePerGasCapForEIP4844 is null and exceeds max cap`
  () {
    val maxFeePerGasCap = BigInteger.valueOf(100)
    val maxFeePerBlobGasCap = BigInteger.valueOf(100)
    val wmaGasProvider =
      WMAGasProvider(
        chainId.toLong(),
        mockFeesFetcher,
        l1PriorityFeeCalculator,
        WMAGasProvider.Config(
          gasLimit = gasLimit,
          maxFeePerGasCap = maxFeePerGasCap,
          maxFeePerBlobGasCap = maxFeePerBlobGasCap
        )
      )

    val calculatedFees = wmaGasProvider.getEIP4844GasFees()
    assertThat(calculatedFees.eip1559GasFees.maxPriorityFeePerGas).isEqualTo(maxFeePerGasCap)
    assertThat(calculatedFees.eip1559GasFees.maxFeePerGas).isEqualTo(maxFeePerGasCap)
    assertThat(calculatedFees.maxFeePerBlobGas).isEqualTo(maxFeePerBlobGasCap)
  }
}
