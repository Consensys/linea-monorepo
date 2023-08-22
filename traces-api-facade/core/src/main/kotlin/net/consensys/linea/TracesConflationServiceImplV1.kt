package net.consensys.linea

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.flatMap
import io.vertx.core.json.JsonObject
import net.consensys.linea.CommonDomainFunctions.batchIntervalString
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

class TracesCountingServiceV1Impl(
  private val repository: TracesRepositoryV1,
  private val tracesCounter: TracesCounter
) : TracesCountingServiceV1 {
  override fun getBlockTracesCounters(
    block: BlockNumberAndHash
  ): SafeFuture<Result<VersionedResult<BlockCounters>, TracesError>> {
    return repository.getTraces(block).thenApply { result ->
      result.flatMap { blockTraces -> tracesCounter.countTraces(blockTraces.traces) }
    }
  }
}

class TracesConflationServiceV1Impl(
  private val repository: TracesRepositoryV1,
  private val tracesConflator: TracesConflator,
  private val conflatedTracesRepository: ConflatedTracesRepository,
  private val tracesVersion: String
) : TracesConflationServiceV1 {
  val log: Logger = LogManager.getLogger(this::class.java)
  override fun getConflatedTraces(
    blocks: List<BlockNumberAndHash>
  ): SafeFuture<Result<VersionedResult<JsonObject>, TracesError>> {
    return repository.getTraces(blocks)
      .thenApply { result ->
        result.flatMap { blocksTraces ->
          tracesConflator.conflateTraces(blocksTraces.sortedBy { it.blockNumber }.map { it.traces })
        }
      }
  }

  override fun generateConflatedTracesToFile(
    blocks: List<BlockNumberAndHash>
  ): SafeFuture<Result<VersionedResult<String>, TracesError>> {
    val blocksSorted = blocks.sortedBy { it.number }
    // we check if we already have the conflation for the given blocks
    return conflatedTracesRepository.findConflatedTraces(
      blocksSorted.first().number,
      blocksSorted.last().number,
      tracesVersion
    ).thenCompose { conflatedTracesFileName: String? ->
      if (conflatedTracesFileName != null) {
        log.info(
          "Reusing conflated traces for batch={}, file={}",
          batchIntervalString(blocksSorted.first().number, blocksSorted.last().number),
          conflatedTracesFileName
        )
        SafeFuture.completedFuture(Ok(VersionedResult(tracesVersion, conflatedTracesFileName)))
      } else {
        getConflatedTraces(blocks).thenCompose { result ->
          when (result) {
            is Ok -> {
              conflatedTracesRepository
                .saveConflatedTraces(
                  TracesConflation(blocksSorted.first().number, blocksSorted.last().number, result.value)
                )
                .thenApply { Ok(VersionedResult(result.value.version, it)) }
            }

            is Err -> SafeFuture.completedFuture(Err(result.error))
          }
        }
      }
    }
  }
}
