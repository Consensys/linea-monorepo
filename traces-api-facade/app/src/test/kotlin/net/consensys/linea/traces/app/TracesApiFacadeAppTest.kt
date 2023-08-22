package net.consensys.linea.traces.app

import io.restassured.RestAssured
import io.restassured.builder.RequestSpecBuilder
import io.restassured.http.ContentType
import io.restassured.response.Response
import io.restassured.specification.RequestSpecification
import io.vertx.core.json.JsonObject
import io.vertx.kotlin.core.json.get
import net.consensys.linea.traces.TracingModule
import net.consensys.linea.traces.app.api.ApiConfig
import net.consensys.linea.traces.repository.ReadTracesCacheConfig
import org.assertj.core.api.Assertions.assertThat
import org.hamcrest.Matchers.equalTo
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.io.TempDir
import java.nio.file.Path
import java.time.Duration

class TracesApiFacadeAppTest {
  private val apiConfig =
    ApiConfig(port = 9394u, observabilityPort = 9395u, path = "/", numberOfVerticles = 1u)
  private val tracesVersion = "0.0.1"
  private val blockL1Size = 1000001U
  private lateinit var appConfig: AppConfig
  private val requestSpecification: RequestSpecification =
    RequestSpecBuilder()
      // enable for debug only
      // .addFilters(listOf(ResponseLoggingFilter(), RequestLoggingFilter()))
      .setBaseUri("http://localhost:${apiConfig.port}/")
      .build()
  private lateinit var app: TracesApiFacadeApp

  @BeforeEach
  fun beforeEach(
    @TempDir conflatedTracesDirectory: Path
  ) {
    appConfig =
      AppConfig(
        inputTracesDirectory = "../../testdata/traces/raw/",
        outputTracesDirectory = conflatedTracesDirectory.toString(),
        tracesVersion = tracesVersion,
        api = apiConfig,
        readTracesCache = ReadTracesCacheConfig(0u, Duration.ofMillis(100)),
        tracesFileExtension = "json.gz"
      )
    app = TracesApiFacadeApp(appConfig)
    app.start().toCompletionStage().toCompletableFuture().get()
  }

  @AfterEach
  fun afterEach() {
    app.stop().toCompletionStage().toCompletableFuture().get()
  }

  @Test()
  @Disabled("Api Method replaced by V1 counterpart")
  fun tracesCountersV0_tracesExist() {
    val jsonRpcRequest = buildRpcJson("rollup_getTracesCountersByBlockNumberV0", "0x1")
    val response = makeJsonRpcRequest(jsonRpcRequest)
    response.then().statusCode(200).contentType("application/json")

    val jsonRpcResponse = JsonObject(response.body.asString())
    assertTracesCountersResponse(jsonRpcResponse, jsonRpcRequest)
  }

  @Test
  @Disabled("Api Method replaced by V1 counterpart")
  fun tracesCountersV0_FileNotFound() {
    val jsonRpcRequest = buildRpcJson("rollup_getTracesCountersByBlockNumberV0", "0xf11ced")
    val response = makeJsonRpcRequest(jsonRpcRequest)
    response.then().statusCode(200).contentType("application/json")

    val jsonRpcResponse = JsonObject(response.body.asString())
    assertThat(jsonRpcResponse.getValue("id")).isEqualTo(jsonRpcRequest.getValue("id"))
    assertThat(jsonRpcResponse.getValue("result")).isNull()

    jsonRpcResponse.getValue("error").let { error ->
      assertThat(error).isNotNull
      error as JsonObject
      assertThat(error.getValue("code")).isEqualTo(-4001)
      assertThat(error.getValue("message")).isEqualTo("Traces not available")
      assertThat(error.getValue("data")).isEqualTo("Traces not available for block 15801581.")
    }
  }

  @Test
  @Disabled("Api Method replaced by V1 counterpart")
  fun generateConflatedTracesToFileV0_success() {
    val jsonRpcRequest = buildRpcJson("rollup_generateConflatedTracesToFileV0", "0x1", "0x3")
    val response = makeJsonRpcRequest(jsonRpcRequest)
    response.then().statusCode(200).contentType("application/json")

    val jsonRpcResponse = JsonObject(response.body.asString())
    assertTracesConflationResponse(jsonRpcResponse, jsonRpcRequest)
  }

