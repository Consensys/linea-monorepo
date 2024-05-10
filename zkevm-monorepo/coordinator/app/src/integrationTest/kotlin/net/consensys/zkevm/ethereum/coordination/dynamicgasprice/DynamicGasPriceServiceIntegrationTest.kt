package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.BoundableFeeCalculator
import net.consensys.linea.jsonrpc.client.RequestRetryConfig
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.linea.web3j.Web3jBlobExtended
import net.consensys.toHexString
import net.consensys.toULong
import net.consensys.zkevm.ethereum.gaspricing.FeesCalculator
import net.consensys.zkevm.ethereum.gaspricing.FeesFetcher
import net.consensys.zkevm.ethereum.gaspricing.GasPriceUpdater
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.AfterAll
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.MethodOrderer
import org.junit.jupiter.api.Order
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
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
import java.math.BigDecimal
import java.math.BigInteger
import java.net.URL
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@TestMethodOrder(MethodOrderer.OrderAnnotation::class)
@Disabled("Disable this test for now as causes issues with other tests because of price updates")
class DynamicGasPriceServiceIntegrationTest {
  private val meterRegistry = SimpleMeterRegistry()
  private val pollingInterval = 2.seconds
  private val feeHistoryBlockCount = 10u
  private val feeHistoryRewardPercentile = 15.0
  private val initialReward = 1000000000uL
  private val initialBaseFeePerGas = 10000000000uL
  private val initialGasUsedRatio = 50u
  private val l2GasPriceUpperBound = BigInteger("fffffffffffffffff", 16)
  private val l2GasPriceLowerBound = BigInteger("f4240", 16)
  private val l2GasPriceFixedCost = BigInteger.ZERO
  private val l2ValidatorRpcEndpoint = "http://localhost:8545"
  private val l2NodeRpcEndpoint = "http://localhost:8845"
  private val gethRecipients = listOf(l2NodeRpcEndpoint)
  private val besuRecipients = listOf(l2ValidatorRpcEndpoint)

  // Set org.web3j.protocol.http to DEBUG in log4j2.xml to debug requests/responses
  private val l1Web3jClient = Web3j.build(HttpService("http://localhost:8445"))
  private val l2ValidatorWeb3jClient = Web3j.build(HttpService(l2ValidatorRpcEndpoint))
  private val l2NodeWeb3jClient = Web3j.build(HttpService(l2NodeRpcEndpoint))

  // WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
  private val l2TestAccountPrivateKey1 = "0x7303d2fadd895018075cbe76d8a700bc65b4a1b8641b97d660533f0e029e3954"
  private val l2TestAccountPrivateKey2 = "0xc5453712de35e7dc2c599b5f86df5d4f0de442d86a2865cfe557acd6d131aa6f"
  private val l2Credentials1 = Credentials.create(l2TestAccountPrivateKey1)
  private val l2Credentials2 = Credentials.create(l2TestAccountPrivateKey2)
  private val l2ChainId = l2ValidatorWeb3jClient.ethChainId().send().chainId.toLong()
  private val l2NodeTxManager = AsyncFriendlyTransactionManager(l2NodeWeb3jClient, l2Credentials1, l2ChainId)
  private val l2NodeTxManager2 = AsyncFriendlyTransactionManager(l2NodeWeb3jClient, l2Credentials2, l2ChainId)
  private val l2MinMinerTipCalculator: FeesCalculator = BoundableFeeCalculator(
    BoundableFeeCalculator.Config(l2GasPriceUpperBound, l2GasPriceLowerBound, l2GasPriceFixedCost),
    GasUsageRatioWeightedAverageFeesCalculator(
      GasUsageRatioWeightedAverageFeesCalculator.Config(
        baseFeeCoefficient = 0.1.toBigDecimal(),
        priorityFeeCoefficient = 0.1.toBigDecimal(),
        baseFeeBlobCoefficient = 0.1.toBigDecimal(),
        blobSubmissionExpectedExecutionGas = BigDecimal(131_000.0),
        expectedBlobGas = BigDecimal(120_000.0)
      )
    )
  )

