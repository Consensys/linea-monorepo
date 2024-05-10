package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import com.github.tomakehurst.wiremock.WireMockServer
import com.github.tomakehurst.wiremock.client.WireMock.containing
import com.github.tomakehurst.wiremock.client.WireMock.ok
import com.github.tomakehurst.wiremock.client.WireMock.post
import com.github.tomakehurst.wiremock.core.WireMockConfiguration.wireMockConfig
import com.github.tomakehurst.wiremock.matching.EqualToJsonPattern
import com.github.tomakehurst.wiremock.matching.EqualToPattern
import com.github.tomakehurst.wiremock.matching.RequestPatternBuilder
import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import io.vertx.core.Vertx
import io.vertx.core.json.JsonObject
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.jsonrpc.client.RequestRetryConfig
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.junit.jupiter.api.extension.ExtendWith
import java.math.BigInteger
import java.net.URL
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.milliseconds

@ExtendWith(VertxExtension::class)
class GasPriceUpdaterImplTest {
  private lateinit var gethRecipients: List<URL>
  private lateinit var besuRecipients: List<URL>
  val requestRetry = RequestRetryConfig(
    maxAttempts = 2u,
    backoffDelay = 10.milliseconds,
    failuresWarningThreshold = 1u
  )
  lateinit var wiremock: WireMockServer
  private lateinit var jsonRpcClientFactory: VertxHttpJsonRpcClientFactory
  val setPriceSuccessResponse =
    JsonObject.of(
      "jsonrpc",
      "2.0",
      "id",
      1,
      "result",
      true
    )
  val expectedGethRequest = JsonObject.of(
    "jsonrpc",
    "2.0",
    "method",
    "miner_setGasPrice",
    "params",
    listOf("0xa")
  )
  val expectedBesuRequest = expectedGethRequest.copy()
    .put("method", "miner_setMinGasPrice")

  @BeforeEach
  fun beforeEach(vertx: Vertx) {
    wiremock = WireMockServer(wireMockConfig().dynamicPort())
    wiremock.start()
    gethRecipients = listOf(
      URL("http://localhost:${wiremock.port()}/geth-1/"),
      URL("http://localhost:${wiremock.port()}/geth-2/"),
      URL("http://localhost:${wiremock.port()}/geth-3/")
    )
    besuRecipients = listOf(
      URL("http://localhost:${wiremock.port()}/besu-1/"),
      URL("http://localhost:${wiremock.port()}/besu-2/"),
      URL("http://localhost:${wiremock.port()}/besu-3/")
    )
    jsonRpcClientFactory = VertxHttpJsonRpcClientFactory(vertx, SimpleMeterRegistry())
  }

  @AfterEach
  fun afterEach() {
    wiremock.stop()
  }

  @Test
  fun throwsExceptionIfNoEndpoints() {
    val exception = assertThrows<IllegalArgumentException> {
      GasPriceUpdaterImpl.Config(
        gethEndpoints = emptyList(),
        besuEndPoints = emptyList(),
        retryConfig = requestRetry
      )
    }

    assertThat(exception)
      .isInstanceOf(IllegalArgumentException::class.java)
      .hasMessageContaining("Must have at least one geth or besu endpoint to update the gas price")

    // works with at least one geth endpoint
    GasPriceUpdaterImpl.Config(
      gethEndpoints = listOf(URL("http://localhost:8545")),
      besuEndPoints = emptyList(),
      retryConfig = requestRetry
    )
    // works with at least one Besu endpoint
    GasPriceUpdaterImpl.Config(
      gethEndpoints = emptyList(),
      besuEndPoints = listOf(URL("http://localhost:8545")),
      retryConfig = requestRetry
    )
  }

  @Test
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun gasPriceUpdaterImpl_setsPriceOnGethAndBesu(vertx: Vertx, testContext: VertxTestContext) {
    testPriceUpdateForEndpoints(testContext, gethRecipients, besuRecipients)
  }

  @Test
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun gasPriceUpdaterImpl_setsPriceOnGethOnly(vertx: Vertx, testContext: VertxTestContext) {
    testPriceUpdateForEndpoints(testContext, gethRecipients = gethRecipients, besuRecipients = emptyList())
  }

  @Test
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun gasPriceUpdaterImpl_setsPriceOnBesuOnly(vertx: Vertx, testContext: VertxTestContext) {
    testPriceUpdateForEndpoints(testContext, gethRecipients = emptyList(), besuRecipients = besuRecipients)
  }

  private fun testPriceUpdateForEndpoints(
    testContext: VertxTestContext,
    gethRecipients: List<URL>,
    besuRecipients: List<URL>
  ) {
    val gasPrice = BigInteger.valueOf(10L)
    gethRecipients.forEach { endpoint -> wiremockStubForPost(wiremock, endpoint, setPriceSuccessResponse) }
    besuRecipients.forEach { endpoint -> wiremockStubForPost(wiremock, endpoint, setPriceSuccessResponse) }

    val l2GasPriceUpdaterImpl =
      GasPriceUpdaterImpl(
        jsonRpcClientFactory,
        GasPriceUpdaterImpl.Config(
          gethEndpoints = gethRecipients,
          besuEndPoints = besuRecipients,
          retryConfig = requestRetry
        )
      )

    l2GasPriceUpdaterImpl.updateMinerGasPrice(gasPrice)
      .thenApply {
        testContext
          .verify {
            gethRecipients.forEach { endpoint -> verifyRequest(wiremock, endpoint, expectedGethRequest) }
            besuRecipients.forEach { endpoint -> verifyRequest(wiremock, endpoint, expectedBesuRequest) }
          }
          .completeNow()
      }
  }

  private fun wiremockStubForPost(wiremock: WireMockServer, requestOriginEndpoint: URL, response: JsonObject) {
    wiremock.stubFor(
      post(requestOriginEndpoint.path)
        .withHeader("Content-Type", containing("application/json"))
        .willReturn(
          ok()
            .withHeader("Content-type", "application/json")
            .withBody(response.toString())
        )
    )
  }

  private fun verifyRequest(
    wiremock: WireMockServer,
    requestOriginEndpoint: URL,
    request: JsonObject
  ) {
    wiremock.verify(
      RequestPatternBuilder.newRequestPattern()
        .withPort(wiremock.port())
        .withUrl(requestOriginEndpoint.path)
        .withHeader("content-type", EqualToPattern("application/json"))
        .withRequestBody(
          EqualToJsonPattern(
            request.toString(), /*ignoreArrayOrder*/
            false, /*ignoreExtraElements*/
            true
          )
        )
    )
  }
}
