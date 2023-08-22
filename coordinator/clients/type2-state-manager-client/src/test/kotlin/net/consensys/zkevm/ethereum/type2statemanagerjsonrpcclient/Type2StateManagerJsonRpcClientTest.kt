package net.consensys.zkevm.ethereum.type2statemanagerjsonrpcclient

import com.fasterxml.jackson.databind.node.ArrayNode
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.tomakehurst.wiremock.WireMockServer
import com.github.tomakehurst.wiremock.client.WireMock.containing
import com.github.tomakehurst.wiremock.client.WireMock.ok
import com.github.tomakehurst.wiremock.client.WireMock.post
import com.github.tomakehurst.wiremock.core.WireMockConfiguration.options
import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import io.vertx.core.Vertx
import io.vertx.core.http.HttpClientOptions
import io.vertx.core.http.HttpVersion
import io.vertx.core.json.JsonObject
import io.vertx.junit5.VertxExtension
import net.consensys.linea.async.get
import net.consensys.linea.errors.ErrorResponse
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClient
import net.consensys.zkevm.coordinator.clients.GetZkEVMStateMerkleProofResponse
import net.consensys.zkevm.coordinator.clients.Type2StateManagerErrorType
import net.consensys.zkevm.coordinator.clients.Type2StateManagerJsonRpcClient
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatIllegalArgumentException
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.net.URI
import java.nio.file.Path

@ExtendWith(VertxExtension::class)
class Type2StateManagerJsonRpcClientTest {
  private lateinit var wiremock: WireMockServer
  private lateinit var type2StateManagerJsonRpcClient: Type2StateManagerJsonRpcClient
  private lateinit var meterRegistry: SimpleMeterRegistry

