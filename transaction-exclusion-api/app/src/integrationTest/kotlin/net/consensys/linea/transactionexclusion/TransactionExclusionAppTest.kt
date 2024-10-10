package net.consensys.linea.transactionexclusion

import com.sksamuel.hoplite.Masked
import io.restassured.RestAssured
import io.restassured.builder.RequestSpecBuilder
import io.restassured.http.ContentType
import io.restassured.response.Response
import io.restassured.specification.RequestSpecification
import io.vertx.core.json.JsonObject
import io.vertx.junit5.VertxExtension
import kotlinx.datetime.Clock
import net.consensys.encodeHex
import net.consensys.linea.async.get
import net.consensys.linea.transactionexclusion.TransactionExclusionServiceV1.SaveRejectedTransactionStatus
import net.consensys.linea.transactionexclusion.app.AppConfig
import net.consensys.linea.transactionexclusion.app.DatabaseConfig
import net.consensys.linea.transactionexclusion.app.DbCleanupConfig
import net.consensys.linea.transactionexclusion.app.DbConnectionConfig
import net.consensys.linea.transactionexclusion.app.PersistenceRetryConfig
import net.consensys.linea.transactionexclusion.app.TransactionExclusionApp
import net.consensys.linea.transactionexclusion.app.api.ApiConfig
import net.consensys.linea.transactionexclusion.test.defaultRejectedTransaction
import net.consensys.toHexString
import net.consensys.trimToMillisecondPrecision
import net.consensys.zkevm.persistence.db.DbHelper
import net.consensys.zkevm.persistence.db.test.CleanDbTestSuiteParallel
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.extension.ExtendWith
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.ValueSource
import java.time.Duration
import kotlin.random.Random

@ExtendWith(VertxExtension::class)
class TransactionExclusionAppTest : CleanDbTestSuiteParallel() {
  init {
    target = "1"
  }

  override var databaseName = DbHelper.generateUniqueDbName("tx-exclusion-api-app-tests")

  private val dbConfig = DatabaseConfig(
    read = DbConnectionConfig(
      host = "localhost",
      port = 5432,
      username = "postgres",
      password = Masked("postgres")
    ),
    write = DbConnectionConfig(
      host = "localhost",
      port = 5432,
      username = "postgres",
      password = Masked("postgres")
    ),
    cleanup = DbCleanupConfig(
      pollingInterval = Duration.parse("PT60S"),
      storagePeriod = Duration.parse("P7D")
    ),
    persistenceRetry = PersistenceRetryConfig(
      backoffDelay = Duration.parse("PT5S"),
      timeout = Duration.parse("PT20S")
    ),
    schema = databaseName
  )

  private lateinit var requestSpecification: RequestSpecification
  private lateinit var app: TransactionExclusionApp

