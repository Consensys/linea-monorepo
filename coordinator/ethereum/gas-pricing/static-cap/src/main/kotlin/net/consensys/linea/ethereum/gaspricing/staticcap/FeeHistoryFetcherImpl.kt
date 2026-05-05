package net.consensys.linea.ethereum.gaspricing.staticcap

import linea.domain.BlockParameter
import linea.domain.FeeHistory
import linea.ethapi.EthApiClient
import linea.kotlin.toIntervalString
import net.consensys.linea.ethereum.gaspricing.FeesFetcher
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.atomic.AtomicReference

class FeeHistoryFetcherImpl(
  private val ethApiClient: EthApiClient,
  private val config: Config,
) : FeesFetcher {
  private val log: Logger = LogManager.getLogger(this::class.java)

  data class Config(
    val feeHistoryBlockCount: UInt,
    val feeHistoryRewardPercentile: Double,
  ) {
    init {
      require(feeHistoryBlockCount > 0u) {
        "feeHistoryBlockCount=$feeHistoryBlockCount must be greater than 0."
      }
      require(feeHistoryRewardPercentile in 0.0..100.0) {
        "feeHistoryRewardPercentile must be within 0.0 and 100.0." +
          " Value=$feeHistoryRewardPercentile"
      }
    }
  }

  private val feeHistoryCache: AtomicReference<Pair<Long, FeeHistory>?> = AtomicReference(null)

  private fun getRecentFees(): SafeFuture<FeeHistory> {
    return ethApiClient.ethBlockNumber()
      .thenCompose { blockNumberResponse ->
        val currentBlockNumber = blockNumberResponse.toLong()
        val cached = feeHistoryCache.get()
        if (cached == null || currentBlockNumber > cached.first) {
          ethApiClient
            .ethFeeHistory(
              blockCount = config.feeHistoryBlockCount.toInt(),
              newestBlock = BlockParameter.Tag.LATEST,
              rewardPercentiles = listOf(config.feeHistoryRewardPercentile),
            )
            .thenApply { feeHistory ->
              log.trace(
                "New Fee History: l1BlockNumber={} l1Blocks={} lastBaseFeePerGas={} reward={} gasUsedRatio={}{}",
                currentBlockNumber,
                feeHistory.blocksRange().toIntervalString(),
                feeHistory.baseFeePerGas[feeHistory.baseFeePerGas.lastIndex],
                feeHistory.reward.map { percentiles -> percentiles[0] },
                feeHistory.gasUsedRatio,
                if (feeHistory.baseFeePerBlobGas.isNotEmpty()) {
                  " baseFeePerBlobGas=${feeHistory.baseFeePerBlobGas[feeHistory.baseFeePerBlobGas.lastIndex]}" +
                    " blobGasUsedRatio=${feeHistory.blobGasUsedRatio}"
                } else {
                  ""
                },
              )
              feeHistoryCache.set(currentBlockNumber to feeHistory)
              feeHistory
            }
        } else {
          SafeFuture.completedFuture(cached.second)
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
