package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import org.web3j.tx.gas.ContractEIP1559GasProvider
import java.math.BigInteger

class WMAGasProvider(
  private val chainId: Long,
  private val feesFetcher: FeesFetcher,
  private val minerTipCalculator: FeesCalculator,
  private val config: Config
) :
  ContractEIP1559GasProvider {

  data class Config(
    val gasLimit: BigInteger,
    val maxFeePerGas: BigInteger
  )
  private data class Fees(
    val maxFeePerGas: BigInteger,
    val maxPriorityFeePerGas: BigInteger
  )

  private fun getRecentFees(): Fees {
    return feesFetcher.getL1EthGasPriceData().thenApply {
        feesData ->
      val minerTip = minerTipCalculator.calculateFees(feesData)
      val maxFeePerGas = feesData.baseFeePerGas[feesData.baseFeePerGas.lastIndex - 1]
        .multiply(BigInteger.TWO).add(minerTip).min(config.maxFeePerGas)
      Fees(maxFeePerGas, minerTip)
    }.get()
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
