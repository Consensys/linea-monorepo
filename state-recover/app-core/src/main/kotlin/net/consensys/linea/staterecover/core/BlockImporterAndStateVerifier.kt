package net.consensys.linea.staterecover.core

import io.vertx.core.Vertx
import net.consensys.linea.BlockInterval
import net.consensys.linea.BlockParameter
import net.consensys.linea.async.AsyncRetryer
import net.consensys.linea.staterecover.clients.elclient.ExecutionLayerClient
import net.consensys.linea.staterecover.domain.BlockL1RecoveredData
import net.consensys.zkevm.coordinator.clients.StateManagerClientV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

data class ImportResult(
  val blockNumber: ULong,
  val zkStateRootHash: ByteArray
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as ImportResult

    if (blockNumber != other.blockNumber) return false
    if (!zkStateRootHash.contentEquals(other.zkStateRootHash)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = blockNumber.hashCode()
    result = 31 * result + zkStateRootHash.contentHashCode()
    return result
  }
}

interface BlockImporterAndStateVerifier {
  fun importBlocks(blocks: List<BlockL1RecoveredData>): SafeFuture<ImportResult>
}

class BlockImporterAndStateVerifierV1(
  private val vertx: Vertx,
  private val elClient: ExecutionLayerClient,
  private val stateManagerClient: StateManagerClientV1,
  private val stateManagerImportTimeoutPerBlock: Duration
) : BlockImporterAndStateVerifier {
  override fun importBlocks(blocks: List<BlockL1RecoveredData>): SafeFuture<ImportResult> {
    return elClient
      .lineaEngineImportBlocksFromBlob(blocks)
      .thenCombine(
        elClient.getBlockNumberAndHash(blockParameter = BlockParameter.Tag.FINALIZED)
      ) { _, elFinalizedBlock ->
        elClient.lineaEngineForkChoiceUpdated(
          finalizedBlockHash = elFinalizedBlock.hash.toArray(),
          headBlockHash = blocks.last().blockHash
        )
      }.thenCompose {
        getBlockStateRootHash(
          blockNumber = blocks.last().blockNumber,
          timeout = stateManagerImportTimeoutPerBlock.times(blocks.size)
        )
      }
      .thenApply { stateRootHash ->
        ImportResult(
          blockNumber = blocks.last().blockNumber,
          zkStateRootHash = stateRootHash
        )
      }
  }

  private fun getBlockStateRootHash(
    blockNumber: ULong,
    timeout: Duration
  ): SafeFuture<ByteArray> {
    return AsyncRetryer
      .retry(
        vertx,
        backoffDelay = 1.seconds,
        timeout = timeout,
        stopRetriesPredicate = { headBlockNumber -> headBlockNumber >= blockNumber },
        action = { stateManagerClient.rollupGetHeadBlockNumber() }
      )
      .thenCompose {
        stateManagerClient.rollupGetStateMerkleProof(BlockInterval(blockNumber, blockNumber))
      }.thenApply { proofResponse -> proofResponse.zkEndStateRootHash.toArray() }
  }
}
