package net.consensys.linea.contract

import net.consensys.linea.Constants.Eip4844BlobSize
import net.consensys.linea.web3j.AtomicContractEIP1559GasProvider
import net.consensys.linea.web3j.EIP4844GasFees
import net.consensys.linea.web3j.EIP4844GasProvider
import net.consensys.linea.web3j.Eip4844Transaction
import net.consensys.linea.web3j.SmartContractErrors
import net.consensys.linea.web3j.informativeEthCall
import net.consensys.toGWei
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCaps
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes
import org.web3j.abi.FunctionEncoder
import org.web3j.abi.TypeReference
import org.web3j.abi.datatypes.DynamicBytes
import org.web3j.abi.datatypes.Function
import org.web3j.abi.datatypes.Type
import org.web3j.abi.datatypes.generated.Uint256
import org.web3j.crypto.Blob
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.RemoteFunctionCall
import org.web3j.protocol.core.Response
import org.web3j.protocol.core.methods.request.Transaction
import org.web3j.protocol.core.methods.response.EthSendTransaction
import org.web3j.protocol.core.methods.response.TransactionReceipt
import org.web3j.protocol.exceptions.TransactionException
import org.web3j.tx.gas.ContractEIP1559GasProvider
import org.web3j.tx.gas.ContractGasProvider
import org.web3j.tx.response.TransactionReceiptProcessor
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.*
import java.util.concurrent.CompletableFuture

