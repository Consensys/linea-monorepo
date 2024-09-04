package net.consensys.linea.transactionexclusion.app.api

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.flatMap
import com.github.michaelbull.result.get
import com.github.michaelbull.result.map
import com.github.michaelbull.result.mapError
import io.vertx.core.Future
import io.vertx.core.json.JsonObject
import io.vertx.ext.auth.User
import net.consensys.encodeHex
import net.consensys.linea.RejectedTransaction
import net.consensys.linea.async.toVertxFuture
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcRequestHandler
import net.consensys.linea.jsonrpc.JsonRpcRequestListParams
import net.consensys.linea.jsonrpc.JsonRpcRequestMapParams
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import net.consensys.linea.transactionexclusion.TransactionExclusionServiceV1
import net.consensys.toHexString

private fun validateParams(request: JsonRpcRequest): Result<JsonRpcRequest, JsonRpcErrorResponse> {
  if (request.params !is Map<*, *> && request.params !is List<*>) {
    return Err(
      JsonRpcErrorResponse.invalidParams(
        request.id,
        "params should be either an object or a list"
      )
    )
  }
  return try {
    if (request.params is Map<*, *>) {
      val jsonRpcRequest = request as JsonRpcRequestMapParams
      if (jsonRpcRequest.params.isEmpty()) {
        return Err(
          JsonRpcErrorResponse.invalidParams(
            request.id,
            "Parameters map is empty!"
          )
        )
      }
    } else if (request.params is List<*>) {
      val jsonRpcRequest = request as JsonRpcRequestListParams
      if (jsonRpcRequest.params.isEmpty()) {
        return Err(
          JsonRpcErrorResponse.invalidParams(
            request.id,
            "Parameters list is empty!"
          )
        )
      }
    }
    Ok(request)
  } catch (e: Exception) {
    Err(JsonRpcErrorResponse.invalidRequest())
  }
}

class SaveRejectedTransactionRequestHandlerV1(
  private val transactionExclusionService: TransactionExclusionServiceV1
) : JsonRpcRequestHandler {
  enum class RequestParams(val paramName: String) {
    TX_REJECTION_STAGE("txRejectionStage"),
    TIMESTAMP("timestamp"),
    REASON_MESSAGE("reasonMessage"),
    TRANSACTION_RLP("transactionRLP"),
    BLOCK_NUMBER("blockNumber"),
    OVERFLOWS("overflows")
  }

  private fun validateMapParamsPresence(requestMapParams: JsonRpcRequestMapParams) {
    RequestParams.entries
      .filter { requestParam ->
        requestParam != RequestParams.BLOCK_NUMBER && requestMapParams.params[requestParam.paramName] == null
      }
      .run {
        if (this.isNotEmpty()) {
          throw IllegalArgumentException(
            "Missing ${this.joinToString(",", "[", "]") { it.paramName }} " +
              "from the given request params"
          )
        }
      }
  }

  private fun validateListParamsPresence(requestListParams: JsonRpcRequestListParams) {
    if (requestListParams.params.size < RequestParams.entries.size) {
      throw IllegalArgumentException(
        "The size of the given request params list is less than ${RequestParams.entries.size}"
      )
    }
  }

  private fun parseAndGetRejectedTransaction(
    txRejectionStage: String,
    timestamp: String,
    blockNumber: String?,
    transactionRLP: String,
    reasonMessage: String,
    overflows: Any
  ): RejectedTransaction {
    return RejectedTransaction(
      txRejectionStage = ArgumentParser.getTxRejectionStage(txRejectionStage),
      timestamp = ArgumentParser.getTimestampFromISO8601(timestamp),
      blockNumber = ArgumentParser.getBlockNumber(blockNumber),
      transactionRLP = ArgumentParser.getTransactionRLPInRawBytes(transactionRLP),
      reasonMessage = ArgumentParser.getReasonMessage(reasonMessage),
      overflows = ArgumentParser.getOverflows(overflows)
    ).also {
      it.transactionInfo = ArgumentParser.getTransactionInfoFromRLP(it.transactionRLP)
    }
  }

  private fun parseMapParamsToRejectedTransaction(validatedRequest: JsonRpcRequestMapParams): RejectedTransaction {
    return validateMapParamsPresence(validatedRequest)
      .run {
        parseAndGetRejectedTransaction(
          txRejectionStage = validatedRequest
            .params[RequestParams.TX_REJECTION_STAGE.paramName].toString(),
          timestamp = validatedRequest
            .params[RequestParams.TIMESTAMP.paramName].toString(),
          blockNumber = validatedRequest
            .params[RequestParams.BLOCK_NUMBER.paramName]?.toString(),
          transactionRLP = validatedRequest
            .params[RequestParams.TRANSACTION_RLP.paramName].toString(),
          reasonMessage = validatedRequest
            .params[RequestParams.REASON_MESSAGE.paramName].toString(),
          overflows = validatedRequest
            .params[RequestParams.OVERFLOWS.paramName]!!
        )
      }
  }

  private fun parseListParamsToRejectedTransaction(validatedRequest: JsonRpcRequestListParams): RejectedTransaction {
    return validateListParamsPresence(validatedRequest).run {
      parseAndGetRejectedTransaction(
        validatedRequest.params[0].toString(),
        validatedRequest.params[1].toString(),
        validatedRequest.params[2].toString(),
        validatedRequest.params[3].toString(),
        validatedRequest.params[4].toString(),
        validatedRequest.params[5]!!
      )
    }
  }
  override fun invoke(
    user: User?,
    request: JsonRpcRequest,
    requestJson: JsonObject
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    val rejectedTransaction = try {
      val parsingResult = validateParams(request).flatMap { validatedRequest ->
        val parsedRejectedTransaction =
          when (validatedRequest) {
            is JsonRpcRequestMapParams -> parseMapParamsToRejectedTransaction(validatedRequest)
            is JsonRpcRequestListParams -> parseListParamsToRejectedTransaction(validatedRequest)
            else -> throw IllegalStateException()
          }
        Ok(parsedRejectedTransaction)
      }
      if (parsingResult is Err) {
        return Future.succeededFuture(parsingResult)
      } else {
        parsingResult.get()!!
      }
    } catch (e: Exception) {
      return Future.succeededFuture(
        Err(
          JsonRpcErrorResponse.invalidParams(
            request.id,
            e.message
          )
        )
      )
    }

    return transactionExclusionService
      .saveRejectedTransaction(rejectedTransaction)
      .thenApply { result ->
        result.map {
          val rpcResult =
            JsonObject()
              .put("txHash", rejectedTransaction.transactionInfo!!.hash.encodeHex())
          JsonRpcSuccessResponse(request.id, rpcResult)
        }.mapError { error ->
          JsonRpcErrorResponse(request.id, jsonRpcError(error))
        }
      }.toVertxFuture()
  }
}

