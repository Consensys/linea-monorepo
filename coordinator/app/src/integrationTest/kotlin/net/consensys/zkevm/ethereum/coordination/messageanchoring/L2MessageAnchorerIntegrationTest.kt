package net.consensys.zkevm.ethereum.coordination.messageanchoring

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.L2MessageService
import net.consensys.linea.contract.ZkEvmV2
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.web3j.abi.EventEncoder
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
class L2MessageAnchorerIntegrationTest {
  private val testZkEvmContractAddress = System.getProperty("ZkEvmV2Address")
  private val testL2MessageManagerContractAddress = System.getProperty("L2MessageService")
  private val l2RecipientAddress = "0x03dfa322A95039BB679771346Ee2dBfEa0e2B773"
  private val l1RpcEndpoint = "http://127.0.0.1:8445"
  private val l2RpcEndpoint = "http://127.0.0.1:8545"

  // WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
  private val testAccountPrivateKey = "202454d1b4e72c41ebf58150030f649648d3cf5590297fb6718e27039ed9c86d"
  private val l2TestAccountPrivateKey = "0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae"
  private val blockRangeLoopLimit = 100u

  private val firstMessage =
    L1EventQuerierIntegrationTest.L1MessageToSend(
      l2RecipientAddress,
      BigInteger.TEN,
      ByteArray(0),
      BigInteger.valueOf(100001)
    )
  private val secondMessage = firstMessage.copy(fee = BigInteger.valueOf(11))

