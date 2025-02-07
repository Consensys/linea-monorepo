package net.consensys.zkevm.ethereum.coordination.proofcreation

import net.consensys.encodeHex
import net.consensys.linea.async.toSafeFuture
import net.consensys.toBigInteger
import net.consensys.zkevm.coordinator.clients.BatchExecutionProofRequestV1
import net.consensys.zkevm.coordinator.clients.ExecutionProverClientV2
import net.consensys.zkevm.coordinator.clients.L2MessageServiceLogsClient
import net.consensys.zkevm.domain.Batch
import net.consensys.zkevm.domain.BlocksConflation
import net.consensys.zkevm.domain.RlpBridgeLogsData
import net.consensys.zkevm.encoding.BlockEncoder
import net.consensys.zkevm.ethereum.coordination.conflation.BlocksTracesConflated
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import tech.pegasys.teku.infrastructure.async.SafeFuture

class ZkProofCreationCoordinatorImpl(
  private val executionProverClient: ExecutionProverClientV2,
  private val l2MessageServiceLogsClient: L2MessageServiceLogsClient,
  private val l2Web3jClient: Web3j,
  private val encoder: BlockEncoder
) : ZkProofCreationCoordinator {
  private val log: Logger = LogManager.getLogger(this::class.java)

  private fun getBlockStateRootHash(blockNumber: ULong): SafeFuture<String> {
    return l2Web3jClient
      .ethGetBlockByNumber(
        DefaultBlockParameter.valueOf(blockNumber.toBigInteger()),
        false
      )
      .sendAsync()
      .thenApply { block -> block.block.stateRoot }
      .toSafeFuture()
  }

  override fun createZkProof(
    blocksConflation: BlocksConflation,
    traces: BlocksTracesConflated
  ): SafeFuture<Batch> {
    val startBlockNumber = blocksConflation.blocks.first().number.toULong()
    val endBlockNumber = blocksConflation.blocks.last().number.toULong()
    val blocksConflationInterval = blocksConflation.intervalString()

    val bridgeLogsSfListFutures = blocksConflation.blocks.map { block ->
      l2MessageServiceLogsClient.getBridgeLogs(blockNumber = block.number.toLong())
        .thenApply { block to it }
    }

    return getBlockStateRootHash(blocksConflation.startBlockNumber - 1UL)
      .thenCompose { previousKeccakStateRootHash ->
        SafeFuture.collectAll(bridgeLogsSfListFutures.stream())
          .thenCompose { blocksAndBridgeLogs ->
            val blocksData = blocksAndBridgeLogs.map { (block, bridgeLogs) ->
              val rlp = encoder.encode(block).encodeHex()
              RlpBridgeLogsData(rlp, bridgeLogs)
            }
            executionProverClient.requestProof(
              BatchExecutionProofRequestV1(
                blocks = blocksConflation.blocks,
                tracesResponse = traces.tracesResponse,
                type2StateData = traces.zkStateTraces,
                blocksData = blocksData,
                keccakParentStateRootHash = previousKeccakStateRootHash
              )
            ).thenApply {
              Batch(
                startBlockNumber = startBlockNumber,
                endBlockNumber = endBlockNumber
              )
            }.whenException {
              log.error("Prover returned for batch={} errorMessage={}", blocksConflationInterval, it.message, it)
            }
          }
      }
  }
}
