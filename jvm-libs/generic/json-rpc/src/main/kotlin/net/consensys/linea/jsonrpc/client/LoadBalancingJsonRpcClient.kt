package net.consensys.linea.jsonrpc.client

import com.github.michaelbull.result.Result
import io.vertx.core.Future
import io.vertx.core.Promise
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.util.LinkedList
import java.util.Queue
import java.util.concurrent.ConcurrentLinkedQueue
import java.util.concurrent.locks.Lock
import java.util.concurrent.locks.ReadWriteLock
import java.util.concurrent.locks.ReentrantReadWriteLock

/**
 * Implements a Client Side LoadBalancer on round-robin for each JsonRpcClient. Each JsonRpcClient
 * can have up to maxInflightRequestsPerClient requests in progress. After all rpc clients reach
 * this limit request will queue and served in a LIFO order.
 *
 * It's expected that each JsonRpcClient represents a different upstream Endpoint e.g
 * prover1.linea.io:8080 and prover2.linea.io:8081
 */
class LoadBalancingJsonRpcClient
private constructor(
  rpcClients: List<JsonRpcClient>,
  private val maxInflightRequestsPerClient: UInt,
  private val requestPriorityComparator: Comparator<JsonRpcRequest>? = null,
) : JsonRpcClient {

  companion object {
    private val loadBalancingJsonRpcClients: ConcurrentLinkedQueue<LoadBalancingJsonRpcClient> = ConcurrentLinkedQueue()

    fun create(
      rpcClients: List<JsonRpcClient>,
      requestLimitPerEndpoint: UInt,
      requestPriorityComparator: Comparator<JsonRpcRequest>? = null,
    ): LoadBalancingJsonRpcClient {
      val loadBalancingJsonRpcClient = LoadBalancingJsonRpcClient(
        rpcClients = rpcClients,
        maxInflightRequestsPerClient = requestLimitPerEndpoint,
        requestPriorityComparator = requestPriorityComparator,
      )
      loadBalancingJsonRpcClients.add(loadBalancingJsonRpcClient)
      return loadBalancingJsonRpcClient
    }

    fun stop() {
      loadBalancingJsonRpcClients.forEach {
        it.close()
      }
    }
  }

  private val log: Logger = LogManager.getLogger(this.javaClass)

  private data class RpcClientContext(val rpcClient: JsonRpcClient, var inflightRequests: UInt)
  private data class RpcRequestContext(
    val request: JsonRpcRequest,
    val promise: Promise<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>>,
    val resultMapper: (Any?) -> Any?,
  )

  private val clientsPool: List<RpcClientContext> = rpcClients.map { RpcClientContext(it, 0u) }
  private val waitingQueue: Queue<RpcRequestContext> =
    if (requestPriorityComparator != null) {
      PriorityQueueWithFIFOFallback<RpcRequestContext>(comparator = { o1, o2 ->
        requestPriorityComparator.compare(o1.request, o2.request)
      })
    } else {
      LinkedList<RpcRequestContext>()
    }

  private val readWriteLock: ReadWriteLock = ReentrantReadWriteLock()
  private val writeLock: Lock = readWriteLock.writeLock()

  internal fun queuedRequests(): List<JsonRpcRequest> {
    try {
      readWriteLock.readLock().lock()
      return waitingQueue.map { it.request }
    } finally {
      readWriteLock.readLock().unlock()
    }
  }

  fun inflightRequestsCount(): Long {
    return clientsPool.fold(0L) { acc, it -> acc + it.inflightRequests.toLong() }
  }

  private fun serveNextWaitingInTheQueue() {
    try {
      readWriteLock.readLock().lock()
      if (waitingQueue.isEmpty()) return
    } finally {
      readWriteLock.readLock().unlock()
    }

    try {
      writeLock.lock()
      val client = clientsPool.sortedBy(RpcClientContext::inflightRequests).first()
      if (client.inflightRequests < maxInflightRequestsPerClient) {
        // fetch waiting request from the queue
        waitingQueue.poll()?.let { request ->
          client.inflightRequests++
          client to request
        }
      } else {
        null
      }
    } finally {
      writeLock.unlock()
    }
      ?.let { (nextAvailableClient, waitingRequest) ->
        log.trace("making request={}", waitingRequest)
        dispatchRequest(nextAvailableClient, waitingRequest)
      }
  }

  private fun enqueueRequest(
    request: JsonRpcRequest,
    resultMapper: (Any?) -> Any?,
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    val resultPromise: Promise<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> =
      Promise.promise()
    try {
      writeLock.lock()
      log.trace("enqueuing request={}", request)
      waitingQueue.add(RpcRequestContext(request, resultPromise, resultMapper))
    } finally {
      writeLock.unlock()
    }
    return resultPromise.future()
  }

  override fun makeRequest(
    request: JsonRpcRequest,
    resultMapper: (Any?) -> Any?,
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    val result = enqueueRequest(request, resultMapper)
    serveNextWaitingInTheQueue()
    return result
  }

  private fun dispatchRequest(
    rpcClientContext: RpcClientContext,
    queuedRequest: RpcRequestContext,
  ) {
    rpcClientContext.rpcClient
      .makeRequest(queuedRequest.request, queuedRequest.resultMapper)
      .onComplete { asyncResult ->
        try {
          writeLock.lock()
          rpcClientContext.inflightRequests--
        } finally {
          writeLock.unlock()
        }
        try {
          queuedRequest.promise.handle(asyncResult)
        } catch (e: Exception) {
          log.error("Response handler threw error:", e)
        } finally {
          serveNextWaitingInTheQueue()
        }
      }
  }

  fun close() {
    try {
      writeLock.lock()
      waitingQueue.clear()
    } finally {
      writeLock.unlock()
    }
  }
}
