package net.consensys.zkevm.ethereum

import net.consensys.linea.toIntervalString
import net.consensys.linea.web3j.blocksRange
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameterName
import org.web3j.tx.gas.ContractEIP1559GasProvider
import java.math.BigDecimal
import java.math.BigInteger

class EIP1559GasProvider(private val l2Client: Web3j, private val config: Config) :
  ContractEIP1559GasProvider {
  private val log: Logger = LogManager.getLogger(this::class.java)

  data class Config(
    val gasLimit: BigInteger,
    val maxFeePerGasCap: BigInteger,
    val feeHistoryBlockCount: UInt,
    val feeHistoryRewardPercentile: Double
  )
  private data class Fees(
    val maxFeePerGas: BigInteger,
    val maxPriorityFeePerGas: BigInteger
  )

  private val chainId: Long = l2Client.ethChainId().send().chainId.toLong()
  private var cacheIsValidForBlockNumber: BigInteger = BigInteger.ZERO
  private var feesCache: Fees = getRecentFees()

  private fun getRecentFees(): Fees {
    val currentBlockNumber = l2Client.ethBlockNumber().send().blockNumber
    if (currentBlockNumber > cacheIsValidForBlockNumber) {
      l2Client
        .ethFeeHistory(
          config.feeHistoryBlockCount.toInt(),
          DefaultBlockParameterName.LATEST,
          listOf(config.feeHistoryRewardPercentile)
        )
        .sendAsync()
        .thenApply {
            feeHistoryResponse ->
          var maxPriorityFeePerGas = feeHistoryResponse.feeHistory.reward.map { BigDecimal(it[0]) }
            .reduce { acc, reward ->
              acc.add(reward)
            }.divide(BigDecimal(feeHistoryResponse.feeHistory.reward.size)).toBigInteger()

          if (maxPriorityFeePerGas > config.maxFeePerGasCap) {
            maxPriorityFeePerGas = config.maxFeePerGasCap
            log.warn(
              "Estimated miner tip of $maxPriorityFeePerGas exceeds configured max " +
                "fee per gas of ${config.maxFeePerGasCap} returning cap instead!"
            )
          }

          cacheIsValidForBlockNumber = currentBlockNumber

          val maxFeePerGas = feeHistoryResponse.feeHistory.baseFeePerGas.last()
            .multiply(BigInteger.TWO)
            .add(maxPriorityFeePerGas)

          if (maxFeePerGas > BigInteger.ZERO && maxPriorityFeePerGas > BigInteger.ZERO) {
            feesCache = Fees(
              maxFeePerGas.min(config.maxFeePerGasCap),
              maxPriorityFeePerGas
            )
            log.debug(
              "New fees estimation: fees={}, l2Blocks={}",
              feeHistoryResponse.feeHistory.blocksRange().toIntervalString(),
              feesCache
            )
          } else {
            feesCache = Fees(
              config.maxFeePerGasCap,
              BigInteger.ZERO
            )
          }
        }
        .get()
    }
    return feesCache
  }

  override fun getGasPrice(contractFunc: String?): BigInteger {
    return getGasPrice()
  }

  @Deprecated("Deprecated in Java")
  override fun getGasPrice(): BigInteger {
    throw NotImplementedError("EIP1559GasProvider only implements EIP1559 specific methods")
  }

  override fun getGasLimit(contractFunc: String?): BigInteger {
    return getGasLimit()
  }

  @Deprecated("Deprecated in Java")
  override fun getGasLimit(): BigInteger {
    return config.gasLimit
  }

  override fun isEIP1559Enabled(): Boolean {
    return true
  }

  override fun getChainId(): Long {
    return chainId
  }

  override fun getMaxFeePerGas(contractFunc: String?): BigInteger {
    return getRecentFees().maxFeePerGas
  }

  override fun getMaxPriorityFeePerGas(contractFunc: String?): BigInteger {
    return getRecentFees().maxPriorityFeePerGas
  }
}
