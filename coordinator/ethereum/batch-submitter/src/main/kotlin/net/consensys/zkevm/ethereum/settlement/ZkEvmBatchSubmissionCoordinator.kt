package net.consensys.zkevm.ethereum.settlement

import io.vertx.core.TimeoutStream
import io.vertx.core.Vertx
import kotlinx.datetime.Clock
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.contract.ZkEvmV2AsyncFriendly
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import net.consensys.zkevm.ethereum.settlement.persistence.BatchesRepository
import net.consensys.zkevm.ethereum.settlement.persistence.DuplicatedBatchException
import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import kotlin.time.Duration

class ZkEvmBatchSubmissionCoordinator(
  private val config: Config,
  private val batchSubmitter: BatchSubmitter,
  private val batchesRepository: BatchesRepository,
  private val zkEvmV2: ZkEvmV2AsyncFriendly,
  private val vertx: Vertx,
  private val clock: Clock
) : BatchSubmissionCoordinatorService {
  private val log = LogManager.getLogger(this::class.java)
  private lateinit var monitorStream: TimeoutStream

  class Config(
    val pollingInterval: Duration,
    val proofSubmissionDelay: Duration
  )

  private fun joinString(batches: List<Batch>): String {
    return batches.joinToString(",", "[", "]") {
      if (it.startBlockNumber == it.endBlockNumber) {
        it.startBlockNumber.toString()
      } else {
        "(${it.startBlockNumber}..${it.endBlockNumber})"
      }
    }
  }

  private fun tick(): SafeFuture<Unit> {
    return zkEvmV2.resetNonce().thenCompose {
      SafeFuture.of(zkEvmV2.currentL2BlockNumber().sendAsync())
    }
      .thenCompose { lastFinalizedBlock ->
        log.debug("Block number from the contract: {}", lastFinalizedBlock)

        val promise = batchesRepository
          .getConsecutiveBatchesFromBlockNumber(
            UInt64.valueOf(lastFinalizedBlock.longValueExact()).increment(),
            clock.now().minus(config.proofSubmissionDelay)
          )
        promise
      }
      .thenCompose(::sendBatches)
      .whenException { log.error("Exception during Submission coordination! errorMessage={}", it.message, it) }
  }

  private fun sendBatches(batches: List<Batch>): SafeFuture<Unit> {
    return vertx.executeBlocking { promise ->
      try {
        log.debug(
          "Batches to send: {} {}",
          batches.size,
          joinString(batches)
        )

        if (batches.isNotEmpty()) {
          submitBatchesAfterEthCall(batches)
            .whenException { error ->
              log.error(
                "Exception while trying to submit batches: startingBlock={} endBlock={}!",
                batches.first().startBlockNumber,
                batches.last().endBlockNumber
              )
              promise.fail(error)
            }
        } else {
          SafeFuture.completedFuture(Unit)
        }
          .whenSuccess {
            promise.complete(Unit)
          }
      } catch (th: Throwable) {
        promise.fail(th)
      }
    }.toSafeFuture()
  }

  override fun start(): SafeFuture<Unit> {
    monitorStream =
      vertx.periodicStream(config.pollingInterval.inWholeMilliseconds).handler {
        try {
          monitorStream.pause()
          tick().whenComplete { _, _ -> monitorStream.resume() }
        } catch (th: Throwable) {
          log.error(th)
          monitorStream.resume()
        }
      }
    return SafeFuture.completedFuture(Unit)
  }

  override fun stop(): SafeFuture<Unit> {
    return if (this::monitorStream.isInitialized) {
      SafeFuture.completedFuture(monitorStream.cancel())
    } else {
      log.warn("Batch submission coordinator hasn't been started to stop it, but Ok")
      SafeFuture.completedFuture(Unit)
    }
  }

  override fun acceptNewBatch(batch: Batch): SafeFuture<Unit> {
    return batchesRepository.saveNewBatch(batch)
      .exceptionally { th ->
        if (th is DuplicatedBatchException) {
          log.debug(
            "Ignoring Batch already persisted error. batch={} errorMessage={}",
            batch.intervalString(),
            th.message
          )
          SafeFuture.completedFuture(Unit)
        } else {
          SafeFuture.failedFuture(th)
        }
      }
  }

  private fun submitBatchesAfterEthCall(batches: List<Batch>): SafeFuture<*> =
    batchSubmitter.submitBatchCall(batches.first())
      .thenCompose {
        val batchesFutures = batches.map { batch ->
          log.debug(
            "Sending batch: batch={} nonce={}",
            batch.intervalString(),
            zkEvmV2.currentNonce()
          )
          batchSubmitter.submitBatch(batch)
        }
        SafeFuture.allOf(batchesFutures.stream())
      }
}
