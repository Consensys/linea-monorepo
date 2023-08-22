package net.consensys.zkevm.coordinator.clients.response

import com.github.michaelbull.result.Result
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.coordinator.clients.ProverErrorType
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64

interface ProverResponsesRepository {
  data class ProverResponseIndex(
    val startBlockNumber: UInt64,
    val endBlockNumber: UInt64,
    val version: String
  )
  fun find(
    index: ProverResponseIndex
  ): SafeFuture<Result<GetProofResponse, ErrorResponse<ProverErrorType>>>
}
