package net.consensys.zkevm.ethereum.coordination.proofcreation

import net.consensys.zkevm.coordinator.clients.ExecutionProverClient
import net.consensys.zkevm.domain.Batch
import net.consensys.zkevm.domain.BlocksConflation
import net.consensys.zkevm.ethereum.coordination.conflation.BlocksTracesConflated
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

class ZkProofCreationCoordinatorImpl(
  private val executionProverClient: ExecutionProverClient
) : ZkProofCreationCoordinator {
  private val log: Logger = LogManager.getLogger(this::class.java)

  override fun createZkProof(
    blocksConflation: BlocksConflation,
    traces: BlocksTracesConflated
  ): SafeFuture<Batch> {
    val startBlockNumber = blocksConflation.blocks.first().blockNumber.toULong()
    val endBlockNumber = blocksConflation.blocks.last().blockNumber.toULong()

    return executionProverClient
      .requestBatchExecutionProof(blocksConflation.blocks, traces.tracesResponse, traces.zkStateTraces)
      .thenApply {
        Batch(
          startBlockNumber = startBlockNumber,
          endBlockNumber = endBlockNumber,
          status = Batch.Status.Proven
        )
      }
      .whenException {
        log.error("Prover returned error: errorMessage={}", it.message, it)
      }
  }
}