class LineaRollupAsyncFriendly(
  contractAddress: String,
  web3j: Web3j,
  private val asyncTransactionManager: AsyncFriendlyTransactionManager,
  contractGasProvider: ContractGasProvider,
  private val smartContractErrors: SmartContractErrors
) : LineaRollup(
  contractAddress,
  web3j,
  asyncTransactionManager,
  contractGasProvider
) {
  private val log: Logger = LogManager.getLogger(this::class.java)
  companion object {
    fun load(
      contractAddress: String,
      web3j: Web3j,
      credentials: Credentials,
      contractGasProvider: ContractGasProvider,
      smartContractErrors: SmartContractErrors
    ): LineaRollupAsyncFriendly {
      return LineaRollupAsyncFriendly(
        contractAddress,
        web3j,
        AsyncFriendlyTransactionManager(web3j, credentials),
        contractGasProvider,
        smartContractErrors
      )
    }

    fun load(
      contractAddress: String,
      web3j: Web3j,
      transactionManager: AsyncFriendlyTransactionManager,
      contractGasProvider: ContractGasProvider,
      smartContractErrors: SmartContractErrors
    ): LineaRollupAsyncFriendly {
      return LineaRollupAsyncFriendly(
        contractAddress,
        web3j,
        transactionManager,
        contractGasProvider,
        smartContractErrors
      )
    }

    fun load(
      contractAddress: String,
      web3j: Web3j,
      credentials: Credentials,
      transactionReceiptProcessor: TransactionReceiptProcessor,
      contractGasProvider: ContractGasProvider,
      smartContractErrors: SmartContractErrors
    ): LineaRollupAsyncFriendly {
      return LineaRollupAsyncFriendly(
        contractAddress,
        web3j,
        AsyncFriendlyTransactionManager(web3j, credentials, transactionReceiptProcessor),
        contractGasProvider,
        smartContractErrors
      )
    }
  }

  private fun buildSubmitDataFunction(
    submissionData: SubmissionData
  ): Function {
    return Function(
      FUNC_SUBMITDATA,
      listOf<Type<*>>(submissionData),
      emptyList<TypeReference<*>>()
    )
  }

  private fun buildSubmitBlobDataFunction(
    supportingSubmissionData: SupportingSubmissionData,
    dataEvaluationClaim: BigInteger,
    kzgCommitment: ByteArray,
    kzgProof: ByteArray
  ): Function {
    val function = Function(
      FUNC_SUBMITBLOBDATA,
      Arrays.asList<Type<*>>(
        supportingSubmissionData,
        Uint256(dataEvaluationClaim),
        DynamicBytes(kzgCommitment),
        DynamicBytes(kzgProof)
      ),
      emptyList<TypeReference<*>>()
    )
    return function
  }

  /**
   * Submits blob data as callData and returns transaction hash
   */
  fun submitDataAndForget(
    submissionData: SubmissionData
  ): String {
    val function = buildSubmitDataFunction(submissionData)
    val result = sendTransaction(function, BigInteger.ZERO)
    throwExceptionIfJsonRpcErrorReturned("eth_sendRawTransaction", result)
    return result.transactionHash
  }

  /**
   * executes eth_call for SubmitData function
   */
  fun submitDataEthCall(
    submissionData: SubmissionData
  ): SafeFuture<String?> {
    val function = buildSubmitDataFunction(submissionData)
    return this.executeEthCall(FunctionEncoder.encode(function))
  }

  /**
   * Sends an EIP4844 blob-carrying transaction to Linea contract
   * It returns the transaction hash
   */
  fun sendBlobData(
    supportingSubmissionData: SupportingSubmissionData,
    dataEvaluationClaim: BigInteger,
    kzgCommitment: ByteArray,
    kzgProof: ByteArray,
    blob: ByteArray,
    gasPriceCaps: GasPriceCaps? = null
  ): SafeFuture<String> {
    require(blob.size == Eip4844BlobSize) { "Blob must have $Eip4844BlobSize bytes, size=${blob.size}bytes" }
    val function = buildSubmitBlobDataFunction(
      supportingSubmissionData,
      dataEvaluationClaim,
      kzgCommitment,
      kzgProof
    )

    return SafeFuture
      .of(sendBlobCarryingTransaction(function, BigInteger.ZERO, listOf(Blob(blob)), gasPriceCaps))
      .thenApply { result ->
        throwExceptionIfJsonRpcErrorReturned("eth_sendRawTransaction", result)
        result.transactionHash
      }
  }

  fun submitBlobDataEthCall(
    supportingSubmissionData: SupportingSubmissionData,
    dataEvaluationClaim: BigInteger,
    kzgCommitment: ByteArray,
    kzgProof: ByteArray,
    blob: ByteArray,
    blobVersionedHashes: List<Bytes>
  ): SafeFuture<String?> {
    val function = buildSubmitBlobDataFunction(
      supportingSubmissionData,
      dataEvaluationClaim,
      kzgCommitment,
      kzgProof
    )
    val gasLimit = gasProvider.getGasLimit(function.name)
    val (_, maxFeePerBlobGas) = getEip4844GasFees()
    val transaction = Eip4844Transaction.createFunctionCallTransaction(
      from = asyncTransactionManager.fromAddress,
      to = contractAddress,
      data = FunctionEncoder.encode(function),
      blobs = listOf(Blob(blob)),
      maxFeePerBlobGas = maxFeePerBlobGas,
      gasLimit = gasLimit,
      blobVersionedHashes = blobVersionedHashes
    )

    return web3j.informativeEthCall(transaction, smartContractErrors)
  }

  private fun buildFinalizeCompressedBlocksWithProofFunction(
    aggregatedProof: ByteArray,
    proofType: BigInteger,
    finalizationData: FinalizationData
  ): Function {
    val function = Function(
      FUNC_FINALIZECOMPRESSEDBLOCKSWITHPROOF,
      Arrays.asList<Type<*>>(
        DynamicBytes(aggregatedProof),
        Uint256(proofType),
        finalizationData
      ),
      emptyList<TypeReference<*>>()
    )
    return function
  }

  fun finalizeAggregation(
    aggregatedProof: ByteArray,
    proofType: BigInteger,
    finalizationData: FinalizationData,
    gasPriceCaps: GasPriceCaps? = null
  ): SafeFuture<TransactionReceipt> {
    val function = buildFinalizeCompressedBlocksWithProofFunction(
      aggregatedProof,
      proofType,
      finalizationData
    )

    return SafeFuture.of(
      sendTransactionAsync(function, BigInteger.ZERO, gasPriceCaps)
    ).thenCompose { result ->
      throwExceptionIfJsonRpcErrorReturned("eth_sendRawTransaction", result)
      asyncTransactionManager.waitForTransaction(
        function,
        FunctionEncoder.encode(function),
        BigInteger.ZERO,
        result
      ).sendAsync()
    }
  }

  data class EIP1559GasFees(
    val maxPriorityFeePerGas: BigInteger,
    val maxFeePerGas: BigInteger
  )

  private fun getEip1559GasFees(
    eip1559GasProvider: ContractEIP1559GasProvider,
    functionName: String
  ): EIP1559GasFees {
    return if (gasProvider is AtomicContractEIP1559GasProvider) {
      val gasFees = (eip1559GasProvider as AtomicContractEIP1559GasProvider).getEIP1559GasFees()
      EIP1559GasFees(gasFees.maxPriorityFeePerGas, gasFees.maxFeePerGas)
    } else {
      EIP1559GasFees(
        maxPriorityFeePerGas = eip1559GasProvider.getMaxPriorityFeePerGas(functionName),
        maxFeePerGas = eip1559GasProvider.getMaxFeePerGas(functionName)
      )
    }
  }

  private fun getEip4844GasFees(): EIP4844GasFees {
    if (gasProvider !is EIP4844GasProvider) {
      throw UnsupportedOperationException("GasProvider does not support EIP4844!")
    }
    return (gasProvider as EIP4844GasProvider).getEIP4844GasFees()
  }

  private fun <T> throwExceptionIfJsonRpcErrorReturned(rpcMethod: String, response: Response<T>) {
    if (response.hasError()) {
      val rpcError = response.error
      var errorMessage =
        "$rpcMethod failed with JsonRpcError: code=${rpcError.code}, message=${rpcError.message}"
      if (rpcError.data != null) {
        errorMessage += ", data=${rpcError.data}"
      }

      throw TransactionException(errorMessage)
    }
  }

  @Synchronized
  private fun sendTransaction(
    function: Function,
    weiValue: BigInteger
  ): EthSendTransaction {
    val encodedData = FunctionEncoder.encode(function)
    val sendRawTransactionResult: EthSendTransaction =
      if (gasProvider is ContractEIP1559GasProvider &&
        (gasProvider as ContractEIP1559GasProvider).isEIP1559Enabled
      ) {
        val eip1559GasProvider = gasProvider as ContractEIP1559GasProvider
        val (maxPriorityFeePerGas, maxFeePerGas) = getEip1559GasFees(eip1559GasProvider, function.name)
        transactionManager.sendEIP1559Transaction(
          eip1559GasProvider.chainId,
          maxPriorityFeePerGas,
          maxFeePerGas,
          eip1559GasProvider.getGasLimit(function.name),
          contractAddress,
          encodedData,
          weiValue,
          false
        )
      } else {
        transactionManager.sendTransaction(
          gasProvider.getGasPrice(function.name),
          gasProvider.getGasLimit(function.name),
          contractAddress,
          encodedData,
          weiValue,
          false
        )
      }

    return sendRawTransactionResult
  }

  @Synchronized
  private fun sendTransactionAsync(
    function: Function,
    weiValue: BigInteger,
    gasPriceCaps: GasPriceCaps? = null
  ): CompletableFuture<EthSendTransaction> {
    val transaction = if (gasProvider is ContractEIP1559GasProvider &&
      (gasProvider as ContractEIP1559GasProvider).isEIP1559Enabled
    ) {
      val eip1559GasProvider = gasProvider as ContractEIP1559GasProvider
      val (maxPriorityFeePerGas, maxFeePerGas) = getEip1559GasFees(eip1559GasProvider, function.name)
      val cappedMaxPriorityFeePerGas = maxPriorityFeePerGas.coerceAtMost(
        gasPriceCaps?.maxPriorityFeePerGasCap ?: maxPriorityFeePerGas
      )
      val cappedMaxFeePerGas = maxFeePerGas.coerceAtMost(
        gasPriceCaps?.maxFeePerGasCap ?: maxFeePerGas
      )

      logGasPriceCapsInfo(
        maxPriorityFeePerGas = maxPriorityFeePerGas,
        maxFeePerGas = maxFeePerGas,
        cappedMaxPriorityFeePerGas = cappedMaxPriorityFeePerGas,
        cappedMaxFeePerGas = cappedMaxFeePerGas,
        gasPriceCaps = gasPriceCaps
      )

      asyncTransactionManager.createRawTransaction(
        chainId = eip1559GasProvider.chainId,
        maxPriorityFeePerGas = cappedMaxPriorityFeePerGas,
        maxFeePerGas = cappedMaxFeePerGas,
        gasLimit = eip1559GasProvider.getGasLimit(function.name),
        to = contractAddress,
        value = weiValue,
        data = FunctionEncoder.encode(function)
      )
    } else {
      asyncTransactionManager.createRawTransaction(
        gasPrice = gasProvider.getGasPrice(function.name),
        gasLimit = gasProvider.getGasLimit(function.name),
        to = contractAddress,
        value = weiValue,
        data = FunctionEncoder.encode(function)
      )
    }
    val signedMessage = asyncTransactionManager.sign(transaction)
    return web3j.ethSendRawTransaction(signedMessage).sendAsync()
  }

  @Synchronized
  private fun sendBlobCarryingTransaction(
    function: Function,
    weiValue: BigInteger,
    blobs: List<Blob>,
    gasPriceCaps: GasPriceCaps? = null
  ): CompletableFuture<EthSendTransaction> {
    val (eip1559fees, maxFeePerBlobGas) = getEip4844GasFees()
    val eip4844GasProvider = gasProvider as EIP4844GasProvider
    val cappedMaxPriorityFeePerGas = eip1559fees.maxPriorityFeePerGas.run {
      this.coerceAtMost(gasPriceCaps?.maxPriorityFeePerGasCap ?: this)
    }
    val cappedMaxFeePerGas = eip1559fees.maxFeePerGas.run {
      this.coerceAtMost(gasPriceCaps?.maxFeePerGasCap ?: this)
    }
    val cappedMaxFeePerBlobGas = maxFeePerBlobGas.run {
      this.coerceAtMost(gasPriceCaps?.maxFeePerBlobGasCap ?: this)
    }

    logGasPriceCapsInfo(
      maxPriorityFeePerGas = eip1559fees.maxPriorityFeePerGas,
      maxFeePerGas = eip1559fees.maxFeePerGas,
      maxFeePerBlobGas = maxFeePerBlobGas,
      cappedMaxPriorityFeePerGas = cappedMaxPriorityFeePerGas,
      cappedMaxFeePerGas = cappedMaxFeePerGas,
      cappedMaxFeePerBlobGas = cappedMaxFeePerBlobGas,
      gasPriceCaps = gasPriceCaps
    )

    val transaction = asyncTransactionManager.createRawTransaction(
      blobs = blobs,
      chainId = eip4844GasProvider.chainId,
      maxPriorityFeePerGas = cappedMaxPriorityFeePerGas,
      maxFeePerGas = cappedMaxFeePerGas,
      gasLimit = eip4844GasProvider.getGasLimit(function.name),
      to = contractAddress,
      data = FunctionEncoder.encode(function),
      value = weiValue,
      maxFeePerBlobGas = cappedMaxFeePerBlobGas
    )
    val signedMessage = asyncTransactionManager.sign(transaction)
    return web3j.ethSendRawTransaction(signedMessage).sendAsync()
  }

  @Synchronized
  override fun executeRemoteCallTransaction(
    function: Function,
    weiValue: BigInteger
  ): RemoteFunctionCall<TransactionReceipt> {
    val encodedData = FunctionEncoder.encode(function)
    val transactionSent = sendTransaction(function, weiValue)
    return asyncTransactionManager.waitForTransaction(function, encodedData, weiValue, transactionSent)
  }

  @Synchronized
  override fun executeRemoteCallTransaction(
    function: Function
  ): RemoteFunctionCall<TransactionReceipt> {
    return executeRemoteCallTransaction(function, BigInteger.ZERO)
  }

  fun executeEthCall(calldata: String): SafeFuture<String?> {
    val gasLimit = gasProvider.getGasLimit()

    val tx = Transaction.createFunctionCallTransaction(
      asyncTransactionManager.fromAddress,
      null,
      null,
      gasLimit,
      contractAddress,
      calldata
    )

    return web3j.informativeEthCall(tx, smartContractErrors)
  }

  fun resetNonce(blockNumber: BigInteger? = null): SafeFuture<Unit> {
    return asyncTransactionManager.resetNonce(blockNumber)
  }

  fun currentNonce(): BigInteger {
    return asyncTransactionManager.currentNonce()
  }

  private fun getCurrentBlock(): SafeFuture<BigInteger> {
    return SafeFuture.of(
      web3j.ethBlockNumber().sendAsync().thenApply {
        it.blockNumber
      }
    )
  }

  fun updateNonceAndReferenceBlockToLastL1Block(): SafeFuture<Unit> {
    return getCurrentBlock().thenCompose { blockNumber ->
      setDefaultBlockParameter(DefaultBlockParameter.valueOf(blockNumber))
      resetNonce(blockNumber)
    }
  }

  private fun logGasPriceCapsInfo(
    maxPriorityFeePerGas: BigInteger,
    maxFeePerGas: BigInteger,
    maxFeePerBlobGas: BigInteger? = null,
    cappedMaxPriorityFeePerGas: BigInteger,
    cappedMaxFeePerGas: BigInteger,
    cappedMaxFeePerBlobGas: BigInteger? = null,
    gasPriceCaps: GasPriceCaps? = null
  ) {
    log.info(
      "Gas price caps: maxPriorityFeePerGas=${maxPriorityFeePerGas.toGWei()} GWei, " +
        "maxPriorityFeePerGasCap=${gasPriceCaps?.maxPriorityFeePerGasCap?.toGWei()} GWei, " +
        "cappedMaxPriorityFeePerGas=${cappedMaxPriorityFeePerGas.toGWei()} GWei, " +
        "priorityFeeCapped=${maxPriorityFeePerGas > cappedMaxPriorityFeePerGas} " +
        "maxFeePerGas=${maxFeePerGas.toGWei()} GWei, " +
        "maxFeePerGasCap=${gasPriceCaps?.maxFeePerGasCap?.toGWei()} GWei, " +
        "cappedMaxFeePerGas=${cappedMaxFeePerGas.toGWei()} GWei, " +
        "feeCapped=${maxFeePerGas > cappedMaxFeePerGas} " +
        if (maxFeePerBlobGas != null && cappedMaxFeePerBlobGas != null) {
          "maxFeePerBlobGas=${maxFeePerBlobGas.toGWei()} GWei, " +
            "maxFeePerBlobGasCap=${gasPriceCaps?.maxFeePerBlobGasCap?.toGWei()} GWei, " +
            "cappedMaxFeePerBlobGas=${cappedMaxFeePerBlobGas.toGWei()} GWei, " +
            "blobFeeCapped=${maxFeePerBlobGas > cappedMaxFeePerBlobGas}"
        } else {
          ""
        }
    )
  }
}
