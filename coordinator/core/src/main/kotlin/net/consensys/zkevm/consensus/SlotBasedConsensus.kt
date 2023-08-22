package net.consensys.zkevm.consensus

import kotlinx.datetime.Instant
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface SlotBasedConsensus {
  data class NewSlot(val timestamp: Instant, val slotNumber: ULong)

  fun setSlotHandler(handler: (NewSlot) -> SafeFuture<Unit>)
}
