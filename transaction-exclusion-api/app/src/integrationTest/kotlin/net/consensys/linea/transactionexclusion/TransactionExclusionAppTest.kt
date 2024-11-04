package net.consensys.linea.transactionexclusion

import com.sksamuel.hoplite.Masked
import io.restassured.RestAssured
import io.restassured.builder.RequestSpecBuilder
import io.restassured.http.ContentType
import io.restassured.specification.RequestSpecification
import io.vertx.junit5.VertxExtension
import kotlinx.datetime.Clock
import net.consensys.encodeHex
import net.consensys.linea.async.get
import net.consensys.linea.transactionexclusion.app.AppConfig
import net.consensys.linea.transactionexclusion.app.DatabaseConfig
import net.consensys.linea.transactionexclusion.app.DbCleanupConfig
import net.consensys.linea.transactionexclusion.app.DbConnectionConfig
import net.consensys.linea.transactionexclusion.app.PersistenceRetryConfig
import net.consensys.linea.transactionexclusion.app.TransactionExclusionApp
import net.consensys.linea.transactionexclusion.app.api.ApiConfig
import net.consensys.linea.transactionexclusion.test.defaultRejectedTransaction
import net.consensys.linea.transactionexclusion.test.rejectedContractDeploymentTransaction
import net.consensys.toHexString
import net.consensys.trimToMillisecondPrecision
import net.consensys.zkevm.persistence.db.DbHelper
import net.consensys.zkevm.persistence.db.test.CleanDbTestSuiteParallel
import net.javacrumbs.jsonunit.assertj.JsonAssertions.assertThatJson
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import java.time.Duration
import kotlin.random.Random

