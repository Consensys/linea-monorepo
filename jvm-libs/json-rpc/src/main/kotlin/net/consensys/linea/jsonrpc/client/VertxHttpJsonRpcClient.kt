package net.consensys.linea.jsonrpc.client

import com.fasterxml.jackson.core.JsonGenerator
import com.fasterxml.jackson.databind.JsonSerializer
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.SerializerProvider
import com.fasterxml.jackson.databind.module.SimpleModule
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import io.micrometer.core.instrument.MeterRegistry
import io.vertx.core.Future
import io.vertx.core.buffer.Buffer
import io.vertx.core.http.HttpClient
import io.vertx.core.http.HttpClientResponse
import io.vertx.core.http.HttpMethod
import io.vertx.core.http.RequestOptions
import io.vertx.core.json.JsonObject
import io.vertx.core.json.jackson.VertxModule
import net.consensys.linea.jsonrpc.JsonRpcError
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import net.consensys.linea.metrics.micrometer.SimpleTimerCapture
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.net.URL
import java.util.HexFormat

class VertxHttpJsonRpcClient(
  private val httpClient: HttpClient,
  private val endpoint: URL,
  private val meterRegistry: MeterRegistry,
  private val requestObjectMapper: ObjectMapper = objectMapper,
  private val responseObjectMapper: ObjectMapper = objectMapper,
  private val log: Logger = LogManager.getLogger(VertxHttpJsonRpcClient::class.java),
  private val requestResponseLogLevel: Level = Level.TRACE,
  private val failuresLogLevel: Level = Level.DEBUG
) : JsonRpcClient {
  private val requestOptions = RequestOptions().apply {
    setMethod(HttpMethod.POST)
    setAbsoluteURI(endpoint)
  }

  override fun makeRequest(
    request: JsonRpcRequest,
    resultMapper: (Any?) -> Any?
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    val json = requestObjectMapper.writeValueAsString(request)

    return httpClient.request(requestOptions).flatMap { httpClientRequest ->
      httpClientRequest.putHeader("Content-Type", "application/json")
      logRequest(json)

      val requestFuture =
        httpClientRequest.send(json).flatMap { response: HttpClientResponse ->
          if (isSuccessStatusCode(response.statusCode())) {
            handleResponse(json, response, resultMapper)
          } else {
            response.body().flatMap { bodyBuffer ->
              logResponse(
                isError = true,
                response = response,
                requestBody = json,
                responseBody = bodyBuffer.toString().lines().firstOrNull() ?: ""
              )
              Future.failedFuture(
                Exception(
                  "HTTP errorCode=${response.statusCode()}, message=${response.statusMessage()}"
                )
              )
            }
          }
        }

      SimpleTimerCapture<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>>(
        meterRegistry,
        "jsonrpc.request"
      )
        .setDescription("Time of Upstream API JsonRpc Requests")
        .setTag("endpoint", endpoint.host)
        .setTag("method", request.method)
        .captureTime(requestFuture)
    }
      .onFailure { th -> logRequestFailure(json, th) }
  }

  private fun handleResponse(
    requestBody: String,
    httpResponse: HttpClientResponse,
    resultMapper: (Any?) -> Any?
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    var isError = false
    var responseBody = ""
    return httpResponse
      .body()
      .flatMap { bodyBuffer: Buffer ->
        responseBody = bodyBuffer.toString()
        try {
          @Suppress("UNCHECKED_CAST")
          val jsonResponse = (responseObjectMapper.readValue(responseBody, Map::class.java) as Map<String, Any>)
            .let(::JsonObject)
          val response =
            when {
              jsonResponse.containsKey("result") ->
                Ok(
                  JsonRpcSuccessResponse(
                    jsonResponse.getValue("id"),
                    resultMapper(jsonResponse.getValue("result"))
                  )
                )

              jsonResponse.containsKey("error") -> {
                isError = true
                Err(
                  JsonRpcErrorResponse(
                    jsonResponse.getValue("id"),
                    jsonResponse.getJsonObject("error").mapTo(JsonRpcError::class.java)
                  )
                )
              }

              else -> throw IllegalArgumentException("Invalid JSON-RPC response without result or error")
            }
          Future.succeededFuture(response)
        } catch (e: Throwable) {
          isError = true
          when (e) {
            is IllegalArgumentException -> Future.failedFuture(e)
            else -> Future.failedFuture(IllegalArgumentException("Invalid JSON-RPC response.", e))
          }
        }
      }
      .andThen { asyncResult ->
        logResponse(isError, httpResponse, requestBody, responseBody, asyncResult.cause())
      }
  }

  private fun logRequest(jsonBody: String, level: Level = requestResponseLogLevel) {
    log.log(level, "--> {} {}", endpoint, jsonBody)
  }

  private fun logResponse(
    isError: Boolean,
    response: HttpClientResponse,
    requestBody: String,
    responseBody: String,
    failureCause: Throwable? = null
  ) {
    val logLevel = if (isError) failuresLogLevel else requestResponseLogLevel
    if (isError && log.level != requestResponseLogLevel) {
      // in case of error, log the request that originated the error
      // to help replicate and debug later
      logRequest(requestBody, logLevel)
    }

    log.log(
      logLevel,
      "<-- {} {} {} {}",
      endpoint,
      response.statusCode(),
      responseBody,
      failureCause?.message ?: ""
    )
  }

  private fun logRequestFailure(
    requestBody: String,
    failureCause: Throwable
  ) {
    log.log(
      failuresLogLevel,
      "<--> {} {} failed with error={}",
      endpoint,
      requestBody,
      failureCause.message,
      failureCause
    )
  }

  private fun isSuccessStatusCode(statusCode: Int): Boolean {
    return statusCode >= 200 && statusCode < 300
  }

  companion object {
    val objectMapper = jacksonObjectMapper()
      .registerModules(VertxModule())
      .registerModules(
        SimpleModule().apply {
          this.addSerializer(ByteArray::class.java, ByteArrayToHexStringSerializer())
        }
      )
  }
}

class ByteArrayToHexStringSerializer : JsonSerializer<ByteArray>() {
  private val hexFormatter = HexFormat.of()
  override fun serialize(value: ByteArray?, gen: JsonGenerator?, serializers: SerializerProvider?) {
    gen?.writeString(value?.let { "0x" + hexFormatter.formatHex(it) })
  }
}
