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
  private fun validateMapParamsPresence(requestMapParams: JsonRpcRequestMapParams) {
    val missingParam = if (requestMapParams.params["stage"] == null) {
      "stage"
    } else if (requestMapParams.params["timestamp"] == null) {
      "timestamp"
    } else if (requestMapParams.params["blockNumber"] == null) {
      "blockNumber"
    } else if (requestMapParams.params["transactionRLP"] == null) {
      "transactionRLP"
    } else if (requestMapParams.params["reasonMessage"] == null) {
      "reasonMessage"
    } else if (requestMapParams.params["overflows"] == null) {
      "overflows"
    } else {
      ""
    }
    if (missingParam.isNotEmpty()) {
      throw IllegalArgumentException("\"$missingParam\" is missing from the given params")
    }
  }

  private fun validateListParamsPresence(requestListParams: JsonRpcRequestListParams) {
    if (requestListParams.params.size < 6) {
      throw IllegalArgumentException("The size of the given params list is less than 6")
    }
  }

  private fun parseAndGetRejectedTransaction(
    stage: String,
    timestamp: String,
    blockNumber: String,
    transactionRLP: String,
    reasonMessage: String,
    overflows: Any
  ): RejectedTransaction {
    return RejectedTransaction(
      stage = ArgumentParser.geRejectedTransactionStage(stage),
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
          validatedRequest.params["stage"].toString(),
          validatedRequest.params["timestamp"].toString(),
          validatedRequest.params["blockNumber"].toString(),
          validatedRequest.params["transactionRLP"].toString(),
          validatedRequest.params["reasonMessage"].toString(),
          validatedRequest.params["overflows"]!!
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
    val missingParam = if (requestMapParams.params["txHash"] == null) "txHash" else ""
    if (missingParam.isNotEmpty()) {
      throw IllegalArgumentException("\"$missingParam\" is missing from the given params")
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
              .put("reason", it.reasonMessage)
              .put("blockNumber", it.blockNumber.toHexString())
              .put("timestamp", it.timestamp.toString())
          JsonRpcSuccessResponse(request.id, rpcResult)
        }.mapError { error ->
          JsonRpcErrorResponse(request.id, jsonRpcError(error))
        }
      }.toVertxFuture()
  }
}
