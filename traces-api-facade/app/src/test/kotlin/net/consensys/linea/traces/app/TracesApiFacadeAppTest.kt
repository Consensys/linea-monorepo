package net.consensys.linea.traces.app

import io.restassured.RestAssured
import io.restassured.builder.RequestSpecBuilder
import io.restassured.http.ContentType
import io.restassured.response.Response
import io.restassured.specification.RequestSpecification
import io.vertx.core.json.JsonObject
import net.consensys.linea.async.get
import net.consensys.linea.traces.TracingModuleV1
import net.consensys.linea.traces.app.api.ApiConfig
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.io.TempDir
import java.nio.file.Path

@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class TracesApiFacadeAppTest {
  private val apiConfig =
    ApiConfig(port = 9394u, observabilityPort = 9395u, path = "/", numberOfVerticles = 1u)
  private val rawTracesVersion = "0.0.1"
  private val tracesApiVersion = "0.0.2"
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
        tracesApiVersion = tracesApiVersion,
        api = apiConfig,
        tracesFileExtension = "json.gz"
      )
    app = TracesApiFacadeApp(appConfig)
    app.start().get()
  }

  @AfterEach
  fun afterEach() {
    app.stop().get()
  }

  private fun assertTracesCountersResponse(
    jsonRpcResponse: JsonObject,
    jsonRpcRequestObject: JsonObject
  ) {
    val jsonRpcRequest = JsonObject(jsonRpcRequestObject.toString())
    assertThat(jsonRpcResponse.getValue("id")).isEqualTo(jsonRpcRequest.getValue("id"))
    assertThat(jsonRpcResponse.getValue("error")).isNull()
    jsonRpcResponse.getValue("result").let { result ->
      assertThat(result).isNotNull
      result as JsonObject
      assertThat(result.getValue("tracesEngineVersion")).isEqualTo(tracesApiVersion)
      assertThat(result.getString("blockNumber")).isEqualTo(
        (
          jsonRpcRequest.getJsonObject("params")
            .getJsonObject("block")
          ).getString("blockNumber")
      )

      result.getValue("tracesCounters").let { tracesCounters ->
        assertThat(tracesCounters).isNotNull
        tracesCounters as JsonObject
        for (traceModule in TracingModuleV1.entries) {
          assertThat(tracesCounters.getValue(traceModule.name) as Int).isGreaterThanOrEqualTo(
            0
          )
        }
      }
      assertThat(result.getString("blockL1Size")).isEqualTo(blockL1Size.toString())
    }
  }

  private fun assertTracesConflationResponse(
    jsonRpcResponse: JsonObject,
    jsonRpcRequestObject: JsonObject
  ) {
    val jsonRpcRequest = JsonObject(jsonRpcRequestObject.toString())
    assertThat(jsonRpcResponse.getValue("id")).isEqualTo(jsonRpcRequest.getValue("id"))
    assertThat(jsonRpcResponse.getValue("error")).isNull()
    jsonRpcResponse.getValue("result").let { result ->
      assertThat(result).isNotNull
      result as JsonObject
      assertThat(result.getValue("tracesEngineVersion")).isEqualTo(tracesApiVersion)
      val blocks = jsonRpcRequest.getJsonObject("params").getJsonArray("blocks")
      assertThat(result.getString("startBlockNumber")).isEqualTo(
        (blocks.first() as JsonObject).getString("blockNumber")
      )
      assertThat(result.getString("endBlockNumber")).isEqualTo(
        (blocks.last() as JsonObject).getString("blockNumber")
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
    val jsonRpcRequest = buildRpcJsonWithNamedParams(
      "rollup_getBlockTracesCountersV1",
      mapOf(
        "block" to mapOf(
          "blockNumber" to "1",
          "blockHash" to "0xab538e7ab831af9442aab00443ee9803907654359dfcdfe1755f1a98fb87eafd"
        ),
        "rawExecutionTracesVersion" to rawTracesVersion,
        "expectedTracesApiVersion" to tracesApiVersion
      )
    )
    val response = makeJsonRpcRequest(jsonRpcRequest)
    response.then().statusCode(200).contentType("application/json")

    assertTracesCountersResponse(JsonObject(response.body.asString()), jsonRpcRequest)
  }

  @Test
  fun tracesCountersV1_incompatibleVersionIsRejected() {
    val incompatibleVersion = "1.0.2"
    val jsonRpcRequest = buildRpcJsonWithNamedParams(
      "rollup_getBlockTracesCountersV1",
      mapOf(
        "block" to mapOf(
          "blockNumber" to "1",
          "blockHash" to "0xab538e7ab831af9442aab00443ee9803907654359dfcdfe1755f1a98fb87eafd"
        ),
        "rawExecutionTracesVersion" to rawTracesVersion,
        "expectedTracesApiVersion" to incompatibleVersion
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
      assertThat(error.getValue("code")).isEqualTo(-32602)
      assertThat(error.getValue("message")).isEqualTo(
        "Client requested version $incompatibleVersion is not compatible to server version $tracesApiVersion"
      )
      assertThat(error.getString("data")).isNull()
    }
  }

  @Test
  fun tracesCountersV1_FileNotFound() {
    val jsonRpcRequest = buildRpcJsonWithNamedParams(
      "rollup_getBlockTracesCountersV1",
      mapOf(
        "block" to mapOf(
          "blockNumber" to "1",
          "blockHash" to "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
        ),
        "rawExecutionTracesVersion" to rawTracesVersion,
        "expectedTracesApiVersion" to tracesApiVersion
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
    val jsonRpcRequest = buildRpcJsonWithNamedParams(
      "rollup_generateConflatedTracesToFileV1",
      mapOf(
        "blocks" to listOf(
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
        ),
        "rawExecutionTracesVersion" to rawTracesVersion,
        "expectedTracesApiVersion" to tracesApiVersion
      )
    )
    val response = makeJsonRpcRequest(jsonRpcRequest)
    response.then().statusCode(200).contentType("application/json")

    val jsonRpcResponse = JsonObject(response.body.asString())
    assertTracesConflationResponse(jsonRpcResponse, jsonRpcRequest)
  }

  @Test
  fun generateConflatedTracesV1_FileNotFound() {
    val jsonRpcRequest = buildRpcJsonWithNamedParams(
      "rollup_generateConflatedTracesToFileV1",
      mapOf(
        "blocks" to listOf(
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
        ),
        "rawExecutionTracesVersion" to rawTracesVersion,
        "expectedTracesApiVersion" to tracesApiVersion
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

  private fun makeJsonRpcRequest(request: JsonObject): Response {
    return RestAssured.given()
      .spec(requestSpecification)
      .accept(ContentType.JSON)
      .body(request.toString())
      .`when`()
      .post("/")
  }

  private fun buildRpcJsonWithNamedParams(method: String, params: Map<String, Any?>): JsonObject {
    return JsonObject()
      .put("id", "1")
      .put("jsonrpc", "2.0")
      .put("method", method)
      .put("params", params)
  }
}
