package net.consensys.linea.web3j

import net.consensys.linea.FeeHistory
import org.web3j.protocol.core.methods.response.EthFeeHistory
import java.math.BigInteger

fun EthFeeHistory.FeeHistory.toLineaDomain(): FeeHistory {
  return FeeHistory(
    oldestBlock = oldestBlock,
    baseFeePerGas = baseFeePerGas,
    reward = reward,
    gasUsedRatio = gasUsedRatio.map { it.toBigDecimal() }
  )
}
fun EthFeeHistory.FeeHistory.blocksRange(): ClosedRange<BigInteger> {
  return oldestBlock..oldestBlock.add((reward.size - 1).toBigInteger())
}
