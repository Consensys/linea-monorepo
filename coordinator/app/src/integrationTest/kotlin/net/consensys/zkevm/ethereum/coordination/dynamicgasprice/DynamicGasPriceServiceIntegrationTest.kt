package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.linea.toHexString
import net.consensys.linea.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.MethodOrderer
import org.junit.jupiter.api.Order
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestMethodOrder
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.ArgumentMatchers
import org.mockito.Mockito
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameterName
import org.web3j.protocol.core.methods.response.EthBlockNumber
import org.web3j.protocol.core.methods.response.EthFeeHistory
import org.web3j.protocol.http.HttpService
import org.web3j.tx.gas.DefaultGasProvider
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.net.URL
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
@TestMethodOrder(MethodOrderer.OrderAnnotation::class)
@Disabled("FIXME: These tests are flaky and need to be fixed")
class DynamicGasPriceServiceIntegrationTest {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private val meterRegistry = SimpleMeterRegistry()
  private val pollingInterval = 2.seconds
  private val feeHistoryBlockCount = 10u
  private val feeHistoryRewardPercentile = 15.0
  private val baseFeeCoefficient = 0.1
  private val priorityFeeCoefficient = 0.1
  private val initialReward = 1000000000uL
  private val initialBaseFeePerGas = 10000000000uL
  private val initialGasUsedRatio = 50u
  private val l2GasPriceCap = BigInteger("fffffffffffffffff", 16)
  private val l2ValidatorRpcEndpoint = "http://localhost:8545"
  private val l2NodeRpcEndpoint = "http://localhost:8845"
  private val minerGasPriceUpdateRecipients = listOf(l2ValidatorRpcEndpoint, l2NodeRpcEndpoint)
  private val l1Web3jClient = Web3j.build(HttpService("http://localhost:8445"))

  private val l2ValidatorWeb3jClient = Web3j.build(HttpService(l2ValidatorRpcEndpoint))
  private val l2NodeWeb3jClient = Web3j.build(HttpService(l2NodeRpcEndpoint))

  // WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
  private val l2TestAccountPrivateKey1 = "0x7303d2fadd895018075cbe76d8a700bc65b4a1b8641b97d660533f0e029e3954"
  private val l2TestAccountPrivateKey2 = "0xc5453712de35e7dc2c599b5f86df5d4f0de442d86a2865cfe557acd6d131aa6f"
  private val l2Credentials1 = Credentials.create(l2TestAccountPrivateKey1)
  private val l2Credentials2 = Credentials.create(l2TestAccountPrivateKey2)
  private val l2ChainId = l2ValidatorWeb3jClient.ethChainId().send().chainId.toLong()
  private val l2ValidatorTxManager = AsyncFriendlyTransactionManager(l2ValidatorWeb3jClient, l2Credentials1, l2ChainId)
  private val l2NodeTxManager = AsyncFriendlyTransactionManager(l2NodeWeb3jClient, l2Credentials1, l2ChainId)
  private val l2NodeTxManager2 = AsyncFriendlyTransactionManager(l2NodeWeb3jClient, l2Credentials2, l2ChainId)

  // Un comment to log Web3j requests and responses
  // init {
  //   Configurator.setLevel(HttpService::class.java,Level.DEBUG)
  // }
  companion object {
    var sendTxnHash = ""
  }

  @Test
  @Order(1)
  @Timeout(5, timeUnit = TimeUnit.MINUTES)
  fun `miner set gas price are sent to recipients correctly and underpriced txn is pending`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    // we need this mocked web3j client because the gas fee history in layer 1 is full of zeros initially
    var l1LatestBlockNumber = BigInteger.valueOf(2)
    val l1Web3jClientMock = mock<Web3j>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
    whenever(
      l1Web3jClientMock
        .ethFeeHistory(
          ArgumentMatchers.eq(feeHistoryBlockCount.toInt()),
          ArgumentMatchers.eq(DefaultBlockParameterName.LATEST),
          ArgumentMatchers.eq(listOf(feeHistoryRewardPercentile))
        )
        .sendAsync()
    )
      .thenAnswer {
        // The following should yield l2 calculated gas fee as 2016233773
        val feeHistoryResponse = getMockedEthFeeHistory(
          l1LatestBlockNumber.toULong() - feeHistoryBlockCount,
          initialReward,
          initialBaseFeePerGas,
          initialGasUsedRatio,
          feeHistoryBlockCount
        )
        SafeFuture.completedFuture(feeHistoryResponse)
      }
    whenever(
      l1Web3jClientMock
        .ethBlockNumber()
        .sendAsync()
    )
      .thenAnswer {
        val ethBlockNumber = EthBlockNumber()
        val l1Response = l1Web3jClient.ethBlockNumber().send()
        ethBlockNumber.result = l1Response.result
        l1LatestBlockNumber = l1Response.blockNumber
        SafeFuture.completedFuture(ethBlockNumber)
      }

