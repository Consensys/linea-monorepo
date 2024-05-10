package net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap

import io.vertx.core.Vertx
import net.consensys.zkevm.PeriodicPollingService
import net.consensys.zkevm.ethereum.gaspricing.FeeHistoriesRepositoryWithCache
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapFeeHistoryFetcher
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.Web3j
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import kotlin.time.Duration

class FeeHistoryCachingService(
  private val config: Config,
  private val vertx: Vertx,
  private val web3jClient: Web3j,
  private val feeHistoryFetcher: GasPriceCapFeeHistoryFetcher,
  private val feeHistoriesRepository: FeeHistoriesRepositoryWithCache,
  private val log: Logger = LogManager.getLogger(FeeHistoryCachingService::class.java)
) : PeriodicPollingService(
  vertx = vertx,
  pollingIntervalMs = config.pollingInterval.inWholeMilliseconds,
  log = log
) {
  data class Config(
    val pollingInterval: Duration,
    val feeHistoryMaxBlockCount: UInt,
    val baseFeePerGasPercentile: Double,
    val feeHistoryWindowInBlocks: UInt,
    val feeHistoryStoragePeriodInBlocks: UInt
  )

  private fun fetchAndSaveFeeHistories(latestL1BlockNumber: BigInteger): SafeFuture<Unit> {
    return feeHistoriesRepository.findHighestBlockNumberWithPercentile(
      config.baseFeePerGasPercentile
    )
      .thenCompose {
        val highestStoredL1BlockNumber = it ?: 0L

        val startBlockNumberInclusive = highestStoredL1BlockNumber.coerceAtLeast(
          latestL1BlockNumber.toLong()
            .minus(config.feeHistoryWindowInBlocks.toLong())
            .coerceAtLeast(0L)
        ).inc()

        val endBlockNumberInclusive = (startBlockNumberInclusive + config.feeHistoryMaxBlockCount.toLong())
          .minus(1L)
          .coerceAtMost(latestL1BlockNumber.toLong())

        if (endBlockNumberInclusive < startBlockNumberInclusive) {
          SafeFuture.completedFuture(Unit)
        } else {
          val fromBlockNumber = (latestL1BlockNumber.toLong() - config.feeHistoryWindowInBlocks.toLong())
            .coerceAtLeast(0L).inc()
          feeHistoryFetcher.getEthFeeHistoryData(
            startBlockNumberInclusive = startBlockNumberInclusive,
            endBlockNumberInclusive = endBlockNumberInclusive
          ).thenCompose { feeHistory ->
            feeHistoriesRepository.saveNewFeeHistory(feeHistory)
          }.thenCompose {
            feeHistoriesRepository.cacheNumOfFeeHistoriesFromBlockNumber(
              rewardPercentile = config.baseFeePerGasPercentile,
              fromBlockNumber = fromBlockNumber
            )
          }.thenCompose {
            if (endBlockNumberInclusive == latestL1BlockNumber.toLong()) {
              feeHistoriesRepository.cachePercentileGasFees(
                config.baseFeePerGasPercentile,
                fromBlockNumber
              )
            } else {
              SafeFuture.completedFuture(Unit)
            }
          }
        }
      }.whenException { th ->
        log.warn(
          "failure occurred when fetch and save fee histories but will be ignored: {}",
          th.message,
          th
        )
      }.alwaysRun {
        feeHistoriesRepository.deleteFeeHistoriesUpToBlockNumber(
          latestL1BlockNumber.minus(
            BigInteger.valueOf(config.feeHistoryStoragePeriodInBlocks.toLong())
          ).coerceAtLeast(BigInteger.ZERO).toLong()
        )
      }
  }

  override fun action(): SafeFuture<Unit> {
    return SafeFuture.of(web3jClient.ethBlockNumber().sendAsync())
      .thenCompose { latestL1BlockNumber ->
        fetchAndSaveFeeHistories(latestL1BlockNumber.blockNumber)
      }.whenException { th ->
        log.error("Failed to fetch and save fee histories", th)
      }
  }

  override fun handleError(error: Throwable) {
    log.error(
      "Error with fee history caching service: errorMessage={}",
      error.message,
      error
    )
  }
}
