package net.consensys.zkevm.ethereum.coordination.proofcreation

import com.github.michaelbull.result.Result
import com.github.michaelbull.result.mapBoth
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.coordinator.clients.ProverClient
import net.consensys.zkevm.coordinator.clients.ProverErrorType
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import net.consensys.zkevm.ethereum.coordination.conflation.BlocksConflation
import net.consensys.zkevm.ethereum.coordination.conflation.BlocksTracesConflated
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

class FileBasedZkProofCreationCoordinator(private val proverClient: ProverClient) :
  ZkProofCreationCoordinator {

  private val log: Logger = LogManager.getLogger(this::class.java)
  override fun createZkProof(
    blocksConflation: BlocksConflation,
    traces: BlocksTracesConflated
  ): SafeFuture<Batch> {
    return proverClient
      .getZkProof(blocksConflation.blocks, traces.tracesResponse, traces.zkStateTraces)
      .thenCompose { result: Result<GetProofResponse, ErrorResponse<ProverErrorType>> ->
        result.mapBoth(
          { SafeFuture.completedFuture(it) },
          {
            log.error("Prover returned error: errorMessage={}", it.message, it)
            SafeFuture.failedFuture(it.asException("Prover error"))
          }
        )
      }
      .thenApply {
        Batch(
          blocksConflation.blocks.first().blockNumber,
          blocksConflation.blocks.last().blockNumber,
          it
        )
      }
  }
}
