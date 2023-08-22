package net.consensys.zkevm.consensus.clock

import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.zkevm.consensus.SlotBasedConsensus
import org.junit.jupiter.api.Test
import org.mockito.Mockito.`when`
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.times
import org.mockito.kotlin.verify
import org.mockito.kotlin.verifyNoInteractions
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.services.timer.TimerService.TIME_TICKER_REFRESH_RATE
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds

class ClockMonarchyTest {
  private val timerTickIntervalMs = (1.0 / TIME_TICKER_REFRESH_RATE * 1000).toLong()

  @Test
  fun clockMonarchyTest() {
    val mockClock: Clock = mock()
    val clockMonarchy =
      ClockMonarchy(
        ClockMonarchy.Config(
          slotInterval = 1.seconds,
          slotStartTime = Instant.fromEpochMilliseconds(0),
          clock = mockClock
        )
      )
    val virtualTimeHandler = mock<(SlotBasedConsensus.NewSlot) -> SafeFuture<Unit>>()
    var currentTime = Instant.fromEpochSeconds(0)
    `when`(mockClock.now()).then { currentTime }

    clockMonarchy.setSlotHandler(virtualTimeHandler)
    clockMonarchy.start().get()
    // After 1 tick we shouldn't produce slot event if clock stands still
    Thread.sleep(timerTickIntervalMs)
    verifyNoInteractions(virtualTimeHandler)
    // If timer service ticks, but clock doesn't move, there shouldn't be a new slot event
    Thread.sleep(timerTickIntervalMs)
    verifyNoInteractions(virtualTimeHandler)
    // If clock moves, there should be a new slot event
    currentTime += 1.seconds
    Thread.sleep(timerTickIntervalMs)
    verify(virtualTimeHandler, times(1))(any())
    currentTime += 1.seconds
    Thread.sleep(timerTickIntervalMs)
    verify(virtualTimeHandler, times(2))(any())
    // If clock moves on less than slot interval, there shouldn't be a new slot event
    currentTime += 500.milliseconds
    Thread.sleep(timerTickIntervalMs)
    verify(virtualTimeHandler, times(2))(any())
    // If we missed a slot, we shouldn't trigger 2 events in the next 2 ticks
    currentTime += 2.seconds
    Thread.sleep(timerTickIntervalMs * 2)
    verify(virtualTimeHandler, times(3))(any())
    clockMonarchy.stop().get()
    // If the service is stopped, there shouldn't be a new slot event
    currentTime += 1.seconds
    Thread.sleep(timerTickIntervalMs)
    verify(virtualTimeHandler, times(3))(any())
  }
}
