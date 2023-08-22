package net.consensys.linea.contract

import org.web3j.abi.FunctionEncoder
import org.web3j.abi.datatypes.Function
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.RemoteFunctionCall
import org.web3j.protocol.core.methods.request.Transaction
import org.web3j.protocol.core.methods.response.EthSendTransaction
import org.web3j.protocol.core.methods.response.TransactionReceipt
import org.web3j.protocol.exceptions.JsonRpcError
import org.web3j.protocol.exceptions.TransactionException
import org.web3j.tx.TransactionManager
import org.web3j.tx.exceptions.ContractCallException
import org.web3j.tx.gas.ContractEIP1559GasProvider
import org.web3j.tx.gas.ContractGasProvider
import org.web3j.tx.response.TransactionReceiptProcessor
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

class ZkEvmV2AsyncFriendly(
  contractAddress: String,
  web3j: Web3j,
  private val asyncTransactionManager: AsyncFriendlyTransactionManager,
  contractGasProvider: ContractGasProvider
) : ZkEvmV2(
  contractAddress,
  web3j,
  asyncTransactionManager,
  contractGasProvider
) {
  companion object {
    fun load(
      contractAddress: String,
      web3j: Web3j,
      credentials: Credentials,
      contractGasProvider: ContractGasProvider
    ): ZkEvmV2AsyncFriendly {
      return ZkEvmV2AsyncFriendly(
        contractAddress,
        web3j,
        AsyncFriendlyTransactionManager(web3j, credentials),
        contractGasProvider
      )
    }

    fun load(
      contractAddress: String,
      web3j: Web3j,
      transactionManager: AsyncFriendlyTransactionManager,
      contractGasProvider: ContractGasProvider
    ): ZkEvmV2AsyncFriendly {
      return ZkEvmV2AsyncFriendly(contractAddress, web3j, transactionManager, contractGasProvider)
    }

    fun load(
      contractAddress: String,
      web3j: Web3j,
      credentials: Credentials,
      transactionReceiptProcessor: TransactionReceiptProcessor,
      contractGasProvider: ContractGasProvider
    ): ZkEvmV2AsyncFriendly {
      return ZkEvmV2AsyncFriendly(
        contractAddress,
        web3j,
        AsyncFriendlyTransactionManager(web3j, credentials, transactionReceiptProcessor),
        contractGasProvider
      )
    }
  }

  @Synchronized
  override fun executeRemoteCallTransaction(
    function: Function,
    weiValue: BigInteger
  ): RemoteFunctionCall<TransactionReceipt> {
    val encodedData = FunctionEncoder.encode(function)

    var transactionSent: EthSendTransaction? = null
    try {
      if (gasProvider is ContractEIP1559GasProvider) {
        val eip1559GasProvider = gasProvider as ContractEIP1559GasProvider
        if (eip1559GasProvider.isEIP1559Enabled) {
          transactionSent = transactionManager.sendEIP1559Transaction(
            eip1559GasProvider.chainId,
            eip1559GasProvider.getMaxPriorityFeePerGas(function.name),
            eip1559GasProvider.getMaxFeePerGas(function.name),
            eip1559GasProvider.getGasLimit(function.name),
            contractAddress,
            encodedData,
            weiValue,
            false
          )
        }
      }
      if (transactionSent == null) {
        transactionSent = transactionManager.sendTransaction(
          gasProvider.getGasPrice(function.name),
          gasProvider.getGasLimit(function.name),
          contractAddress,
          encodedData,
          weiValue,
          false
        )
      }
    } catch (error: JsonRpcError) {
      if (error.data != null) {
        throw TransactionException(error.data.toString())
      } else {
        throw TransactionException(
          String.format(
            "JsonRpcError thrown with code %d. Message: %s",
            error.code,
            error.message
          )
        )
      }
    }

    return asyncTransactionManager.waitForTransaction(function, encodedData, BigInteger.ZERO, transactionSent!!)
  }

  @Synchronized
  override fun executeRemoteCallTransaction(
    function: Function
  ): RemoteFunctionCall<TransactionReceipt> {
    return executeRemoteCallTransaction(function, BigInteger.ZERO)
  }

  fun executeEthCall(calldata: String): SafeFuture<String?> {
    val gasLimit = gasProvider.getGasLimit()
    val ethCallFuture = web3j.ethCall(
      Transaction.createFunctionCallTransaction(
        asyncTransactionManager.fromAddress,
        null,
        null,
        gasLimit,
        contractAddress,
        calldata
      ),
      defaultBlockParameter
    )
      .sendAsync()
    return SafeFuture.of(ethCallFuture).thenApply { ethCall ->
      if (ethCall.isReverted) {
        val exceptionMessage =
          String.format(TransactionManager.REVERT_ERR_STR, ethCall.revertReason) + " Data : ${ethCall.error?.data}"
        throw ContractCallException(exceptionMessage)
      } else {
        ethCall.value
      }
    }
  }

  fun resetNonce(): SafeFuture<Unit> {
    return asyncTransactionManager.resetNonce()
  }

  fun currentNonce(): BigInteger {
    return asyncTransactionManager.currentNonce()
  }
}
