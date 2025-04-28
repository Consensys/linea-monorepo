package net.consensys.zkevm.ethereum.coordination.aggregation

import kotlinx.datetime.Instant
import linea.contract.l2.L2MessageServiceSmartContractClientReadOnly
import linea.domain.BlockParameter.Companion.toBlockParameter
import linea.ethapi.EthApiClient
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface AggregationL2StateProvider {
  fun getAggregationL2State(blockNumber: Long): SafeFuture<AggregationL2State>
}

class AggregationL2StateProviderImpl(
  private val ethApiClient: EthApiClient,
  private val messageService: L2MessageServiceSmartContractClientReadOnly
) : AggregationL2StateProvider {

  override fun getAggregationL2State(blockNumber: Long): SafeFuture<AggregationL2State> {
    val blockParameter = blockNumber.toBlockParameter()
    return messageService
      .getLastAnchoredL1MessageNumber(block = blockParameter)
      .thenCompose { lastAnchoredMessageNumber ->
        messageService.getRollingHashByL1MessageNumber(
          block = blockParameter,
          l1MessageNumber = lastAnchoredMessageNumber
        )
          .thenApply { rollingHash -> lastAnchoredMessageNumber to rollingHash }
      }
      .thenCombine(
        ethApiClient.getBlockByNumberWithoutTransactionsData(blockParameter)
      ) { (messageNumber, rollingHash), block ->
        if (block == null) {
          throw IllegalStateException("Block $blockNumber not found")
        }

        AggregationL2State(
          parentAggregationLastBlockTimestamp = Instant.fromEpochSeconds(block.timestamp.toLong()),
          parentAggregationLastL1RollingHashMessageNumber = messageNumber,
          parentAggregationLastL1RollingHash = rollingHash
        )
      }
  }
}
