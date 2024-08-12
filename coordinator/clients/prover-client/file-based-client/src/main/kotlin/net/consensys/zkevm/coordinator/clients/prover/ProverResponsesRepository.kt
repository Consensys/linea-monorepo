package net.consensys.zkevm.coordinator.clients.prover

import net.consensys.zkevm.domain.ProofIndex
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface ProverResponsesRepository {
  fun find(index: ProofIndex): SafeFuture<Unit>
  fun monitor(index: ProofIndex): SafeFuture<Unit>
}
