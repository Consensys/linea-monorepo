package net.consensys.linea.transactionexclusion

import com.sksamuel.hoplite.Masked
import io.restassured.RestAssured
import io.restassured.builder.RequestSpecBuilder
import io.restassured.http.ContentType
import io.restassured.specification.RequestSpecification
import io.vertx.junit5.VertxExtension
import kotlinx.datetime.Clock
import net.consensys.linea.async.get
import net.consensys.linea.transactionexclusion.app.AppConfig
import net.consensys.linea.transactionexclusion.app.DatabaseConfig
import net.consensys.linea.transactionexclusion.app.DbCleanupConfig
import net.consensys.linea.transactionexclusion.app.DbConnectionConfig
import net.consensys.linea.transactionexclusion.app.PersistenceRetryConfig
import net.consensys.linea.transactionexclusion.app.TransactionExclusionApp
import net.consensys.linea.transactionexclusion.app.api.ApiConfig
import net.consensys.trimToMillisecondPrecision
import net.consensys.zkevm.persistence.db.DbHelper
import net.consensys.zkevm.persistence.db.test.CleanDbTestSuiteParallel
import net.javacrumbs.jsonunit.assertj.JsonAssertions.assertThatJson
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import java.time.Duration

@ExtendWith(VertxExtension::class)
class TransactionExclusionAppV2Test : CleanDbTestSuiteParallel() {
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

  @BeforeEach()
  fun beforeEach() {
    app = TransactionExclusionApp(
      config = AppConfig(
        api = ApiConfig(
          port = 0, // port will be assigned under os
          observabilityPort = 0, // port will be assigned under os
          numberOfVerticles = 1
        ),
        database = dbConfig,
        dataQueryableWindowSinceRejectedTimestamp = Duration.parse("P7D")
      )
    )
    app.start().get()

    requestSpecification = RequestSpecBuilder()
      .setBaseUri("http://localhost:${app.apiBindedPort}/")
      .build()
  }

  @AfterEach
  fun afterEach() {
    app.stop().get()
  }

  private fun makeRequestJsonResponse(request: String): String {
    return RestAssured.given()
      .spec(requestSpecification)
      .accept(ContentType.JSON)
      .body(request)
      .`when`()
      .post("/")
      .then()
      .statusCode(200)
      .contentType("application/json")
      .extract()
      .asString()
  }

  @Test
  fun `when transaction is saved it should be retrievable`() {
    val rejectionTimeStamp = Clock.System.now()
      .trimToMillisecondPrecision()
      .toString()

    val saveTxJonRequest = """
      {
        "id": "123",
        "jsonrpc": "2.0",
        "method": "linea_saveRejectedTransactionV1",
        "params": [{
            "txRejectionStage": "SEQUENCER",
            "timestamp": "$rejectionTimeStamp",
            "transactionRLP": "0x02f8388204d2648203e88203e88203e8941195cf65f83b3a5768f3c496d3a05ad6412c64b38203e88c666d93e9cc5f73748162cea9c0017b8201c8",
            "blockNumber": "10000",
            "reasonMessage": "Transaction line count for module ADD=402 is above the limit 70",
            "overflows": [
              { "module": "ADD", "count": 402, "limit": 70 },
              { "module": "MUL", "count": 587, "limit": 401}
            ]
          }]
      }
    """.trimIndent()

    assertThatJson(makeRequestJsonResponse(saveTxJonRequest))
      .isEqualTo(
        """{
        "jsonrpc":"2.0",
        "id":"123",
        "result": {"status":"SAVED","txHash":"0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350ad7"}
        }"""
      )

    val getTxJsonRequest = """{
        "id": "124",
        "jsonrpc": "2.0",
        "method": "linea_getTransactionExclusionStatusV1",
        "params": ["0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350ad7"]
      }
    """.trimIndent()

    assertThatJson(makeRequestJsonResponse(getTxJsonRequest))
      .isEqualTo(
        """{
        "jsonrpc":"2.0",
        "id":"124",
        "result": {
          "txHash": "0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350ad7",
          "from": "0x4d144d7b9c96b26361d6ac74dd1d8267edca4fc2",
          "nonce": "0x64",
          "txRejectionStage": "SEQUENCER",
          "reasonMessage": "Transaction line count for module ADD=402 is above the limit 70",
          "timestamp": "$rejectionTimeStamp",
          "blockNumber": "0x2710"
        }}"""
      )
  }

  @Test
  fun `when transaction request is invalid shall return error`() {
    val saveTxJonRequest = """
      {
        "id": "123",
        "jsonrpc": "2.0",
        "method": "linea_saveRejectedTransactionV1",
        "params": [{
            "txRejectionStage": "SEQUENCER",
            "transactionRLP": "0x02f8388204d2648203e88203e88203e8941195cf65f83b3a5768f3c496d3a05ad6412c64b38203e88c666d93e9cc5f73748162cea9c0017b8201c8",
            "blockNumber": "10000",
            "reasonMessage": "Transaction line count for module ADD=402 is above the limit 70",
          }
        ]
      }
    """.trimIndent()
    assertThatJson(makeRequestJsonResponse(saveTxJonRequest))
      .isEqualTo(
        """{
        "jsonrpc":"2.0",
        "id":"123",
        "error": {
          "code":-32602,
          "message":"Missing [timestamp,reasonMessage] from the given request params"
        }
        }"""
      )
  }
}
