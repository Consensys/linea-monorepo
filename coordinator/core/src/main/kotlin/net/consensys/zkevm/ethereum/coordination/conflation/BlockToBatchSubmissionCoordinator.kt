package net.consensys.zkevm.ethereum.coordination.conflation

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.getOrElse
import com.github.michaelbull.result.map
import com.github.michaelbull.result.runCatching
import io.vertx.core.Vertx
import io.vertx.core.json.JsonObject
import kotlinx.datetime.Instant
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.GetTracesCountersResponse
import net.consensys.zkevm.coordinator.clients.TracesCountersClientV1
import net.consensys.zkevm.coordinator.clients.TracesServiceErrorType
import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.encoding.ExecutionPayloadV1Encoder
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockCreated
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockCreationListener
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.Callable

class BlockToBatchSubmissionCoordinator(
  private val conflationService: ConflationService,
  private val tracesCountersClient: TracesCountersClientV1,
  private val vertx: Vertx,
  private val payloadEncoder: ExecutionPayloadV1Encoder,
  private val log: Logger = LogManager.getLogger(BlockToBatchSubmissionCoordinator::class.java)
) : BlockCreationListener {
  private fun getTracesCounters(
    blockEvent: BlockCreated
  ): SafeFuture<GetTracesCountersResponse> {
    return tracesCountersClient
      .rollupGetTracesCounters(
        BlockNumberAndHash(
          blockEvent.executionPayload.blockNumber.toULong(),
          blockEvent.executionPayload.blockHash.toArray()
        )
      )
      .thenCompose { result ->
        when (result) {
          is Err<ErrorResponse<TracesServiceErrorType>> -> {
            SafeFuture.failedFuture(result.error.asException("Traces api error: "))
          }

          is Ok<GetTracesCountersResponse> -> {
            runCatching {
              parseTracesCountersResponseToJson(
                blockEvent.executionPayload.blockNumber.longValue(),
                blockEvent.executionPayload.blockHash.toHexString(),
                result.value
              )
            }.map {
              log.info("Traces counters returned in JSON: {}", it)
            }.getOrElse {
              log.error(
                "Error when parsing traces counters to JSON for block {}-{}: {}",
                blockEvent.executionPayload.blockNumber.longValue(),
                blockEvent.executionPayload.blockHash.toHexString(),
                it.message
              )
            }
            SafeFuture.completedFuture(result.value)
          }
        }
      }
  }

  override fun acceptBlock(blockEvent: BlockCreated): SafeFuture<Unit> {
    log.debug("Accepting new block={}", blockEvent.executionPayload.blockNumber)
    vertx.executeBlocking(
      Callable {
        payloadEncoder.encode(blockEvent.executionPayload)
      }
    ).toSafeFuture().thenCombine(getTracesCounters(blockEvent)) { blockRLPEncoded, traces ->
      conflationService.newBlock(
        blockEvent.executionPayload,
        BlockCounters(
          blockNumber = blockEvent.executionPayload.blockNumber.toULong(),
          blockTimestamp = Instant.fromEpochSeconds(blockEvent.executionPayload.timestamp.longValue()),
          tracesCounters = traces.tracesCounters,
          blockRLPEncoded = blockRLPEncoded
        )
      )
    }.whenException { th ->
      log.error(
        "Failed to conflate block={} errorMessage={}",
        blockEvent.executionPayload.blockNumber,
        th.message,
        th
      )
    }

    // This is to parallelize `getTracesCounters` requests which would otherwise be sent sequentially
    return SafeFuture.completedFuture(Unit)
  }

  internal companion object {
    fun parseTracesCountersResponseToJson(
      blockNumber: Long,
      blockHash: String,
      tcResponse: GetTracesCountersResponse
    ): JsonObject {
      return JsonObject.of(
        "tracesEngineVersion",
        tcResponse.tracesEngineVersion,
        "blockNumber",
        blockNumber,
        "blockHash",
        blockHash,
        "tracesCounters",
        tcResponse.tracesCounters
          .entries()
          .map { it.first.name to it.second.toLong() }
          .sortedBy { it.first }
          .toMap()
      )
    }
  }
}
