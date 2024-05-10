package net.consensys.zkevm.ethereum.coordination.aggregation

import io.vertx.core.Handler
import io.vertx.core.Vertx
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.zkevm.LongRunningService
import net.consensys.zkevm.domain.BlobCounters
import net.consensys.zkevm.ethereum.coordination.blockcreation.SafeBlockProvider
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.CompletableFuture
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

class AggregationTriggerCalculatorByDeadline(
  private val config: Config,
  private val clock: Clock = Clock.System,
  private val latestBlockProvider: SafeBlockProvider
) : DeferredAggregationTriggerCalculator {
  data class Config(val aggregationDeadline: Duration)

  // Timestamp of first block in aggregation
  private var aggregationStartTimestamp: Instant? = null
  private var lastBlockNumber: ULong? = null
  private var aggregationTriggerHandler = AggregationTriggerHandler.NOOP_HANDLER
  private val log: Logger = LogManager.getLogger(this::class.java)

  @Synchronized
  fun checkAggregation(): SafeFuture<Unit> {
    val now = clock.now()
    log.trace(
      "Checking deadline: aggregationStartTimestamp={} timeElapsed={} deadline={}",
      aggregationStartTimestamp,
      aggregationStartTimestamp?.let { now.minus(it) } ?: 0.seconds,
      config.aggregationDeadline
    )

    val deadlineReached =
      aggregationStartTimestamp != null && now > aggregationStartTimestamp!!.plus(config.aggregationDeadline)

    if (!deadlineReached) {
      return SafeFuture.completedFuture(Unit)
    }

    return latestBlockProvider.getLatestSafeBlockHeader().thenApply {
      if (it.number == lastBlockNumber) {
        log.trace("Aggregation Deadline reached at block {}", lastBlockNumber)
        aggregationTriggerHandler.onAggregationTrigger(AggregationTriggerType.TIME_LIMIT)
          .whenComplete { _, error ->
            error?.let {
              log.error("Aggregation trigger handler failed: errorMessage: {}", error.message, error)
            }
            reset()
          }
      }
    }.whenException { th ->
      log.warn(
        "SafeBlock request failed. Will retry aggregation deadline on next tick errorMessage={}",
        th.message,
        th
      )
    }
  }

  @Synchronized
  override fun newBlob(blobCounters: BlobCounters): SafeFuture<*> {
    if (aggregationStartTimestamp == null) {
      aggregationStartTimestamp = blobCounters.startBlockTimestamp
    }
    lastBlockNumber = blobCounters.endBlockNumber
    return SafeFuture.completedFuture(Unit)
  }

  override fun onAggregationTrigger(aggregationTriggerHandler: AggregationTriggerHandler) {
    this.aggregationTriggerHandler = aggregationTriggerHandler
  }

  @Synchronized
  override fun reset() {
    this.aggregationStartTimestamp = null
    this.lastBlockNumber = null
  }
}

class AggregationTriggerCalculatorByDeadlineRunner(
  private val vertx: Vertx,
  private val config: Config,
  private val aggregationTriggerByDeadline: AggregationTriggerCalculatorByDeadline
) : DeferredAggregationTriggerCalculator by aggregationTriggerByDeadline, LongRunningService {
  data class Config(val deadlineCheckInterval: Duration)

  private val log: Logger = LogManager.getLogger(this::class.java)

  private var deadlineCheckerTimerId: Long? = null
  private lateinit var deadlineCheckerAction: Handler<Long>

  override fun start(): CompletableFuture<Unit> {
    if (deadlineCheckerTimerId == null) {
      deadlineCheckerAction = Handler<Long> {
        aggregationTriggerByDeadline.checkAggregation().whenComplete { _, error ->
          error?.let {
            log.error("Error in checking for aggregation deadline: errorMessage={}", error.message, error)
          }
          deadlineCheckerTimerId = vertx.setTimer(
            config.deadlineCheckInterval.inWholeMilliseconds,
            deadlineCheckerAction
          )
        }
      }
      deadlineCheckerTimerId = vertx.setTimer(config.deadlineCheckInterval.inWholeMilliseconds, deadlineCheckerAction)
    }
    return SafeFuture.completedFuture(Unit)
  }

  override fun stop(): CompletableFuture<Unit> {
    if (deadlineCheckerTimerId != null) {
      vertx.cancelTimer(deadlineCheckerTimerId!!)
      deadlineCheckerAction = Handler<Long> {}
    }
    return SafeFuture.completedFuture(Unit)
  }
}
