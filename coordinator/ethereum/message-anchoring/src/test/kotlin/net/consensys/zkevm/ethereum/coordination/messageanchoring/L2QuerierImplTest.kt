package net.consensys.zkevm.ethereum.coordination.messageanchoring

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.contract.L2MessageService
import net.consensys.linea.contract.L2MessageService.L1L2MESSAGEHASHESADDEDTOINBOX_EVENT
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.RepeatedTest
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.Mockito
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import org.web3j.abi.EventEncoder
import org.web3j.abi.FunctionEncoder
import org.web3j.abi.FunctionReturnDecoder
import org.web3j.abi.datatypes.DynamicArray
import org.web3j.crypto.Credentials
import org.web3j.crypto.Keys
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.methods.response.EthBlockNumber
import org.web3j.protocol.core.methods.response.EthCall
import org.web3j.protocol.core.methods.response.EthLog
import org.web3j.protocol.core.methods.response.Log
import org.web3j.tx.gas.DefaultGasProvider
import java.math.BigInteger
import java.util.*
import java.util.concurrent.CompletableFuture
import java.util.concurrent.TimeUnit
import kotlin.math.max

@ExtendWith(VertxExtension::class)
class L2QuerierImplTest {
  private val testContractAddress = "0x6d976c9b8ceee705d4fe8699b44e5eb58242f484"
  private val blockNumber = 13
  private val keyPair = Keys.createEcKeyPair()

  @RepeatedTest(10)
  @Timeout(5, timeUnit = TimeUnit.SECONDS)
  fun findLastFinalizedAnchoredEvent_returnsTheLastEvent(vertx: Vertx, testContext: VertxTestContext) {
    val l2ClientMock = mock<Web3j>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
    whenever(l2ClientMock.ethBlockNumber().send().blockNumber)
      .thenReturn(BigInteger.valueOf(blockNumber.toLong()))
    val randomEvents =
      listOf(
        createRandomEventWithHashes(1),
        createRandomEventWithHashes(2),
        createRandomEventWithHashes(3)
      )
    val lastEventData = randomEvents.last().data
    val expectedHash =
      lastEventData.substring(lastEventData.length - Bytes32.ZERO.toUnprefixedHexString().length)

    val mockLogs = mock<EthLog>()
    val logResults: List<EthLog.LogResult<Log>> = randomEvents.map { EthLog.LogResult { it } }
    whenever(mockLogs.logs).thenReturn(logResults)
    whenever(l2ClientMock.ethGetLogs(any()).send()).thenReturn(mockLogs)

    val credentials = Credentials.create(keyPair)
    val messageManager =
      L2MessageService.load(testContractAddress, l2ClientMock, credentials, DefaultGasProvider())
    val l2Querier =
      L2QuerierImpl(
        l2Client = l2ClientMock,
        messageService = messageManager,
        config = L2QuerierImpl.Config(
          blocksToFinalizationL2 = 1u,
          lastHashSearchWindow = 1u,
          contractAddressToListen = testContractAddress
        ),
        vertx = vertx
      )
    l2Querier.findLastFinalizedAnchoredEvent().thenApply {
      testContext
        .verify {
          assertThat(it).isNotNull
          assertThat(it!!.messageHash).isEqualTo(Bytes32.fromHexString(expectedHash))
        }
        .completeNow()
    }.whenException { testContext.failNow(it) }
  }

