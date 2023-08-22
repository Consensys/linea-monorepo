package net.consensys.zkevm.ethereum.executionclient

import com.github.michaelbull.result.Result
import net.consensys.linea.errors.ErrorResponse
import net.consensys.linea.traces.TracesCounters
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceUpdatedResult
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadAttributesV1
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadStatusV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes8
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import kotlin.time.Duration

enum class EngineErrorType {
  Unknown
}

typealias EngineError = ErrorResponse<EngineErrorType>

interface ExecutionEngineClient {
  fun getPayloadV1(
    payloadId: Bytes8
  ): SafeFuture<Result<ExecutionPayloadV1, ErrorResponse<EngineErrorType>>>
  fun newPayloadV1(
    executionPayload: ExecutionPayloadV1
  ): SafeFuture<Result<PayloadStatusV1, ErrorResponse<EngineErrorType>>>
  fun forkChoiceUpdatedV1(
    forkChoiceState: ForkChoiceStateV1,
    payloadAttributes: PayloadAttributesV1? = null
  ): SafeFuture<Result<ForkChoiceUpdatedResult, ErrorResponse<EngineErrorType>>>
}

data class RollupCreatePayloadResult(
  val payloadStatus: PayloadStatusV1,
  val payload: ExecutionPayloadV1,
  val tracesCounters: TracesCounters
)

interface RollupEngineClient : ExecutionEngineClient {
  fun rollupCreatePayloadV0(
    forkChoiceState: ForkChoiceStateV1,
    payloadAttributes: PayloadAttributesV1,
    timeout: Duration
  ): SafeFuture<Result<RollupCreatePayloadResult, EngineError>>

  fun rollupGetTracesCountersByBlockNumberV0(
    blockNumber: UInt64,
    tracesEngineVersion: String?
  ): SafeFuture<Result<TracesCounters, EngineError>>
}
