package net.consensys.zkevm.ethereum

import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.EIP1559GasProvider
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.linea.contract.l2.L2MessageServiceGasLimitEstimate
import net.consensys.linea.web3j.SmartContractErrors
import org.slf4j.LoggerFactory
import org.web3j.protocol.Web3j
import org.web3j.tx.gas.ContractEIP1559GasProvider
import org.web3j.tx.gas.StaticEIP1559GasProvider
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.TimeUnit

data class LineaRollupDeploymentResult(
  val contractAddress: String,
  val contractDeploymentBlockNumber: ULong,
  val rollupOperators: List<AccountTransactionManager>,
  val rollupOperatorClient: LineaRollupAsyncFriendly
) {
  val rollupOperator: AccountTransactionManager
    get() = rollupOperators.first()
}

data class L2MessageServiceDeploymentResult(
  val contractAddress: String,
  val contractDeploymentBlockNumber: ULong,
  val anchorerOperator: AccountTransactionManager
)

data class ContactsDeploymentResult(
  val lineaRollup: LineaRollupDeploymentResult,
  val l2MessageService: L2MessageServiceDeploymentResult
)

interface ContractsManager {
  /**
   * Deploys a linea rollup contract with specified number of operators.
   * Operator accounts are generated on the fly and funded from whale account in genesis file.
   */
  fun deployLineaRollup(
    numberOfOperators: Int = 1
  ): SafeFuture<LineaRollupDeploymentResult>

  fun deployL2MessageService(): SafeFuture<L2MessageServiceDeploymentResult>

  fun deployRollupAndL2MessageService(
    dataCompressionAndProofAggregationMigrationBlock: ULong = 1000UL,
    numberOfOperators: Int = 1
  ): SafeFuture<ContactsDeploymentResult>

  fun connectToLineaRollupContract(
    contractAddress: String,
    transactionManager: AsyncFriendlyTransactionManager,
    gasProvider: ContractEIP1559GasProvider = StaticEIP1559GasProvider(
      L1AccountManager.chainId,
      /*maxFeePerGas*/BigInteger.valueOf(11_000L),
      /*maxPriorityFeePerGas*/ BigInteger.valueOf(10_000L),
      /*gasLimit*/BigInteger.valueOf(1_000_000L)
    )
  ): LineaRollupAsyncFriendly

  fun connectToLineaRollupContract(
    contractAddress: String,
    transactionManager: AsyncFriendlyTransactionManager,
    maxFeePerGas: ULong = 11_000UL,
    maxPriorityFeePerGas: ULong = 10_000UL,
    gasLimit: ULong = 1_000_000UL
  ): LineaRollupAsyncFriendly

  fun connectL2MessageService(
    contractAddress: String,
    web3jClient: Web3j = L2AccountManager.web3jClient,
    transactionManager: AsyncFriendlyTransactionManager,
    gasProvider: EIP1559GasProvider = EIP1559GasProvider(
      web3jClient,
      EIP1559GasProvider.Config(
        gasLimit = 1_000_000.toBigInteger(),
        maxFeePerGasCap = BigInteger.valueOf(10_000),
        feeHistoryBlockCount = 5u,
        feeHistoryRewardPercentile = 0.15
      )
    ),
    smartContractErrors: SmartContractErrors = emptyMap()
  ): L2MessageServiceGasLimitEstimate

  companion object {
    // TODO: think of better get the Instance
    fun get(): ContractsManager = MakeFileDelegatedContractsManger
  }
}

