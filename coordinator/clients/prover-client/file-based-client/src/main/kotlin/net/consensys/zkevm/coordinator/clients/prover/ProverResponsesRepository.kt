package net.consensys.zkevm.coordinator.clients.prover

import com.github.michaelbull.result.Result
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.ProverErrorType
import net.consensys.zkevm.domain.ProofIndex
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface ProverResponsesRepository {
  fun find(index: ProofIndex): SafeFuture<Result<ExecutionProofResponse, ErrorResponse<ProverErrorType>>>
  fun monitor(index: ProofIndex): SafeFuture<Result<ExecutionProofResponse, ErrorResponse<ProverErrorType>>>
}
