package net.consensys.zkevm.ethereum.settlement

import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface BatchSubmitter {
  fun submitBatch(batch: Batch): SafeFuture<*>

  fun submitBatchCall(batch: Batch): SafeFuture<*>
}
