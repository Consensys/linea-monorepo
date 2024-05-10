package net.consensys.zkevm.ethereum.submission

import net.consensys.encodeHex
import net.consensys.linea.Constants.Eip4844BlobSize
import net.consensys.linea.contract.LineaRollup
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.toBigInteger
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapProvider
import net.consensys.zkevm.ethereum.settlement.BlobSubmitter
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

class BlobSubmitterAsEIP4844(
  private val contract: LineaRollupAsyncFriendly,
  private val gasPriceCapProvider: GasPriceCapProvider
) : BlobSubmitter {
  private val log: Logger = LogManager.getLogger(this::class.java)

  private fun buildSupportingSubmissionData(blobRecord: BlobRecord): LineaRollup.SupportingSubmissionData {
    val blobCompressionProof = blobRecord.blobCompressionProof!!
    return LineaRollup.SupportingSubmissionData(
      /*parentStateRootHash*/ blobCompressionProof.parentStateRootHash,
      /*dataParentHash*/ blobCompressionProof.parentDataHash,
      /*finalStateRootHash*/ blobCompressionProof.finalStateRootHash,
      /*firstBlockInData*/ blobRecord.startBlockNumber.toBigInteger(),
      /*finalBlockInData*/ blobRecord.endBlockNumber.toBigInteger(),
      /*snarkHash*/ blobCompressionProof.snarkHash
    )
  }

  private fun padBlobToCorrectSize(blob: ByteArray): ByteArray {
    return ByteArray(Eip4844BlobSize).apply { blob.copyInto(this) }
  }

  override fun submitBlob(blobRecord: BlobRecord): SafeFuture<String> {
    log.debug(
      "submitting blob: blob={} dataHash={}",
      blobRecord.intervalString(),
      blobRecord.blobHash.encodeHex()
    )

    return gasPriceCapProvider.getGasPriceCaps(blobRecord.startBlockNumber.toLong())
      .thenCompose { gasPriceCaps ->
        contract
          .sendBlobData(
            supportingSubmissionData = buildSupportingSubmissionData(blobRecord),
            dataEvaluationClaim = BigInteger(blobRecord.blobCompressionProof!!.expectedY),
            kzgCommitment = blobRecord.blobCompressionProof!!.commitment,
            kzgProof = blobRecord.blobCompressionProof!!.kzgProofContract,
            blob = padBlobToCorrectSize(blobRecord.blobCompressionProof!!.compressedData),
            gasPriceCaps = gasPriceCaps
          )
      }
  }

  override fun submitBlobCall(blobRecord: BlobRecord): SafeFuture<*> {
    return contract.submitBlobDataEthCall(
      supportingSubmissionData = buildSupportingSubmissionData(blobRecord),
      dataEvaluationClaim = BigInteger(blobRecord.blobCompressionProof!!.expectedY),
      kzgCommitment = blobRecord.blobCompressionProof!!.commitment,
      kzgProof = blobRecord.blobCompressionProof!!.kzgProofContract,
      blob = padBlobToCorrectSize(blobRecord.blobCompressionProof!!.compressedData),
      blobVersionedHashes = listOf(Bytes.wrap(blobRecord.blobHash))
    )
  }
}
