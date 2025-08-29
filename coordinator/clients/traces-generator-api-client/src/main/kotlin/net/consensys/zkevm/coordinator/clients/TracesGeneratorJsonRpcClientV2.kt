package net.consensys.zkevm.coordinator.clients

import com.github.michaelbull.result.Result
import com.github.michaelbull.result.mapEither
import io.vertx.core.Vertx
import io.vertx.core.json.JsonObject
import io.vertx.kotlin.core.json.get
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.errors.ErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcRequestListParams
import net.consensys.linea.jsonrpc.client.JsonRpcClient
import net.consensys.linea.jsonrpc.client.JsonRpcRequestRetryer
import net.consensys.linea.jsonrpc.client.RequestRetryConfig
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.atomic.AtomicInteger

class TracesGeneratorJsonRpcClientV2(
  private val rpcClient: JsonRpcClient,
  private val config: Config,
) :
  TracesCountersClientV2, TracesConflationClientV2 {
  constructor(
    vertx: Vertx,
    rpcClient: JsonRpcClient,
    config: Config,
    retryConfig: RequestRetryConfig,
    log: Logger = LogManager.getLogger(TracesGeneratorJsonRpcClientV2::class.java),
  ) : this(
    JsonRpcRequestRetryer(
      vertx,
      rpcClient,
      config = JsonRpcRequestRetryer.Config(
        methodsToRetry = retryableMethods,
        requestRetry = retryConfig,
      ),
      log = log,
    ),
    config,
  )

  data class Config(
    val expectedTracesApiVersion: String,
  )

  private var requestBuilder = RequestBuilder(config.expectedTracesApiVersion)

  override fun getTracesCounters(
    blockNumber: ULong,
  ): SafeFuture<Result<GetTracesCountersResponse, ErrorResponse<TracesServiceErrorType>>> {
    val jsonRequest = requestBuilder.buildGetTracesCountersV2Request(blockNumber)

    return rpcClient.makeRequest(jsonRequest).toSafeFuture()
      .thenApply { responseResult ->
        responseResult.mapEither(
          TracesClientResponsesParser::parseTracesCounterResponseV2,
          TracesClientResponsesParser::mapErrorResponseV2,
        )
      }
  }

  override fun generateConflatedTracesToFile(
    startBlockNumber: ULong,
    endBlockNumber: ULong,
  ): SafeFuture<Result<GenerateTracesResponse, ErrorResponse<TracesServiceErrorType>>> {
    val jsonRequest = requestBuilder.buildGenerateConflatedTracesToFileV2Request(
      startBlockNumber,
      endBlockNumber,
    )

    return rpcClient.makeRequest(jsonRequest).toSafeFuture()
      .thenApply { responseResult ->
        responseResult.mapEither(
          TracesClientResponsesParser::parseConflatedTracesToFileResponse,
          TracesClientResponsesParser::mapErrorResponseV2,
        )
      }
  }

  internal class RequestBuilder(
    private val expectedTracesEngineVersion: String,
    private val id: AtomicInteger = AtomicInteger(0),
  ) {
    fun buildGetTracesCountersV2Request(
      blockNumber: ULong,
    ): JsonRpcRequest {
      return JsonRpcRequestListParams(
        "2.0",
        id.incrementAndGet(),
        "linea_getBlockTracesCountersV2",
        listOf(
          JsonObject.of(
            "blockNumber",
            blockNumber,
            "expectedTracesEngineVersion",
            expectedTracesEngineVersion,
          ),
        ),
      )
    }

    fun buildGenerateConflatedTracesToFileV2Request(
      startBlockNumber: ULong,
      endBlockNumber: ULong,
    ): JsonRpcRequest {
      return JsonRpcRequestListParams(
        "2.0",
        id.incrementAndGet(),
        "linea_generateConflatedTracesToFileV2",
        listOf(
          JsonObject.of(
            "startBlockNumber",
            startBlockNumber,
            "endBlockNumber",
            endBlockNumber,
            "expectedTracesEngineVersion",
            expectedTracesEngineVersion,
          ),
        ),
      )
    }
  }

  companion object {
    internal val retryableMethods = setOf("linea_getBlockTracesCountersV2", "linea_generateConflatedTracesToFileV2")

    val requestPriorityComparator = object : Comparator<JsonRpcRequest> {
      @Suppress("UNCHECKED_CAST")
      override fun compare(
        o1: JsonRpcRequest,
        o2: JsonRpcRequest,
      ): Int {
        // linea_generateConflatedTracesToFileV2 is always fired after linea_getBlockTracesCountersV2
        // has successfully completed, so we should prioritize it.
        return when {
          o1.method == "linea_generateConflatedTracesToFileV2" && o2.method == "linea_getBlockTracesCountersV2" -> -1
          o1.method == "linea_getBlockTracesCountersV2" && o2.method == "linea_generateConflatedTracesToFileV2" -> 1
          o1.method == "linea_getBlockTracesCountersV2" && o2.method == "linea_getBlockTracesCountersV2" -> {
            val bn1 = (o1.params as List<JsonObject>).first().get<ULong>("blockNumber")
            val bn2 = (o2.params as List<JsonObject>).first().get<ULong>("blockNumber")
            return (bn1 - bn2).toInt()
          }

          o1.method == "linea_generateConflatedTracesToFileV2" &&
            o2.method == "linea_generateConflatedTracesToFileV2" -> {
            val bn1 = (o1.params as List<JsonObject>).first().get<ULong>("startBlockNumber")
            val bn2 = (o2.params as List<JsonObject>).first().get<ULong>("startBlockNumber")
            return (bn1 - bn2).toInt()
          }

          else -> 0
        }
      }
    }
  }
}
