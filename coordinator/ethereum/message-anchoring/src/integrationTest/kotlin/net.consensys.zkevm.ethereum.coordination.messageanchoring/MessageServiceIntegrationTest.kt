package net.consensys.zkevm.ethereum.coordination.messageanchoring

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.async.AsyncRetryer
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.L2MessageService
import net.consensys.linea.contract.LineaRollup
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.linea.contract.RollupSmartContractClientWeb3JImpl
import net.consensys.linea.contract.Web3JLogsClient
import net.consensys.toBigInteger
import net.consensys.toULong
import net.consensys.zkevm.ethereum.ContractsManager
import net.consensys.zkevm.ethereum.L1AccountManager
import net.consensys.zkevm.ethereum.L2AccountManager
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.web3j.abi.EventEncoder
import org.web3j.tx.gas.DefaultGasProvider
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.CompletableFuture
import java.util.concurrent.TimeUnit
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
class MessageServiceIntegrationTest {
  private val log: Logger = LogManager.getLogger(this::class.java)

  private val l2RecipientAddress = "0x03dfa322A95039BB679771346Ee2dBfEa0e2B773"
  private val l1Web3Client = L1AccountManager.web3jClient
  private val l2Web3jClient = L2AccountManager.web3jClient
  private lateinit var l2TransactionManager: AsyncFriendlyTransactionManager

  private val messagePollingInterval = 200.milliseconds
  private val maxScrapingTime = 2.seconds
  private val maxMessagesToAnchor = 5u
  private val blockRangeLoopLimit = 100u
  private val receiptPollingInterval = 500.milliseconds

  private lateinit var l1Contract: LineaRollupAsyncFriendly
  private lateinit var l2Contract: L2MessageService

  private fun deployContracts() {
    val l1RollupDeploymentResult = ContractsManager.get().deployLineaRollup().get()
    l1Contract = ContractsManager.get().connectToLineaRollupContract(
      l1RollupDeploymentResult.contractAddress,
      l1RollupDeploymentResult.rollupOperator.txManager
    )
    val l2MessageServiceDeploymentResult = ContractsManager.get().deployL2MessageService().get()
    l2Contract = ContractsManager.get().connectL2MessageService(
      contractAddress = l2MessageServiceDeploymentResult.contractAddress,
      transactionManager = l2MessageServiceDeploymentResult.anchorerOperator.txManager
    )
    l2TransactionManager = l2MessageServiceDeploymentResult.anchorerOperator.txManager
  }

