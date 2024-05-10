package net.consensys.zkevm.ethereum.coordination.aggregation

import io.vertx.core.Vertx
import kotlinx.datetime.Instant
import net.consensys.linea.async.AsyncRetryer
import net.consensys.zkevm.coordinator.clients.L2MessageServiceClient
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration.Companion.milliseconds

interface AggregationL2StateProvider {
  fun getAggregationL2State(blockNumber: Long): SafeFuture<AggregationL2State>
}

class AggregationL2StateProviderImpl(
  vertx: Vertx,
  private val l2web3jClient: Web3j,
  private val l2MessageServiceClient: L2MessageServiceClient
) : AggregationL2StateProvider {
  private val retryer = AsyncRetryer.retryer<AggregationL2State>(
    vertx,
    backoffDelay = 500.milliseconds,
    maxRetries = 10
  )

  override fun getAggregationL2State(blockNumber: Long): SafeFuture<AggregationL2State> {
    return retryer.retry { getAggregationL2StateInternal(blockNumber) }
  }

  private fun getAggregationL2StateInternal(
    blockNumber: Long
  ): SafeFuture<AggregationL2State> {
    return l2MessageServiceClient.getLastAnchoredMessageUpToBlock(blockNumber)
      .thenCombine(getBlockTimestamp(blockNumber)) { event, timestamp ->
        AggregationL2State(
          parentAggregationLastBlockTimestamp = timestamp,
          parentAggregationLastL1RollingHashMessageNumber = event.messageNumber,
          parentAggregationLastL1RollingHash = event.messageRollingHash
        )
      }
  }

  private fun getBlockTimestamp(blockNumber: Long): SafeFuture<Instant> {
    return SafeFuture.of(
      l2web3jClient.ethGetBlockByNumber(
        DefaultBlockParameter.valueOf(blockNumber.toBigInteger()),
        false
      ).sendAsync().thenApply { block -> Instant.fromEpochSeconds(block.block.timestamp.toLong()) }
    )
  }
}