  @RepeatedTest(10)
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun findLastFinalizedAnchoredEvent_isAbleToFindEventsInThePast(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val l2ClientMock = mock<Web3j>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
    whenever(l2ClientMock.ethBlockNumber().send().blockNumber)
      .thenReturn(BigInteger.valueOf(blockNumber.toLong()))

    val randomEventsForRequests = createRandomEventBatches(6, 4, 6)
    val lastEventData = randomEventsForRequests.last().last().data

    val expectedHash =
      lastEventData.substring(lastEventData.length - Bytes32.ZERO.toUnprefixedHexString().length)

    val mockLogs = mock<EthLog>()
    val logResults: List<EthLog.LogResult<Log>> = randomEventsForRequests.last().map { EthLog.LogResult { it } }
    whenever(mockLogs.logs).thenReturn(logResults)

    val emptyEvents: List<EthLog.LogResult<Log>> = listOf()
    val emptyMockLogs = mock<EthLog>()
    whenever(emptyMockLogs.logs).thenReturn(emptyEvents)

    whenever(l2ClientMock.ethGetLogs(any()).send()).thenAnswer {
      emptyMockLogs
    }.thenAnswer {
      mockLogs
    }

    val credentials = Credentials.create(keyPair)
    val messageManager =
      L2MessageService.load(testContractAddress, l2ClientMock, credentials, DefaultGasProvider())

    val l2Querier =
      L2QuerierImpl(
        l2Client = l2ClientMock,
        messageService = messageManager,
        config = L2QuerierImpl.Config(
          blocksToFinalizationL2 = 1u,
          lastHashSearchWindow = 5u,
          contractAddressToListen = testContractAddress
        ),
        vertx = vertx
      )
    l2Querier.findLastFinalizedAnchoredEvent().thenApply {
      testContext
        .verify {
          assertThat(it).isNotNull
          assertThat(it!!.messageHash).isEqualTo(Bytes32.fromHexString(expectedHash))
        }
        .completeNow()
    }.whenException { testContext.failNow(it) }
  }

  @Test
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun getMessageHashStatus(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val l2ClientMock = mock<Web3j>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
    val credentials = Credentials.create(keyPair)
    val messageManager =
      L2MessageService.load(testContractAddress, l2ClientMock, credentials, DefaultGasProvider())

    val mockBlockNumberReturn = mock<EthBlockNumber>()
    whenever(mockBlockNumberReturn.blockNumber).thenReturn(BigInteger.valueOf(blockNumber.toLong()))
    whenever(l2ClientMock.ethBlockNumber().sendAsync())
      .thenReturn(CompletableFuture.completedFuture(mockBlockNumberReturn))

    val l2Querier =
      L2QuerierImpl(
        l2Client = l2ClientMock,
        messageService = messageManager,
        config = L2QuerierImpl.Config(
          blocksToFinalizationL2 = 1u,
          lastHashSearchWindow = 1u,
          contractAddressToListen = testContractAddress
        ),
        vertx = vertx
      )

    val messageHash = Bytes32.random()
    val mockEthCall = mock<EthCall>()
    whenever(mockEthCall.value).thenReturn("0x0000000000000000000000000000000000000000000000000000000000000001")
    whenever(l2ClientMock.ethCall(any(), any()).send()).thenReturn(mockEthCall)

    l2Querier.getMessageHashStatus(messageHash).thenApply {
      testContext
        .verify {
          assertThat(it).isNotNull
          assertThat(it!!).isEqualTo(BigInteger.valueOf(1))
        }
        .completeNow()
    }.whenException { testContext.failNow(it) }
  }

  private fun createRandomEventBatches(
    numberOfBatches: Int,
    maxEventsPerBatch: Int,
    maxHashesPerEvent: Int
  ): List<List<Log>> {
    return (1..numberOfBatches).map {
      val eventsToGenerate = max(Random().nextInt(maxEventsPerBatch), 1)
      (1..eventsToGenerate).map { createRandomEventWithHashes(maxHashesPerEvent) }
    }
  }

  private fun createRandomEventWithHashes(numberOfRandomHashes: Int): Log {
    val log = Log()
    val randomHashes =
      (0..numberOfRandomHashes)
        .map { Bytes32.random() }
        .map { org.web3j.abi.datatypes.generated.Bytes32(it.toArray()) }
    val eventSignature = EventEncoder.encode(L1L2MESSAGEHASHESADDEDTOINBOX_EVENT)

    log.topics = listOf(eventSignature)
    val data = DynamicArray(org.web3j.abi.datatypes.generated.Bytes32::class.java, randomHashes)
    log.data = FunctionEncoder.encodeConstructor(listOf(data))
    FunctionReturnDecoder.decode(log.data, L1L2MESSAGEHASHESADDEDTOINBOX_EVENT.nonIndexedParameters)
    return log
  }
}
