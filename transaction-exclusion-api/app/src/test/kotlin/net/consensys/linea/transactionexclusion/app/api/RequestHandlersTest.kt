package net.consensys.linea.transactionexclusion.app.api

import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.get
import io.vertx.core.json.JsonObject
import net.consensys.encodeHex
import net.consensys.linea.async.get
import net.consensys.linea.jsonrpc.JsonRpcRequestListParams
import net.consensys.linea.jsonrpc.JsonRpcRequestMapParams
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import net.consensys.linea.transactionexclusion.TransactionExclusionServiceV1
import net.consensys.linea.transactionexclusion.defaultRejectedTransaction
import net.consensys.toHexString
import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture

class RequestHandlersTest {
  private val transactionExclusionServiceMock = mock<TransactionExclusionServiceV1>().also {
    whenever(
      it.saveRejectedTransaction(any())
    )
      .thenReturn(
        SafeFuture.completedFuture(
          Ok(TransactionExclusionServiceV1.SaveRejectedTransactionStatus.SAVED)
        )
      )

    whenever(
      it.getTransactionExclusionStatus(any())
    )
      .thenReturn(
        SafeFuture.completedFuture(
          Ok(defaultRejectedTransaction)
        )
      )
  }

  @Test
  fun SaveRejectedTransactionRequestHandlerV1_invoke_acceptsValidRequestMap() {
    val request = JsonRpcRequestMapParams(
      "2.0",
      "1",
      "linea_saveRejectedTransactionV1",
      mapOf(
        "txRejectionStage" to "SEQUENCER",
        "timestamp" to "2024-09-05T09:22:52Z",
        "blockNumber" to "12345",
        "transactionRLP" to defaultRejectedTransaction.transactionRLP.encodeHex(),
        "reasonMessage" to defaultRejectedTransaction.reasonMessage,
        "overflows" to
          "[{\"module\":\"ADD\",\"count\":402,\"limit\":70},{\"module\":\"MUL\",\"count\":587,\"limit\":400}]"
      )
    )

    val saveRequestHandlerV1 = SaveRejectedTransactionRequestHandlerV1(
      transactionExclusionServiceMock
    )

    val expectedResult = JsonObject()
      .put("status", TransactionExclusionServiceV1.SaveRejectedTransactionStatus.SAVED)
      .put("txHash", defaultRejectedTransaction.transactionInfo!!.hash.encodeHex())
      .let {
        JsonRpcSuccessResponse(request.id, it)
      }

    val result = saveRequestHandlerV1.invoke(
      user = null,
      request = request,
      requestJson = JsonObject()
    ).get()

    Assertions.assertEquals(expectedResult, result.get())
  }

  @Test
  fun SaveRejectedTransactionRequestHandlerV1_invoke_acceptsValidRequestMap_without_blockNumber() {
    val request = JsonRpcRequestMapParams(
      "2.0",
      "1",
      "linea_saveRejectedTransactionV1",
      mapOf(
        "txRejectionStage" to "SEQUENCER",
        "timestamp" to "2024-09-05T09:22:52Z",
        "transactionRLP" to defaultRejectedTransaction.transactionRLP.encodeHex(),
        "reasonMessage" to defaultRejectedTransaction.reasonMessage,
        "overflows" to
          "[{\"module\":\"ADD\",\"count\":402,\"limit\":70},{\"module\":\"MUL\",\"count\":587,\"limit\":400}]"
      )
    )

    val saveTxRequestHandlerV1 = SaveRejectedTransactionRequestHandlerV1(
      transactionExclusionServiceMock
    )

    val expectedResult = JsonObject()
      .put("status", TransactionExclusionServiceV1.SaveRejectedTransactionStatus.SAVED)
      .put("txHash", defaultRejectedTransaction.transactionInfo!!.hash.encodeHex())
      .let {
        JsonRpcSuccessResponse(request.id, it)
      }

    val result = saveTxRequestHandlerV1.invoke(
      user = null,
      request = request,
      requestJson = JsonObject()
    ).get()

    Assertions.assertEquals(expectedResult, result.get())
  }

  @Test
  fun GetTransactionExclusionStatusRequestHandlerV1_invoke_acceptsValidRequestList() {
    val request = JsonRpcRequestListParams(
      "2.0",
      "1",
      "linea_getTransactionExclusionStatusV1",
      listOf(
        defaultRejectedTransaction.transactionInfo!!.hash.encodeHex()
      )
    )

    val getTxStatusRequestHandlerV1 = GetTransactionExclusionStatusRequestHandlerV1(
      transactionExclusionServiceMock
    )

    val expectedResult = JsonObject()
      .put("txHash", defaultRejectedTransaction.transactionInfo!!.hash.encodeHex())
      .put("from", defaultRejectedTransaction.transactionInfo!!.from.encodeHex())
      .put("nonce", defaultRejectedTransaction.transactionInfo!!.nonce.toHexString())
      .put("txRejectionStage", defaultRejectedTransaction.txRejectionStage.name)
      .put("reasonMessage", defaultRejectedTransaction.reasonMessage)
      .put("timestamp", defaultRejectedTransaction.timestamp.toString())
      .put("blockNumber", defaultRejectedTransaction.blockNumber!!.toHexString())
      .let {
        JsonRpcSuccessResponse(request.id, it)
      }

    val result = getTxStatusRequestHandlerV1.invoke(
      user = null,
      request = request,
      requestJson = JsonObject()
    ).get()

    Assertions.assertEquals(result.get(), expectedResult)
  }
}
