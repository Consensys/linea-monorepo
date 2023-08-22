package net.consensys.zkevm.coordinator.blockcreation

import net.consensys.zkevm.ethereum.coordination.blockcreation.SafeBlockProvider
import org.web3j.protocol.core.DefaultBlockParameterName
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture

class GethCliqueSafeBlockProvider(
  private val extendedWeb3j: ExtendedWeb3J,
  private val config: Config
) : SafeBlockProvider {
  data class Config(
    val blocksToFinalization: Long
  )

  override fun getLatestSafeBlock(): SafeFuture<ExecutionPayloadV1> {
    return SafeFuture.of(
      extendedWeb3j.web3jClient
        .ethGetBlockByNumber(DefaultBlockParameterName.LATEST, false).sendAsync()
    )
      .thenCompose { block ->
        val safeBlockNumber = (block.block.number.toLong() - config.blocksToFinalization).coerceAtLeast(0)
        extendedWeb3j.ethGetExecutionPayloadByNumber(safeBlockNumber)
      }
  }
}
