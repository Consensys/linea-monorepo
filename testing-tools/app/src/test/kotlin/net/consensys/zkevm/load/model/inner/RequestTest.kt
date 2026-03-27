package net.consensys.zkevm.load.model.inner

import net.consensys.zkevm.load.model.JSON
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertNotNull
import org.junit.jupiter.api.Test
import java.io.InputStreamReader
import kotlin.test.assertFailsWith

class RequestTest {

  @Test
  fun translateCreateContractWithoutGasLimitFailsWithClearError() {
    // Arrange
    val requestJson =
      """
      {
        "id": 1,
        "name": "deploy contract",
        "context": {
          "chainId": 59141,
          "url": "https://rpc.sepolia.linea.build",
          "nbOfExecutions": 1
        },
        "calls": [
          {
            "nbOfExecution": 1,
            "scenario": {
              "scenarioType": "ContractCall",
              "wallet": "source",
              "contract": {
                "contractCallType": "CreateContract",
                "name": "dummyContract",
                "byteCode": "0x00"
              }
            }
          }
        ]
      }
      """.trimIndent()

    val swaggerRequest =
      JSON.createGson().create().fromJson(requestJson, net.consensys.zkevm.load.swagger.Request::class.java)

    // Act
    val exception = assertFailsWith<IllegalArgumentException> {
      Request.translate(swaggerRequest)
    }

    // Assert
    assertEquals("CreateContract `dummyContract` is missing required gasLimit.", exception.message)
  }

  @Test
  fun translateSepoliaLoadTestRequestSucceeds() {
    // Arrange
    val resource = checkNotNull(this.javaClass.classLoader.getResource("sepolia/load-test.json"))
    val swaggerRequest =
      JSON.createGson().create().fromJson(
        InputStreamReader(resource.openStream()),
        net.consensys.zkevm.load.swagger.Request::class.java,
      )

    // Act
    val request = Request.translate(swaggerRequest)

    // Assert
    assertNotNull(request)
  }
}