@ExtendWith(VertxExtension::class)
class TransactionExclusionAppTest : CleanDbTestSuiteParallel() {
  init {
    target = "2"
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

  private fun saveFirstRejectedTransaction() {
    val saveTxJonRequest = """{
      "jsonrpc": "2.0",
      "id": 123,
      "method": "linea_saveRejectedTransactionV1",
      "params": {
        "txRejectionStage": "P2P",
        "timestamp": "${Clock.System.now().trimToMillisecondPrecision()}",
        "transactionRLP": "${defaultRejectedTransaction.transactionRLP.encodeHex()}",
        "reasonMessage": "Transaction line count for module ADD=402 is above the limit 70",
        "overflows": [
          { "module": "ADD", "count": 402, "limit": 70 },
          { "module": "MUL", "count": 587, "limit": 401 }
        ]
      }
    }
    """.trimIndent()

    // Check the save response and ensure the rejected txn was saved
    assertThatJson(makeRequestJsonResponse(saveTxJonRequest))
      .isEqualTo(
        """{
          "jsonrpc": "2.0",
          "id": 123,
          "result": {"status":"SAVED","txHash":"${defaultRejectedTransaction.transactionInfo.hash.encodeHex()}"}
        }"""
      )
  }

  @Test
  fun `Should save the rejected tx from P2P and then from SEQUENCER with same txHash but different reason message`() {
    // Save the first rejected tx from P2P without rejected block number
    saveFirstRejectedTransaction()

    // Save the rejected tx from SEQUENCER with rejected block number and different
    // rejected reason and a more recent rejected timestamp
    val rejectionTimeStamp = Clock.System.now()
      .trimToMillisecondPrecision()
      .toString()

    val saveTxJonRequest = """{
      "jsonrpc": "2.0",
      "id": 124,
      "method": "linea_saveRejectedTransactionV1",
      "params": [{
        "txRejectionStage": "SEQUENCER",
        "timestamp": "$rejectionTimeStamp",
        "transactionRLP": "${defaultRejectedTransaction.transactionRLP.encodeHex()}",
        "blockNumber": "10000",
        "reasonMessage": "Transaction line count for module ADD=402 is above the limit 70 (from Sequencer)",
        "overflows": [
          { "module": "ADD", "count": 402, "limit": 70 },
          { "module": "MUL", "count": 587, "limit": 401 }
        ]
      }]
    }
    """.trimIndent()

    // Check the save response and ensure the rejected txn was saved
    assertThatJson(makeRequestJsonResponse(saveTxJonRequest))
      .isEqualTo(
        """{
          "jsonrpc": "2.0",
          "id": 124,
          "result": {"status":"SAVED","txHash":"${defaultRejectedTransaction.transactionInfo.hash.encodeHex()}"}
        }"""
      )

    // Send the get request for the rejected transaction
    val getTxJsonRequest = """{
      "jsonrpc": "2.0",
      "id": 125,
      "method": "linea_getTransactionExclusionStatusV1",
      "params": ["${defaultRejectedTransaction.transactionInfo.hash.encodeHex()}"]
    }
    """.trimIndent()

    // Check the get response is corresponding to the rejected txn from SEQUENCER
    assertThatJson(makeRequestJsonResponse(getTxJsonRequest))
      .isEqualTo(
        """{
          "jsonrpc": "2.0",
          "id": 125,
          "result": {
            "txHash": "${defaultRejectedTransaction.transactionInfo.hash.encodeHex()}",
            "from": "${defaultRejectedTransaction.transactionInfo.from.encodeHex()}",
            "nonce": "0x64",
            "txRejectionStage": "SEQUENCER",
            "reasonMessage": "Transaction line count for module ADD=402 is above the limit 70 (from Sequencer)",
            "timestamp": "$rejectionTimeStamp",
            "blockNumber": "0x2710"
          }
        }"""
      )
  }

  @Test
  fun `Should save the rejected contract deployment tx from RPC`() {
    // Save the rejected contract deployment tx from RPC (without block number and sender address)
    val rejectionTimeStamp = Clock.System.now()
      .trimToMillisecondPrecision()
      .toString()

    val saveTxJonRequest = """{
      "jsonrpc": "2.0",
      "id": 124,
      "method": "linea_saveRejectedTransactionV1",
      "params": [{
        "txRejectionStage": "RPC",
        "timestamp": "$rejectionTimeStamp",
        "transactionRLP": "${rejectedContractDeploymentTransaction.transactionRLP.encodeHex()}",
        "reasonMessage": "${rejectedContractDeploymentTransaction.reasonMessage}",
        "overflows": []
      }]
    }
    """.trimIndent()

    // Check the save response and ensure the rejected txn was saved
    assertThatJson(makeRequestJsonResponse(saveTxJonRequest))
      .isEqualTo(
        """{
          "jsonrpc": "2.0",
          "id": 124,
          "result": {"status":"SAVED","txHash":"${
        rejectedContractDeploymentTransaction.transactionInfo.hash.encodeHex()
        }"}
        }"""
      )

    // Send the get request for the rejected transaction
    val getTxJsonRequest = """{
      "jsonrpc": "2.0",
      "id": 125,
      "method": "linea_getTransactionExclusionStatusV1",
      "params": ["${rejectedContractDeploymentTransaction.transactionInfo.hash.encodeHex()}"]
    }
    """.trimIndent()

    // Check the get response is corresponding to the rejected txn from SEQUENCER
    assertThatJson(makeRequestJsonResponse(getTxJsonRequest))
      .isEqualTo(
        """{
          "jsonrpc": "2.0",
          "id": 125,
          "result": {
            "txHash": "${rejectedContractDeploymentTransaction.transactionInfo.hash.encodeHex()}",
            "from": "${rejectedContractDeploymentTransaction.transactionInfo.from.encodeHex()}",
            "nonce": "${rejectedContractDeploymentTransaction.transactionInfo.nonce.toHexString()}",
            "txRejectionStage": "RPC",
            "reasonMessage": "${rejectedContractDeploymentTransaction.reasonMessage}",
            "timestamp": "$rejectionTimeStamp"
          }
        }"""
      )
  }

  @Test
  fun `Should return DUPLICATE_ALREADY_SAVED_BEFORE when saving rejected tx with same txHash and reason message`() {
    // Save the first rejected tx from P2P without rejected block number
    saveFirstRejectedTransaction()

    // Save the same rejected tx from SEQUENCER with rejected block number and a more recent rejected timestamp
    val rejectionTimeStamp = Clock.System.now()
      .trimToMillisecondPrecision()
      .toString()

    val saveTxJonRequest = """{
      "jsonrpc": "2.0",
      "id": 124,
      "method": "linea_saveRejectedTransactionV1",
      "params": [{
        "txRejectionStage": "SEQUENCER",
        "timestamp": "$rejectionTimeStamp",
        "transactionRLP": "${defaultRejectedTransaction.transactionRLP.encodeHex()}",
        "blockNumber": "10000",
        "reasonMessage": "Transaction line count for module ADD=402 is above the limit 70",
        "overflows": [
          { "module": "ADD", "count": 402, "limit": 70 },
          { "module": "MUL", "count": 587, "limit": 401 }
        ]
      }]
    }
    """.trimIndent()

    // Check the save response and ensure the status is "duplicated already saved before"
    assertThatJson(makeRequestJsonResponse(saveTxJonRequest))
      .isEqualTo(
        """{
          "jsonrpc": "2.0",
          "id": 124,
          "result": {"status":"DUPLICATE_ALREADY_SAVED_BEFORE","txHash":"${
        defaultRejectedTransaction.transactionInfo.hash.encodeHex()
        }"}
        }"""
      )
  }

  @Test
  fun `Should return result as null when getting the rejected tx with random transaction hash`() {
    // Save the first rejected tx from P2P without rejected block number
    saveFirstRejectedTransaction()

    // Send the get request with a random txn hash
    val getTxJsonRequest = """{
      "jsonrpc": "2.0",
      "id": 124,
      "method": "linea_getTransactionExclusionStatusV1",
      "params": ["${Random.nextBytes(32).encodeHex()}"]
    }
    """.trimIndent()

    // Check the get response and ensure the result is null
    assertThatJson(makeRequestJsonResponse(getTxJsonRequest))
      .isEqualTo(
        """{
          "jsonrpc": "2.0",
          "id": 124,
          "result": null
        }"""
      )
  }

  @Test
  fun `when transaction request is invalid shall return error`() {
    val saveTxJonRequest = """{
      "jsonrpc": "2.0",
      "id": 123,
      "method": "linea_saveRejectedTransactionV1",
      "params": [{
        "txRejectionStage": "SEQUENCER",
        "transactionRLP": "${defaultRejectedTransaction.transactionRLP.encodeHex()}",
        "blockNumber": "10000",
        "reasonMessage": "Transaction line count for module ADD=402 is above the limit 70"
      }]
    }
    """.trimIndent()

    assertThatJson(makeRequestJsonResponse(saveTxJonRequest))
      .isEqualTo(
        """{
          "jsonrpc": "2.0",
          "id": 123,
          "error": {
            "code": -32602,
            "message": "Missing [timestamp,overflows] from the given request params"
          }
        }"""
      )
  }
}