  @Test
  @Disabled("Api Method replaced by V1 counterpart")
  fun generateConflatedTracesV0_FileNotFound() {
    val jsonRpcRequest = buildRpcJson("rollup_generateConflatedTracesToFileV0", "0x1", "0x5")
    val response = makeJsonRpcRequest(jsonRpcRequest)
    response.then().statusCode(200).contentType("application/json")

    val jsonRpcResponse = JsonObject(response.body.asString())
    assertThat(jsonRpcResponse.getValue("id")).isEqualTo(jsonRpcRequest.getValue("id"))
    assertThat(jsonRpcResponse.getValue("result")).isNull()

    jsonRpcResponse.getValue("error").let { error ->
      assertThat(error).isNotNull
      error as JsonObject
      assertThat(error.getValue("code")).isEqualTo(-4001)
      assertThat(error.getValue("message")).isEqualTo("Traces not available")
      assertThat(error.getValue("data")).isEqualTo("Traces not available for block 4.")
    }
  }

  private fun assertTracesCountersResponse(jsonRpcResponse: JsonObject, jsonRpcRequestObject: JsonObject) {
    val jsonRpcRequest = JsonObject(jsonRpcRequestObject.toString())
    assertThat(jsonRpcResponse.getValue("id")).isEqualTo(jsonRpcRequest.getValue("id"))
    assertThat(jsonRpcResponse.getValue("error")).isNull()
    jsonRpcResponse.getValue("result").let { result ->
      assertThat(result).isNotNull
      result as JsonObject
      assertThat(result.getValue("tracesEngineVersion")).isEqualTo(tracesVersion)
      assertThat(result.getString("blockNumber")).isEqualTo(
        (jsonRpcRequest.getJsonArray("params").get(0) as JsonObject).getString("blockNumber")
      )

      result.getValue("tracesCounters").let { tracesCounters ->
        assertThat(tracesCounters).isNotNull
        tracesCounters as JsonObject
        for (traceModule in TracingModule.values()) {
          assertThat(tracesCounters.getValue(traceModule.name) as Int).isGreaterThanOrEqualTo(0)
        }
      }
      assertThat(result.getString("blockL1Size")).isEqualTo(blockL1Size.toString())
    }
  }

  private fun assertTracesConflationResponse(jsonRpcResponse: JsonObject, jsonRpcRequestObject: JsonObject) {
    val jsonRpcRequest = JsonObject(jsonRpcRequestObject.toString())
    assertThat(jsonRpcResponse.getValue("id")).isEqualTo(jsonRpcRequest.getValue("id"))
    assertThat(jsonRpcResponse.getValue("error")).isNull()
    jsonRpcResponse.getValue("result").let { result ->
      assertThat(result).isNotNull
      result as JsonObject
      assertThat(result.getValue("tracesEngineVersion")).isEqualTo(tracesVersion)
      assertThat(result.getString("startBlockNumber")).isEqualTo(
        (jsonRpcRequest.getJsonArray("params").first() as JsonObject).getString("blockNumber")
      )
      assertThat(result.getString("endBlockNumber")).isEqualTo(
        (jsonRpcRequest.getJsonArray("params").last() as JsonObject).getString("blockNumber")
      )

      result.getValue("conflatedTracesFileName").let { fileName ->
        assertThat(fileName).isNotNull
        assertThat(fileName).isInstanceOf(String::class.java)
        fileName as String
        assertThat(fileName).matches(".*.json.gz")
        assertThat(Path.of(appConfig.outputTracesDirectory, fileName)).isRegularFile()
      }
    }
  }

  @Test
  fun tracesCountersV1_tracesExist() {
    val jsonRpcRequest = buildRpcJson(
      "rollup_getBlockTracesCountersV1",
      mapOf(
        "blockNumber" to "1",
        "blockHash" to "0xab538e7ab831af9442aab00443ee9803907654359dfcdfe1755f1a98fb87eafd"
      )
    )
    val response = makeJsonRpcRequest(jsonRpcRequest)
    response.then().statusCode(200).contentType("application/json")

    assertTracesCountersResponse(JsonObject(response.body.asString()), jsonRpcRequest)
  }

  @Test
  fun tracesCountersV1_FileNotFound() {
    val jsonRpcRequest = buildRpcJson(
      "rollup_getBlockTracesCountersV1",
      mapOf(
        "blockNumber" to "1",
        "blockHash" to "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
      )
    )
    val response = makeJsonRpcRequest(jsonRpcRequest)
    response.then().statusCode(200).contentType("application/json")

    val jsonRpcResponse = JsonObject(response.body.asString())
    assertThat(jsonRpcResponse.getValue("id")).isEqualTo(jsonRpcRequest.getValue("id"))
    assertThat(jsonRpcResponse.getValue("result")).isNull()

    jsonRpcResponse.getValue("error").let { error ->
      assertThat(error).isNotNull
      error as JsonObject
      assertThat(error.getValue("code")).isEqualTo(-4001)
      assertThat(error.getValue("message")).isEqualTo("Traces not available")
      assertThat(error.getString("data")).startsWith("Traces not available for block 1.")
    }
  }

