package net.consensys.zkevm.ethereum.coordination.messageanchoring

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.contract.L2MessageService
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.http.HttpService
import org.web3j.tx.RawTransactionManager
import org.web3j.tx.gas.DefaultGasProvider
import org.web3j.tx.response.PollingTransactionReceiptProcessor
import org.web3j.utils.Async
import java.math.BigInteger
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
class L2QuerierIntegrationTest {
  private val testL2MessageManagerContractAddress = System.getProperty("L2MessageService")
  private val l2RpcEndpoint = "http://localhost:8545"

  // WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
  private val l2TestAccountPrivateKey = "0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae"

  @Test
  @Timeout(1, timeUnit = TimeUnit.MINUTES)
  fun `anchored hashes are returned correctly`(vertx: Vertx, testContext: VertxTestContext) {
    val l2Web3jClient = Web3j.build(HttpService(l2RpcEndpoint), 1000, Async.defaultExecutorService())
    val l2credentials = Credentials.create(l2TestAccountPrivateKey)
    val l2ChainId = l2Web3jClient.ethChainId().send().chainId.toLong()
    val pollingTransactionReceiptProcessor = PollingTransactionReceiptProcessor(l2Web3jClient, 1000, 40)
    val l2TransactionManager = RawTransactionManager(
      l2Web3jClient,
      l2credentials,
      l2ChainId,
      pollingTransactionReceiptProcessor
    )

    val l2Contract = L2MessageService.load(
      testL2MessageManagerContractAddress,
      l2Web3jClient,
      l2TransactionManager,
      DefaultGasProvider()
    )

    val config = L2QuerierImpl.Config(0u, 2u, 2u, testL2MessageManagerContractAddress)
    val l2Querier = L2QuerierImpl(l2Web3jClient, l2Contract, config, vertx)
    val hashes = createRandomEventWithHashes(10)

    val l2MessageAnchorerImpl = L2MessageAnchorerImpl(
      vertx,
      l2Web3jClient,
      l2Contract,
      L2MessageAnchorerImpl.Config(12.seconds, 8u, 0)
    )

    l2MessageAnchorerImpl.anchorMessages(hashes)
      .thenCompose { _ ->
        l2Querier.findLastFinalizedAnchoredEvent().thenCompose {
          l2Querier.getMessageHashStatus(it!!.messageHash).thenApply {
            testContext.verify {
              Assertions.assertThat(it).isNotNull
              Assertions.assertThat(it).isEqualTo(BigInteger.valueOf(1))
            }.completeNow()
          }
        }.whenException(testContext::failNow)
      }
  }

  private fun createRandomEventWithHashes(numberOfRandomHashes: Int): List<Bytes32> {
    return (0..numberOfRandomHashes)
      .map { Bytes32.random() }
  }
}
