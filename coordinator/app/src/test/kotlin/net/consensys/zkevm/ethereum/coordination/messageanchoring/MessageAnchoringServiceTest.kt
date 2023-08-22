package net.consensys.zkevm.ethereum.coordination.messageanchoring

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.L2MessageService
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.Mockito.RETURNS_DEEP_STUBS
import org.mockito.Mockito.atLeastOnce
import org.mockito.Mockito.verify
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.never
import org.mockito.kotlin.whenever
import org.web3j.protocol.core.methods.response.TransactionReceipt
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.milliseconds

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class MessageAnchoringServiceTest {
  @BeforeAll
  fun init() {
    // To warmup assertions otherwise first test may fail
    Assertions.assertThat(true).isTrue()
  }

  @Test
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun start_startsPollingProcess(vertx: Vertx, testContext: VertxTestContext) {
    val pollingInterval = 10.milliseconds
    val maxMessagesToAnchor = 100u

    val l2MessageService = mock<L2MessageService>(defaultAnswer = RETURNS_DEEP_STUBS)
    whenever(l2MessageService.INBOX_STATUS_UNKNOWN().send()).thenReturn(BigInteger.valueOf(0))

    val mockTransactionManager = mock<AsyncFriendlyTransactionManager>(defaultAnswer = RETURNS_DEEP_STUBS)
    val mockL1Querier = mock<L1EventQuerier>(defaultAnswer = RETURNS_DEEP_STUBS)
    val mockL2MessageAnchorer = mock<L2MessageAnchorer>(defaultAnswer = RETURNS_DEEP_STUBS)
    val mockL2Querier = mock<L2Querier>(defaultAnswer = RETURNS_DEEP_STUBS)

    val foundAnchoredHashEvent = MessageHashAnchoredEvent(Bytes32.random())
    val events = listOf(SendMessageEvent(Bytes32.random()))
    val mockTransactionReceipt = mock<TransactionReceipt>()

    whenever(mockL2Querier.findLastFinalizedAnchoredEvent()).thenReturn(
      SafeFuture.completedFuture(foundAnchoredHashEvent)
    )
    whenever(mockL1Querier.getSendMessageEventsForAnchoredMessage(foundAnchoredHashEvent)).thenReturn(
      SafeFuture.completedFuture(events)
    )
    whenever(mockL2Querier.getMessageHashStatus(events.first().messageHash)).thenReturn(
      SafeFuture.completedFuture(BigInteger.valueOf(0))
    )
    whenever(mockL2MessageAnchorer.anchorMessages(events.map { it.messageHash })).thenReturn(
      SafeFuture.completedFuture(mockTransactionReceipt)
    )

    whenever(mockTransactionReceipt.transactionHash).thenReturn(
      Bytes32.random().toHexString()
    )

    whenever(mockTransactionManager.resetNonce()).thenAnswer { SafeFuture.completedFuture(Unit) }

    val monitor =
      MessageAnchoringService(
        MessageAnchoringService.Config(
          pollingInterval,
          maxMessagesToAnchor
        ),
        vertx,
        mockL1Querier,
        mockL2MessageAnchorer,
        mockL2Querier,
        l2MessageService,
        mockTransactionManager
      )
    monitor.start().thenApply {
      vertx.setTimer(
        100
      ) {
        monitor.stop().thenApply {
          testContext.verify {
            verify(mockTransactionReceipt, atLeastOnce()).transactionHash
            verify(mockL2Querier, atLeastOnce()).getMessageHashStatus(events.first().messageHash)
            verify(mockL2Querier, atLeastOnce()).findLastFinalizedAnchoredEvent()
            verify(mockL1Querier, atLeastOnce()).getSendMessageEventsForAnchoredMessage(foundAnchoredHashEvent)
            verify(mockTransactionManager, atLeastOnce()).resetNonce()
          }
            .completeNow()
        }
      }
    }
  }

  @Test
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun start_startsPollingProcessAndLimitsReturnedEvents(vertx: Vertx, testContext: VertxTestContext) {
    val pollingInterval = 10.milliseconds
    val maxMessagesToAnchor = 2u

    val l2MessageService = mock<L2MessageService>(defaultAnswer = RETURNS_DEEP_STUBS)
    whenever(l2MessageService.INBOX_STATUS_UNKNOWN().send()).thenReturn(BigInteger.valueOf(0))

    val mockTransactionManager = mock<AsyncFriendlyTransactionManager>(defaultAnswer = RETURNS_DEEP_STUBS)
    val mockL1Querier = mock<L1EventQuerier>(defaultAnswer = RETURNS_DEEP_STUBS)
    val mockL2MessageAnchorer = mock<L2MessageAnchorer>(defaultAnswer = RETURNS_DEEP_STUBS)
    val mockL2Querier = mock<L2Querier>(defaultAnswer = RETURNS_DEEP_STUBS)

    val foundAnchoredHashEvent = MessageHashAnchoredEvent(Bytes32.random())
    val events = listOf(
      SendMessageEvent(Bytes32.random()),
      SendMessageEvent(Bytes32.random()),
      SendMessageEvent(Bytes32.random()),
      SendMessageEvent(Bytes32.random())
    )

    val mockTransactionReceipt = mock<TransactionReceipt>()

    whenever(mockL2Querier.findLastFinalizedAnchoredEvent()).thenReturn(
      SafeFuture.completedFuture(foundAnchoredHashEvent)
    )
    whenever(mockL1Querier.getSendMessageEventsForAnchoredMessage(foundAnchoredHashEvent)).thenReturn(
      SafeFuture.completedFuture(events)
    )
    whenever(mockL2Querier.getMessageHashStatus(any())).thenReturn(
      SafeFuture.completedFuture(BigInteger.valueOf(0))
    )
    whenever(mockL2MessageAnchorer.anchorMessages(any())).thenReturn(
      SafeFuture.completedFuture(mockTransactionReceipt)
    )
    whenever(mockTransactionReceipt.transactionHash).thenReturn(
      Bytes32.random().toHexString()
    )

    whenever(mockTransactionManager.resetNonce()).thenAnswer { SafeFuture.completedFuture(Unit) }

    val monitor =
      MessageAnchoringService(
        MessageAnchoringService.Config(
          pollingInterval,
          maxMessagesToAnchor
        ),
        vertx,
        mockL1Querier,
        mockL2MessageAnchorer,
        mockL2Querier,
        l2MessageService,
        mockTransactionManager
      )
    monitor.start().thenApply {
      vertx.setTimer(
        100
      ) {
        monitor.stop().thenApply {
          testContext.verify {
            verify(mockL2MessageAnchorer, atLeastOnce()).anchorMessages(events.take(2).map { it.messageHash })
            verify(mockL2MessageAnchorer, never()).anchorMessages(events.take(4).map { it.messageHash })
            verify(mockL2Querier, atLeastOnce()).getMessageHashStatus(events.first().messageHash)
            verify(mockL2Querier, atLeastOnce()).getMessageHashStatus(events[1].messageHash)
            verify(mockL2Querier, atLeastOnce()).getMessageHashStatus(events[2].messageHash)
            verify(mockL2Querier, atLeastOnce()).getMessageHashStatus(events[3].messageHash)
            verify(mockL2Querier, atLeastOnce()).findLastFinalizedAnchoredEvent()
            verify(mockL1Querier, atLeastOnce()).getSendMessageEventsForAnchoredMessage(foundAnchoredHashEvent)
            verify(mockTransactionReceipt, atLeastOnce()).transactionHash
            verify(mockTransactionManager, atLeastOnce()).resetNonce()
          }
            .completeNow()
        }
      }
    }
  }
}
