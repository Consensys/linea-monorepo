package net.consensys.zkevm.ethereum.coordination.proofcreation

import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import net.consensys.zkevm.ethereum.coordination.conflation.BlocksConflation
import net.consensys.zkevm.ethereum.coordination.conflation.BlocksTracesConflated
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface ZkProofCreationCoordinator {
  fun createZkProof(
    blocksConflation: BlocksConflation,
    traces: BlocksTracesConflated
  ): SafeFuture<Batch>
}
