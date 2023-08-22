package net.consensys.zkevm.ethereum.coordination.messageanchoring

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.async.retryWithInterval
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.L2MessageService
import net.consensys.linea.contract.ZkEvmV2
import net.consensys.zkevm.ethereum.EIP1559GasProvider
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.web3j.abi.EventEncoder
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.http.HttpService
import org.web3j.tx.gas.DefaultGasProvider
import org.web3j.tx.response.PollingTransactionReceiptProcessor
import org.web3j.utils.Async
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.CompletableFuture
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
class MessageServiceIntegrationTest {
  private val log: Logger = LogManager.getLogger(this::class.java)

  private val testZkEvmContractAddress = System.getProperty("ZkEvmV2Address")
  private val testL2MessageManagerContractAddress = System.getProperty("L2MessageService")
  private val l2RecipientAddress = "0x03dfa322A95039BB679771346Ee2dBfEa0e2B773"
  private val l1RpcEndpoint = "http://localhost:8445"
  private val l2RpcEndpoint = "http://localhost:8545"

  // WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
  private val testAccountPrivateKey = "202454d1b4e72c41ebf58150030f649648d3cf5590297fb6718e27039ed9c86d"
  private val l2TestAccountPrivateKey = "0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae"

  private val l1Web3Client = Web3j.build(HttpService(l1RpcEndpoint), 1000, Async.defaultExecutorService())
  private val credentials = Credentials.create(testAccountPrivateKey)
  private val chainId = l1Web3Client.ethChainId().send().chainId.toLong()
  private val l1PollingTransactionReceiptProcessor = PollingTransactionReceiptProcessor(l1Web3Client, 1000, 40)
  private val transactionManager = AsyncFriendlyTransactionManager(
    l1Web3Client,
    credentials,
    chainId,
    l1PollingTransactionReceiptProcessor
  )
  private val contract = ZkEvmV2.load(testZkEvmContractAddress, l1Web3Client, transactionManager, DefaultGasProvider())

  private val l2Web3jClient = Web3j.build(HttpService(l2RpcEndpoint), 1000, Async.defaultExecutorService())
  private val l2credentials = Credentials.create(l2TestAccountPrivateKey)
  private val l2ChainId = l2Web3jClient.ethChainId().send().chainId.toLong()
  private val l2PollingTransactionReceiptProcessor = PollingTransactionReceiptProcessor(l2Web3jClient, 1000, 40)
  private val l2TransactionManager = AsyncFriendlyTransactionManager(
    l2Web3jClient,
    l2credentials,
    l2ChainId,
    l2PollingTransactionReceiptProcessor
  )

  private val l2Contract = L2MessageService.load(
    testL2MessageManagerContractAddress,
    l2Web3jClient,
    l2TransactionManager,
    DefaultGasProvider()
  )

  private val pollingInterval = 2.seconds
  private val sendMessagePollingInterval = 2.seconds
  private val maxScrapingTime = 30.seconds
  private val earliestBlock = BigInteger.valueOf(0)
  private val maxMessagesToAnchor = 5u
  private val blockRangeLoopLimit = 100u
  private val receiptPollingInterval = 2.seconds

  private val messageAnchoringServiceConfig = MessageAnchoringService.Config(
    pollingInterval,
    maxMessagesToAnchor
  )

  @Test
  @Timeout(1, timeUnit = TimeUnit.MINUTES)
  fun `anchored hashes are returned correctly`(vertx: Vertx, testContext: VertxTestContext) {
    val messageAnchoringService = initialiseServices(vertx)

    val sentMessages = sendMessages()
    messageAnchoringService.start()
      .thenCompose { _ ->
        allHashesHaveBeenSet(sentMessages, 60, vertx).thenCompose { success ->
          messageAnchoringService.stop().thenApply {
            testContext.verify {
              Assertions.assertThat(success).isTrue()
            }.completeNow()
          }
        }
      }.whenException(testContext::failNow)
  }

  private fun initialiseServices(vertx: Vertx): MessageAnchoringService {
    val l2MessageService =
      L2MessageService.load(
        testL2MessageManagerContractAddress,
        l2Web3jClient,
        l2TransactionManager,
        EIP1559GasProvider(
          l2Web3jClient,
          EIP1559GasProvider.Config(
            BigInteger.valueOf(10000000),
            BigInteger.valueOf(100000000000),
            4u,
            0.5
          )
        )
      )

    val l1EventQuerier = L1EventQuerierImpl(
      vertx,
      L1EventQuerierImpl.Config(
        sendMessagePollingInterval,
        maxScrapingTime,
        earliestBlock,
        maxMessagesToAnchor,
        testZkEvmContractAddress,
        "latest",
        blockRangeLoopLimit
      ),
      l1Web3Client
    )

    val l2Querier = L2QuerierImpl(
      l2Web3jClient,
      l2MessageService,
      L2QuerierImpl.Config(
        0u,
        5u,
        2u,
        testL2MessageManagerContractAddress
      ),
      vertx
    )

    val l2MessageAnchorer: L2MessageAnchorer = L2MessageAnchorerImpl(
      vertx,
      l2Web3jClient,
      l2MessageService,
      L2MessageAnchorerImpl.Config(
        receiptPollingInterval,
        10u,
        0
      )
    )

    return MessageAnchoringService(
      messageAnchoringServiceConfig,
      vertx,
      l1EventQuerier,
      l2MessageAnchorer,
      l2Querier,
      l2Contract,
      l2TransactionManager
    )
  }

  private fun allHashesHaveBeenSet(
    messageHashes: List<Bytes32>,
    maxAttempts: Int,
    vertx: Vertx
  ): SafeFuture<Boolean> {
    return retryWithInterval(maxAttempts, 2.seconds, vertx, { it }) {
      val statuses: List<CompletableFuture<Boolean>> = messageHashes.map { messageHash ->
        l2Contract.inboxL1L2MessageStatus(messageHash.toArray()).sendAsync()
          .thenApply {
            it == BigInteger.valueOf(1)
          }
      }
      statuses.fold(SafeFuture.completedFuture(true)) { acc, newFuture ->
        acc.thenCompose { allPreviousComplete ->
          if (allPreviousComplete) {
            newFuture.thenApply { allPreviousComplete && it }
          } else {
            SafeFuture.completedFuture(false)
          }
        }
      }
    }
  }

  private fun sendMessages(): List<Bytes32> {
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
      baseMessageToSend.copy(value = BigInteger.valueOf(100001)),
      baseMessageToSend.copy(value = BigInteger.valueOf(100001)),
      baseMessageToSend.copy(value = BigInteger.valueOf(100001))
    )

    val emittedEvents = messagesToSend.map {
      contract.sendMessage(it.recipient, it.fee, it.calldata, it.value).send()
    }.map {
      log.debug("Message has been sent in block {}", it.blockNumber)
      ZkEvmV2.staticExtractEventParameters(
        ZkEvmV2.MESSAGESENT_EVENT,
        it.logs.first { log ->
          log.topics.contains(EventEncoder.encode(ZkEvmV2.MESSAGESENT_EVENT))
        }
      )
    }.map { Bytes32.wrap(it.indexedValues[2].value as ByteArray) }

    return emittedEvents
  }
}
