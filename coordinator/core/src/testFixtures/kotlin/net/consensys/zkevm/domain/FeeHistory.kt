package net.consensys.zkevm.domain

import net.consensys.linea.FeeHistory
import net.consensys.toBigInteger
import java.math.BigInteger

fun createFeeHistory(
  oldestBlockNumber: ULong,
  initialReward: ULong,
  initialBaseFeePerGas: ULong,
  initialGasUsedRatio: UInt,
  initialBaseFeePerBlobGas: ULong,
  initialBlobGasUsedRatio: UInt,
  feeHistoryBlockCount: UInt,
  rewardPercentilesCount: Int
): FeeHistory {
  return FeeHistory(
    oldestBlock = BigInteger.valueOf(oldestBlockNumber.toLong()),
    baseFeePerGas = (initialBaseFeePerGas until initialBaseFeePerGas + feeHistoryBlockCount + 1u)
      .map { it.toBigInteger() },
    reward = (initialReward until initialReward + feeHistoryBlockCount)
      .map { reward -> (1..rewardPercentilesCount).map { reward.times(it.toUInt()).toBigInteger() } },
    gasUsedRatio = (initialGasUsedRatio until initialGasUsedRatio + feeHistoryBlockCount)
      .map { (it.toDouble() / 100.0).toBigDecimal() },
    baseFeePerBlobGas = (initialBaseFeePerBlobGas until initialBaseFeePerBlobGas + feeHistoryBlockCount + 1u)
      .map { it.toBigInteger() },
    blobGasUsedRatio = (initialBlobGasUsedRatio until initialBlobGasUsedRatio + feeHistoryBlockCount)
      .map { (it.toDouble() / 100.0).toBigDecimal() }
  )
}
