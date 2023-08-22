package net.consensys.linea.jsonrpc.client

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import io.micrometer.core.instrument.MeterRegistry
import io.vertx.core.Future
import io.vertx.core.Vertx
import io.vertx.core.buffer.Buffer
import io.vertx.core.http.HttpClient
import io.vertx.core.http.HttpClientOptions
import io.vertx.core.http.HttpClientRequest
import io.vertx.core.http.HttpClientResponse
import io.vertx.core.http.HttpMethod
import io.vertx.core.http.HttpVersion
import io.vertx.core.json.JsonObject
import net.consensys.linea.async.get
import net.consensys.linea.jsonrpc.JsonRpcError
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import net.consensys.linea.metrics.monitoring.SimpleTimerCapture
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.net.URL

open class VertxHttpJsonRpcClientFactory(
  private val vertx: Vertx,
  private val meterRegistry: MeterRegistry
) {
  fun create(
    endpoint: URL,
    maxPoolSize: Int? = null,
    httpVersion: HttpVersion? = null,
    log: Logger = LogManager.getLogger(VertxHttpJsonRpcClient::class.java),
    requestDefaultLogLevel: Level = Level.TRACE,
    requestFailureLogLevel: Level = Level.DEBUG
  ): VertxHttpJsonRpcClient {
    val clientOptions =
      HttpClientOptions()
        .setKeepAlive(true)
        .setDefaultHost(endpoint.host)
        .setDefaultPort(endpoint.port)
    maxPoolSize?.let(clientOptions::setMaxPoolSize)
    httpVersion?.let(clientOptions::setProtocolVersion)
    val httpClient = vertx.createHttpClient(clientOptions)
    return VertxHttpJsonRpcClient(
      httpClient,
      endpoint.path,
      meterRegistry,
      log,
      requestDefaultLogLevel = requestDefaultLogLevel,
      requestFailureLogLevel = requestFailureLogLevel
    )
  }
}

class VertxHttpJsonRpcClient(
  private val httpClient: HttpClient,
  private val apiPath: String,
  private val meterRegistry: MeterRegistry,
  private val log: Logger = LogManager.getLogger(VertxHttpJsonRpcClient::class.java),
  private val requestDefaultLogLevel: Level = Level.TRACE,
  private val requestFailureLogLevel: Level = Level.DEBUG
) : JsonRpcClient {
  override fun makeRequest(
    request: JsonRpcRequest,
    resultMapper: (Any?) -> Any?
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    return httpClient.request(HttpMethod.POST, apiPath).flatMap { httpClientRequest ->
      httpClientRequest.putHeader("Content-Type", "application/json")

      val json = JsonObject.mapFrom(request).encode()
      logRequest(httpClientRequest, json)

      val requestFuture =
        httpClientRequest.send(json).flatMap { response: HttpClientResponse ->
          if (isSuccessStatusCode(response.statusCode())) {
            handleResponse(httpClientRequest, json, response, resultMapper)
          } else {
            response.body().flatMap { bodyBuffer ->
              log.warn(
                "{}/{} {} {} rpc_method={} responseBody='{}'",
                httpClientRequest.host,
                httpClientRequest.path(),
                response.statusCode(),
                response.statusMessage(),
                request.method,
                bodyBuffer.toString()
              )
              Future.failedFuture(
                Exception(
                  "HTTP error code=${response.statusCode()}, message=${response.statusMessage()}"
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
        .setTag("endpoint", httpClientRequest.host)
        .setTag("method", request.method)
        .captureTime(requestFuture)
    }
  }

  private fun handleResponse(
    request: HttpClientRequest,
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
          val jsonResponse = JsonObject(responseBody)
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

              else ->
                throw IllegalArgumentException(
                  "Invalid JSON-RPC response without result or error"
                )
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
      .andThen { logResponse(isError, request, httpResponse, requestBody, responseBody) }
  }

  // somehow HttpClientRequest.uri does not work, so building manually
  private fun HttpClientRequest.logUri(): String = "${this.host}:${this.port}${this.path()}"

  private fun logRequest(
    request: HttpClientRequest,
    json: String,
    level: Level = requestDefaultLogLevel
  ) {
    log.log(level, " --> {} {}", request.logUri(), json)
  }

  private fun logResponse(
    isError: Boolean,
    request: HttpClientRequest,
    response: HttpClientResponse,
    requestBody: String,
    responseBody: String
  ) {
    val logLevel = if (isError) requestFailureLogLevel else requestDefaultLogLevel
    if (isError && log.level != requestDefaultLogLevel) {
      // in case of error, log the request that originated the error
      // to help replicate and debug later
      logRequest(request, requestBody, logLevel)
    }

    log.log(logLevel, " <-- {} {} {}", request.logUri(), response.statusCode(), responseBody)
  }

  private fun isSuccessStatusCode(statusCode: Int): Boolean {
    return statusCode >= 200 && statusCode < 300
  }
}
