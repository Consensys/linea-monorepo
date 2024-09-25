package net.consensys.zkevm.ethereum.coordination

import net.consensys.zkevm.domain.BlobSubmittedEvent
import net.consensys.zkevm.domain.FinalizationSubmittedEvent
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.atomic.AtomicReference
import java.util.function.Supplier

abstract class SubmittedEventTracker<T>(submission: T?) : (T?) -> SafeFuture<*>, Supplier<T?> {
  private val submissionCache = AtomicReference(submission)

  override fun invoke(trackable: T?): SafeFuture<*> {
    submissionCache.set(trackable)
    return SafeFuture.completedFuture(Unit)
  }

  override fun get(): T? {
    return submissionCache.get()
  }
}

class BlobSubmittedEventTracker(
  blobSubmittedEvent: BlobSubmittedEvent? = null
) : SubmittedEventTracker<BlobSubmittedEvent>(blobSubmittedEvent) {
  fun getEndBlockNumber(): ULong {
    return get()?.blobs?.last()?.endBlockNumber ?: 0UL
  }
}

class FinalizationSubmittedEventTracker(
  finalizationSubmittedEvent: FinalizationSubmittedEvent? = null
) : SubmittedEventTracker<FinalizationSubmittedEvent>(finalizationSubmittedEvent) {
  fun getEndBlockNumber(): ULong {
    return get()?.aggregationProof?.finalBlockNumber?.toULong() ?: 0UL
  }
}
