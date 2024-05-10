package net.consensys.zkevm.ethereum.coordination.proofcreation

import com.github.michaelbull.result.Result
import com.github.michaelbull.result.mapBoth
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.ExecutionProverClient
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.coordinator.clients.ProverErrorType
import net.consensys.zkevm.domain.Batch
import net.consensys.zkevm.domain.BlocksConflation
import net.consensys.zkevm.ethereum.coordination.conflation.BlocksTracesConflated
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

class ZkProofCreationCoordinatorImpl(
  private val executionProverClient: ExecutionProverClient,
  private val config: Config
) : ZkProofCreationCoordinator {
  private val log: Logger = LogManager.getLogger(this::class.java)

  data class Config(
    val conflationCalculatorVersion: String
  )

  override fun createZkProof(
    blocksConflation: BlocksConflation,
    traces: BlocksTracesConflated
  ): SafeFuture<Batch> {
    val startBlockNumber = blocksConflation.blocks.first().blockNumber.toULong()
    val endBlockNumber = blocksConflation.blocks.last().blockNumber.toULong()

    return executionProverClient
      .getBatchProof(blocksConflation.blocks, traces.tracesResponse, traces.zkStateTraces)
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
          startBlockNumber = startBlockNumber,
          endBlockNumber = endBlockNumber,
          status = Batch.Status.Proven,
          conflationVersion = config.conflationCalculatorVersion
        )
      }
  }
}
