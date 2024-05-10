package net.consensys.zkevm.ethereum.settlement

import net.consensys.zkevm.domain.BlobRecord
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface BlobSubmitter {
  fun submitBlob(blobRecord: BlobRecord): SafeFuture<String>

  fun submitBlobCall(blobRecord: BlobRecord): SafeFuture<*>
}
