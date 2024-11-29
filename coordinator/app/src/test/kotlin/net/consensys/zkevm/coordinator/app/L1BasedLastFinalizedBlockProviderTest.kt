package net.consensys.zkevm.coordinator.app

import build.linea.contract.LineaRollupV5
import io.vertx.core.Vertx
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import org.web3j.protocol.core.RemoteFunctionCall
import java.math.BigInteger
import java.util.concurrent.CompletableFuture
import kotlin.time.Duration.Companion.milliseconds

class L1BasedLastFinalizedBlockProviderTest {
  private lateinit var lineaRollupSmartContractWeb3jClient: LineaRollupV5

  @BeforeEach
  fun beforeEach() {
    lineaRollupSmartContractWeb3jClient = mock()
  }

  @Test
  fun `shall wait number of blocks before returning for consistency`() {
    val replies = listOf(
      mockRemoteFnCallWithBlockNumber(100),
      mockRemoteFnCallWithBlockNumber(100),
      mockRemoteFnCallWithBlockNumber(101),
      mockRemoteFnCallWithBlockNumber(101),
      mockRemoteFnCallWithBlockNumber(101),
      mockRemoteFnCallWithBlockNumber(101)
    )
    whenever(lineaRollupSmartContractWeb3jClient.currentL2BlockNumber())
      .thenReturn(replies[0], *replies.subList(1, replies.size).toTypedArray())

    val resumerCalculator = L1BasedLastFinalizedBlockProvider(
      Vertx.vertx(),
      lineaRollupSmartContractWeb3jClient,
      consistentNumberOfBlocksOnL1 = 3u,
      numberOfRetries = 50u,
      pollingInterval = 10.milliseconds
    )

    assertThat(resumerCalculator.getLastFinalizedBlock().get()).isEqualTo(101.toULong())
  }

  private fun mockRemoteFnCallWithBlockNumber(blockNumber: Long): RemoteFunctionCall<BigInteger> {
    return mock<RemoteFunctionCall<BigInteger>>() {
      on { send() } doReturn (BigInteger.valueOf(blockNumber))
      on { sendAsync() } doReturn (CompletableFuture.completedFuture(BigInteger.valueOf(blockNumber)))
    }
  }
}
