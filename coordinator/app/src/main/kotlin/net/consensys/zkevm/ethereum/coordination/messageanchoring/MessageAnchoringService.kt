package net.consensys.zkevm.ethereum.coordination.messageanchoring

import io.vertx.core.TimeoutStream
import io.vertx.core.Vertx
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.L2MessageService
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import kotlin.time.Duration

class MessageAnchoringService(
  private val config: Config,
  private val vertx: Vertx,
  private val l1EventQuerier: L1EventQuerier,
  private val l2MessageAnchorer: L2MessageAnchorer,
  private val l2Querier: L2Querier,
  l2MessageService: L2MessageService,
  private val transactionManager: AsyncFriendlyTransactionManager
) {
  class Config(
    val pollingInterval: Duration,
    val maxMessagesToAnchor: UInt
  )

  private val log: Logger = LogManager.getLogger(this::class.java)

  @Volatile
  private lateinit var monitorStream: TimeoutStream
  private val inboxStatusUnknown: BigInteger = l2MessageService.INBOX_STATUS_UNKNOWN().send()

  private fun tick(): SafeFuture<Unit> {
    return l2Querier
      .findLastFinalizedAnchoredEvent()
      .thenCompose(l1EventQuerier::getSendMessageEventsForAnchoredMessage)
      .thenCompose { sentMessages ->
        SafeFuture.collectAll(
          sentMessages
            .map { sendMessageEvent ->
              l2Querier.getMessageHashStatus(sendMessageEvent.messageHash).thenApply { status
                ->
                sendMessageEvent to status
              }
            }
            .stream()
        )
      }
      .thenApply { eventAndStatusPairs ->
        eventAndStatusPairs
          .filter { eventAndStatus -> eventAndStatus.second == inboxStatusUnknown }
          .map { it.first }
      }
      .thenCompose { eventsToAnchor ->
        if (eventsToAnchor.isNotEmpty()) {
          log.debug("Found {} un-anchored events", eventsToAnchor.count())
          transactionManager.resetNonce().thenCompose {
            l2MessageAnchorer.anchorMessages(
              eventsToAnchor.take(config.maxMessagesToAnchor.toInt())
                .map { transactionReceipt -> transactionReceipt.messageHash }
            )
          }.thenApply { transactionReceipt ->
            log.info("Message anchoring transactionHash=${transactionReceipt.transactionHash}")
          }
        } else {
          log.debug("Skipping anchoring as there are no hashes")
          SafeFuture.completedFuture(Unit)
        }
      }
  }

  fun start(): SafeFuture<Unit> {
    monitorStream =
      vertx.periodicStream(config.pollingInterval.inWholeMilliseconds).handler {
        try {
          monitorStream.pause()
          tick()
            .whenComplete { _, error ->
              error?.let {
                log.error("Failed to anchor messages: errorMessage={}", error.message, error)
              }
              monitorStream.resume()
            }
        } catch (th: Throwable) {
          log.error("Failed to trigger message anchoring: errorMessage={}", th.message, th)
          monitorStream.resume()
        }
      }
    return SafeFuture.completedFuture(Unit)
  }

  fun stop(): SafeFuture<Unit> {
    if (this::monitorStream.isInitialized) {
      return SafeFuture.completedFuture(monitorStream.cancel())
    } else {
      throw IllegalStateException("Message Anchoring Service hasn't been started to stop it!")
    }
  }
}
