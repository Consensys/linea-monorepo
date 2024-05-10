package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.FeeHistory
import net.consensys.linea.contract.BoundableFeeCalculator
import net.consensys.zkevm.ethereum.gaspricing.FeesCalculator
import net.consensys.zkevm.ethereum.gaspricing.FeesFetcher
import net.consensys.zkevm.ethereum.gaspricing.GasPriceUpdater
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.Mockito.atLeastOnce
import org.mockito.Mockito.verify
import org.mockito.kotlin.any
import org.mockito.kotlin.doAnswer
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.eq
import org.mockito.kotlin.mock
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.milliseconds

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class DynamicGasPriceServiceTest {
  private val feeHistory = FeeHistory(
    oldestBlock = BigInteger.valueOf(100),
    baseFeePerGas = listOf(100, 110, 120, 130, 140).map { it.toBigInteger() },
    reward = listOf(1000, 1100, 1200, 1300).map { listOf(it.toBigInteger()) },
    gasUsedRatio = listOf(0.25.toBigDecimal(), 0.5.toBigDecimal(), 0.75.toBigDecimal(), 0.9.toBigDecimal()),
    baseFeePerBlobGas = listOf(100, 110, 120, 130, 140).map { it.toBigInteger() },
    blobGasUsedRatio = listOf(0.25.toBigDecimal(), 0.5.toBigDecimal(), 0.75.toBigDecimal(), 0.9.toBigDecimal())
  )

  @Test
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun start_startsPollingProcess(vertx: Vertx, testContext: VertxTestContext) {
    val pollingInterval = 10.milliseconds
    val gasPriceUpperBound = BigInteger("9000000000000")
    val gasPriceLowerBound = BigInteger("1000000")
    val gasPriceFixedCost = BigInteger("17")
    val calculatedL2GasPrice = gasPriceUpperBound.minus(BigInteger("100"))
    val expectedGasPrice = calculatedL2GasPrice + gasPriceFixedCost

    val mockFeesFetcher = mock<FeesFetcher>() {
      on { getL1EthGasPriceData() } doReturn SafeFuture.completedFuture(feeHistory)
    }
    val mockFeesCalculator = mock<FeesCalculator> {
      on { calculateFees(eq(feeHistory)) } doReturn calculatedL2GasPrice
    }
    val boundableFeeCalculator = BoundableFeeCalculator(
      BoundableFeeCalculator.Config(gasPriceUpperBound, gasPriceLowerBound, gasPriceFixedCost),
      mockFeesCalculator
    )
    val mockGasPriceUpdater = mock<GasPriceUpdater>() {
      on { updateMinerGasPrice(any()) } doAnswer { SafeFuture.completedFuture(Unit) }
    }
    val monitor =
      DynamicGasPriceService(
        DynamicGasPriceService.Config(pollingInterval),
        vertx,
        mockFeesFetcher,
        boundableFeeCalculator,
        mockGasPriceUpdater
      )
    monitor.start().thenApply {
      vertx.setTimer(100) {
        testContext
          .verify {
            monitor.stop()
            verify(mockFeesFetcher, atLeastOnce()).getL1EthGasPriceData()
            verify(mockFeesCalculator, atLeastOnce()).calculateFees(feeHistory)
            verify(mockGasPriceUpdater, atLeastOnce()).updateMinerGasPrice(expectedGasPrice)
          }
          .completeNow()
      }
    }
  }

  @Test
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun updatePrice_defaultsToPriceCapCalculatedPriceGoesOverMaxLimit(vertx: Vertx) {
    val pollingInterval = 10.milliseconds
    val gasPriceUpperBound = BigInteger.valueOf(10_000_000_000L) // 10 GWei
    val gasPriceLowerBound = BigInteger.valueOf(1_000_000L) // 0.001 GWei
    val gasPriceFixedCost = BigInteger.ZERO
    val calculatedL2GasPrice = gasPriceUpperBound.plus(BigInteger.ONE)

    val mockFeesFetcher = mock<FeesFetcher>() {
      on { getL1EthGasPriceData() } doReturn SafeFuture.completedFuture(feeHistory)
    }
    val mockFeesCalculator = mock<FeesCalculator> {
      on { calculateFees(any()) } doReturn calculatedL2GasPrice
    }
    val boundableFeeCalculator = BoundableFeeCalculator(
      BoundableFeeCalculator.Config(gasPriceUpperBound, gasPriceLowerBound, gasPriceFixedCost),
      mockFeesCalculator
    )
    val mockGasPriceUpdater = mock<GasPriceUpdater>() {
      on { updateMinerGasPrice(any()) } doAnswer { SafeFuture.completedFuture(Unit) }
    }

    val monitor =
      DynamicGasPriceService(
        DynamicGasPriceService.Config(pollingInterval),
        vertx,
        mockFeesFetcher,
        boundableFeeCalculator,
        mockGasPriceUpdater
      )

    monitor.action().get()
    verify(mockGasPriceUpdater).updateMinerGasPrice(gasPriceUpperBound)
  }

  @Test
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun updatePrice_defaultsToPriceCapCalculatedPriceGoesUnderMinLimit(vertx: Vertx) {
    val pollingInterval = 10.milliseconds
    val gasPriceUpperBound = BigInteger.valueOf(10_000_000_000L) // 10 GWei
    val gasPriceLowerBound = BigInteger.valueOf(90_000_000L) // 0.09 GWei
    val gasPriceFixedCost = BigInteger.ZERO
    val calculatedL2GasPrice = gasPriceLowerBound.minus(BigInteger.ONE)

    val mockFeesFetcher = mock<FeesFetcher>() {
      on { getL1EthGasPriceData() } doReturn SafeFuture.completedFuture(feeHistory)
    }
    val mockFeesCalculator = mock<FeesCalculator> {
      on { calculateFees(any()) } doReturn calculatedL2GasPrice
    }
    val boundableFeeCalculator = BoundableFeeCalculator(
      BoundableFeeCalculator.Config(gasPriceUpperBound, gasPriceLowerBound, gasPriceFixedCost),
      mockFeesCalculator
    )
    val mockGasPriceUpdater = mock<GasPriceUpdater>() {
      on { updateMinerGasPrice(any()) } doAnswer { SafeFuture.completedFuture(Unit) }
    }

    val monitor =
      DynamicGasPriceService(
        DynamicGasPriceService.Config(pollingInterval),
        vertx,
        mockFeesFetcher,
        boundableFeeCalculator,
        mockGasPriceUpdater
      )

    monitor.action().get()
    verify(mockGasPriceUpdater).updateMinerGasPrice(gasPriceLowerBound)
  }
}
