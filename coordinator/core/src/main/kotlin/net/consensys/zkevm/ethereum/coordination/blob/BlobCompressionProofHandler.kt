package net.consensys.zkevm.ethereum.coordination.blob

import net.consensys.zkevm.coordinator.clients.BlobCompressionProof
import net.consensys.zkevm.domain.BlockInterval
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class BlobCompressionProofUpdate(
  val blockInterval: BlockInterval,
  val blobCompressionProof: BlobCompressionProof
)

fun interface BlobCompressionProofHandler {
  fun acceptNewBlobCompressionProof(blobCompressionProofUpdate: BlobCompressionProofUpdate): SafeFuture<*>
}
