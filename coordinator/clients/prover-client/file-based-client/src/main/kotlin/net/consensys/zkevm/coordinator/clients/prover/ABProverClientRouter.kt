package net.consensys.zkevm.coordinator.clients.prover

import net.consensys.zkevm.coordinator.clients.ProverClient
import net.consensys.zkevm.domain.BlockInterval
import tech.pegasys.teku.infrastructure.async.SafeFuture

class StartBlockNumberBasedSwitchPredicate<ProofRequest>(
  val switchStartBlockNumberInclusive: ULong
) where ProofRequest : BlockInterval {
  fun invoke(proofRequest: ProofRequest): Boolean = proofRequest.startBlockNumber >= switchStartBlockNumberInclusive
}

class ABProverClientRouter<ProofRequest, ProofResponse>(
  val proverA: ProverClient<ProofRequest, ProofResponse>,
  val proverB: ProverClient<ProofRequest, ProofResponse>,
  val switchToProverBPredicate: (ProofRequest) -> Boolean
) : ProverClient<ProofRequest, ProofResponse> {

  override fun requestProof(proofRequest: ProofRequest): SafeFuture<ProofResponse> {
    if (switchToProverBPredicate(proofRequest)) {
      return proverB.requestProof(proofRequest)
    } else {
      return proverA.requestProof(proofRequest)
    }
  }
}
