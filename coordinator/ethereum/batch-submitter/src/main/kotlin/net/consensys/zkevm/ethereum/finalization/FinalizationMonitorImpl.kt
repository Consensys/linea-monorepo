package net.consensys.zkevm.ethereum.finalization

import io.vertx.core.TimeoutStream
import io.vertx.core.Vertx
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.contract.ZkEvmV2AsyncFriendly
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.math.BigInteger
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.atomic.AtomicReference
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds

class FinalizationMonitorImpl(
  private val config: Config,
  private val contract: ZkEvmV2AsyncFriendly,
  private val l1Client: Web3j,
  private val l2Client: Web3j,
  private val vertx: Vertx
) : FinalizationMonitor {
  private val log: Logger = LogManager.getLogger(this::class.java)

  data class Config(
    val pollingInterval: Duration = 500.milliseconds,
    val blocksToFinalization: UInt
  )

  private val finalizationHandlers:
    MutableMap<String, (FinalizationMonitor.FinalizationUpdate) -> SafeFuture<*>> =
      ConcurrentHashMap()
  private lateinit var monitorStream: TimeoutStream
  private val lastFinalizationUpdate = AtomicReference<FinalizationMonitor.FinalizationUpdate>(null)

  fun start(): SafeFuture<Unit> {
    if (this::monitorStream.isInitialized) {
      this.stop()
    }

    return getFinalizationState().thenApply {
      lastFinalizationUpdate.set(it)
      onUpdate(it)
      monitorStream =
        vertx.periodicStream(config.pollingInterval.inWholeMilliseconds).handler {
          try {
            monitorStream.pause()
            tick().whenComplete { _, _ -> monitorStream.resume() }
          } catch (_: Throwable) {
            monitorStream.resume()
          }
        }
    }
  }

  fun stop(): SafeFuture<Unit> {
    val result =
      if (this::monitorStream.isInitialized) {
        monitorStream.cancel()
      } else {
        Unit
      }
    return SafeFuture.completedFuture(result)
  }

  private fun tick(): SafeFuture<Unit> {
    log.debug("Checking finalization updates")
    return getFinalizationState().thenApply { currentState ->
      if (lastFinalizationUpdate.get() != currentState) {
        log.debug("Old finalization state: {}", currentState)
        log.info("Finalization update: ${lastFinalizationUpdate.get()}")
        lastFinalizationUpdate.set(currentState)
        onUpdate(currentState)
      }
    }
  }

  private fun getFinalizationState(): SafeFuture<FinalizationMonitor.FinalizationUpdate> {
    return vertx
      .executeBlocking { promise ->
        try {
          val latestBlockNumber = l1Client.ethBlockNumber().send().blockNumber
          val safeBlockNumber =
            DefaultBlockParameter.valueOf(
              latestBlockNumber.minus(
                BigInteger.valueOf(config.blocksToFinalization.toLong())
              )
            )
          contract.setDefaultBlockParameter(safeBlockNumber)
          val blockNumber = contract.currentL2BlockNumber().send()

          val finalizedBlock =
            l2Client
              .ethGetBlockByNumber(DefaultBlockParameter.valueOf(blockNumber), false)
              .send()

          val stateRootHash = contract.stateRootHashes(blockNumber).send()
          val result =
            FinalizationMonitor.FinalizationUpdate(
              Bytes32.wrap(Bytes.wrap(stateRootHash)),
              UInt64.valueOf(blockNumber),
              Bytes32.fromHexString(finalizedBlock.block.hash)
            )
          promise.complete(result)
        } catch (th: Throwable) {
          promise.fail(th)
        }
      }
      .toSafeFuture()
  }

  private fun onUpdate(finalizationUpdate: FinalizationMonitor.FinalizationUpdate) {
    for ((handlerName, finalizationHandler) in finalizationHandlers.entries) {
      try {
        finalizationHandler(finalizationUpdate)
      } catch (th: Throwable) {
        log.error("Finalization callback {} failed. errorMessage={}", handlerName, th.message, th)
      }
    }
  }

  override fun getLastFinalizationUpdate(): FinalizationMonitor.FinalizationUpdate {
    return lastFinalizationUpdate.get()
  }

  override fun addFinalizationHandler(
    handlerName: String,
    handler: (FinalizationMonitor.FinalizationUpdate) -> SafeFuture<*>
  ) {
    finalizationHandlers[handlerName] = handler
  }

  override fun removeFinalizationHandler(handlerName: String) {
    finalizationHandlers.remove(handlerName)
  }
}
