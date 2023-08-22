package net.consensys.zkevm.ethereum.coordination.conflation

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.getOrElse
import com.github.michaelbull.result.map
import com.github.michaelbull.result.mapBoth
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.GenerateTracesResponse
import net.consensys.zkevm.coordinator.clients.GetZkEVMStateMerkleProofResponse
import net.consensys.zkevm.coordinator.clients.TracesConflationClientV1
import net.consensys.zkevm.coordinator.clients.TracesServiceErrorType
import net.consensys.zkevm.coordinator.clients.Type2StateManagerClient
import net.consensys.zkevm.coordinator.clients.Type2StateManagerErrorType
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64

class TracesConflationCoordinatorImpl(
  private val tracesConflationClient: TracesConflationClientV1,
  private val zkStateClient: Type2StateManagerClient
) : TracesConflationCoordinator {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private fun requestConflatedTraces(
    blocks: List<BlockNumberAndHash>
  ): SafeFuture<GenerateTracesResponse> {
    return tracesConflationClient
      .rollupGenerateConflatedTracesToFile(blocks)
      .thenCompose { result: Result<GenerateTracesResponse, ErrorResponse<TracesServiceErrorType>>
        ->
        result.mapBoth(
          { SafeFuture.completedFuture(it) },
          {
            log.error("Conflation service returned error: errorMessage={}", it.message, it)
            SafeFuture.failedFuture(it.asException("Conflation service error"))
          }
        )
      }
  }

  private fun requestStateMerkleProof(
    startBlockNumber: ULong,
    endBlockNumber: ULong
  ): SafeFuture<GetZkEVMStateMerkleProofResponse> {
    return zkStateClient
      .rollupGetZkEVMStateMerkleProof(
        UInt64.valueOf(startBlockNumber.toLong()),
        UInt64.valueOf(endBlockNumber.toLong())
      )
      .thenCompose { result:
            Result<GetZkEVMStateMerkleProofResponse, ErrorResponse<Type2StateManagerErrorType>>
        ->
        result.mapBoth(
          { SafeFuture.completedFuture(it) },
          {
            log.error("Type2State manager returned error={}", it)
            SafeFuture.failedFuture(it.asException("State manager error: ${it.message}"))
          }
        )
      }
  }

  override fun conflateExecutionTraces(
    blocks: List<BlockNumberAndHash>
  ): SafeFuture<BlocksTracesConflated> {
    return assertBlocksList(blocks).map { sortedByNumber ->
      requestConflatedTraces(blocks).thenCombine(
        requestStateMerkleProof(sortedByNumber.first().number, sortedByNumber.last().number)
      ) { tracesConflationResult: GenerateTracesResponse,
        zkStateUpdateResult: GetZkEVMStateMerkleProofResponse ->
        BlocksTracesConflated(tracesConflationResult, zkStateUpdateResult)
      }
    }.getOrElse { SafeFuture.failedFuture<BlocksTracesConflated>(it) }
  }
}

internal fun assertBlocksList(
  blocks: List<BlockNumberAndHash>
): Result<List<BlockNumberAndHash>, IllegalArgumentException> {
  if (blocks.isEmpty()) {
    return Err(IllegalArgumentException("Empty list of blocs"))
  }

  if (blocks.size == 1) {
    return Ok(blocks)
  }

  val sortedByNumber = blocks.sortedBy { it.number }
  var prevBlockNumber = sortedByNumber.first().number
  var gapFound = false
  for (i in 1 until sortedByNumber.size) {
    val block = sortedByNumber[i]
    if (block.number != prevBlockNumber + 1u) {
      gapFound = true
      break
    }
    prevBlockNumber = block.number
  }
  gapFound

  if (gapFound) {
    return Err(IllegalArgumentException("Conflated blocks list has non consecutive blocks!"))
  }
  return Ok(sortedByNumber)
}
