package net.consensys.zkevm.ethereum.submission

import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.ethereum.settlement.BlobSubmitter
import tech.pegasys.teku.infrastructure.async.SafeFuture

class Eip4844SwitchAwareBlobSubmitter(
  private val blobSubmitterAsCallData: BlobSubmitter,
  private val blobSubmitterAsEIP4844: BlobSubmitter
) : BlobSubmitter {

  private fun getBlobSubmitter(blobRecord: BlobRecord): BlobSubmitter {
    return if (blobRecord.blobCompressionProof!!.eip4844Enabled) {
      blobSubmitterAsEIP4844
    } else {
      blobSubmitterAsCallData
    }
  }

  override fun submitBlob(blobRecord: BlobRecord): SafeFuture<String> {
    return getBlobSubmitter(blobRecord).submitBlob(blobRecord)
  }

  override fun submitBlobCall(blobRecord: BlobRecord): SafeFuture<*> {
    return getBlobSubmitter(blobRecord).submitBlobCall(blobRecord)
  }
}