  @AfterEach
  fun afterEach() {
    app.stop().get()
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

  private fun buildRpcJsonWithListParams(method: String, params: List<Any>): JsonObject {
    return JsonObject()
      .put("id", "1")
      .put("jsonrpc", "2.0")
      .put("method", method)
      .put("params", params)
  }

  private fun buildSaveJsonRpcRequestWithNamedParams(
    txRejectionStage: RejectedTransaction.Stage = defaultRejectedTransaction.txRejectionStage,
    timestamp: String = defaultRejectedTransaction.timestamp.toString(),
    transactionRLP: String = defaultRejectedTransaction.transactionRLP.encodeHex(),
    blockNumber: String? = defaultRejectedTransaction.blockNumber?.toString(),
    reasonMessage: String = defaultRejectedTransaction.reasonMessage,
    overflows: List<ModuleOverflow> = defaultRejectedTransaction.overflows
  ): JsonObject {
    return buildRpcJsonWithNamedParams(
      "linea_saveRejectedTransactionV1",
      buildSaveRequestMapParams(
        txRejectionStage = txRejectionStage,
        timestamp = timestamp,
        transactionRLP = transactionRLP,
        blockNumber = blockNumber,
        reasonMessage = reasonMessage,
        overflows = overflows
      )
    )
  }

  private fun buildSaveJsonRpcRequestWithListParams(
    txRejectionStage: RejectedTransaction.Stage = defaultRejectedTransaction.txRejectionStage,
    timestamp: String = defaultRejectedTransaction.timestamp.toString(),
    transactionRLP: String = defaultRejectedTransaction.transactionRLP.encodeHex(),
    blockNumber: String? = defaultRejectedTransaction.blockNumber?.toString(),
    reasonMessage: String = defaultRejectedTransaction.reasonMessage,
    overflows: List<ModuleOverflow> = defaultRejectedTransaction.overflows
  ): JsonObject {
    return buildRpcJsonWithListParams(
      "linea_saveRejectedTransactionV1",
      listOf(
        buildSaveRequestMapParams(
          txRejectionStage = txRejectionStage,
          timestamp = timestamp,
          transactionRLP = transactionRLP,
          blockNumber = blockNumber,
          reasonMessage = reasonMessage,
          overflows = overflows
        )
      )
    )
  }

  private fun buildSaveRequestMapParams(
    txRejectionStage: RejectedTransaction.Stage = defaultRejectedTransaction.txRejectionStage,
    timestamp: String = defaultRejectedTransaction.timestamp.toString(),
    transactionRLP: String = defaultRejectedTransaction.transactionRLP.encodeHex(),
    blockNumber: String? = defaultRejectedTransaction.blockNumber?.toString(),
    reasonMessage: String = defaultRejectedTransaction.reasonMessage,
    overflows: List<ModuleOverflow> = defaultRejectedTransaction.overflows
  ): Map<String, Any?> {
    return mapOf(
      "txRejectionStage" to txRejectionStage,
      "timestamp" to timestamp,
      "transactionRLP" to transactionRLP,
      "blockNumber" to blockNumber,
      "reasonMessage" to reasonMessage,
      "overflows" to overflows
    )
  }

  private fun buildGetJsonRpcRequest(
    txHash: String
  ): JsonObject {
    return buildRpcJsonWithListParams(
      "linea_getTransactionExclusionStatusV1",
      listOf(txHash)
    )
  }

  private fun assertSaveJsonRpcResponse(
    saveJsonRpcResponse: JsonObject,
    expectedStatus: SaveRejectedTransactionStatus,
    expectedTxHash: String
  ) {
    saveJsonRpcResponse.getValue("result").let { result ->
      assertThat(result).isNotNull
      result as JsonObject
      assertThat(result.getString("status")).isEqualTo(expectedStatus.toString())
      assertThat(result.getString("txHash")).isEqualTo(expectedTxHash)
    }
  }

  private fun assertGetJsonRpcResponse(
    getJsonRpcResponse: JsonObject,
    expectedTxRejectionStage: RejectedTransaction.Stage = defaultRejectedTransaction.txRejectionStage,
    expectedTxHash: String = defaultRejectedTransaction.transactionInfo.hash.encodeHex(),
    expectedReasonMessage: String = defaultRejectedTransaction.reasonMessage,
    expectedFrom: String = defaultRejectedTransaction.transactionInfo.from.encodeHex(),
    expectedNonce: String = defaultRejectedTransaction.transactionInfo.nonce.toHexString(),
    expectedBlockNumber: String? = defaultRejectedTransaction.blockNumber?.toHexString(),
    expectedTimestamp: String = defaultRejectedTransaction.timestamp.toString()
  ) {
    getJsonRpcResponse.getValue("result").let { result ->
      assertThat(result).isNotNull
      result as JsonObject
      assertThat(result.getString("txHash")).isEqualTo(expectedTxHash)
      assertThat(result.getString("txRejectionStage")).isEqualTo(expectedTxRejectionStage.toString())
      assertThat(result.getString("reasonMessage")).isEqualTo(expectedReasonMessage)
      assertThat(result.getString("from")).isEqualTo(expectedFrom)
      assertThat(result.getString("nonce")).isEqualTo(expectedNonce)
      assertThat(result.getString("blockNumber")).isEqualTo(expectedBlockNumber)
      assertThat(result.getString("timestamp")).isEqualTo(expectedTimestamp)
    }
  }

  private fun startTransactionExclusionApp(apiPort: Int) {
    requestSpecification = RequestSpecBuilder()
      .setBaseUri("http://localhost:$apiPort/")
      .build()

    app = TransactionExclusionApp(
      config = AppConfig(
        api = ApiConfig(
          port = apiPort,
          observabilityPort = apiPort + 10000,
          numberOfVerticles = 1
        ),
        database = dbConfig,
        dataQueryableWindowSinceRejectedTimestamp = Duration.parse("P7D")
      )
    )
    app.start().get()
  }

  @ParameterizedTest
  @ValueSource(ints = [8082])
  fun `Should save the rejected tx from P2P and then from SEQUENCER with same txHash but different reason message`
  (apiPort: Int) {
    // Start the transaction exclusion app with the given api port
    startTransactionExclusionApp(apiPort)

    // Save the rejected tx from P2P without rejected block number
    var rejectedTimestampISOString = Clock.System.now().trimToMillisecondPrecision().toString()
    var saveJsonRpcRequest = buildSaveJsonRpcRequestWithNamedParams(
      txRejectionStage = RejectedTransaction.Stage.P2P,
      timestamp = rejectedTimestampISOString,
      blockNumber = null // P2P node has no info on rejected block number
    )

    // Send the save request and ensure it returns OK status code
    var saveResponse = makeJsonRpcRequest(saveJsonRpcRequest)
    saveResponse.then().statusCode(200).contentType("application/json")

    // Check the save response and ensure the rejected txn was saved
    var saveJsonRpcResponse = JsonObject(saveResponse.body.asString())
    assertSaveJsonRpcResponse(
      saveJsonRpcResponse = saveJsonRpcResponse,
      expectedStatus = SaveRejectedTransactionStatus.SAVED,
      expectedTxHash = defaultRejectedTransaction.transactionInfo.hash.encodeHex()
    )

    // Send the get request and ensure it returns OK status code
    var getJsonRpcRequest = buildGetJsonRpcRequest(
      defaultRejectedTransaction.transactionInfo.hash.encodeHex()
    )
    var getResponse = makeJsonRpcRequest(getJsonRpcRequest)
    getResponse.then().statusCode(200).contentType("application/json")

    // Check the get response is corresponding to the rejected txn from P2P
    var getJsonRpcResponse = JsonObject(getResponse.body.asString())
    assertGetJsonRpcResponse(
      getJsonRpcResponse = getJsonRpcResponse,
      expectedTxRejectionStage = RejectedTransaction.Stage.P2P,
      expectedBlockNumber = null,
      expectedTimestamp = rejectedTimestampISOString
    )

    // Save the rejected tx from SEQUENCER with rejected block number and different
    // rejected reason and a more recent rejected timestamp
    rejectedTimestampISOString = Clock.System.now().trimToMillisecondPrecision().toString()
    val rejectedReasonMessage = "Transaction line count for module MUL=587 is above the limit 400 (from e2e test)"
    saveJsonRpcRequest = buildSaveJsonRpcRequestWithListParams(
      txRejectionStage = RejectedTransaction.Stage.SEQUENCER,
      timestamp = rejectedTimestampISOString,
      reasonMessage = rejectedReasonMessage
    )

    // Send the save request and ensure it returns OK status code
    saveResponse = makeJsonRpcRequest(saveJsonRpcRequest)
    saveResponse.then().statusCode(200).contentType("application/json")

    // Check the save response and ensure the rejected txn was saved
    saveJsonRpcResponse = JsonObject(saveResponse.body.asString())
    assertSaveJsonRpcResponse(
      saveJsonRpcResponse = saveJsonRpcResponse,
      expectedStatus = SaveRejectedTransactionStatus.SAVED,
      expectedTxHash = defaultRejectedTransaction.transactionInfo.hash.encodeHex()
    )

    // Send the get request and ensure it returns OK status code
    getJsonRpcRequest = buildGetJsonRpcRequest(
      defaultRejectedTransaction.transactionInfo.hash.encodeHex()
    )
    getResponse = makeJsonRpcRequest(getJsonRpcRequest)
    getResponse.then().statusCode(200).contentType("application/json")

    // Check the get response is corresponding to the rejected txn from SEQUENCER
    getJsonRpcResponse = JsonObject(getResponse.body.asString())
    assertGetJsonRpcResponse(
      getJsonRpcResponse = getJsonRpcResponse,
      expectedTxRejectionStage = RejectedTransaction.Stage.SEQUENCER,
      expectedReasonMessage = rejectedReasonMessage,
      expectedTimestamp = rejectedTimestampISOString
    )
  }

  @ParameterizedTest
  @ValueSource(ints = [8083])
  fun `Should return DUPLICATE_ALREADY_SAVED_BEFORE when saving rejected tx with same txHash and reason message`
  (apiPort: Int) {
    // Start the transaction exclusion app with the given api port
    startTransactionExclusionApp(apiPort)

    // Save the rejected tx from P2P without rejected block number
    var rejectedTimestampISOString = Clock.System.now().trimToMillisecondPrecision().toString()
    var saveJsonRpcRequest = buildSaveJsonRpcRequestWithNamedParams(
      txRejectionStage = RejectedTransaction.Stage.P2P,
      timestamp = rejectedTimestampISOString,
      blockNumber = null
    )

    // Send the save request and ensure it returns OK status code
    var saveResponse = makeJsonRpcRequest(saveJsonRpcRequest)
    saveResponse.then().statusCode(200).contentType("application/json")

    // Check the save response and ensure the rejected txn was saved
    var saveJsonRpcResponse = JsonObject(saveResponse.body.asString())
    assertSaveJsonRpcResponse(
      saveJsonRpcResponse = saveJsonRpcResponse,
      expectedStatus = SaveRejectedTransactionStatus.SAVED,
      expectedTxHash = defaultRejectedTransaction.transactionInfo.hash.encodeHex()
    )

    // Save the same rejected tx from SEQUENCER with rejected block number and a more recent rejected timestamp
    rejectedTimestampISOString = Clock.System.now().trimToMillisecondPrecision().toString()
    saveJsonRpcRequest = buildSaveJsonRpcRequestWithNamedParams(
      txRejectionStage = RejectedTransaction.Stage.SEQUENCER,
      timestamp = rejectedTimestampISOString
    )

    // Send the save request and ensure it returns OK status code
    saveResponse = makeJsonRpcRequest(saveJsonRpcRequest)
    saveResponse.then().statusCode(200).contentType("application/json")

    // Check the save response and ensure the status is "duplicated already saved before"
    saveJsonRpcResponse = JsonObject(saveResponse.body.asString())
    assertSaveJsonRpcResponse(
      saveJsonRpcResponse = saveJsonRpcResponse,
      expectedStatus = SaveRejectedTransactionStatus.DUPLICATE_ALREADY_SAVED_BEFORE,
      expectedTxHash = defaultRejectedTransaction.transactionInfo.hash.encodeHex()
    )
  }

  @ParameterizedTest
  @ValueSource(ints = [8084])
  fun `Should return result as null when getting the rejected tx with random transaction hash`(apiPort: Int) {
    // Start the transaction exclusion app with the given api port
    startTransactionExclusionApp(apiPort)

    // Save the rejected tx from P2P without rejected block number
    var rejectedTimestampISOString = Clock.System.now().trimToMillisecondPrecision().toString()
    var saveJsonRpcRequest = buildSaveJsonRpcRequestWithNamedParams(
      txRejectionStage = RejectedTransaction.Stage.P2P,
      timestamp = rejectedTimestampISOString,
      blockNumber = null
    )

    // Send the save request and ensure it returns OK status code
    var saveResponse = makeJsonRpcRequest(saveJsonRpcRequest)
    saveResponse.then().statusCode(200).contentType("application/json")

    // Check the save response and ensure the rejected txn was saved
    var saveJsonRpcResponse = JsonObject(saveResponse.body.asString())
    assertSaveJsonRpcResponse(
      saveJsonRpcResponse = saveJsonRpcResponse,
      expectedStatus = SaveRejectedTransactionStatus.SAVED,
      expectedTxHash = defaultRejectedTransaction.transactionInfo.hash.encodeHex()
    )

    // Send the get request and ensure it returns OK status code
    val getJsonRpcRequest = buildGetJsonRpcRequest(
      Random.nextBytes(32).encodeHex()
    )
    val getResponse = makeJsonRpcRequest(getJsonRpcRequest)
    getResponse.then().statusCode(200).contentType("application/json")

    // Check the get response and ensure the result is null
    val getJsonRpcResponse = JsonObject(getResponse.body.asString())
    getJsonRpcResponse.getValue("result").let { result ->
      assertThat(result).isNull()
    }
  }

  @ParameterizedTest
  @ValueSource(ints = [8085])
  fun `Should return error result when saving rejected tx without rejected timestamp and reason message`(apiPort: Int) {
    // Start the transaction exclusion app with the given api port
    startTransactionExclusionApp(apiPort)

    // Save the rejected tx from P2P without rejected timestamp and reason message
    val saveJsonRpcRequest = buildRpcJsonWithNamedParams(
      "linea_saveRejectedTransactionV1",
      mapOf(
        "txRejectionStage" to RejectedTransaction.Stage.P2P,
        "transactionRLP" to defaultRejectedTransaction.transactionRLP.encodeHex(),
        "overflows" to defaultRejectedTransaction.overflows
      )
    )

    // Send the save request and ensure it returns OK status code
    val saveResponse = makeJsonRpcRequest(saveJsonRpcRequest)
    saveResponse.then().statusCode(200).contentType("application/json")

    // Check the save response and ensure the error with proper code and message
    val saveJsonRpcResponse = JsonObject(saveResponse.body.asString())
    saveJsonRpcResponse.getValue("result").let { result ->
      assertThat(result).isNull()
    }
    saveJsonRpcResponse.getValue("error").let { error ->
      error as JsonObject
      assertThat(error.getInteger("code")).isEqualTo(-32602)
      assertThat(error.getString("message")).isEqualTo(
        "Missing [timestamp,reasonMessage] from the given request params"
      )
    }
  }
}
