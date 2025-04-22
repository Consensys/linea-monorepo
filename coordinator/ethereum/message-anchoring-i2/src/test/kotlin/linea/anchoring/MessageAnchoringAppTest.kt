package linea.anchoring

import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import linea.anchoring.events.L1RollingHashUpdatedEvent
import linea.anchoring.events.L2RollingHashUpdatedEvent
import linea.anchoring.events.MessageSentEvent
import linea.anchoring.fakes.FakeL2MessageService
import linea.domain.BlockParameter
import linea.domain.RetryConfig
import linea.ethapi.FakeEthApiClient
import linea.kotlin.decodeHex
import linea.kotlin.encodeHex
import linea.kotlin.toHexStringUInt256
import linea.log4j.configureLoggers
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

@ExtendWith(VertxExtension::class)
class MessageAnchoringAppTest {

  private val L1_CONTRACT_ADDRESS = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa01"
  private val L2_CONTRACT_ADDRESS = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa02"
  private lateinit var l1Client: FakeEthApiClient
  private lateinit var l2MessageService: FakeL2MessageService
  private val l1MessageSentLogs = listOf(
    createL1MessageSentV1Logs(
      blockNumber = 100UL,
      contractAddress = L1_CONTRACT_ADDRESS,
      messageNumber = 1UL,
      messageHash = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa01".decodeHex(),
      rollingHash = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb01".decodeHex()
    ).asList(),
    createL1MessageSentV1Logs(
      blockNumber = 200UL,
      contractAddress = L1_CONTRACT_ADDRESS,
      messageNumber = 2UL,
      messageHash = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa02".decodeHex(),
      rollingHash = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb02".decodeHex()
    ).asList(),
    createL1MessageSentV1Logs(
      blockNumber = 200UL,
      contractAddress = L1_CONTRACT_ADDRESS,
      messageNumber = 3UL,
      messageHash = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa03".decodeHex(),
      rollingHash = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb03".decodeHex()
    ).asList(),
    createL1MessageSentV1Logs(
      blockNumber = 400UL,
      contractAddress = L1_CONTRACT_ADDRESS,
      messageNumber = 4UL,
      messageHash = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa04".decodeHex(),
      rollingHash = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb04".decodeHex()
    ).asList()
  )
  private lateinit var vertx: Vertx

  @BeforeEach
  fun setUp(vertx: Vertx) {
    this.vertx = vertx
    l2MessageService = FakeL2MessageService(contractAddress = L2_CONTRACT_ADDRESS)
    l1Client = FakeEthApiClient(
      initialLogsDb = emptySet(),
      topicsTranslation = mapOf(
        L1RollingHashUpdatedEvent.topic to "L1RollingHashUpdatedEvent",
        MessageSentEvent.topic to "MessageSentEvent",
        L2RollingHashUpdatedEvent.topic to "L2RollingHashUpdatedEvent"
      ),
      log = LogManager.getLogger("FakeEthApiClient")
    )

    configureLoggers(
      rootLevel = Level.INFO,
      "FakeEthApiClient" to Level.INFO
    )
  }

  private fun createApp(
    anchoringTickInterval: Duration = 100.milliseconds,
    l1PollingInterval: Duration = 100.milliseconds,
    l1EventSearchBlockChunk: UInt = 1000u,
    l1EventPollingTimeout: Duration = 2.seconds,
    l1SuccessBackoffDelay: Duration = 1.milliseconds,
    messageQueueCapacity: UInt = 100u,
    maxMessagesToAnchorPerL2Transaction: UInt = 10u
  ): MessageAnchoringApp {
    return MessageAnchoringApp(
      vertx = vertx,
      config = MessageAnchoringApp.Config(
        l1PollingInterval = l1PollingInterval,
        l1SuccessBackoffDelay = l1SuccessBackoffDelay,
        l1ContractAddress = L1_CONTRACT_ADDRESS,
        l2HighestBlockTag = BlockParameter.Tag.LATEST,
        anchoringTickInterval = anchoringTickInterval,
        l1RequestRetryConfig = RetryConfig.noRetries,
        l1EventPollingTimeout = l1EventPollingTimeout,
        l1EventSearchBlockChunk = l1EventSearchBlockChunk,
        messageQueueCapacity = messageQueueCapacity,
        maxMessagesToAnchorPerL2Transaction = maxMessagesToAnchorPerL2Transaction
      ),
      l1EthApiClient = l1Client,
      l2MessageService = l2MessageService
    )
  }

  @Test
  fun `should anchor messages from fresh deployment`() {
    l1Client.setLogs(l1MessageSentLogs.flatten())

    val anchoringApp = createApp(anchoringTickInterval = 100.milliseconds, l1EventSearchBlockChunk = 100u)
    anchoringApp.start().get()
    l1Client.setLatestBlockTag(l1MessageSentLogs.last().last().blockNumber + 10UL)
    l1Client.setFinalizedBlockTag(l1MessageSentLogs.last().last().blockNumber + 10UL)

    await()
      .atMost(5.seconds.toJavaDuration())
      .untilAsserted {
        assertThat(l2MessageService.getLastAnchoredL1MessageNumber(block = BlockParameter.Tag.LATEST).get())
          .isEqualTo(4UL)
        assertThat(l2MessageService.getLastAnchoredRollingHash())
          .isEqualTo("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb04".decodeHex())
      }

    anchoringApp.stop().get()
  }

