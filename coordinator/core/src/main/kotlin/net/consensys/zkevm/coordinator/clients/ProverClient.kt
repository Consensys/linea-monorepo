package net.consensys.zkevm.coordinator.clients

import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.ProofsToAggregate
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface ProverClient<ProofRequest, ProofResponse> {
  fun requestProof(proofRequest: ProofRequest): SafeFuture<ProofResponse>
}

typealias BlobCompressionProverClientV2 = ProverClient<BlobCompressionProofRequest, BlobCompressionProof>
typealias ProofAggregationProverClientV2 = ProverClient<ProofsToAggregate, ProofToFinalize>
typealias ExecutionProverClientV2 = ProverClient<BatchExecutionProofRequestV1, BatchExecutionProofResponse>