  @Test
  @Timeout(2, timeUnit = TimeUnit.MINUTES)
  fun `can send a message on L1 and see the hash on L2`(vertx: Vertx, testContext: VertxTestContext) {
    val l2Web3jClient = Web3j.build(HttpService(l2RpcEndpoint), 1000, Async.defaultExecutorService())
    val l2credentials = Credentials.create(l2TestAccountPrivateKey)
    val l2ChainId = l2Web3jClient.ethChainId().send().chainId.toLong()
    val pollingTransactionReceiptProcessor = PollingTransactionReceiptProcessor(l2Web3jClient, 1000, 40)
    val l2TransactionManager = AsyncFriendlyTransactionManager(
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

    // we need to wait a sufficient time for block processing
    val l2MessageAnchorerImpl = L2MessageAnchorerImpl(
      vertx,
      l2Web3jClient,
      l2Contract,
      L2MessageAnchorerImpl.Config(12.seconds, 5u, 0)
    )

    val expectedHash = Bytes32.random()

    l2MessageAnchorerImpl.anchorMessages(listOf(expectedHash))
      .thenCompose { txReceipt ->
        l2Contract.sendMessage(firstMessage.recipient, firstMessage.fee, firstMessage.calldata, firstMessage.value)
          .sendAsync()
          .thenApply { txReceipt }
      }
      .thenCompose { txReceipt ->
        l2Contract.sendMessage(secondMessage.recipient, secondMessage.fee, secondMessage.calldata, secondMessage.value)
          .sendAsync()
          .thenApply { txReceipt }
      }
      .thenApply { transactionReceipt ->
        testContext.verify {
          Assertions.assertThat(transactionReceipt.logs).isNotNull
          Assertions.assertThat(transactionReceipt.logs).isNotEmpty
          Assertions.assertThat(transactionReceipt.logs.count()).isEqualTo(1)
          Assertions.assertThat(transactionReceipt.logs[0].topics[0]).isEqualTo(
            EventEncoder.encode(L2MessageService.L1L2MESSAGEHASHESADDEDTOINBOX_EVENT)
          )
          Assertions.assertThat(transactionReceipt.logs[0].topics.count()).isEqualTo(1)
          Assertions.assertThat(transactionReceipt.logs[0].data).contains(
            expectedHash.toString().removePrefix("0x")
          )
          Assertions.assertThat(transactionReceipt.blockNumber).isNotNull
        }.completeNow()
      }.whenException(testContext::failNow)
  }

  @Test
  @Timeout(1, timeUnit = TimeUnit.MINUTES)
  fun `all hashes found are anchored`(vertx: Vertx, testContext: VertxTestContext) {
    val l1Web3jClient = Web3j.build(HttpService(l1RpcEndpoint), 1000, Async.defaultExecutorService())
    val credentials = Credentials.create(testAccountPrivateKey)
    val chainId = l1Web3jClient.ethChainId().send().chainId.toLong()
    val l1PollingTransactionReceiptProcessor = PollingTransactionReceiptProcessor(l1Web3jClient, 1000, 40)
    val transactionManager = RawTransactionManager(
      l1Web3jClient,
      credentials,
      chainId,
      l1PollingTransactionReceiptProcessor
    )

    val contract = ZkEvmV2.load(testZkEvmContractAddress, l1Web3jClient, transactionManager, DefaultGasProvider())

    val l2Web3jClient = Web3j.build(HttpService(l2RpcEndpoint), 1000, Async.defaultExecutorService())
    val l2credentials = Credentials.create(l2TestAccountPrivateKey)
    val l2ChainId = l2Web3jClient.ethChainId().send().chainId.toLong()
    val l2PollingTransactionReceiptProcessor = PollingTransactionReceiptProcessor(l2Web3jClient, 1000, 40)
    val l2TransactionManager = RawTransactionManager(
      l2Web3jClient,
      l2credentials,
      l2ChainId,
      l2PollingTransactionReceiptProcessor
    )

    val l2Contract = L2MessageService.load(
      testL2MessageManagerContractAddress,
      l2Web3jClient,
      l2TransactionManager,
      DefaultGasProvider()
    )

    val baseMessageToSend =
      L1EventQuerierIntegrationTest.L1MessageToSend(
        l2RecipientAddress,
        BigInteger.TEN,
        ByteArray(0),
        BigInteger.valueOf(100001)
      )
    val messagesToSend = listOf(
      baseMessageToSend,
      baseMessageToSend.copy(fee = BigInteger.valueOf(11)),
      baseMessageToSend.copy(value = BigInteger.valueOf(100001))
    )
    val emittedEvents = messagesToSend.map {
      contract.sendMessage(it.recipient, it.fee, it.calldata, it.value).send()
    }.map {
      ZkEvmV2.staticExtractEventParameters(
        ZkEvmV2.MESSAGESENT_EVENT,
        it.logs.first { log ->
          log.topics.contains(EventEncoder.encode(ZkEvmV2.MESSAGESENT_EVENT))
        }
      )
    }

    val hashInTheMiddle = emittedEvents[1].let {
      Bytes32.wrap(it.indexedValues[2].value as ByteArray)
    }

    val l1QuerierImpl = L1EventQuerierImpl(
      vertx,
      L1EventQuerierImpl.Config(
        2.seconds,
        30.seconds,
        BigInteger.ZERO,
        100u,
        testZkEvmContractAddress,
        "latest",
        blockRangeLoopLimit
      ),
      l1Web3jClient
    )

    // we need to wait a sufficient time for block processing
    val l2MessageAnchorerImpl = L2MessageAnchorerImpl(
      vertx,
      l2Web3jClient,
      l2Contract,
      L2MessageAnchorerImpl.Config(12.seconds, 5u, 0)
    )

    val expectedHash = Bytes32.wrap(emittedEvents.last().indexedValues[2].value as ByteArray)

    l1QuerierImpl.getSendMessageEventsForAnchoredMessage(MessageHashAnchoredEvent(hashInTheMiddle))
      .thenApply { events ->
        l2MessageAnchorerImpl.anchorMessages(events.map { it.messageHash })
          .thenApply { transactionReceipt ->
            testContext.verify {
              Assertions.assertThat(transactionReceipt.logs).isNotNull
              Assertions.assertThat(transactionReceipt.logs).isNotEmpty
              Assertions.assertThat(transactionReceipt.logs.count()).isEqualTo(1)
              Assertions.assertThat(transactionReceipt.logs[0].topics[0]).isEqualTo(
                EventEncoder.encode(L2MessageService.L1L2MESSAGEHASHESADDEDTOINBOX_EVENT)
              )
              Assertions.assertThat(transactionReceipt.logs[0].topics.count()).isEqualTo(1)
              Assertions.assertThat(transactionReceipt.logs[0].data).contains(
                expectedHash.toString().removePrefix("0x")
              )
              Assertions.assertThat(transactionReceipt.blockNumber).isNotNull
            }.completeNow()
          }
      }.whenException(testContext::failNow)
  }
}
