package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.RepeatedTest
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.ArgumentMatchers
import org.mockito.Mockito
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.methods.response.EthBlockNumber
import org.web3j.protocol.core.methods.response.EthFeeHistory
import org.web3j.protocol.core.methods.response.EthGasPrice
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.*
import java.util.concurrent.CompletableFuture
import java.util.concurrent.TimeUnit

@ExtendWith(VertxExtension::class)
class FeeHistoryFetcherImplTest {
  private val feeHistoryBlockCount = 10u
  private val feeHistoryRewardPercentile = 15.0

  @RepeatedTest(1)
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun feeHistoryFetcherImpl_returnsFeeHistoryData(testContext: VertxTestContext) {
    val l1ClientMock = createMockedWeb3jClient()

    val feeHistoryFetcherImpl =
      FeeHistoryFetcherImpl(
        l1ClientMock,
        FeeHistoryFetcherImpl.Config(
          feeHistoryBlockCount,
          feeHistoryRewardPercentile
        )
      )

    feeHistoryFetcherImpl.getL1EthGasPriceData()
      .thenApply {
        testContext
          .verify {
            assertThat(it).isNotNull
            assertThat(it.baseFeePerGas.size.toUInt()).isEqualTo(feeHistoryBlockCount + 1u)
            assertThat(it.reward.size.toUInt()).isEqualTo(feeHistoryBlockCount)
            assertThat(it.gasUsedRatio.size.toUInt()).isEqualTo(feeHistoryBlockCount)
            it.reward.forEach {
              assertThat(it.size).isEqualTo(1)
            }
          }
          .completeNow()
      }
  }

  private fun createMockedWeb3jClient(): Web3j {
    val web3jClient = mock<Web3j>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
    whenever(
      web3jClient
        .ethFeeHistory(
          ArgumentMatchers.eq(feeHistoryBlockCount.toInt()),
          ArgumentMatchers.eq(DefaultBlockParameter.valueOf("latest")),
          ArgumentMatchers.eq(listOf(feeHistoryRewardPercentile))
        )
        .sendAsync()
    )
      .thenAnswer {
        val feeHistoryResponse = EthFeeHistory()
        val feeHistory = EthFeeHistory.FeeHistory()
        feeHistory.setReward((1000 until 1010).map { listOf(it.toString()) })
        feeHistory.setBaseFeePerGas((10000 until 10011).map { it.toString() })
        feeHistory.gasUsedRatio = (10 until 20).map { it / 100.0 }
        feeHistoryResponse.result = feeHistory
        feeHistory.setOldestBlock("0x16")
        SafeFuture.completedFuture(feeHistoryResponse)
      }
    whenever(
      web3jClient
        .ethGasPrice()
        .sendAsync()
    )
      .thenAnswer {
        val gasPriceResponse = EthGasPrice()
        gasPriceResponse.result = "0x100"
        SafeFuture.completedFuture(gasPriceResponse)
      }
    val mockBlockNumberReturn = mock<EthBlockNumber>()
    whenever(mockBlockNumberReturn.blockNumber).thenReturn(BigInteger.valueOf(13L))
    whenever(web3jClient.ethBlockNumber().sendAsync())
      .thenReturn(CompletableFuture.completedFuture(mockBlockNumberReturn))

    return web3jClient
  }
}
