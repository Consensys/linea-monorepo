package net.consensys.zkevm.coordinator.clients.prover

import com.github.michaelbull.result.Result
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.coordinator.clients.ProverErrorType
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface ProverResponsesRepository {
  data class ProverResponseIndex(
    val startBlockNumber: ULong,
    val endBlockNumber: ULong,
    val version: String
  )
  fun find(index: ProverResponseIndex): SafeFuture<Result<GetProofResponse, ErrorResponse<ProverErrorType>>>
  fun monitor(index: ProverResponseIndex): SafeFuture<Result<GetProofResponse, ErrorResponse<ProverErrorType>>>
}
