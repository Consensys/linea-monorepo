package net.consensys.zkevm.ethereum.finalization

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.contract.ZkEvmV2AsyncFriendly
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.RepeatedTest
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.ArgumentMatchers.eq
import org.mockito.Mockito.RETURNS_DEEP_STUBS
import org.mockito.Mockito.atLeast
import org.mockito.Mockito.atLeastOnce
import org.mockito.Mockito.verify
import org.mockito.Mockito.verifyNoMoreInteractions
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import org.web3j.protocol.Web3j
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.milliseconds

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class FinalizationMonitorImplTest {
  private val expectedBlockNumber = BigInteger.TWO
  private val pollingInterval = 20.milliseconds
  private val blocksToFinalization = 2u
  private val config = FinalizationMonitorImpl.Config(pollingInterval, blocksToFinalization)

  @BeforeAll
  fun init() {
    // To warmup assertions otherwise first test may fail
    assertThat(true).isTrue()
  }

  @RepeatedTest(10)
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun start_startsPollingProcess(vertx: Vertx, testContext: VertxTestContext) {
    val mockL2Client = mock<Web3j>(defaultAnswer = RETURNS_DEEP_STUBS)
    val mockL1Client = mock<Web3j>(defaultAnswer = RETURNS_DEEP_STUBS)
    val contractMock = mock<ZkEvmV2AsyncFriendly>(defaultAnswer = RETURNS_DEEP_STUBS)
    whenever(mockL1Client.ethBlockNumber().send().blockNumber).thenReturn(BigInteger.TWO)
    whenever(contractMock.currentL2BlockNumber().send()).thenReturn(expectedBlockNumber)
    val expectedStateRootHash = Bytes32.random()
    whenever(contractMock.stateRootHashes(eq(expectedBlockNumber)).send())
      .thenReturn(expectedStateRootHash.toArray())
    val expectedBlockHash = Bytes32.random()
    whenever(mockL2Client.ethGetBlockByNumber(any(), any()).send().block.hash)
      .thenReturn(expectedBlockHash.toHexString())
    val finalizationMonitorImpl =
      FinalizationMonitorImpl(
        config = config,
        contract = contractMock,
        l1Client = mockL1Client,
        l2Client = mockL2Client,
        vertx = vertx
      )
    finalizationMonitorImpl
      .start()
      .thenApply {
        testContext.verify {
          verify(mockL1Client.ethBlockNumber().send(), atLeastOnce()).blockNumber
          verify(mockL2Client.ethGetBlockByNumber(null, false).send().block, atLeastOnce()).hash
          verify(contractMock.currentL2BlockNumber(), atLeastOnce()).send()
          verify(contractMock.stateRootHashes(expectedBlockNumber), atLeastOnce()).send()
        }
      }
      .thenCompose { finalizationMonitorImpl.stop() }
      .thenApply { testContext.completeNow() }
      .whenException(testContext::failNow)
  }

  @RepeatedTest(10)
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun stop_worksCorrectly(vertx: Vertx, testContext: VertxTestContext) {
    val mockL2Client = mock<Web3j>(defaultAnswer = RETURNS_DEEP_STUBS)
    val mockL1Client = mock<Web3j>(defaultAnswer = RETURNS_DEEP_STUBS)
    val contractMock = mock<ZkEvmV2AsyncFriendly>(defaultAnswer = RETURNS_DEEP_STUBS)
    whenever(mockL1Client.ethBlockNumber().send().blockNumber).thenReturn(BigInteger.TWO)
    whenever(contractMock.currentL2BlockNumber().send()).thenReturn(expectedBlockNumber)
    whenever(contractMock.stateRootHashes(eq(expectedBlockNumber)).send())
      .thenReturn(Bytes32.random().toArray())
    whenever(mockL2Client.ethGetBlockByNumber(any(), eq(false)).send().block.hash)
      .thenReturn(Bytes32.random().toHexString())
    val finalizationMonitorImpl =
      FinalizationMonitorImpl(
        config = config,
        contract = contractMock,
        l1Client = mockL1Client,
        l2Client = mockL2Client,
        vertx = vertx
      )
    val afterStopAssertions = testContext.checkpoint()
    // Doesn't fail if called without start
    finalizationMonitorImpl.stop().thenCompose { finalizationMonitorImpl.start() }
    vertx.setTimer(pollingInterval.inWholeMilliseconds * 3) {
      finalizationMonitorImpl
        .stop()
        .thenCompose {
          // Stop doesn't wait for possible in progress updates
          vertx.setTimer(pollingInterval.inWholeMilliseconds) {
            afterStopAssertions.flag()
            testContext.verify {
              verify(mockL1Client.ethBlockNumber().send(), atLeast(3)).blockNumber
              // it shouldn't poll faster than it should
              verify(mockL2Client.ethGetBlockByNumber(null, false).send().block, atLeast(3)).hash
              verify(contractMock.currentL2BlockNumber(), atLeast(3)).send()
              verify(contractMock.stateRootHashes(expectedBlockNumber), atLeast(3)).send()
            }
          }
          // Idempotency of stop
          finalizationMonitorImpl.stop()
        }
        .thenApply {
          vertx.setTimer(pollingInterval.inWholeMilliseconds * 2) {
            // When it stopped, there are no new polls
            testContext.verify {
              verifyNoMoreInteractions(mockL2Client.ethGetBlockByNumber(null, false).send().block)
              verifyNoMoreInteractions(contractMock.currentL2BlockNumber())
              verifyNoMoreInteractions(contractMock.stateRootHashes(expectedBlockNumber))
            }
            testContext.completeNow()
          }
        }
        .whenException(testContext::failNow)
    }
  }

  @RepeatedTest(10)
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun finalizationUpdatesAreSentToTheRightHandlers(vertx: Vertx, testContext: VertxTestContext) {
    val mockL2Client = mock<Web3j>(defaultAnswer = RETURNS_DEEP_STUBS)
    val mockL1Client = mock<Web3j>(defaultAnswer = RETURNS_DEEP_STUBS)
    val contractMock = mock<ZkEvmV2AsyncFriendly>(defaultAnswer = RETURNS_DEEP_STUBS)
    var blockNumber = 0

    whenever(mockL1Client.ethBlockNumber().send().blockNumber).thenReturn(BigInteger.TWO)
    whenever(contractMock.currentL2BlockNumber().send()).thenAnswer {
      blockNumber += 1
      BigInteger.valueOf(blockNumber.toLong())
    }
    whenever(contractMock.stateRootHashes(any()).send()).thenAnswer {
      intToBytes32(blockNumber).toArray()
    }
    whenever(mockL2Client.ethGetBlockByNumber(any(), eq(false)).send().block.hash).thenAnswer {
      intToBytes32(blockNumber).toHexString()
    }
    val finalizationMonitorImpl =
      FinalizationMonitorImpl(
        config = config,
        contract = contractMock,
        l1Client = mockL1Client,
        l2Client = mockL2Client,
        vertx = vertx
      )
    val updatesReceived1 = mutableListOf<FinalizationMonitor.FinalizationUpdate>()
    val updatesReceived2 = mutableListOf<FinalizationMonitor.FinalizationUpdate>()
    finalizationMonitorImpl.addFinalizationHandler("handler1") {
      SafeFuture.completedFuture(updatesReceived1.add(it))
    }
    val secondFinalisationHandlerSet = testContext.checkpoint()
    finalizationMonitorImpl
      .start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 2) {
          secondFinalisationHandlerSet.flag()
          finalizationMonitorImpl.addFinalizationHandler("handler2") {
            SafeFuture.completedFuture(updatesReceived2.add(it))
          }
        }
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 4) {
          finalizationMonitorImpl
            .stop()
            .thenApply {
              testContext.verify {
                assertThat(updatesReceived1).isNotEmpty
                assertThat(updatesReceived2).isNotEmpty
                assertThat(updatesReceived1).isNotEqualTo(updatesReceived2)
                assertThat(updatesReceived2).contains(updatesReceived1.last())
              }
              testContext.completeNow()
            }
            .whenException(testContext::failNow)
        }
      }
      .whenException(testContext::failNow)
  }

  @RepeatedTest(10)
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun finalizationUpdatesDontCrashTheWholeMonitorInCaseOfErrors(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val mockL2Client = mock<Web3j>(defaultAnswer = RETURNS_DEEP_STUBS)
    val mockL1Client = mock<Web3j>(defaultAnswer = RETURNS_DEEP_STUBS)
    val contractMock = mock<ZkEvmV2AsyncFriendly>(defaultAnswer = RETURNS_DEEP_STUBS)
    var blockNumber = 0
    whenever(mockL1Client.ethBlockNumber().send().blockNumber).thenReturn(BigInteger.TWO)
    whenever(contractMock.currentL2BlockNumber().send()).thenAnswer {
      blockNumber += 1
      BigInteger.valueOf(blockNumber.toLong())
    }
    whenever(contractMock.stateRootHashes(any()).send()).thenAnswer {
      intToBytes32(blockNumber).toArray()
    }
    whenever(mockL2Client.ethGetBlockByNumber(any(), eq(false)).send().block.hash).thenAnswer {
      intToBytes32(blockNumber).toHexString()
    }
    val finalizationMonitorImpl =
      FinalizationMonitorImpl(
        config = config,
        contract = contractMock,
        l1Client = mockL1Client,
        l2Client = mockL2Client,
        vertx = vertx
      )
    val updatesReceived = mutableListOf<FinalizationMonitor.FinalizationUpdate>()
    val updatesWhenFailureHappened: MutableList<FinalizationMonitor.FinalizationUpdate> =
      mutableListOf()
    finalizationMonitorImpl.addFinalizationHandler("handler1") {
      SafeFuture.completedFuture(updatesReceived.add(it))
    }
    val failingFinalizationHandlerSet = testContext.checkpoint()
    finalizationMonitorImpl
      .start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 2) {
          updatesWhenFailureHappened.addAll(updatesReceived)
          finalizationMonitorImpl.addFinalizationHandler("handler2") {
            throw Exception("Finalization callback failure for the testing!")
          }
          failingFinalizationHandlerSet.flag()
        }
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 6) {
          finalizationMonitorImpl
            .stop()
            .thenApply {
              testContext.verify {
                assertThat(updatesReceived).isNotEmpty
                // Difference is only 2 because actual execution might be slower
                assertThat(updatesReceived).hasSizeGreaterThanOrEqualTo(2)
                assertThat(updatesReceived.size - updatesWhenFailureHappened.size)
                  .isGreaterThanOrEqualTo(2)
              }
              testContext.completeNow()
            }
            .whenException(testContext::failNow)
        }
      }
      .whenException(testContext::failNow)
  }

  @RepeatedTest(10)
  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  fun monitorDoesntCrashInCaseOfWeb3Exceptions(vertx: Vertx, testContext: VertxTestContext) {
    val mockL2Client = mock<Web3j>(defaultAnswer = RETURNS_DEEP_STUBS)
    val contractMock = mock<ZkEvmV2AsyncFriendly>(defaultAnswer = RETURNS_DEEP_STUBS)
    val mockL1Client = mock<Web3j>(defaultAnswer = RETURNS_DEEP_STUBS)
    var blockNumber = 0
    whenever(mockL1Client.ethBlockNumber().send().blockNumber).thenReturn(BigInteger.TWO)
    whenever(contractMock.currentL2BlockNumber().send()).thenAnswer {
      blockNumber += 1
      BigInteger.valueOf(blockNumber.toLong())
    }
    whenever(contractMock.stateRootHashes(any()).send()).thenAnswer {
      intToBytes32(blockNumber).toArray()
    }
    whenever(mockL2Client.ethGetBlockByNumber(any(), eq(false)).send().block.hash).thenAnswer {
      intToBytes32(blockNumber).toHexString()
    }
    val finalizationMonitorImpl =
      FinalizationMonitorImpl(
        config = config,
        contract = contractMock,
        l1Client = mockL1Client,
        l2Client = mockL2Client,
        vertx = vertx
      )
    val updatesReceived = mutableListOf<FinalizationMonitor.FinalizationUpdate>()
    val updatesWhenFailureHappened: MutableList<FinalizationMonitor.FinalizationUpdate> =
      mutableListOf()
    val failingMockSet = testContext.checkpoint()
    finalizationMonitorImpl.addFinalizationHandler("handler1") {
      SafeFuture.completedFuture(updatesReceived.add(it))
    }
    finalizationMonitorImpl
      .start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 2) {
          failingMockSet.flag()
          updatesWhenFailureHappened.addAll(updatesReceived)
          whenever(contractMock.currentL2BlockNumber().send())
            .thenThrow(Exception("Web3 client failure for the testing!"))
        }
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 6) {
          finalizationMonitorImpl
            .stop()
            .thenApply {
              testContext.verify {
                assertThat(updatesReceived).isNotEmpty
                assertThat(updatesReceived).isEqualTo(updatesWhenFailureHappened)
              }
              testContext.completeNow()
            }
            .whenException(testContext::failNow)
        }
      }
      .whenException(testContext::failNow)
  }

  private fun intToBytes32(i: Int): Bytes32 {
    return Bytes32.leftPad(Bytes.wrap(listOf(i.toByte()).toByteArray()))
  }
}
