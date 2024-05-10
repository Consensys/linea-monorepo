package net.consensys.zkevm.coordinator.clients

import com.github.michaelbull.result.Result
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.errors.ErrorResponse
import net.consensys.linea.traces.TracesCounters
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64

enum class TracesServiceErrorType {
  BLOCK_MISSING_IN_CHAIN,
  BLOCK_RANGE_TOO_LARGE,
  INVALID_TRACES_VERSION,
  UNKNOWN_ERROR
}

data class GetTracesCountersResponse(
  val blockL1Size: UInt,
  val tracesCounters: TracesCounters,
  val tracesEngineVersion: String
)

data class GenerateTracesResponse(val tracesFileName: String, val tracesEngineVersion: String)

interface TracesCountersClient {
  fun rollupGetTracesCounters(
    blockNumber: UInt64
  ): SafeFuture<Result<GetTracesCountersResponse, ErrorResponse<TracesServiceErrorType>>>
}

interface TracesConflationClient {
  fun rollupGenerateConflatedTracesToFile(
    startBlockNumber: UInt64,
    endBlockNumber: UInt64
  ): SafeFuture<Result<GenerateTracesResponse, ErrorResponse<TracesServiceErrorType>>>
}

interface TracesCountersClientV1 {
  fun rollupGetTracesCounters(
    block: BlockNumberAndHash
  ): SafeFuture<Result<GetTracesCountersResponse, ErrorResponse<TracesServiceErrorType>>>
}

interface TracesConflationClientV1 {
  fun rollupGenerateConflatedTracesToFile(
    blocks: List<BlockNumberAndHash>
  ): SafeFuture<Result<GenerateTracesResponse, ErrorResponse<TracesServiceErrorType>>>
}

interface TracesWatcher {
  fun waitRawTracesGenerationOf(blockNumber: UInt64, blockHash: Bytes32): SafeFuture<String>
}
