package net.consensys.zkevm.ethereum.coordination.messageanchoring

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.ZkEvmV2
import net.consensys.linea.contract.ZkEvmV2AsyncFriendly
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.web3j.abi.EventEncoder
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.http.HttpService
import org.web3j.tx.gas.DefaultGasProvider
import org.web3j.tx.response.PollingTransactionReceiptProcessor
import org.web3j.utils.Async
import java.math.BigInteger
import java.util.*
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
class L1EventQuerierIntegrationTest {
  private val testZkEvmContractAddress = System.getProperty("ZkEvmV2Address")
  private val l2RecipientAddress = "0x03dfa322A95039BB679771346Ee2dBfEa0e2B773"
  private val l1RpcEndpoint = "http://localhost:8445"

  // WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
  private val testAccountPrivateKey = "202454d1b4e72c41ebf58150030f649648d3cf5590297fb6718e27039ed9c86d"
  private val blockRangeLoopLimit = 5u
  private val maxMessagesToCollect = 100u

  data class L1MessageToSend(
    val recipient: String,
    val fee: BigInteger,
    val calldata: ByteArray,
    val value: BigInteger
  ) {
    override fun equals(other: Any?): Boolean {
      if (this === other) return true
      if (javaClass != other?.javaClass) return false

      other as L1MessageToSend

      if (recipient != other.recipient) return false
      if (fee != other.fee) return false
      if (!calldata.contentEquals(other.calldata)) return false
      return value == other.value
    }

    override fun hashCode(): Int {
      var result = recipient.hashCode()
      result = 31 * result + fee.hashCode()
      result = 31 * result + calldata.contentHashCode()
      result = 31 * result + value.hashCode()
      return result
    }
  }

  private fun connectToZkevmContract(web3Client: Web3j): ZkEvmV2AsyncFriendly {
    val credentials = Credentials.create(testAccountPrivateKey)
    val chainId = web3Client.ethChainId().send().chainId.toLong()
    val pollingTransactionReceiptProcessor = PollingTransactionReceiptProcessor(web3Client, 1000, 40)
    val transactionManager = AsyncFriendlyTransactionManager(
      web3Client,
      credentials,
      chainId,
      pollingTransactionReceiptProcessor
    )

    return ZkEvmV2AsyncFriendly.load(testZkEvmContractAddress, web3Client, transactionManager, DefaultGasProvider())
  }

  @Test
  @Timeout(45, timeUnit = TimeUnit.SECONDS)
  fun `l1Event querier returns events from the given hash`(vertx: Vertx, testContext: VertxTestContext) {
    val web3Client = Web3j.build(HttpService(l1RpcEndpoint), 1000, Async.defaultExecutorService())
    val contract = connectToZkevmContract(web3Client)

    val baseMessageToSend =
      L1MessageToSend(l2RecipientAddress, BigInteger.ZERO, ByteArray(0), BigInteger.valueOf(100001))
    val messagesToSend = listOf(
      baseMessageToSend,
      baseMessageToSend.copy(value = BigInteger.valueOf(100001)),
      baseMessageToSend.copy(value = BigInteger.valueOf(100002)),
      baseMessageToSend.copy(value = BigInteger.valueOf(100003)),
      baseMessageToSend.copy(value = BigInteger.valueOf(100005))
    )

    val earlierEmittedParallelEvents = messagesToSend.map {
      contract.sendMessage(it.recipient, it.fee, it.calldata, it.value).sendAsync()
        .thenApply { receipt ->
          Pair(
            ZkEvmV2.staticExtractEventParameters(
              ZkEvmV2.MESSAGESENT_EVENT,
              receipt.logs.first { log ->
                log.topics.contains(EventEncoder.encode(ZkEvmV2.MESSAGESENT_EVENT))
              }
            ),
            receipt
          )
        }
    }

    val earlierEvents = earlierEmittedParallelEvents.map { it.get() }

    val hashIndexToQueryFrom = 2
    val hashInTheMiddle = earlierEvents[hashIndexToQueryFrom].let {
      Bytes32.wrap(it.first.indexedValues[2].value as ByteArray)
    }

    val laterEmittedEvents = messagesToSend.map {
      contract.sendMessage(it.recipient, it.fee, it.calldata, it.value).sendAsync()
        .thenApply { receipt ->
          Pair(
            ZkEvmV2.staticExtractEventParameters(
              ZkEvmV2.MESSAGESENT_EVENT,
              receipt.logs.first { log ->
                log.topics.contains(EventEncoder.encode(ZkEvmV2.MESSAGESENT_EVENT))
              }
            ),
            receipt
          )
        }
    }

    val laterEvents = laterEmittedEvents.map { it.get() }
    val multiPostBlockLoopLimit = 20u
    val l1QuerierImpl = L1EventQuerierImpl(
      vertx,
      L1EventQuerierImpl.Config(
        2.seconds,
        16.seconds,
        BigInteger.ZERO,
        maxMessagesToCollect,
        testZkEvmContractAddress,
        "latest",
        multiPostBlockLoopLimit
      ),
      web3Client
    )

    // we should have events 4 and 5 (count = 5, less hashIndexToQueryFrom, less 1 (0 based))
    val allExpectedHashesInOrder: MutableList<SendMessageEvent> =
      earlierEvents.drop(hashIndexToQueryFrom + 1).take(maxMessagesToCollect.toInt()).map {
        SendMessageEvent(
          Bytes32.wrap(it.first.indexedValues[2].value as ByteArray)
        )
      }.toMutableList()

    l1QuerierImpl.getSendMessageEventsForAnchoredMessage(MessageHashAnchoredEvent(hashInTheMiddle)).thenApply {
      // we should have events 1,2,3,4 and 5
      val expectedHashes = laterEvents.map {
        SendMessageEvent(
          Bytes32.wrap(it.first.indexedValues[2].value as ByteArray)
        )
      }.toMutableList()

      // we should have all 7
      allExpectedHashesInOrder.addAll(expectedHashes)

      testContext.verify {
        assertThat(it).isEqualTo(allExpectedHashesInOrder)
      }.completeNow()
    }.whenException(testContext::failNow)
  }

