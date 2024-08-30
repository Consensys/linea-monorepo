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
import kotlinx.datetime.Instant
import net.consensys.decodeHex
import net.consensys.encodeHex
import net.consensys.linea.ModuleOverflow
import net.consensys.linea.RejectedTransaction
import net.consensys.linea.TransactionExclusionServiceV1
import net.consensys.linea.async.toVertxFuture
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcRequestHandler
import net.consensys.linea.jsonrpc.JsonRpcRequestListParams
import net.consensys.linea.jsonrpc.JsonRpcRequestMapParams
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import net.consensys.toHexString
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

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

private fun parseMapParamsToRejectedTransaction(validatedRequest: JsonRpcRequestMapParams): RejectedTransaction {
  return RejectedTransaction(
    stage = ArgumentParser.geRejectedTransactionStage(validatedRequest.params["stage"].toString()),
    timestamp = Instant.parse(validatedRequest.params["timestamp"].toString()),
    blockNumber = validatedRequest.params["blockNumber"].toString().toULong(),
    transactionRLP = validatedRequest.params["transactionRLP"].toString().decodeHex(),
    reasonMessage = validatedRequest.params["reasonMessage"].toString(),
    overflows = ModuleOverflow.parseListFromJsonString(
      ModuleOverflow.parseToJsonString(validatedRequest.params["overflows"]!!)
    )
  )
}

private fun parseListParamsToRejectedTransaction(validatedRequest: JsonRpcRequestListParams): RejectedTransaction {
  return RejectedTransaction(
    stage = ArgumentParser.geRejectedTransactionStage(validatedRequest.params[0].toString()),
    timestamp = Instant.parse(validatedRequest.params[1].toString()),
    blockNumber = validatedRequest.params[2].toString().toULong(),
    transactionRLP = validatedRequest.params[3].toString().decodeHex(),
    reasonMessage = validatedRequest.params[4].toString(),
    overflows = ModuleOverflow.parseListFromJsonString(
      ModuleOverflow.parseToJsonString(validatedRequest.params[5]!!)
    )
  )
}

private fun parseMapParamsToTxHash(validatedRequest: JsonRpcRequestMapParams): ByteArray {
  return (validatedRequest.params["txHash"] as String).decodeHex()
}

private fun parseListParamsToTxHash(validatedRequest: JsonRpcRequestListParams): ByteArray {
  return (validatedRequest.params[0] as String).decodeHex()
}

class SaveRejectedTransactionRequestHandlerV1(
  private val transactionExclusionService: TransactionExclusionServiceV1
) : JsonRpcRequestHandler {
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
  private val log: Logger = LogManager.getLogger(this::class.java)

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