  private fun wiremockStubForPost(response: JsonObject) {
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

  @BeforeEach
  fun setup(vertx: Vertx) {
    wiremock = WireMockServer(options().dynamicPort())
    wiremock.start()

    val uri = URI("http://127.0.0.1:" + wiremock.port())
    val clientOptions =
      HttpClientOptions()
        .setKeepAlive(true)
        .setProtocolVersion(HttpVersion.HTTP_1_1)
        .setMaxPoolSize(10)
        .setDefaultHost(uri.host)
        .setDefaultPort(uri.port)
        .setLogActivity(true)
    val httpClient = vertx.createHttpClient(clientOptions)
    meterRegistry = SimpleMeterRegistry()
    val vertxHttpJsonRpcClient = VertxHttpJsonRpcClient(httpClient, "/", meterRegistry)
    val clientConfig = Type2StateManagerJsonRpcClient.Config("0.0.1-dev-3e607237")
    type2StateManagerJsonRpcClient =
      Type2StateManagerJsonRpcClient(vertxHttpJsonRpcClient, clientConfig)
  }

  @AfterEach
  fun tearDown(vertx: Vertx) {
    val vertxStopFuture = vertx.close()
    wiremock.stop()
    vertxStopFuture.get()
  }

  @Test
  fun getZkEVMStateMerkleProof() {
    val testFilePath = "../../../testdata/type2state-manager/state-proof.json"
    val json = jacksonObjectMapper().readTree(Path.of(testFilePath).toFile())
    val zkStateManagerVersion = json.get("zkStateManagerVersion").asText()
    val zkStateMerkleProof = json.get("zkStateMerkleProof") as ArrayNode
    val zkParentStateRootHash = json.get("zkParentStateRootHash").asText()
    val startBlockNumber = 50L
    val endBlockNumber = 100L

    val response =
      JsonObject.of(
        "jsonrpc",
        "2.0",
        "id",
        "1",
        "result",
        mapOf(
          "zkParentStateRootHash" to zkParentStateRootHash,
          "zkStateMerkleProof" to zkStateMerkleProof,
          "zkStateManagerVersion" to zkStateManagerVersion
        )
      )

    wiremockStubForPost(response)

    val resultFuture =
      type2StateManagerJsonRpcClient.rollupGetZkEVMStateMerkleProof(
        UInt64.valueOf(startBlockNumber),
        UInt64.valueOf(endBlockNumber)
      )
    resultFuture.get()

    assertThat(resultFuture)
      .isCompletedWithValue(
        Ok(
          GetZkEVMStateMerkleProofResponse(
            zkStateManagerVersion = zkStateManagerVersion,
            zkStateMerkleProof = zkStateMerkleProof,
            zkParentStateRootHash = Bytes32.fromHexString(zkParentStateRootHash)
          )
        )
      )
  }

  @Test
  fun error_block_missing_getZkEVMStateMerkleProof() {
    val errorMessage = "BLOCK_MISSING_IN_CHAIN - block 1 is missing"
    val startBlockNumber = 50L
    val endBlockNumber = 100L

    val response =
      JsonObject.of(
        "jsonrpc",
        "2.0",
        "id",
        "1",
        "error",
        mapOf("code" to "-32600", "message" to errorMessage)
      )

    wiremockStubForPost(response)

    val resultFuture =
      type2StateManagerJsonRpcClient.rollupGetZkEVMStateMerkleProof(
        UInt64.valueOf(startBlockNumber),
        UInt64.valueOf(endBlockNumber)
      )
    resultFuture.get()

    assertThat(resultFuture)
      .isCompletedWithValue(
        Err(ErrorResponse(Type2StateManagerErrorType.BLOCK_MISSING_IN_CHAIN, errorMessage))
      )
  }

  @Test
  fun error_unsupported_version_getZkEVMStateMerkleProof() {
    val startBlockNumber = 50L
    val endBlockNumber = 100L
    val errorMessage = "UNSUPPORTED_VERSION"
    val errorData =
      mapOf(
        "requestedVersion" to "0.0.1-dev-3e607217",
        "supportedVersion" to "0.0.1-dev-3e607237"
      )

    val response =
      JsonObject.of(
        "jsonrpc",
        "2.0",
        "id",
        "1",
        "error",
        mapOf("code" to "-32602", "message" to errorMessage, "data" to errorData)
      )

    wiremockStubForPost(response)

    val resultFuture =
      type2StateManagerJsonRpcClient.rollupGetZkEVMStateMerkleProof(
        UInt64.valueOf(startBlockNumber),
        UInt64.valueOf(endBlockNumber)
      )
    resultFuture.get()

    assertThat(resultFuture)
      .isCompletedWithValue(
        Err(
          ErrorResponse(
            Type2StateManagerErrorType.UNSUPPORTED_VERSION,
            "$errorMessage: $errorData"
          )
        )
      )
  }

  @Test
  fun error_unknown_getZkEVMStateMerkleProof() {
    val startBlockNumber = 50L
    val endBlockNumber = 100L
    val errorMessage = "BRA_BRA_BRA_SOME_UNKNOWN_ERROR"
    val errorData = mapOf("xyz" to "1234", "abc" to 100L)

    val response =
      JsonObject.of(
        "jsonrpc",
        "2.0",
        "id",
        "1",
        "error",
        mapOf("code" to "-999", "message" to errorMessage, "data" to errorData)
      )

    wiremockStubForPost(response)

    val resultFuture =
      type2StateManagerJsonRpcClient.rollupGetZkEVMStateMerkleProof(
        UInt64.valueOf(startBlockNumber),
        UInt64.valueOf(endBlockNumber)
      )
    resultFuture.get()

    assertThat(resultFuture)
      .isCompletedWithValue(
        Err(ErrorResponse(Type2StateManagerErrorType.UNKNOWN, "$errorMessage: $errorData"))
      )
  }

  @Test
  fun error_invalid_start_end_block_number_getZkEVMStateMerkleProof() {
    val startBlockNumber = 100L
    val endBlockNumber = 50L

    assertThatIllegalArgumentException()
      .isThrownBy {
        val resultFuture =
          type2StateManagerJsonRpcClient.rollupGetZkEVMStateMerkleProof(
            UInt64.valueOf(startBlockNumber),
            UInt64.valueOf(endBlockNumber)
          )
        resultFuture.get()
      }
      .withMessageContaining("endBlockNumber must be greater than startBlockNumber")
  }
}