  @Test
  fun generateConflatedTracesToFileV1_success() {
    val jsonRpcRequest = buildRpcJson(
      "rollup_generateConflatedTracesToFileV1",
      mapOf(
        "blockNumber" to "1",
        "blockHash" to "0xab538e7ab831af9442aab00443ee9803907654359dfcdfe1755f1a98fb87eafd"
      ),
      mapOf(
        "blockNumber" to "2",
        "blockHash" to "0x0b68ebab401c813394f4c6139c743b6e5c72fe2da68c660c7a7616e2519a66a7"
      ),
      mapOf(
        "blockNumber" to "3",
        "blockHash" to "0x833d27bb3b09544c2de8ddf7e4e1c95557ebafdba0a308d59ba016e793eac568"
      )
    )
    val response = makeJsonRpcRequest(jsonRpcRequest)
    response.then().statusCode(200).contentType("application/json")

    val jsonRpcResponse = JsonObject(response.body.asString())
    assertTracesConflationResponse(jsonRpcResponse, jsonRpcRequest)
  }

  @Test
  fun generateConflatedTracesV1_FileNotFound() {
    val jsonRpcRequest = buildRpcJson(
      "rollup_generateConflatedTracesToFileV1",
      mapOf(
        "blockNumber" to "1",
        "blockHash" to "0xab538e7ab831af9442aab00443ee9803907654359dfcdfe1755f1a98fb87eafd"
      ),
      mapOf(
        "blockNumber" to "2",
        "blockHash" to "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
      ),
      mapOf(
        "blockNumber" to "3",
        "blockHash" to "0x833d27bb3b09544c2de8ddf7e4e1c95557ebafdba0a308d59ba016e793eac568"
      )
    )
    val response = makeJsonRpcRequest(jsonRpcRequest)
    response.then().statusCode(200).contentType("application/json")

    val jsonRpcResponse = JsonObject(response.body.asString())
    assertThat(jsonRpcResponse.getValue("id")).isEqualTo(jsonRpcRequest.getValue("id"))
    assertThat(jsonRpcResponse.getValue("result")).isNull()

    jsonRpcResponse.getValue("error").let { error ->
      assertThat(error).isNotNull
      error as JsonObject
      assertThat(error.getValue("code")).isEqualTo(-4001)
      assertThat(error.getValue("message")).isEqualTo("Traces not available")
      assertThat(error.getString("data")).startsWith("Traces not available for block 2.")
    }
  }

  private val monitorRequestSpecification =
    RequestSpecBuilder()
      // enable fro debug only
      // .addFilters(listOf(ResponseLoggingFilter(), RequestLoggingFilter()))
      .setBaseUri("http://localhost:${apiConfig.observabilityPort}/")
      .build()

  @Test
  fun exposesLiveEndpoint() {
    RestAssured.given()
      .spec(monitorRequestSpecification)
      .`when`()
      .get("/live")
      .then()
      .statusCode(200)
      .contentType(ContentType.JSON)
      .body("status", equalTo("OK"))
  }

  @Test
  fun exposesHealthEndpoint() {
    RestAssured.given()
      .spec(monitorRequestSpecification)
      .`when`()
      .get("/live")
      .then()
      .statusCode(200)
      .contentType(ContentType.JSON)
      .body("status", equalTo("OK"))
  }

  @Test
  fun exposesMetricsEndpoint() {
    RestAssured.given()
      .spec(monitorRequestSpecification)
      .accept(ContentType.JSON)
      .`when`()
      .get("/metrics")
      .also { assertThat(it.body.asString()).contains("vertx_http_server_response_bytes_bucket") }
      .then()
      .statusCode(200)
      .contentType(ContentType.TEXT)
  }

  @Test
  private fun makeJsonRpcRequest(request: JsonObject): Response {
    return RestAssured.given()
      .spec(requestSpecification)
      .accept(ContentType.JSON)
      .body(request.toString())
      .`when`()
      .post("/")
  }

  private fun buildRpcJson(method: String, vararg params: Any): JsonObject {
    return JsonObject()
      .put("id", "1")
      .put("jsonrpc", "2.0")
      .put("method", method)
      .put("params", params)
  }
}