class GetTransactionExclusionStatusRequestHandlerV1(
  private val transactionExclusionService: TransactionExclusionServiceV1
) : JsonRpcRequestHandler {
  private fun validateMapParamsPresence(requestMapParams: JsonRpcRequestMapParams) {
    if (requestMapParams.params["txHash"] == null) {
      throw IllegalArgumentException("Missing txHash from the given request param")
    }
  }

  private fun parseMapParamsToTxHash(validatedRequest: JsonRpcRequestMapParams): ByteArray {
    return validateMapParamsPresence(validatedRequest).run {
      ArgumentParser.getTxHashInRawBytes(validatedRequest.params["txHash"].toString())
    }
  }

  private fun parseListParamsToTxHash(validatedRequest: JsonRpcRequestListParams): ByteArray {
    return ArgumentParser.getTxHashInRawBytes(validatedRequest.params[0].toString())
  }

  override fun invoke(
    user: User?,
    request: JsonRpcRequest,
    requestJson: JsonObject
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    val txHash = try {
      val parsingResult = validateParams(request).flatMap { validatedRequest ->
        val parsedTxHash =
          when (validatedRequest) {
            is JsonRpcRequestMapParams -> parseMapParamsToTxHash(validatedRequest)
            is JsonRpcRequestListParams -> parseListParamsToTxHash(validatedRequest)
            else -> throw IllegalStateException()
          }
        Ok(parsedTxHash)
      }
      if (parsingResult is Err) {
        return Future.succeededFuture(parsingResult)
      } else {
        parsingResult.get()!!
      }
    } catch (e: Exception) {
      return Future.succeededFuture(
        Err(
          JsonRpcErrorResponse.invalidParams(
            request.id,
            e.message
          )
        )
      )
    }

    return transactionExclusionService
      .getTransactionExclusionStatus(txHash)
      .thenApply { result ->
        result.map {
          val rpcResult =
            JsonObject()
              .put("txHash", it.transactionInfo!!.hash.encodeHex())
              .put("from", it.transactionInfo!!.from.encodeHex())
              .put("nonce", it.transactionInfo!!.nonce.toHexString())
              .put("txRejectionStage", it.txRejectionStage.name)
              .put("reason", it.reasonMessage)
              .put("timestamp", it.timestamp.toString())
              .also { jsonObject ->
                if (it.blockNumber != null) {
                  jsonObject.put("blockNumber", it.blockNumber!!.toHexString())
                }
              }
          JsonRpcSuccessResponse(request.id, rpcResult)
        }.mapError { error ->
          JsonRpcErrorResponse(request.id, jsonRpcError(error))
        }
      }.toVertxFuture()
  }
}
