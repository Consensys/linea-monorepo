package net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap

import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapProvider
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCaps
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.Mockito.RETURNS_DEEP_STUBS
import org.mockito.kotlin.any
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

@ExtendWith(VertxExtension::class)
class GasPriceCapProviderForDataSubmissionTest {
  private val defaultGasPriceCaps = GasPriceCaps(
    maxFeePerGasCap = BigInteger("10000000000"), // 10GWei
    maxPriorityFeePerGasCap = BigInteger("10000000000"), // 10GWei
    maxFeePerBlobGasCap = BigInteger("1000000000") // 1GWei
  )
  private val returnGasPriceCaps = GasPriceCaps(
    maxFeePerGasCap = BigInteger("7000000000"), // 7GWei
    maxPriorityFeePerGasCap = BigInteger("7000000000"), // 7GWei
    maxFeePerBlobGasCap = BigInteger("500000000") // 0.5GWei
  )

  private lateinit var gasPriceCapProviderForDataSubmission: GasPriceCapProvider
  private lateinit var mockedGasPriceCapProvider: GasPriceCapProvider
  private lateinit var mockedMetricsFacade: MetricsFacade

  @BeforeEach
  fun beforeEach(vertx: Vertx) {
    mockedGasPriceCapProvider = mock<GasPriceCapProvider> {
      on { getGasPriceCaps(any()) } doReturn SafeFuture.completedFuture(returnGasPriceCaps)
      on { getDefaultGasPriceCaps() } doReturn defaultGasPriceCaps
    }
    mockedMetricsFacade = mock<MetricsFacade>(defaultAnswer = RETURNS_DEEP_STUBS)
    gasPriceCapProviderForDataSubmission = GasPriceCapProviderForDataSubmission(
      gasPriceCapProvider = mockedGasPriceCapProvider,
      metricsFacade = mockedMetricsFacade
    )
  }

  @Test
  fun `gas price caps of data submission should be returned correctly`() {
    assertThat(
      gasPriceCapProviderForDataSubmission.getGasPriceCaps(100L).get()
    ).isEqualTo(
      returnGasPriceCaps
    )
  }

  @Test
  fun `default gas price caps of data submission should be returned correctly`() {
    assertThat(
      gasPriceCapProviderForDataSubmission.getDefaultGasPriceCaps()
    ).isEqualTo(
      defaultGasPriceCaps
    )
  }
}
