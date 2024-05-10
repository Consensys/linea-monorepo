package net.consensys.zkevm.ethereum.submission

import net.consensys.linea.contract.LineaRollup.SubmissionData
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.toBigInteger
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.ethereum.error.handling.SubmissionException
import net.consensys.zkevm.ethereum.settlement.BlobSubmitter
import tech.pegasys.teku.infrastructure.async.SafeFuture

class BlobSubmitterAsCallData(
  private val contract: LineaRollupAsyncFriendly
) : BlobSubmitter {
  private fun buildSubmissionDataObj(blobRecord: BlobRecord): SubmissionData {
    val blobCompressionProof = blobRecord.blobCompressionProof!!
    return SubmissionData(
      blobCompressionProof.parentStateRootHash,
      blobCompressionProof.parentDataHash,
      blobCompressionProof.finalStateRootHash,
      blobRecord.startBlockNumber.toBigInteger(),
      blobRecord.endBlockNumber.toBigInteger(),
      blobCompressionProof.snarkHash,
      blobCompressionProof.compressedData
    )
  }

  override fun submitBlob(blobRecord: BlobRecord): SafeFuture<String> {
    return try {
      val transactionHash = contract.submitDataAndForget(buildSubmissionDataObj(blobRecord))
      SafeFuture.completedFuture(transactionHash)
    } catch (e: Throwable) {
      SafeFuture.failedFuture(
        SubmissionException("Blob submission failed: blob=${blobRecord.intervalString()}, message='${e.message}'", e)
      )
    }
  }

  @Synchronized
  override fun submitBlobCall(blobRecord: BlobRecord): SafeFuture<*> {
    return contract.submitDataEthCall(buildSubmissionDataObj(blobRecord))
  }
}
