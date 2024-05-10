package net.consensys.zkevm.ethereum.coordination.messageanchoring

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.EIP1559GasProvider
import net.consensys.linea.contract.L2MessageService
import net.consensys.linea.contract.LineaRollup
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.linea.contract.RollupSmartContractClientWeb3JImpl
import net.consensys.linea.contract.Web3JLogsClient
import net.consensys.zkevm.coordinator.clients.RollupSmartContractClient
import net.consensys.zkevm.ethereum.ContractsManager
import net.consensys.zkevm.ethereum.L1AccountManager
import net.consensys.zkevm.ethereum.L2AccountManager
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions
import org.assertj.core.data.Offset
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.junit.jupiter.api.extension.ExtendWith
import org.web3j.abi.EventEncoder
import org.web3j.protocol.Web3j
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.ExecutionException
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
class L2MessageAnchorerIntegrationTest {
  private lateinit var testLineaRollupContractAddress: String
  private lateinit var testL2MessageManagerContractAddress: String
  private val l2RecipientAddress = "0x03dfa322A95039BB679771346Ee2dBfEa0e2B773"
  private val blockRangeLoopLimit = 100u

  private val gasLimit = BigInteger.valueOf(2_500_000)
  private val feeHistoryBlockCount = 4u
  private val feeHistoryRewardPercentile = 15.0
  private val maxFeePerGasCap = BigInteger.valueOf(10000)
  private lateinit var l1Web3jClient: Web3j
  private lateinit var l2Web3jClient: Web3j
  private lateinit var l2TransactionManager: AsyncFriendlyTransactionManager
  private lateinit var l2Contract: L2MessageService
  private lateinit var l1Contract: LineaRollupAsyncFriendly
  private lateinit var rollupSmartContractClient: RollupSmartContractClient
  private lateinit var eip1559GasProvider: EIP1559GasProvider
  private val messageAnchorerConfig = L2MessageAnchorerImpl.Config(
    receiptPollingInterval = 200.milliseconds,
    maxReceiptRetries = 100u,
    blocksToFinalisation = 0
  )
  private lateinit var l2MessageAnchorer: L2MessageAnchorerImpl

  @BeforeEach
  fun beforeEach(
    vertx: Vertx
  ) {
    val deploymentResult = ContractsManager.get().deployRollupAndL2MessageService().get()
    testLineaRollupContractAddress = deploymentResult.lineaRollup.contractAddress
    testL2MessageManagerContractAddress = deploymentResult.l2MessageService.contractAddress
    l1Web3jClient = L1AccountManager.web3jClient
    l2Web3jClient = L2AccountManager.web3jClient
    l2TransactionManager = deploymentResult.l2MessageService.anchorerOperator.txManager
    l1Contract = ContractsManager.get().connectToLineaRollupContract(
      deploymentResult.lineaRollup.contractAddress,
      deploymentResult.lineaRollup.rollupOperator.txManager
    )

    eip1559GasProvider = EIP1559GasProvider(
      l2Web3jClient,
      EIP1559GasProvider.Config(
        gasLimit = gasLimit,
        maxFeePerGasCap = maxFeePerGasCap,
        feeHistoryBlockCount = feeHistoryBlockCount,
        feeHistoryRewardPercentile = feeHistoryRewardPercentile
      )
    )

    l2Contract = ContractsManager.get().connectL2MessageService(
      testL2MessageManagerContractAddress,
      l2Web3jClient,
      l2TransactionManager,
      gasProvider = eip1559GasProvider,
      mapOf("3B174434" to "MessageHashesListLengthHigherThanOneHundred")
    )

    l2MessageAnchorer = L2MessageAnchorerImpl(
      vertx,
      l2Web3jClient,
      l2Contract,
      messageAnchorerConfig
    )
    val logsClient = Web3JLogsClient(vertx, l1Web3jClient)
    rollupSmartContractClient = RollupSmartContractClientWeb3JImpl(logsClient, l1Contract)
  }

  @Test
  @Timeout(2, timeUnit = TimeUnit.MINUTES)
  fun `can send a message on L1 and see the hash on L2`(vertx: Vertx, testContext: VertxTestContext) {
    val expectedHash = Bytes32.random()

    l2MessageAnchorer.anchorMessages(createRandomSendMessageEvents(2UL), Bytes32.random().toArray())
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
    val emittedEvents: List<LineaRollup.MessageSentEventResponse> =
      SafeFuture.collectAll(
        messagesToSend.map { message ->
          SafeFuture.of(
            l1Contract.sendMessage(message.recipient, message.fee, message.calldata, message.value).sendAsync()
          )
            .thenApply { txReceipt ->
              LineaRollup.getMessageSentEventFromLog(
                txReceipt.logs.first { log ->
                  log.topics.contains(EventEncoder.encode(LineaRollup.MESSAGESENT_EVENT))
                }
              )
            }
        }.stream()
      ).get()

    val hashInTheMiddle = emittedEvents[1].let { Bytes32.wrap(it._messageHash) }

    val l1QuerierImpl = L1EventQuerierImpl(
      vertx,
      L1EventQuerierImpl.Config(
        pollingInterval = 200.milliseconds,
        maxEventScrapingTime = 2.seconds,
        earliestL1Block = BigInteger.ZERO,
        maxMessagesToCollect = 100u,
        l1MessageServiceAddress = testLineaRollupContractAddress,
        finalized = "latest",
        blockRangeLoopLimit = blockRangeLoopLimit
      ),
      l1Web3jClient = l1Web3jClient
    )
    val expectedHash = Bytes32.wrap(emittedEvents.last()._messageHash)

    l1QuerierImpl.getSendMessageEventsForAnchoredMessage(MessageHashAnchoredEvent(hashInTheMiddle))
      .thenApply { events ->
        val rollingHash = rollupSmartContractClient.getMessageRollingHash(
          events.last().messageNumber.toLong()
        ).get()
        l2MessageAnchorer.anchorMessages(events, rollingHash)
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

  @Test
  @Timeout(2, timeUnit = TimeUnit.MINUTES)
  fun `anchor messages and gas limit set to estimation`(vertx: Vertx, testContext: VertxTestContext) {
    val expectedHash = Bytes32.random()

    l2MessageAnchorer
      .anchorMessages(createRandomSendMessageEvents(10UL), Bytes32.random().toArray())
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
          Assertions.assertThat(transactionReceipt.gasUsed)
            .isCloseTo(BigInteger.valueOf(56_000L), Offset.offset(BigInteger.valueOf(5_000L)))
          Assertions.assertThat(eip1559GasProvider.getGasLimit("addL1L2MessageHashes"))
            .isEqualTo(gasLimit)
        }.completeNow()
      }.whenException(testContext::failNow)
  }

  @Test
  @Timeout(20, timeUnit = TimeUnit.SECONDS)
  fun `anchor messages gas estimation returns informative error`(vertx: Vertx) {
    val exception = assertThrows<ExecutionException> {
      l2MessageAnchorer.anchorMessages(createRandomSendMessageEvents(101UL), Bytes32.random().toArray()).get()
    }
    Assertions.assertThat(exception.message).contains(
      "3b174434",
      "MessageHashesListLengthHigherThanOneHundred",
      "Execution reverted"
    )
  }
}