object MakeFileDelegatedContractsManger : ContractsManager {
  val log = LoggerFactory.getLogger(MakeFileDelegatedContractsManger::class.java)
  override fun deployLineaRollup(
    numberOfOperators: Int
  ): SafeFuture<LineaRollupDeploymentResult> {
    val newAccounts = L1AccountManager.generateAccounts(numberOfOperators)
    val contractDeploymentAccount = newAccounts.first()
    val operatorsAccounts = newAccounts.drop(1)
    log.debug(
      "going deploy LineaRollup: deployerAccount={} rollupOperators={}",
      contractDeploymentAccount.address,
      operatorsAccounts.map { it.address }
    )
    val future = makeDeployLineaRollup(
      deploymentPrivateKey = contractDeploymentAccount.privateKey,
      operatorsAddresses = operatorsAccounts.map { it.address }
    )
      .thenApply { deploymentResult ->
        log.debug(
          "LineaRollup deployed: address={} blockNumber={} deployerAccount={} " +
            "rollupOperators={}",
          deploymentResult.address,
          deploymentResult.blockNumber,
          contractDeploymentAccount.address,
          operatorsAccounts.map { it.address }
        )
        val accountsTxManagers = operatorsAccounts.map {
          AccountTransactionManager(it, L1AccountManager.getTransactionManager(it))
        }
        val lineaRollupContract = connectToLineaRollupContract(
          deploymentResult.address,
          accountsTxManagers.first().txManager
        )
        LineaRollupDeploymentResult(
          contractAddress = deploymentResult.address,
          contractDeploymentBlockNumber = deploymentResult.blockNumber.toULong(),
          rollupOperators = accountsTxManagers,
          rollupOperatorClient = lineaRollupContract
        )
      }
    future.get(60, TimeUnit.SECONDS)
    return future
  }

  override fun deployL2MessageService(): SafeFuture<L2MessageServiceDeploymentResult> {
    val (deployerAccount, anchorerAccount) = L2AccountManager.generateAccounts(2)
    return makeDeployL2MessageService(
      deploymentPrivateKey = deployerAccount.privateKey,
      anchorOperatorAddresses = anchorerAccount.address
    )
      .thenApply {
        L2MessageServiceDeploymentResult(
          contractAddress = it.address,
          contractDeploymentBlockNumber = it.blockNumber.toULong(),
          anchorerOperator = AccountTransactionManager(
            account = anchorerAccount,
            txManager = L2AccountManager.getTransactionManager(anchorerAccount)
          )
        )
      }
  }

  override fun deployRollupAndL2MessageService(
    dataCompressionAndProofAggregationMigrationBlock: ULong,
    numberOfOperators: Int
  ): SafeFuture<ContactsDeploymentResult> {
    return deployLineaRollup(numberOfOperators)
      .thenCombine(deployL2MessageService()) { lineaRollupDeploymentResult, l2MessageServiceDeploymentResult ->
        ContactsDeploymentResult(
          lineaRollup = lineaRollupDeploymentResult,
          l2MessageService = l2MessageServiceDeploymentResult
        )
      }
  }

  override fun connectToLineaRollupContract(
    contractAddress: String,
    transactionManager: AsyncFriendlyTransactionManager,
    gasProvider: ContractEIP1559GasProvider
  ): LineaRollupAsyncFriendly {
    return LineaRollupAsyncFriendly.load(
      contractAddress,
      L1AccountManager.web3jClient,
      transactionManager,
      gasProvider,
      emptyMap()
    )
  }

  override fun connectToLineaRollupContract(
    contractAddress: String,
    transactionManager: AsyncFriendlyTransactionManager,
    maxFeePerGas: ULong,
    maxPriorityFeePerGas: ULong,
    gasLimit: ULong
  ): LineaRollupAsyncFriendly {
    return connectToLineaRollupContract(
      contractAddress,
      transactionManager,
      StaticEIP1559GasProvider(
        L1AccountManager.chainId,
        /*maxFeePerGas*/BigInteger.valueOf(maxFeePerGas.toLong()),
        /*maxPriorityFeePerGas*/ BigInteger.valueOf(maxPriorityFeePerGas.toLong()),
        /*gasLimit*/BigInteger.valueOf(gasLimit.toLong())
      )
    )
  }

  override fun connectL2MessageService(
    contractAddress: String,
    web3jClient: Web3j,
    transactionManager: AsyncFriendlyTransactionManager,
    gasProvider: EIP1559GasProvider,
    smartContractErrors: SmartContractErrors
  ): L2MessageServiceGasLimitEstimate {
    return L2MessageServiceGasLimitEstimate.load(
      contractAddress,
      web3jClient,
      transactionManager,
      gasProvider,
      smartContractErrors
    )
  }
}
