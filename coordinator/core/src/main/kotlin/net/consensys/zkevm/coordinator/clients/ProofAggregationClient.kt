package net.consensys.zkevm.coordinator.clients

import com.github.michaelbull.result.Result
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.ProofsToAggregate
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface ProofAggregationClient {
  fun getAggregatedProof(aggregation: ProofsToAggregate):
    SafeFuture<Result<ProofToFinalize, ErrorResponse<ProverErrorType>>>
}
