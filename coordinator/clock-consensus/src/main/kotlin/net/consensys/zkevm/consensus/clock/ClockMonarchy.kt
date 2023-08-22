package net.consensys.zkevm.consensus.clock

import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.zkevm.consensus.SlotBasedConsensus
import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.services.timer.TimerService
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

class ClockMonarchy(private val config: Config) : SlotBasedConsensus {
  private val log = LogManager.getLogger(this::class.java)
  data class Config(
    val slotInterval: Duration = 6.seconds,
    val initialSlotNumber: ULong = 0u,
    val clock: Clock = Clock.System,
    val slotStartTime: Instant = clock.now()
  )

  private var slotNumber = config.initialSlotNumber
  private val timerService: TimerService = TimerService { onTick() }
  private var slotHandler: ((SlotBasedConsensus.NewSlot) -> SafeFuture<Unit>) =
    { _: SlotBasedConsensus.NewSlot ->
      SafeFuture.completedFuture(Unit)
    }
  private var nextExpectedSlotTimeStamp = config.slotStartTime + config.slotInterval

  fun start(): SafeFuture<*> {
    return timerService.start()
  }

  fun stop(): SafeFuture<*> {
    return timerService.stop()
  }

  private fun onTick() {
    val currentTime = config.clock.now()
    if (currentTime >= nextExpectedSlotTimeStamp) {
      slotNumber += 1u
      slotHandler(SlotBasedConsensus.NewSlot(currentTime, slotNumber))
      nextExpectedSlotTimeStamp = currentTime + config.slotInterval
      log.info(
        "New slot # {} at {} next slot {}",
        slotNumber,
        currentTime,
        nextExpectedSlotTimeStamp
      )
    } else {
      log.trace("Tick on {} slot # {} skipping", currentTime, slotNumber)
    }
  }

  override fun setSlotHandler(handler: (SlotBasedConsensus.NewSlot) -> SafeFuture<Unit>) {
    slotHandler = handler
  }
}
