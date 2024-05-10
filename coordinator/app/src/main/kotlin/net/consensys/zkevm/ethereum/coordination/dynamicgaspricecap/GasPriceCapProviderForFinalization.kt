package net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap

import net.consensys.linea.metrics.LineaMetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapProvider
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCaps
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigDecimal
import java.util.concurrent.atomic.AtomicReference

class GasPriceCapProviderForFinalization(
  private val config: Config,
  private val gasPriceCapProvider: GasPriceCapProvider,
  metricsFacade: MetricsFacade
) : GasPriceCapProvider {
  data class Config(
    val gasPriceCapMultiplier: Double = 1.0
  )
  private var lastGasPriceCap: AtomicReference<GasPriceCaps>

  init {
    require(config.gasPriceCapMultiplier >= 0.0) {
      "gasPriceCapMultiplier must be no less than 0.0. Value=${config.gasPriceCapMultiplier}"
    }

    lastGasPriceCap = AtomicReference(
      multiplyGasPriceCaps(gasPriceCapProvider.getDefaultGasPriceCaps())
    )

    metricsFacade.createGauge(
      category = LineaMetricsCategory.GAS_PRICE_CAP,
      name = "l1.finalization.maxfeepergascap",
      description = "Max fee per gas cap for finalization on L1",
      measurementSupplier = { lastGasPriceCap.get().maxFeePerGasCap }
    )
  }

  private fun multiplyGasPriceCaps(gasPriceCaps: GasPriceCaps): GasPriceCaps {
    return GasPriceCaps(
      maxFeePerGasCap = gasPriceCaps.maxFeePerGasCap.toBigDecimal()
        .multiply(BigDecimal.valueOf(config.gasPriceCapMultiplier))
        .toBigInteger(),
      maxPriorityFeePerGasCap = gasPriceCaps.maxPriorityFeePerGasCap.toBigDecimal()
        .multiply(BigDecimal.valueOf(config.gasPriceCapMultiplier))
        .toBigInteger(),
      maxFeePerBlobGasCap = gasPriceCaps.maxFeePerBlobGasCap
    )
  }

  override fun getGasPriceCaps(targetL2BlockNumber: Long): SafeFuture<GasPriceCaps> {
    return gasPriceCapProvider.getGasPriceCaps(targetL2BlockNumber)
      .thenApply(this::multiplyGasPriceCaps)
      .thenPeek {
        this.lastGasPriceCap.set(it)
      }
  }

  override fun getDefaultGasPriceCaps(): GasPriceCaps {
    return multiplyGasPriceCaps(gasPriceCapProvider.getDefaultGasPriceCaps())
  }
}