    val dynamicGasPriceService = initialiseServices(vertx, l1Web3jClientMock, BigInteger.valueOf(1000000000))

    val l2ValidatorGasPrice = l2ValidatorWeb3jClient.ethGasPrice().send().gasPrice // should be 1000000000
    val l2NodeGasPrice = l2NodeWeb3jClient.ethGasPrice().send().gasPrice // should be 1000000000

    dynamicGasPriceService.start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 2) {
          dynamicGasPriceService.stop().thenApply {
            val updatedL2ValidatorGasPrice =
              l2ValidatorWeb3jClient.ethGasPrice().send().gasPrice // should see 2016233773
            val updatedL2NodeGasPrice = l2NodeWeb3jClient.ethGasPrice().send().gasPrice // should see 2016233773
            testContext.verify {
              Assertions.assertThat(updatedL2ValidatorGasPrice).isGreaterThan(l2ValidatorGasPrice)
              Assertions.assertThat(updatedL2NodeGasPrice).isGreaterThan(l2NodeGasPrice)
            }

            val sendResp = l2ValidatorTxManager.sendEIP1559Transaction(
              l2ChainId,
              updatedL2ValidatorGasPrice.subtract(BigInteger.valueOf(1)),
              updatedL2ValidatorGasPrice.subtract(BigInteger.valueOf(1)),
              DefaultGasProvider().gasLimit,
              l2ValidatorTxManager.fromAddress,
              "0x",
              BigInteger.valueOf(1000),
              false
            )
            // save the txn hash in the static variable to be retrieved in subsequent tests
            sendTxnHash = sendResp.transactionHash
            if (sendTxnHash == null) {
              // if node returns transaction already known it won't have hash
              testContext.failNow("Transaction hash is null")
            }
            vertx.setTimer(pollingInterval.inWholeMilliseconds * 2) {
              // should see exception as the underpriced txn would be keeping in txn pool
              testContext.verify {
                val receipt = l2ValidatorWeb3jClient.ethGetTransactionReceipt(sendResp.transactionHash).send()
                Assertions.assertThat(receipt.transactionReceipt.isPresent).isFalse()
              }.completeNow()
            }
          }
        }
      }.whenException(testContext::failNow)
  }

  @Test
  @Order(2)
  @Timeout(90, timeUnit = TimeUnit.SECONDS)
  fun `underpriced txn is mined after miner gas price set to a lower value`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    var l1LatestBlockNumber = BigInteger.valueOf(2)
    val l1Web3jClientMock = mock<Web3j>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
    whenever(
      l1Web3jClientMock
        .ethFeeHistory(
          ArgumentMatchers.eq(feeHistoryBlockCount.toInt()),
          ArgumentMatchers.eq(DefaultBlockParameterName.LATEST),
          ArgumentMatchers.eq(listOf(feeHistoryRewardPercentile))
        )
        .sendAsync()
    )
      .thenAnswer {
        // The following should yield l2 calculated gas fee as 201623383
        val feeHistoryResponse = getMockedEthFeeHistory(
          l1LatestBlockNumber.toULong() - feeHistoryBlockCount,
          (initialReward / 10u),
          (initialBaseFeePerGas / 10u),
          initialGasUsedRatio,
          feeHistoryBlockCount
        )
        SafeFuture.completedFuture(feeHistoryResponse)
      }
    whenever(
      l1Web3jClientMock
        .ethBlockNumber()
        .sendAsync()
    )
      .thenAnswer {
        val ethBlockNumber = EthBlockNumber()
        val l1Response = l1Web3jClient.ethBlockNumber().send()
        ethBlockNumber.result = l1Response.result
        l1LatestBlockNumber = l1Response.blockNumber
        SafeFuture.completedFuture(ethBlockNumber)
      }

    val dynamicGasPriceService = initialiseServices(vertx, l1Web3jClientMock)

    val l2ValidatorGasPrice = l2ValidatorWeb3jClient.ethGasPrice().send().gasPrice // should be 2016233773
    val l2NodeGasPrice = l2NodeWeb3jClient.ethGasPrice().send().gasPrice // should be 2016233773

    dynamicGasPriceService.start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 2) {
          dynamicGasPriceService.stop().thenApply {
            sendTransactionWithGasPrice(l2ValidatorGasPrice)
            val updatedL2ValidatorGasPrice =
              l2ValidatorWeb3jClient.ethGasPrice().send().gasPrice // should see 201623383
            val updatedL2NodeGasPrice = l2NodeWeb3jClient.ethGasPrice().send().gasPrice // should see 201623383
            testContext.verify {
              Assertions.assertThat(updatedL2ValidatorGasPrice).isLessThan(l2ValidatorGasPrice)
              Assertions.assertThat(updatedL2NodeGasPrice).isLessThan(l2NodeGasPrice)
            }
            vertx.setTimer(pollingInterval.inWholeMilliseconds * 4) {
              testContext.verify {
                val receipt = l2ValidatorWeb3jClient.ethGetTransactionReceipt(sendTxnHash).send()
                Assertions.assertThat(receipt.transactionReceipt.get().status.equals("0x1"))
              }.completeNow()
            }
          }
        }
      }.whenException(testContext::failNow)
  }

  @Test
  @Order(3)
  @Timeout(90, timeUnit = TimeUnit.SECONDS)
  fun `txn with max fee per gas as current gas price is sent to l2-node and is mined correctly`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val l2ValidatorGasPrice = l2ValidatorWeb3jClient.ethGasPrice().send().gasPrice

    vertx.setTimer(pollingInterval.inWholeMilliseconds * 2) {
      val sendResp = l2NodeTxManager.sendEIP1559Transaction(
        l2ChainId,
        l2ValidatorGasPrice,
        l2ValidatorGasPrice,
        DefaultGasProvider().gasLimit,
        l2NodeTxManager.fromAddress,
        "0x",
        BigInteger.valueOf(1000),
        false
      )
      sendTxnHash = sendResp.transactionHash

      vertx.setTimer(pollingInterval.inWholeMilliseconds * 3) {
        testContext.verify {
          val receipt = l2NodeWeb3jClient.ethGetTransactionReceipt(sendTxnHash).send()
          Assertions.assertThat(receipt.transactionReceipt.get().status.equals("0x1"))
        }.completeNow()
      }
    }
  }

  private fun sendTransactionWithGasPrice(gasPrice: BigInteger) {
    l2NodeTxManager2.sendTransaction(
      gasPrice,
      BigInteger.valueOf(25000),
      l2Credentials2.address,
      "",
      BigInteger.ZERO
    )
  }

  private fun getMockedEthFeeHistory(
    oldestBlockNumber: ULong,
    initialReward: ULong,
    initialBaseFeePerGas: ULong,
    initialGasUsedRatio: UInt,
    feeHistoryBlockCount: UInt
  ): EthFeeHistory {
    val feeHistoryResponse = EthFeeHistory()
    val feeHistory = EthFeeHistory.FeeHistory()
    feeHistory.setReward((initialReward until initialReward + feeHistoryBlockCount).map { listOf(it.toString()) })
    feeHistory.setBaseFeePerGas(
      (initialBaseFeePerGas until initialBaseFeePerGas + feeHistoryBlockCount + 1u)
        .map { it.toString() }
    )
    feeHistory.gasUsedRatio =
      (initialGasUsedRatio until initialGasUsedRatio + feeHistoryBlockCount).map { it.toDouble() / 100.0 }
    feeHistory.setOldestBlock(oldestBlockNumber.toHexString())
    feeHistoryResponse.result = feeHistory
    return feeHistoryResponse
  }

  private fun initialiseServices(
    vertx: Vertx,
    l1Web3jClient: Web3j,
    initialGasPrice: BigInteger? = null
  ): DynamicGasPriceService {
    val feesFetcher: FeesFetcher = FeeHistoryFetcherImpl(
      l1Web3jClient,
      FeeHistoryFetcherImpl.Config(
        feeHistoryBlockCount,
        feeHistoryRewardPercentile
      )
    )

    val l2MinMinerTipCalculator: FeesCalculator = GasUsageRatioWeightedAverageFeesCalculator(
      GasUsageRatioWeightedAverageFeesCalculator.Config(
        baseFeeCoefficient.toBigDecimal(),
        priorityFeeCoefficient.toBigDecimal()
      )
    )

    val l2SetGasPriceUpdater: GasPriceUpdater = GasPriceUpdaterImpl(
      VertxHttpJsonRpcClientFactory(vertx, meterRegistry),
      GasPriceUpdaterImpl.Config(
        minerGasPriceUpdateRecipients.map { URL(it) }
      )
    )

    if (initialGasPrice != null) {
      l2SetGasPriceUpdater.updateMinerGasPrice(initialGasPrice).get()
    }

    return DynamicGasPriceService(
      DynamicGasPriceService.Config(
        pollingInterval,
        l2GasPriceCap
      ),
      vertx,
      feesFetcher,
      l2MinMinerTipCalculator,
      l2SetGasPriceUpdater
    )
  }
}
