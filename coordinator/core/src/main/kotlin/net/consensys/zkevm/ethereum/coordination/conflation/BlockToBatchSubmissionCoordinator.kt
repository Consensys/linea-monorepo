package net.consensys.zkevm.ethereum.coordination.conflation

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import io.vertx.core.Vertx
import kotlinx.datetime.Instant
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.async.retryWithInterval
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.GetTracesCountersResponse
import net.consensys.zkevm.coordinator.clients.TracesCountersClientV1
import net.consensys.zkevm.coordinator.clients.TracesServiceErrorType
import net.consensys.zkevm.coordinator.clients.TracesWatcher
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockCreated
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockCreationListener
import net.consensys.zkevm.ethereum.coordination.proofcreation.ZkProofCreationCoordinator
import net.consensys.zkevm.ethereum.settlement.BatchSubmissionCoordinator
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration.Companion.seconds

class BlockToBatchSubmissionCoordinator(
  private val conflationService: ConflationService,
  private val tracesFileManager: TracesWatcher,
  private val tracesCountersClient: TracesCountersClientV1,
  private val tracesProductionCoordinator: TracesConflationCoordinator,
  private val zkProofProductionCoordinator: ZkProofCreationCoordinator,
  private val batchSubmissionCoordinator: BatchSubmissionCoordinator,
  private val vertx: Vertx
) : BlockCreationListener {
  private val log: Logger = LogManager.getLogger(this::class.java)

  init {
    conflationService.onConflatedBatch(::conflatedBatchHandler)
  }

  override fun acceptBlock(blockEvent: BlockCreated): SafeFuture<Unit> {
    log.debug("Accepting new block={}", blockEvent.executionPayload.blockNumber)
    tracesFileManager
      .waitRawTracesGenerationOf(
        blockEvent.executionPayload.blockNumber,
        blockEvent.executionPayload.blockHash
      )
      .thenCompose {
        log.trace("Traces file generated: block={}", blockEvent.executionPayload.blockNumber)
        tracesCountersClient.rollupGetTracesCounters(
          BlockNumberAndHash(
            blockEvent.executionPayload.blockNumber.longValue().toULong(),
            blockEvent.executionPayload.blockHash
          )
        )
      }
      .thenCompose { result ->
        when (result) {
          is Err<ErrorResponse<TracesServiceErrorType>> -> {
            SafeFuture.failedFuture(result.error.asException("Traces api error: "))
          }

          is Ok<GetTracesCountersResponse> -> {
            log.trace(
              "Traces counters returned: block={}, blockL1Size={} bytes, counters={}",
              blockEvent.executionPayload.blockNumber,
              result.value.blockL1Size,
              result.value.tracesCounters
            )
            runCatching {
              conflationService.newBlock(
                blockEvent.executionPayload,
                BlockCounters(
                  blockEvent.executionPayload.blockNumber.toULong(),
                  Instant.fromEpochSeconds(blockEvent.executionPayload.timestamp.longValue()),
                  result.value.tracesCounters,
                  result.value.blockL1Size
                )
              )
            }.map { SafeFuture.completedFuture(it) }
              .getOrElse { error -> SafeFuture.failedFuture(error) }
          }
        }
      }.whenException { th ->
        log.error(
          "Failed to conflate block={}, errorMessage={}",
          blockEvent.executionPayload.blockNumber,
          th.message,
          th
        )
      }

    return SafeFuture.completedFuture(Unit)
  }

  private fun conflatedBatchHandler(conflation: BlocksConflation): SafeFuture<Unit> {
    return runCatching {
      log.info(
        "Starting conflation: batch={}, trigger={}, dataL1Size={} bytes, tracesCounters={}",
        conflation.conflationResult.intervalString(),
        conflation.conflationResult.conflationTrigger,
        conflation.conflationResult.dataL1Size,
        conflation.conflationResult.tracesCounters
      )
      retryWithInterval(3, 5.seconds, vertx) { conflationToBatchSubmission(conflation) }
    }.getOrElse { error -> SafeFuture.failedFuture(error) }
      .whenException { th ->
        log.error(
          "Traces conflation or proof failed for: batch={}, errorMessage={}",
          conflation.conflationResult.intervalString(),
          th.message,
          th
        )
      }
  }

  private fun conflationToBatchSubmission(conflation: BlocksConflation): SafeFuture<Unit> {
    val blockNumbersAndHash = conflation.blocks.map {
      BlockNumberAndHash(it.blockNumber.toULong(), it.blockHash)
    }
    return tracesProductionCoordinator
      .conflateExecutionTraces(blockNumbersAndHash)
      .whenException { th ->
        log.error(
          "Traces conflation failed: batch={} errorMessage={}",
          conflation.conflationResult.intervalString(),
          th.message,
          th
        )
      }
      .thenCompose { blocksTracesConflated: BlocksTracesConflated ->
        log.info(
          "Requesting proof for: batch={} tracesFile={}",
          conflation.conflationResult.intervalString(),
          blocksTracesConflated.tracesResponse.tracesFileName
        )
        zkProofProductionCoordinator
          .createZkProof(conflation, blocksTracesConflated)
          .thenPeek {
            log.info("Batch proof created for: batch={}", conflation.conflationResult.intervalString())
          }
          .whenException { th ->
            log.error(
              "Batch proof failure: batch={} {}",
              conflation.conflationResult.intervalString(),
              th.message,
              th
            )
          }
      }
      .thenCompose(batchSubmissionCoordinator::acceptNewBatch)
  }
}
