package net.consensys.zkevm.ethereum.finalization

import net.consensys.linea.contract.LineaRollup
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.ethereum.error.handling.ErrorHandling.handleError
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapProvider
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.abi.FunctionEncoder
import org.web3j.abi.TypeReference
import org.web3j.abi.datatypes.DynamicBytes
import org.web3j.abi.datatypes.Function
import org.web3j.abi.datatypes.Type
import org.web3j.abi.datatypes.generated.Uint256
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

class AggregationFinalizationAsCallData(
  private val contract: LineaRollupAsyncFriendly,
  private val gasPriceCapProvider: GasPriceCapProvider
) : AggregationFinalization {
  private val log: Logger = LogManager.getLogger(this::class.java)

  override fun finalizeAggregation(
    aggregationProof: ProofToFinalize
  ): SafeFuture<*> {
    return gasPriceCapProvider.getGasPriceCaps(aggregationProof.firstBlockNumber)
      .thenCompose { gasPriceCaps ->
        val finalizationFuture = contract.finalizeAggregation(
          aggregationProof.aggregatedProof,
          BigInteger.valueOf(aggregationProof.aggregatedVerifierIndex.toLong()),
          createFinalizationData(aggregationProof),
          gasPriceCaps
        )

        log.debug(
          "aggregation proof submitted to L1: endBlockNumber={}",
          aggregationProof.finalBlockNumber
        )

        finalizationFuture
          .thenApply { transactionReceipt ->
            log.info(
              "aggregation proof accepted: endBlockNumber={} tx_hash={}",
              aggregationProof.finalBlockNumber,
              transactionReceipt.transactionHash
            )
            transactionReceipt
          }.exceptionallyCompose { th ->
            handleError(
              messagePrefix =
              "aggregation proof submission failed: endBlockNumber=${aggregationProof.finalBlockNumber} ",
              error = th
            ) {
              finalizeAggregationEthCall(aggregationProof)
            }
          }
      }
  }

  override fun finalizeAggregationEthCall(aggregationProof: ProofToFinalize): SafeFuture<*> {
    val finalizationData = createFinalizationData(aggregationProof)
    val function = Function(
      LineaRollup.FUNC_FINALIZECOMPRESSEDBLOCKSWITHPROOF,
      listOf<Type<*>>(
        DynamicBytes(aggregationProof.aggregatedProof),
        Uint256(aggregationProof.aggregatedVerifierIndex.toLong()),
        finalizationData
      ),
      emptyList<TypeReference<*>>()
    )
    val calldata = FunctionEncoder.encode(function)
    return contract.executeEthCall(calldata)
  }

  private fun createFinalizationData(aggregationProof: ProofToFinalize) =
    LineaRollup.FinalizationData(
      aggregationProof.parentStateRootHash,
      aggregationProof.dataHashes,
      aggregationProof.dataParentHash,
      BigInteger.valueOf(aggregationProof.finalBlockNumber),
      BigInteger.valueOf(aggregationProof.parentAggregationLastBlockTimestamp.epochSeconds),
      BigInteger.valueOf(aggregationProof.finalTimestamp.epochSeconds),
      aggregationProof.l1RollingHash,
      BigInteger.valueOf(aggregationProof.l1RollingHashMessageNumber),
      aggregationProof.l2MerkleRoots,
      BigInteger.valueOf(aggregationProof.l2MerkleTreesDepth.toLong()),
      aggregationProof.l2MessagingBlocksOffsets
    )
}
