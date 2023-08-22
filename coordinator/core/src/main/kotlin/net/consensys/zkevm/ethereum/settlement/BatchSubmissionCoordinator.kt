package net.consensys.zkevm.ethereum.settlement

import net.consensys.zkevm.LongRunningService
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface BatchSubmissionCoordinator {
  fun acceptNewBatch(batch: Batch): SafeFuture<Unit>
}

interface BatchSubmissionCoordinatorService : BatchSubmissionCoordinator, LongRunningService