  @Disabled
  @Timeout(1, timeUnit = TimeUnit.MINUTES)
  fun `l1Event querier returns events with more than 10000 messages`(vertx: Vertx, testContext: VertxTestContext) {
    val web3Client = Web3j.build(HttpService(l1RpcEndpoint), 1000, Async.defaultExecutorService())
    val contract = connectToZkevmContract(web3Client)

    val baseMessageToSend =
      L1MessageToSend(l2RecipientAddress, BigInteger.ZERO, ByteArray(0), BigInteger.valueOf(100001))

    val messagesToSend = (1..10025).map { baseMessageToSend.copy(value = BigInteger.valueOf(100001)) }

    val earlierEmittedParallelEvents = messagesToSend.map {
      contract.sendMessage(it.recipient, it.fee, it.calldata, it.value).sendAsync()
        .thenApply { receipt ->
          Pair(
            ZkEvmV2.staticExtractEventParameters(
              ZkEvmV2.MESSAGESENT_EVENT,
              receipt.logs.first { log ->
                log.topics.contains(EventEncoder.encode(ZkEvmV2.MESSAGESENT_EVENT))
              }
            ),
            receipt
          )
        }
    }

    val events = earlierEmittedParallelEvents.map { it.get() }

    val hashIndexToQueryFrom = 2
    val hashToTake = events[hashIndexToQueryFrom].let {
      Bytes32.wrap(it.first.indexedValues[2].value as ByteArray)
    }

    val l1QuerierImpl = L1EventQuerierImpl(
      vertx,
      L1EventQuerierImpl.Config(
        2.seconds,
        16.seconds,
        BigInteger.ZERO,
        maxMessagesToCollect,
        testZkEvmContractAddress,
        "latest",
        blockRangeLoopLimit
      ),
      web3Client
    )

    val allExpectedHashesInOrder =
      events.drop(hashIndexToQueryFrom + 1).take(maxMessagesToCollect.toInt()).map {
        Bytes32.wrap(it.first.indexedValues[2].value as ByteArray)
      }

    l1QuerierImpl.getSendMessageEventsForAnchoredMessage(MessageHashAnchoredEvent(hashToTake)).thenApply {
      val foundHashes = it.map { evt ->
        evt.messageHash
      }

      // we should have all the hashes meeting the maxMessagesToCollect
      testContext.verify {
        assertThat(it.count()).isEqualTo(maxMessagesToCollect)
        assertThat(foundHashes).isEqualTo(allExpectedHashesInOrder)
      }.completeNow()
    }.whenException(testContext::failNow)
  }
}