  companion object {
    var sendTxnHash = ""
  }

  @AfterAll
  fun afterAll(vertx: Vertx) {
    val l2SetGasPriceUpdater: GasPriceUpdater = createGasPriceUpdater(vertx)
    l2SetGasPriceUpdater.updateMinerGasPrice(BigInteger.valueOf(1322222229)).get()
  }

  // eth_gasPrice returns network gas price capped by admin_setPrice ± a margin of x% (e.g. 15%)
  // so for new admin_setPrice to observable from eth_gasPrice,
  // we need to call admin_setPrice outside of the margin.
  // Target: L2 price = L1 price / 10. We need to increase L2 price by 20x to have new L2 price increased by 2x
  // after applying l2 coefficient of 0.1
  @Test
  @Order(1)
  @Timeout(5, timeUnit = TimeUnit.MINUTES)
  fun `miner set gas price are sent to recipients correctly and underpriced txn is pending`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    // we need this mocked web3j client because the gas fee history in layer 1 is full of zeros initially
    val l1Web3jClientMock = mock<Web3j>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
    var l1LatestBlockNumber = BigInteger.valueOf(2)
    whenever(
      l1Web3jClientMock
        .ethBlockNumber()
        .sendAsync()
    )
      .thenAnswer {
        val l1Response = l1Web3jClient.ethBlockNumber().send()
        l1LatestBlockNumber = l1Response.blockNumber
        val ethBlockNumber = EthBlockNumber().apply {
          this.result = l1Response.result
        }
        SafeFuture.completedFuture(ethBlockNumber)
      }
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
        val lastFakeFeeHistory = buildFakeEthFeeHistory(
          l1LatestBlockNumber.toULong() - feeHistoryBlockCount,
          initialReward,
          initialBaseFeePerGas = initialBaseFeePerGas.times(2u),
          initialGasUsedRatio,
          feeHistoryBlockCount
        )
        SafeFuture.completedFuture(lastFakeFeeHistory)
      }

    val dynamicGasPriceService = initialiseServices(
      vertx,
      l1Web3jClientMock,
      BigInteger.valueOf(initialBaseFeePerGas.div(10u).toLong())
    )

    dynamicGasPriceService.start()
      .thenApply {
        val initialL2NodePrice = l2NodeWeb3jClient.ethGasPrice().send().gasPrice // 1200000000

        vertx.setTimer(pollingInterval.inWholeMilliseconds * 2) {
          dynamicGasPriceService.stop().thenApply {
            val updatedL2NodeGasPrice = l2NodeWeb3jClient.ethGasPrice().send().gasPrice // 2100000001*1.2
            testContext.verify {
              Assertions.assertThat(updatedL2NodeGasPrice).isGreaterThanOrEqualTo(initialL2NodePrice)
            }

            // Here we tried to calculate the original set mineable gas price from current gas price, we
            // divide by 1.5 (i.e. ZKGETH_UPPER_GAS_MARGIN_PERCENTS) to get the safe gas price that would be
            // underpriced by subtracting 1
            val setMineableGasPrice = BigDecimal.valueOf(updatedL2NodeGasPrice.toDouble() / 1.5).toBigInteger()
            val sendResp = l2NodeTxManager.sendEIP1559Transaction(
              l2ChainId,
              setMineableGasPrice.subtract(BigInteger.valueOf(1)), // 1679999999
              /*maxFeePerGas*/updatedL2NodeGasPrice, // 2520000001
              DefaultGasProvider().gasLimit,
              l2NodeTxManager.fromAddress,
              Bytes.random(32).toHexString(), // avoid tx already known error
              BigInteger.valueOf(1000),
              false
            )
            println("maxPriorityGasFee: ${setMineableGasPrice.subtract(BigInteger.valueOf(1))}")
            // save the txn hash in the static variable to be retrieved in subsequent tests
            sendTxnHash = sendResp.transactionHash
            if (sendTxnHash == null) {
              // if node returns transaction already known it won't have hash
              testContext.failNow("Transaction hash is null")
            }
            vertx.setTimer(pollingInterval.inWholeMilliseconds * 2) {
              // should see exception as the underpriced txn would be keeping in txn pool
              testContext.verify {
                val receipt = l2NodeWeb3jClient.ethGetTransactionReceipt(sendResp.transactionHash).send()
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
        // The following should yield l2 calculated gas fee as 110,000,001
        val feeHistoryResponse = buildFakeEthFeeHistory(
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

    val l2NodeGasPrice = l2NodeWeb3jClient.ethGasPrice().send().gasPrice // 2520000001

    dynamicGasPriceService.start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 2) {
          dynamicGasPriceService.stop().thenApply {
            sendTransactionWithGasPrice(l2NodeGasPrice)
            val updatedL2NodeGasPrice = l2NodeWeb3jClient.ethGasPrice().send().gasPrice // 165000001
            testContext.verify {
              Assertions.assertThat(updatedL2NodeGasPrice).isLessThan(l2NodeGasPrice)
            }
            vertx.setTimer(pollingInterval.inWholeMilliseconds * 4) {
              testContext.verify {
                val receipt = l2NodeWeb3jClient.ethGetTransactionReceipt(sendTxnHash).send()
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
    val l2NodeGasPrice = l2NodeWeb3jClient.ethGasPrice().send().gasPrice // 165000001

    vertx.setTimer(pollingInterval.inWholeMilliseconds * 2) {
      val sendResp = l2NodeTxManager.sendEIP1559Transaction(
        l2ChainId,
        l2NodeGasPrice,
        l2NodeGasPrice,
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

  private fun buildFakeEthFeeHistory(
    oldestBlockNumber: ULong,
    initialReward: ULong,
    initialBaseFeePerGas: ULong,
    initialGasUsedRatio: UInt,
    feeHistoryBlockCount: UInt
  ): EthFeeHistory {
    val feeHistory = EthFeeHistory.FeeHistory()
    feeHistory.setReward((initialReward until initialReward + feeHistoryBlockCount).map { listOf(it.toString()) })
    feeHistory.setBaseFeePerGas(
      (initialBaseFeePerGas until initialBaseFeePerGas + feeHistoryBlockCount + 1u)
        .map { it.toString() }
    )
    feeHistory.gasUsedRatio =
      (initialGasUsedRatio until initialGasUsedRatio + feeHistoryBlockCount).map { it.toDouble() / 100.0 }
    feeHistory.setOldestBlock(oldestBlockNumber.toHexString())

    val feeHistoryResponse = EthFeeHistory()
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
      Web3jBlobExtended(
        HttpService(System.getProperty("L1_RPC_URL", "http://localhost:8445"))
      ),
      FeeHistoryFetcherImpl.Config(
        feeHistoryBlockCount,
        feeHistoryRewardPercentile
      )
    )

    val l2SetGasPriceUpdater: GasPriceUpdater = createGasPriceUpdater(vertx)

    if (initialGasPrice != null) {
      l2SetGasPriceUpdater.updateMinerGasPrice(initialGasPrice).get()
    }

    return DynamicGasPriceService(
      DynamicGasPriceService.Config(pollingInterval),
      vertx,
      feesFetcher,
      l2MinMinerTipCalculator,
      l2SetGasPriceUpdater
    )
  }

  private fun createGasPriceUpdater(vertx: Vertx) = GasPriceUpdaterImpl(
    VertxHttpJsonRpcClientFactory(vertx, meterRegistry),
    GasPriceUpdaterImpl.Config(
      gethEndpoints = gethRecipients.map { URL(it) },
      besuEndPoints = besuRecipients.map { URL(it) },
      retryConfig = RequestRetryConfig(
        maxAttempts = 3u,
        backoffDelay = 1.seconds
      )
    )
  )
}
