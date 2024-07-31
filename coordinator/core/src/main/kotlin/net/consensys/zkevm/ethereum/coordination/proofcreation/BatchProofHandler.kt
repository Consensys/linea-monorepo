package net.consensys.zkevm.ethereum.coordination.proofcreation

import net.consensys.zkevm.domain.Batch
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface BatchProofHandler {
  fun acceptNewBatch(batch: Batch): SafeFuture<*>
}
