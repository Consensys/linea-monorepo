package net.consensys.zkevm.ethereum.executionclient

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import net.consensys.linea.errors.ErrorResponse
import tech.pegasys.teku.ethereum.executionclient.ExecutionEngineClient
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceUpdatedResult
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadAttributesV1
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadStatusV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes8
import java.util.Optional

class TekuImplBasedEngineClient(private val delegate: ExecutionEngineClient) :
  net.consensys.zkevm.ethereum.executionclient.ExecutionEngineClient {
  override fun getPayloadV1(
    payloadId: Bytes8
  ): SafeFuture<Result<ExecutionPayloadV1, ErrorResponse<EngineErrorType>>> {
    return delegate.getPayloadV1(payloadId).thenApply(this::unwrapResponse)
  }

  override fun newPayloadV1(
    executionPayload: ExecutionPayloadV1
  ): SafeFuture<Result<PayloadStatusV1, ErrorResponse<EngineErrorType>>> {
    return delegate.newPayloadV1(executionPayload).thenApply(this::unwrapResponse)
  }

  override fun forkChoiceUpdatedV1(
    forkChoiceState: ForkChoiceStateV1,
    payloadAttributes: PayloadAttributesV1?
  ): SafeFuture<Result<ForkChoiceUpdatedResult, ErrorResponse<EngineErrorType>>> {
    return delegate
      .forkChoiceUpdatedV1(forkChoiceState, Optional.ofNullable(payloadAttributes))
      .thenApply(this::unwrapResponse)
  }

  private fun <T> unwrapResponse(
    response: tech.pegasys.teku.ethereum.executionclient.schema.Response<T>
  ): Result<T, ErrorResponse<EngineErrorType>> {
    return if (response.isFailure) {
      Err(ErrorResponse(EngineErrorType.Unknown, response.errorMessage))
    } else {
      Ok(response.payload)
    }
  }
}
