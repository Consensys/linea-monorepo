package net.consensys.zkevm.ethereum.coordination.conflation

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.getOrElse
import com.github.michaelbull.result.getOrThrow
import com.github.michaelbull.result.runCatching
import io.vertx.core.Vertx
import linea.domain.Block
import net.consensys.linea.async.AsyncRetryer
import net.consensys.zkevm.domain.Batch
import net.consensys.zkevm.domain.BlocksConflation
import net.consensys.zkevm.ethereum.coordination.proofcreation.BatchProofHandler
import net.consensys.zkevm.ethereum.coordination.proofcreation.ZkProofCreationCoordinator
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration

class ProofGeneratingConflationHandlerImpl(
  private val tracesProductionCoordinator: TracesConflationCoordinator,
  private val zkProofProductionCoordinator: ZkProofCreationCoordinator,
  private val batchProofHandler: BatchProofHandler,
  private val batchAlreadyProvenSupplier: (Batch) -> SafeFuture<Boolean>,
  private val vertx: Vertx,
  private val config: Config,
) : ConflationHandler {
  private val log: Logger = LogManager.getLogger(this::class.java)

  data class Config(
    val conflationAndProofGenerationRetryInterval: Duration,
  )

  override fun handleConflatedBatch(conflation: BlocksConflation): SafeFuture<*> {
    val blockIntervalString = conflation.conflationResult.intervalString()
    return runCatching {
      log.info(
        "new batch: batch={} trigger={} tracesCounters={}",
        blockIntervalString,
        conflation.conflationResult.conflationTrigger,
        conflation.conflationResult.tracesCounters,
      )
      AsyncRetryer.retry(
        vertx = vertx,
        backoffDelay = config.conflationAndProofGenerationRetryInterval,
        exceptionConsumer = {
          // log failure as warning, but keeps on retrying...
          log.warn(
            "conflation and proof creation flow failed batch={} errorMessage={}",
            blockIntervalString,
            it.message,
          )
        },
      ) {
        conflationToProofCreation(conflation)
      }
    }.getOrElse { error -> SafeFuture.failedFuture<Unit>(error) }
      .whenException { th ->
        log.error(
          "traces conflation or proof request failed: batch={} errorMessage={}",
          blockIntervalString,
          th.message,
          th,
        )
      }
  }

  private fun conflationToProofCreation(conflation: BlocksConflation): SafeFuture<*> {
    val blockIntervalString = conflation.conflationResult.intervalString()
    return assertConsecutiveBlocksRange(conflation.blocks)
      .getOrThrow().let { blocksRange ->
        val batch = Batch(conflation.startBlockNumber, conflation.endBlockNumber)
        batchAlreadyProvenSupplier(batch)
          .thenCompose { responseAlreadyDone ->
            if (responseAlreadyDone) {
              log.info("skipping conflation and proof request: batch={} already proven", blockIntervalString)
              SafeFuture.completedFuture(batch)
            } else {
              tracesProductionCoordinator
                .conflateExecutionTraces(blocksRange)
                .whenException { th ->
                  log.debug(
                    "traces conflation failed: batch={} errorMessage={}",
                    blockIntervalString,
                    th.message,
                    th,
                  )
                }
                .thenCompose { blocksTracesConflated: BlocksTracesConflated ->
                  log.debug(
                    "requesting execution proof: batch={} tracesFile={}",
                    blockIntervalString,
                    blocksTracesConflated.tracesResponse.tracesFileName,
                  )
                  zkProofProductionCoordinator
                    .createZkProof(conflation, blocksTracesConflated)
                    .thenPeek {
                      log.info("execution proof generated: batch={}", blockIntervalString)
                    }
                    .whenException { th ->
                      log.debug(
                        "execution proof failure: batch={} errorMessage={}",
                        blockIntervalString,
                        th.message,
                        th,
                      )
                    }
                }
            }
          }
          .thenCompose { batchProofHandler.acceptNewBatch(batch) }
      }
  }
}

internal fun assertConsecutiveBlocksRange(
  blocks: List<Block>,
): Result<ULongRange, IllegalArgumentException> {
  if (blocks.isEmpty()) {
    return Err(IllegalArgumentException("Empty list of blocks"))
  }

  if (blocks.size == 1) {
    return Ok(blocks.first().number..blocks.last().number)
  }

  val sortedByNumber = blocks.sortedBy { it.number }
  val gapFound = sortedByNumber
    .zipWithNext { a, b -> b.number - a.number }
    .any { it != 1UL }

  if (gapFound) {
    return Err(IllegalArgumentException("Conflated blocks list has non consecutive blocks!"))
  }
  return Ok(sortedByNumber.first().number..sortedByNumber.last().number)
}