  @Test
  @Timeout(90, timeUnit = TimeUnit.SECONDS)
  fun `test anchoring with RollingHash`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    deployContracts()
    testAnchoredHashesAreReturnedCorrectly(
      l1Contract,
      l2Contract,
      vertx,
      testContext
    )
  }

  fun testAnchoredHashesAreReturnedCorrectly(
    l1Contract: LineaRollupAsyncFriendly,
    l2Contract: L2MessageService,
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val sentMessages = sendMessages(l1Contract)
    val messageAnchoringService = initialiseServices(
      vertx,
      l1Contract.contractAddress,
      l2Contract.contractAddress,
      earliestL1Block = sentMessages.first().l1BlockNumber
    )

    messageAnchoringService.start()
      .thenCompose { _ -> allHashesHaveBeenSet(sentMessages.map { it.messageHash }, 65.seconds, vertx, l2Contract) }
      .thenCompose { success ->
        messageAnchoringService.stop()
          .thenApply {
            testContext.verify {
              assertThat(success).isTrue()
            }.completeNow()
          }
      }.whenException(testContext::failNow)
  }

  private fun initialiseServices(
    vertx: Vertx,
    l1ContractAddress: String,
    l2ContractAddress: String,
    earliestL1Block: ULong
  ): MessageAnchoringService {
    val l1EventQuerier = L1EventQuerierImpl(
      vertx,
      L1EventQuerierImpl.Config(
        pollingInterval = messagePollingInterval,
        maxEventScrapingTime = maxScrapingTime,
        earliestL1Block = earliestL1Block.toBigInteger(),
        maxMessagesToCollect = maxMessagesToAnchor,
        l1MessageServiceAddress = l1ContractAddress,
        "latest",
        blockRangeLoopLimit = blockRangeLoopLimit
      ),
      l1Web3Client
    )

    val l2Querier = L2QuerierImpl(
      l2Web3jClient,
      l2Contract,
      L2QuerierImpl.Config(
        0u,
        5u,
        2u,
        l2ContractAddress
      ),
      vertx
    )

    val l2MessageAnchorer: L2MessageAnchorer = L2MessageAnchorerImpl(
      vertx,
      l2Web3jClient,
      l2Contract,
      L2MessageAnchorerImpl.Config(
        receiptPollingInterval,
        10u,
        0
      )
    )

    val rollupSmartContractClient = RollupSmartContractClientWeb3JImpl(
      Web3JLogsClient(vertx, l1Web3Client),
      lineaRollup = l1Contract
    )

    return MessageAnchoringService(
      MessageAnchoringService.Config(
        pollingInterval = 500.milliseconds,
        maxMessagesToAnchor
      ),
      vertx,
      l1EventQuerier,
      l2MessageAnchorer,
      l2Querier,
      rollupSmartContractClient,
      L2MessageService.load(
        l2ContractAddress,
        l2Web3jClient,
        l2TransactionManager,
        DefaultGasProvider()
      ),
      l2TransactionManager
    )
  }

  private fun allHashesHaveBeenSet(
    messageHashes: List<ByteArray>,
    timeout: Duration,
    vertx: Vertx,
    contract: L2MessageService
  ): SafeFuture<Boolean> {
    return AsyncRetryer.retry(
      vertx,
      timeout = timeout,
      backoffDelay = 200.milliseconds,
      stopRetriesPredicate = { it }
    ) {
      val statuses: List<CompletableFuture<Boolean>> = messageHashes.map { messageHash ->
        contract.inboxL1L2MessageStatus(messageHash).sendAsync()
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

  data class MessageSentResult(
    val l1BlockNumber: ULong,
    val messageHash: ByteArray
  )

  private fun sendMessages(contract: LineaRollupAsyncFriendly): List<MessageSentResult> {
    val baseMessageToSend =
      L1EventQuerierIntegrationTest.L1MessageToSend(
        l2RecipientAddress,
        BigInteger.TEN,
        ByteArray(0),
        BigInteger.valueOf(200001)
      )
    val messagesToSend = listOf(
      baseMessageToSend,
      baseMessageToSend.copy(fee = BigInteger.valueOf(21)),
      baseMessageToSend.copy(value = BigInteger.valueOf(200001)),
      baseMessageToSend.copy(value = BigInteger.valueOf(200001)),
      baseMessageToSend.copy(value = BigInteger.valueOf(200001)),
      baseMessageToSend,
      baseMessageToSend.copy(fee = BigInteger.valueOf(21)),
      baseMessageToSend.copy(value = BigInteger.valueOf(200001)),
      baseMessageToSend.copy(value = BigInteger.valueOf(200001)),
      baseMessageToSend.copy(value = BigInteger.valueOf(200001)),
      baseMessageToSend,
      baseMessageToSend.copy(fee = BigInteger.valueOf(21)),
      baseMessageToSend.copy(value = BigInteger.valueOf(200001)),
      baseMessageToSend.copy(value = BigInteger.valueOf(200001)),
      baseMessageToSend.copy(value = BigInteger.valueOf(200001))
    )

    val futures = messagesToSend.map {
      SafeFuture
        .of(contract.sendMessage(it.recipient, it.fee, it.calldata, it.value).sendAsync())
        .thenApply { transactionReceipt ->
          log.debug("Message has been sent in block {}", transactionReceipt.blockNumber)
          val eventValues = LineaRollup.staticExtractEventParameters(
            LineaRollup.MESSAGESENT_EVENT,
            transactionReceipt.logs.first { log ->
              log.topics.contains(EventEncoder.encode(LineaRollup.MESSAGESENT_EVENT))
            }
          )
          MessageSentResult(
            l1BlockNumber = transactionReceipt.blockNumber.toULong(),
            messageHash = eventValues.indexedValues[2].value as ByteArray
          )
        }
    }
    val emittedEvents: List<MessageSentResult> = SafeFuture.collectAll(futures.stream()).get()

    return emittedEvents
  }
}