  @Test
  fun `should anchor messages resuming from latest anchored message on L2`() {
    l2MessageService.setLastAnchoredL1Message(
      l1MessageNumber = 1UL,
      rollingHash = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb01".decodeHex()
    )
    val anchoringApp = createApp(anchoringTickInterval = 100.milliseconds, l1EventSearchBlockChunk = 100u)
    anchoringApp.start().get()
    l1Client.setLogs(l1MessageSentLogs.flatten())
    l1Client.setFinalizedBlockTag(l1MessageSentLogs.last().last().blockNumber + 10UL)

    await()
      .atMost(5.seconds.toJavaDuration())
      .untilAsserted {
        assertThat(l2MessageService.getLastAnchoredL1MessageNumber(block = BlockParameter.Tag.LATEST).get())
          .isEqualTo(4UL)
        assertThat(l2MessageService.getLastAnchoredRollingHash())
          .isEqualTo("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb04".decodeHex())

        assertThat(l2MessageService.getAnchoredMessageHashes().map { it.encodeHex() })
          .isEqualTo(
            listOf(
              "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa02",
              "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa03",
              "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa04"
            )
          )
      }

    anchoringApp.stop().get()
  }

  @Test
  fun `should anchor messages in chunks when too many messages are waiting on L1`() {
    val ethLogs = (1UL..100UL).mapIndexed { index, blockNumber ->
      val messageNumber = (index + 1).toULong()
      createL1MessageSentV1Logs(
        blockNumber = blockNumber,
        contractAddress = L1_CONTRACT_ADDRESS,
        messageNumber = messageNumber,
        messageHash = messageNumber.toHexStringUInt256().decodeHex(),
        rollingHash = messageNumber.toHexStringUInt256().decodeHex()
      ).asList()
    }
    val anchoringApp = createApp(
      l1EventSearchBlockChunk = 10u,
      maxMessagesToAnchorPerL2Transaction = 20u,
      anchoringTickInterval = 1.seconds
    )
    anchoringApp.start().get()

    l1Client.setLogs(ethLogs.flatten())
    l1Client.setFinalizedBlockTag(ethLogs.last().last().blockNumber + 10UL)

    val lastMessageSentEvent = MessageSentEvent.fromEthLog(ethLogs.last().last())
    await()
      .atMost(10.seconds.toJavaDuration())
      .untilAsserted {
        assertThat(l2MessageService.getLastAnchoredL1MessageNumber(block = BlockParameter.Tag.LATEST).get())
          .isEqualTo(lastMessageSentEvent.event.messageNumber)
      }

    anchoringApp.stop().get()
  }

  @Test
  fun `should be resilient to low activity or downtimes and find messages on L1 spread apart and anchor them`() {
    // put Log every 1000 Blocks on L1
    // search chunk is 10 so it needs 100 queries to find 1 log on L1
    // slow timeout 1
    val ethLogs = (1UL..100_000UL).step(1000).mapIndexed { index, blockNumber ->
      val messageNumber = index.toULong() + 1UL
      createL1MessageSentV1Logs(
        blockNumber = blockNumber,
        contractAddress = L1_CONTRACT_ADDRESS,
        messageNumber = messageNumber,
        messageHash = messageNumber.toHexStringUInt256().decodeHex(),
        rollingHash = messageNumber.toHexStringUInt256().decodeHex()
      ).asList()
    }
    val anchoringApp = createApp(
      l1PollingInterval = 1.milliseconds,
      l1EventSearchBlockChunk = 10u,
      l1EventPollingTimeout = 50.milliseconds,
      l1SuccessBackoffDelay = 20.milliseconds,
      maxMessagesToAnchorPerL2Transaction = 50u,
      anchoringTickInterval = 1.seconds
    )

    l1Client.setLogs(ethLogs.flatten())
    l1Client.setFinalizedBlockTag(ethLogs.last().last().blockNumber + 10UL)

    anchoringApp.start().get()
    val lastMessageSentEvent = MessageSentEvent.fromEthLog(ethLogs.last().last())
    await()
      .atMost(10.minutes.toJavaDuration())
      .untilAsserted {
        assertThat(l2MessageService.getLastAnchoredL1MessageNumber(block = BlockParameter.Tag.LATEST).get())
          .isEqualTo(lastMessageSentEvent.event.messageNumber)
      }

    assertThat(l2MessageService.getAnchoredMessageHashes())
      .isEqualTo(ethLogs.map { MessageSentEvent.fromEthLog(it.last()).event.messageHash })

    anchoringApp.stop().get()
  }
}
