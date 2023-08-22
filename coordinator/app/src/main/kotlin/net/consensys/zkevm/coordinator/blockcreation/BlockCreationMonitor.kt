package net.consensys.zkevm.coordinator.blockcreation

import io.vertx.core.TimeoutStream
import io.vertx.core.Vertx
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockCreated
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockCreationListener
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.Response
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.atomic.AtomicBoolean
import java.util.concurrent.atomic.AtomicLong
import java.util.concurrent.atomic.AtomicReference
import kotlin.time.Duration

/**
 * Web3J is quite complex to Mock and Unit test, this is just a tinny interface to make
 * BlockCreationMonitor more testable
 */
interface ExtendedWeb3J {
  val web3jClient: Web3j
  fun ethBlockNumber(): SafeFuture<BigInteger>
  fun ethGetExecutionPayloadByNumber(blockNumber: Long): SafeFuture<ExecutionPayloadV1>
}

class ExtendedWeb3JImpl(override val web3jClient: Web3j) : ExtendedWeb3J {

  private fun buildException(error: Response.Error): Exception =
    Exception("${error.code}: ${error.message}")

  override fun ethBlockNumber(): SafeFuture<BigInteger> {
    return SafeFuture.of(web3jClient.ethBlockNumber().sendAsync()).thenCompose { response ->
      if (response.hasError()) {
        SafeFuture.failedFuture(buildException(response.error))
      } else {
        SafeFuture.completedFuture(response.blockNumber)
      }
    }
  }

  override fun ethGetExecutionPayloadByNumber(blockNumber: Long): SafeFuture<ExecutionPayloadV1> {
    return SafeFuture.of(
      web3jClient
        .ethGetBlockByNumber(
          DefaultBlockParameter.valueOf(BigInteger.valueOf(blockNumber)),
          true
        )
        .sendAsync()
    )
      .thenCompose { response ->
        if (response.hasError()) {
          SafeFuture.failedFuture(buildException(response.error))
        } else {
          SafeFuture.completedFuture(response.block.toExecutionPayloadV1())
        }
      }
  }
}

class BlockCreationMonitor(
  private val vertx: Vertx,
  private val extendedWeb3j: ExtendedWeb3J,
  startingBlockNumberInclusive: Long,
  expectedParentRooHash: Bytes32,
  private val blockCreationListener: BlockCreationListener,
  private val config: Config,
  private val log: Logger = LogManager.getLogger(BlockCreationMonitor::class.java)
) {
  data class Config(
    val pollingInterval: Duration,
    val blocksToFinalization: Long
  )

  private val _nexBlockNumberToFetch: AtomicLong = AtomicLong(startingBlockNumberInclusive)
  private val expectedParentRooHash: AtomicReference<Bytes32> = AtomicReference(expectedParentRooHash)
  private val reorgDetected: AtomicBoolean = AtomicBoolean(false)

  @Volatile
  private lateinit var monitorStream: TimeoutStream

  val nexBlockNumberToFetch: Long
    get() = _nexBlockNumberToFetch.get()

  @Synchronized
  fun start(): SafeFuture<Unit> {
    if (reorgDetected.get()) {
      return SafeFuture.failedFuture(IllegalStateException("Reorg detect. Cannot restart"))
    }
    if (this::monitorStream.isInitialized) {
      this.monitorStream.cancel()
    }

    monitorStream =
      vertx.periodicStream(config.pollingInterval.inWholeMilliseconds.coerceAtLeast(1L))
        .handler {
          try {
            monitorStream.pause()
            tick().whenComplete { _, _ ->
              contunueOrStopIfReorg()
            }
          } catch (th: Throwable) {
            log.error(th)
            contunueOrStopIfReorg()
          }
        }

    return SafeFuture.completedFuture(Unit)
  }

  @Synchronized
  fun stop(): SafeFuture<Unit> {
    if (this::monitorStream.isInitialized) {
      this.monitorStream.cancel()
    }
    return SafeFuture.completedFuture(Unit)
  }

  private fun tick(): SafeFuture<*> {
    log.trace("tick start")
    return getNetNextSafeBlock()
      .thenCompose { payload ->
        if (payload != null) {
          if (payload.parentHash == expectedParentRooHash.get()) {
            notifyListener(payload)
              .whenSuccess {
                log.debug(
                  "updating nexBlockNumberToFetch from {} --> {}",
                  _nexBlockNumberToFetch.get(),
                  _nexBlockNumberToFetch.incrementAndGet()
                )
                expectedParentRooHash.set(payload.blockHash)
              }
          } else {
            reorgDetected.set(true)
            log.error(
              "Shooting down conflation poller, " +
                "chain reorg detected: block (number={}, hash={}, parentHash={}) should have parentHash={}",
              payload.blockNumber.longValue(),
              payload.blockHash.toHexString().subSequence(0, 8),
              payload.parentHash.toHexString().subSequence(0, 8),
              expectedParentRooHash.get().toHexString().subSequence(0, 8)
            )
            SafeFuture.failedFuture(IllegalStateException("Reorg detected on block ${payload.blockNumber}"))
          }
        } else {
          SafeFuture.completedFuture(Unit)
        }
      }
      .whenComplete { _, error ->
        log.error("Block creation monitor failed: errorMessage={}", error.message, error)
        log.trace("tick end")
      }
  }

  private fun contunueOrStopIfReorg() {
    if (!reorgDetected.get()) {
      monitorStream.resume()
    } else {
      stop()
    }
  }

  private fun notifyListener(payload: ExecutionPayloadV1): SafeFuture<Unit> {
    return blockCreationListener.acceptBlock(BlockCreated(payload))
      .thenApply {
        log.debug(
          "blockCreationListener blockNumber={} resolved with success",
          payload.blockNumber
        )
      }
      .whenException { throwable ->
        log.warn(
          "Failed to notify blockCreationListener: blockNumber={} errorMessage={}",
          payload.blockNumber.bigIntegerValue(),
          throwable.message,
          throwable
        )
      }
  }

  private fun getNetNextSafeBlock(): SafeFuture<ExecutionPayloadV1?> {
    return extendedWeb3j
      .ethBlockNumber()
      .whenException { log.error("eth_blockNumber failed: errorMessage={}", it.message, it) }
      .thenCompose { latestBlockNumber ->
        // Check if is safe to fetch nextWaitingBlockNumber
        log.trace(
          "nexBlockNumberToFetch={}, blocksToFinalization={}, latestBlockNumber={}",
          _nexBlockNumberToFetch.get(),
          config.blocksToFinalization,
          latestBlockNumber
        )
        if (latestBlockNumber.toLong() >=
          _nexBlockNumberToFetch.get() + config.blocksToFinalization
        ) {
          val blockNumber = _nexBlockNumberToFetch.get()
          extendedWeb3j.ethGetExecutionPayloadByNumber(blockNumber)
            .whenException {
              log.error(
                "eth_getBlockByNumber({}) failed: errorMessage={}",
                blockNumber,
                it.message,
                it
              )
            }
        } else {
          SafeFuture.completedFuture(null)
        }
      }
  }
}
