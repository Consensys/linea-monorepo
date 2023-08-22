package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import net.consensys.linea.FeeHistory
import net.consensys.linea.web3j.toLineaDomain
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameterName
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

class FeeHistoryFetcherImpl(
  private val web3jClient: Web3j,
  private val config: Config
) : FeesFetcher {
  private val log: Logger = LogManager.getLogger(this::class.java)

  data class Config(
    val feeHistoryBlockCount: UInt,
    val feeHistoryRewardPercentile: Double
  ) {
    init {
      require(feeHistoryRewardPercentile in 0.0..100.0) {
        throw IllegalArgumentException(
          "feeHistoryRewardPercentile must be within 0.0 and 100.0." +
            " Value=$feeHistoryRewardPercentile"
        )
      }
    }
  }

  private var cacheIsValidForBlockNumber: BigInteger = BigInteger.ZERO
  private var feesCache: FeeHistory = getRecentFees().get()

  private fun getRecentFees(): SafeFuture<FeeHistory> {
    val blockNumberFuture = web3jClient.ethBlockNumber().sendAsync()
    return SafeFuture.of(blockNumberFuture)
      .thenCompose { blockNumberResponse ->
        val currentBlockNumber = blockNumberResponse.blockNumber
        if (currentBlockNumber > cacheIsValidForBlockNumber) {
          web3jClient
            .ethFeeHistory(
              config.feeHistoryBlockCount.toInt(),
              DefaultBlockParameterName.LATEST,
              listOf(config.feeHistoryRewardPercentile)
            )
            .sendAsync()
            .thenApply { it ->
              val feeHistory = it.feeHistory.toLineaDomain()
              cacheIsValidForBlockNumber = currentBlockNumber
              log.trace(
                "New Fee History: l1BlockNumber={}, lastBaseFeePerGas={} reward={}, gasUsedRatio={}",
                currentBlockNumber,
                feeHistory.baseFeePerGas[feeHistory.baseFeePerGas.lastIndex - 1],
                feeHistory.reward.map { percentiles -> percentiles[0] },
                feeHistory.gasUsedRatio
              )
              feesCache = feeHistory
              feesCache
            }
        } else {
          SafeFuture.completedFuture(feesCache)
        }
      }
  }

  override fun getL1EthGasPriceData(): SafeFuture<FeeHistory> {
    return getRecentFees()
      .whenException { th ->
        log.error("Get L1 gas price data failure: {}", th.message, th)
      }
  }
}
