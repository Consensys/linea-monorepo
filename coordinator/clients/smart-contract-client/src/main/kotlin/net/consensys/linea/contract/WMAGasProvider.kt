package net.consensys.linea.contract

import net.consensys.linea.web3j.AtomicContractEIP1559GasProvider
import net.consensys.linea.web3j.EIP1559GasFees
import net.consensys.linea.web3j.EIP4844GasFees
import net.consensys.linea.web3j.EIP4844GasProvider
import net.consensys.zkevm.ethereum.gaspricing.FeesCalculator
import net.consensys.zkevm.ethereum.gaspricing.FeesFetcher
import java.math.BigInteger

class WMAGasProvider(
  private val chainId: Long,
  private val feesFetcher: FeesFetcher,
  private val priorityFeeCalculator: FeesCalculator,
  private val config: Config
) :
  AtomicContractEIP1559GasProvider, EIP4844GasProvider {
  data class Config(
    val gasLimit: BigInteger,
    val maxFeePerGasCap: BigInteger,
    val maxFeePerBlobGasCap: BigInteger,
    var maxFeePerGasCapForEIP4844: BigInteger? = null
  ) {
    init {
      maxFeePerGasCapForEIP4844 = maxFeePerGasCapForEIP4844 ?: maxFeePerGasCap
    }
  }

  private fun getRecentFees(): EIP4844GasFees {
    return feesFetcher.getL1EthGasPriceData().thenApply {
        feesData ->
      val maxPriorityFeePerGas = priorityFeeCalculator.calculateFees(feesData)
      val maxFeePerGas = feesData.baseFeePerGas.last()
        .multiply(BigInteger.TWO).add(maxPriorityFeePerGas)
      val maxFeePerBlobGas = config.maxFeePerBlobGasCap
      EIP4844GasFees(EIP1559GasFees(maxPriorityFeePerGas, maxFeePerGas), maxFeePerBlobGas)
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
    return getRecentFees().eip1559GasFees.maxFeePerGas.min(config.maxFeePerGasCap)
  }

  override fun getMaxPriorityFeePerGas(contractFunc: String?): BigInteger {
    return getRecentFees().eip1559GasFees.maxPriorityFeePerGas.min(config.maxFeePerGasCap)
  }

  override fun getEIP1559GasFees(): EIP1559GasFees {
    return getRecentFees().run {
      EIP1559GasFees(
        maxPriorityFeePerGas = this.eip1559GasFees.maxPriorityFeePerGas.min(config.maxFeePerGasCap),
        maxFeePerGas = this.eip1559GasFees.maxFeePerGas.min(config.maxFeePerGasCap)
      )
    }
  }

  override fun getEIP4844GasFees(): EIP4844GasFees {
    return getRecentFees().run {
      EIP4844GasFees(
        eip1559GasFees = EIP1559GasFees(
          maxPriorityFeePerGas = this.eip1559GasFees.maxPriorityFeePerGas
            .min(config.maxFeePerGasCapForEIP4844),
          maxFeePerGas = this.eip1559GasFees.maxFeePerGas
            .min(config.maxFeePerGasCapForEIP4844)
        ),
        maxFeePerBlobGas = this.maxFeePerBlobGas.min(config.maxFeePerBlobGasCap)
      )
    }
  }
}
