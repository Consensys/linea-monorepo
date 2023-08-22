package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import com.github.tomakehurst.wiremock.WireMockServer
import com.github.tomakehurst.wiremock.client.WireMock.containing
import com.github.tomakehurst.wiremock.client.WireMock.ok
import com.github.tomakehurst.wiremock.client.WireMock.post
import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import io.vertx.core.Vertx
import io.vertx.core.http.HttpClientOptions
import io.vertx.core.http.HttpVersion
import io.vertx.core.json.JsonObject
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClient
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import org.apache.logging.log4j.Logger
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.RepeatedTest
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.Mockito
import org.mockito.kotlin.any
import org.mockito.kotlin.anyOrNull
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import java.math.BigInteger
import java.net.URI
import java.net.URL
import java.util.concurrent.TimeUnit

@ExtendWith(VertxExtension::class)
class GasPriceUpdaterImplTest {
  private val recipientHostnames = listOf(
    "http://localhost:18545/",
    "http://localhost:18546/",
    "http://localhost:18647/"
  )

  private fun wiremockStubForPost(response: JsonObject, wiremock: WireMockServer) {
    wiremock.stubFor(
      post("/")
        .withHeader("Content-Type", containing("application/json"))
        .willReturn(
          ok()
            .withHeader("Content-type", "application/json")
            .withBody(response.toString().toByteArray())
        )
    )
  }

  private fun createMockedHttpJsonRpcClientFactory(vertx: Vertx): VertxHttpJsonRpcClientFactory {
    val httpJsonRpcClientFactory = mock<VertxHttpJsonRpcClientFactory>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)

    whenever(
      httpJsonRpcClientFactory
        .create(
          any<URL>(),
          anyOrNull<Int>(),
          anyOrNull<HttpVersion>(),
          any<Logger>(),
          anyOrNull(),
          anyOrNull()
        )
    ).thenAnswer { invocation ->
      val endpointUrl = invocation.getArgument<URL>(0)
      val wiremock = WireMockServer(endpointUrl.port)
      wiremock.start()

      val response =
        JsonObject.of(
          "jsonrpc",
          "2.0",
          "id",
          1,
          "result",
          endpointUrl.port % 2 == 0
        )

      wiremockStubForPost(response, wiremock)

      val uri = URI(endpointUrl.toString())
      val clientOptions =
        HttpClientOptions()
          .setKeepAlive(true)
          .setProtocolVersion(HttpVersion.HTTP_1_1)
          .setMaxPoolSize(10)
          .setDefaultHost(uri.host)
          .setDefaultPort(uri.port)
          .setLogActivity(true)
      val httpClient = vertx.createHttpClient(clientOptions)
      val meterRegistry = SimpleMeterRegistry()
      VertxHttpJsonRpcClient(httpClient, "/", meterRegistry)
    }

    return httpJsonRpcClientFactory
  }

  @RepeatedTest(1)
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun gasPriceUpdaterImpl_sendsMinerSetGasPriceToRecipients(vertx: Vertx, testContext: VertxTestContext) {
    val httpJsonRpcClientFactoryMock = createMockedHttpJsonRpcClientFactory(vertx)

    val l2GasPriceUpdaterImpl =
      GasPriceUpdaterImpl(
        httpJsonRpcClientFactoryMock,
        GasPriceUpdaterImpl.Config(
          this.recipientHostnames.map { URL(it) }
        )
      )

    l2GasPriceUpdaterImpl.updateMinerGasPrice(BigInteger.valueOf(3000000000L))
      .thenApply {
        testContext
          .verify {
            assertThat(it).isEqualTo((0 until this.recipientHostnames.size).map {})
          }
          .completeNow()
      }
  }
}
