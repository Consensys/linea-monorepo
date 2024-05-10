package net.consensys.zkevm.persistence.feehistory

import net.consensys.linea.FeeHistory
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

interface FeeHistoriesRepository {
  fun saveNewFeeHistory(feeHistory: FeeHistory): SafeFuture<Unit>

  fun findBaseFeePerGasAtPercentile(
    percentile: Double,
    fromBlockNumber: Long
  ): SafeFuture<BigInteger?>

  fun findBaseFeePerBlobGasAtPercentile(
    percentile: Double,
    fromBlockNumber: Long
  ): SafeFuture<BigInteger?>

  fun findAverageRewardAtPercentile(
    rewardPercentile: Double,
    fromBlockNumber: Long
  ): SafeFuture<BigInteger?>

  fun findHighestBlockNumberWithPercentile(
    rewardPercentile: Double
  ): SafeFuture<Long?>

  fun getNumOfFeeHistoriesFromBlockNumber(
    rewardPercentile: Double,
    fromBlockNumber: Long
  ): SafeFuture<Int>

  fun deleteFeeHistoriesUpToBlockNumber(
    blockNumberInclusive: Long
  ): SafeFuture<Int>
}
