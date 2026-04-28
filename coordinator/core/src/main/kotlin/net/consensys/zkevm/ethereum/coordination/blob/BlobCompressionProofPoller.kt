package net.consensys.zkevm.ethereum.coordination.blob

import io.vertx.core.Vertx
import linea.clients.BlobCompressionProverClientV2
import linea.domain.BlobCompressionProof
import linea.domain.BlobRecord
import linea.domain.CompressionProofIndex
import linea.metrics.LineaMetricsCategory
import linea.timer.TimerSchedule
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.zkevm.ethereum.coordination.AbstractProofPoller
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration.Companion.milliseconds

class BlobCompressionProofPoller(
  private val blobCompressionProverClient: BlobCompressionProverClientV2,
  private val blobCompressionProofHandler: BlobCompressionProofHandler,
  vertx: Vertx,
  log: Logger,
  pollingIntervalMs: Long = 100.milliseconds.inWholeMilliseconds,
  name: String = "BlobCompressionProofPoller",
  timerSchedule: TimerSchedule = TimerSchedule.FIXED_DELAY,
  metricsFacade: MetricsFacade,
) : AbstractProofPoller<CompressionProofIndex, BlobRecord, BlobCompressionProof>(
  vertx = vertx,
  pollingIntervalMs = pollingIntervalMs,
  log = log,
  name = name,
  timerSchedule = timerSchedule,
) {
  init {
    metricsFacade.createGauge(
      category = LineaMetricsCategory.BLOB,
      name = "prover.pendingproofs",
      description = "Number of blob compression proof waiting responses",
      measurementSupplier = { proofRequestsInProgress.size },
    )
  }

  override fun findProofResponse(proofIndex: CompressionProofIndex): SafeFuture<BlobCompressionProof?> =
    blobCompressionProverClient.findProofResponse(proofIndex)

  override fun handleProof(
    proofIndex: CompressionProofIndex,
    unprovenItem: BlobRecord,
    proof: BlobCompressionProof,
  ): SafeFuture<*> {
    log.info("blob compression proof generated: blob={}", unprovenItem.intervalString())
    return blobCompressionProofHandler.acceptNewBlobCompressionProof(
      unprovenItem.copy(blobCompressionProof = proof),
    )
  }
}
