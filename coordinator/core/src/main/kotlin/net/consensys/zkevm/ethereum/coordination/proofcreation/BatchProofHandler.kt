package net.consensys.zkevm.ethereum.coordination.proofcreation

import linea.domain.Batch
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface BatchProofHandler {
  fun acceptNewBatch(batch: Batch): SafeFuture<*>
}
