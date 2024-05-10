package net.consensys.zkevm.coordinator.blockcreation

import io.vertx.core.TimeoutStream
import io.vertx.core.Vertx
import net.consensys.linea.async.AsyncRetryer
import net.consensys.zkevm.PeriodicPollingService
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockCreated
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockCreationListener
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.Response
import org.web3j.protocol.core.methods.response.EthBlock
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.atomic.AtomicBoolean
import java.util.concurrent.atomic.AtomicLong
import java.util.concurrent.atomic.AtomicReference
import kotlin.time.Duration
import kotlin.time.Duration.Companion.days

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
          response.block?.let {
            SafeFuture.completedFuture(response.block.toExecutionPayloadV1())
          } ?: SafeFuture.failedFuture(Exception("Block $blockNumber not found!"))
        }
      }
  }
}

class BlockCreationMonitor(
  private val vertx: Vertx,
  private val extendedWeb3j: ExtendedWeb3J,
  private val startingBlockNumberExclusive: Long,
  private val blockCreationListener: BlockCreationListener,
  private val lastProvenBlockNumberProviderAsync: LastProvenBlockNumberProviderAsync,
  private val config: Config,
  private val log: Logger = LogManager.getLogger(BlockCreationMonitor::class.java)
) : PeriodicPollingService(
  vertx = vertx,
  pollingIntervalMs = config.pollingInterval.inWholeMilliseconds,
  log = log
) {
  data class Config(
    val pollingInterval: Duration,
    val blocksToFinalization: Long,
    val blocksFetchLimit: Long,
    val startingBlockWaitTimeout: Duration = 14.days
  )

  private val _nexBlockNumberToFetch: AtomicLong = AtomicLong(startingBlockNumberExclusive + 1)
  private val expectedParentBlockHash: AtomicReference<Bytes32> = AtomicReference(null)
  private val reorgDetected: AtomicBoolean = AtomicBoolean(false)
  private var statingBlockAvailabilityFuture: SafeFuture<*>? = null
  private lateinit var monitorStream: TimeoutStream

  val nexBlockNumberToFetch: Long
    get() = _nexBlockNumberToFetch.get()

  override fun handleError(error: Throwable) {
    log.error("Error with block creation monitor: errorMessage={}", error.message, error)
  }

  @Synchronized
  override fun start(): SafeFuture<Unit> {
    if (reorgDetected.get()) {
      return SafeFuture.failedFuture(IllegalStateException("Reorg detect. Cannot restart"))
    }
    if (this::monitorStream.isInitialized) {
      this.monitorStream.cancel()
    }

    return awaitStartingBlockToBePresent()
      .thenApply {
        super.start()
      }
  }

  @Synchronized
  fun awaitStartingBlockToBePresent(): SafeFuture<*> {
    if (statingBlockAvailabilityFuture == null) {
      log.info("Awaiting for block {} to be present", startingBlockNumberExclusive)
      statingBlockAvailabilityFuture = AsyncRetryer.retry(
        vertx,
        backoffDelay = config.pollingInterval,
        timeout = config.startingBlockWaitTimeout,
        stopRetriesPredicate = { block: EthBlock ->
          if (block.block == null) {
            log.warn(
              "Block {} not found yet. Retrying in {}",
              startingBlockNumberExclusive,
              config.pollingInterval
            )
            false
          } else {
            log.info("Block {} found. Resuming block monitor", startingBlockNumberExclusive)
            expectedParentBlockHash.set(Bytes32.fromHexString(block.block.hash))
            true
          }
        }
      ) {
        SafeFuture.of(
          extendedWeb3j.web3jClient
            .ethGetBlockByNumber(
              DefaultBlockParameter.valueOf(startingBlockNumberExclusive.toBigInteger()),
              false
            )
            .sendAsync()
        )
      }
    }

    return statingBlockAvailabilityFuture!!
  }

  override fun action(): SafeFuture<*> {
    log.trace("tick start")
    return lastProvenBlockNumberProviderAsync.getLastProvenBlockNumber()
      .thenCompose { lastProvenBlockNumber ->
        if (!nextBlockNumberWithinLimit(lastProvenBlockNumber)) {
          log.warn(
            "Gap between highest consecutive proven block and L2 block is too big: lastProvenBlock={} " +
              "nextBlockToFetch={} gapOverflow={} gapLimit={}",
            lastProvenBlockNumber,
            _nexBlockNumberToFetch.get(),
            _nexBlockNumberToFetch.get() - lastProvenBlockNumber,
            config.blocksFetchLimit
          )
          SafeFuture.COMPLETE
        } else {
          getNetNextSafeBlock()
            .thenCompose { payload ->
              if (payload != null) {
                if (payload.parentHash == expectedParentBlockHash.get()) {
                  notifyListener(payload)
                    .whenSuccess {
                      log.debug(
                        "updating nexBlockNumberToFetch from {} --> {}",
                        _nexBlockNumberToFetch.get(),
                        _nexBlockNumberToFetch.incrementAndGet()
                      )
                      expectedParentBlockHash.set(payload.blockHash)
                    }
                } else {
                  reorgDetected.set(true)
                  log.error(
                    "Shooting down conflation poller, " +
                      "chain reorg detected: block { blockNumber={} hash={} parentHash={} } should have parentHash={}",
                    payload.blockNumber.longValue(),
                    payload.blockHash.toHexString().subSequence(0, 8),
                    payload.parentHash.toHexString().subSequence(0, 8),
                    expectedParentBlockHash.get().toHexString().subSequence(0, 8)
                  )
                  SafeFuture.failedFuture(IllegalStateException("Reorg detected on block ${payload.blockNumber}"))
                }
              } else {
                SafeFuture.completedFuture(Unit)
              }
            }
            .whenException { error ->
              log.warn("Block creation monitor failed: errorMessage={}", error.message, error)
            }.whenComplete { _, _ ->
              log.trace("tick end")
            }
        }
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
      .thenCompose { latestBlockNumber ->
        // Check if is safe to fetch nextWaitingBlockNumber
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

  private fun nextBlockNumberWithinLimit(lastProvenBlockNumber: Long): Boolean {
    return _nexBlockNumberToFetch.get() - lastProvenBlockNumber <= config.blocksFetchLimit
  }
}
