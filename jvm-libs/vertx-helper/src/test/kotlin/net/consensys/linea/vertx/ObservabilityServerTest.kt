package net.consensys.linea.vertx

import io.restassured.RestAssured
import io.restassured.builder.RequestSpecBuilder
import io.restassured.http.ContentType
import io.restassured.module.kotlin.extensions.When
import io.restassured.specification.RequestSpecification
import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import net.consensys.linea.async.get
import org.assertj.core.api.Assertions
import org.hamcrest.Matchers
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import kotlin.random.Random

@ExtendWith(VertxExtension::class)
class ObservabilityServerTest {
  private var port: Int = 0
  private lateinit var monitorRequestSpecification: RequestSpecification
  private lateinit var observabilityServerDeploymentId: String

  @BeforeEach
  fun beforeEach(vertx: Vertx) {
    val (newDeploymentId, newPort) = runServerOnARandomPort(vertx)
    observabilityServerDeploymentId = newDeploymentId
    port = newPort
    monitorRequestSpecification =
      RequestSpecBuilder()
        .setBaseUri("http://localhost:$port/")
        .build()
  }

  @AfterEach
  fun afterEach(vertx: Vertx) {
    vertx.undeploy(observabilityServerDeploymentId).get()
  }

  @Test
  fun exposesLiveEndpoint() {
    RestAssured.given()
      .spec(monitorRequestSpecification)
      .When {
        get("/live")
      }
      .then()
      .statusCode(200)
      .contentType(ContentType.JSON)
      .body("status", Matchers.equalTo("OK"))
  }

  @Test
  fun exposesHealthEndpoint() {
    RestAssured.given()
      .spec(monitorRequestSpecification)
      .When {
        get("/health")
      }
      .then()
      .statusCode(200)
      .contentType(ContentType.JSON)
      .body("status", Matchers.equalTo("UP"))
  }

  @Test
  fun exposesMetricsEndpoint() {
    // If no request is sent, no server metrics will appear
    RestAssured.given()
      .spec(monitorRequestSpecification)
      .accept(ContentType.JSON)
      .When {
        get("/metrics")
      }

    RestAssured.given()
      .spec(monitorRequestSpecification)
      .accept(ContentType.JSON)
      .When {
        get("/metrics")
      }
      .also { Assertions.assertThat(it.body.asString()).contains("vertx_http_server_response_bytes_bucket") }
      .then()
      .statusCode(200)
      .contentType(ContentType.TEXT)
  }

  private fun runServerOnARandomPort(vertx: Vertx): Pair<String, Int> {
    val maxAttempts = 3
    var attempts = 0
    while (attempts < maxAttempts) {
      val randomPort = Random.nextInt(1000, UShort.MAX_VALUE.toInt())
      try {
        val observabilityServer = ObservabilityServer(
          ObservabilityServer.Config(applicationName = "test", port = randomPort)
        )
        val deploymentId = vertx.deployVerticle(observabilityServer).get()
        return deploymentId to randomPort
      } catch (_: Exception) {
        attempts += 1
      }
    }
    throw RuntimeException("Couldn't start observability server on a random port after $maxAttempts attempts!")
  }
}
