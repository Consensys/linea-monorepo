package net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap

import net.consensys.linea.metrics.LineaMetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapProvider
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCaps
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.atomic.AtomicReference

class GasPriceCapProviderForDataSubmission(
  private val gasPriceCapProvider: GasPriceCapProvider,
  metricsFacade: MetricsFacade
) : GasPriceCapProvider {
  private var lastGasPriceCap: AtomicReference<GasPriceCaps> =
    AtomicReference(gasPriceCapProvider.getDefaultGasPriceCaps())

  init {
    metricsFacade.createGauge(
      category = LineaMetricsCategory.GAS_PRICE_CAP,
      name = "l1.blobsubmission.maxfeepergascap",
      description = "Max fee per gas cap for blob submission on L1",
      measurementSupplier = { lastGasPriceCap.get().maxFeePerGasCap }
    )
    metricsFacade.createGauge(
      category = LineaMetricsCategory.GAS_PRICE_CAP,
      name = "l1.blobsubmission.maxfeeperblobgascap",
      description = "Max fee per blob gas cap for blob submission on L1",
      measurementSupplier = { lastGasPriceCap.get().maxFeePerBlobGasCap }
    )
  }

  override fun getGasPriceCaps(targetL2BlockNumber: Long): SafeFuture<GasPriceCaps> {
    return gasPriceCapProvider.getGasPriceCaps(targetL2BlockNumber)
      .thenPeek {
        this.lastGasPriceCap.set(it)
      }
  }

  override fun getDefaultGasPriceCaps(): GasPriceCaps {
    return gasPriceCapProvider.getDefaultGasPriceCaps()
  }
}
