package net.consensys.linea.web3j

import org.web3j.tx.gas.ContractEIP1559GasProvider
import java.math.BigInteger

data class EIP1559GasFees(
  val maxPriorityFeePerGas: BigInteger,
  val maxFeePerGas: BigInteger
)

interface AtomicContractEIP1559GasProvider : ContractEIP1559GasProvider {
  fun getEIP1559GasFees(): EIP1559GasFees
}
