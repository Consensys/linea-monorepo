package net.consensys.linea.contract.l1

import net.consensys.linea.contract.LineaRollup
import net.consensys.linea.contract.LineaRollup.BlobSubmissionData
import net.consensys.toBigInteger
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.ProofToFinalize
import org.web3j.abi.TypeReference
import org.web3j.abi.datatypes.DynamicArray
import org.web3j.abi.datatypes.DynamicBytes
import org.web3j.abi.datatypes.Function
import org.web3j.abi.datatypes.Type
import org.web3j.abi.datatypes.generated.Bytes32
import org.web3j.abi.datatypes.generated.Uint256
import java.math.BigInteger
import java.util.Arrays

internal fun buildSubmitBlobsFunction(
  blobs: List<BlobRecord>
): Function {
  val blobsSubmissionData = blobs.map { blob ->
    val blobCompressionProof = blob.blobCompressionProof!!
    val supportingSubmissionData = LineaRollup.SupportingSubmissionDataV2(
      /*finalStateRootHash*/ blobCompressionProof.finalStateRootHash,
      /*firstBlockInData*/ blob.startBlockNumber.toBigInteger(),
      /*finalBlockInData*/ blob.endBlockNumber.toBigInteger(),
      /*snarkHash*/ blobCompressionProof.snarkHash
    )

    BlobSubmissionData(
      /*submissionData*/ supportingSubmissionData,
      /*dataEvaluationClaim*/ BigInteger(blobCompressionProof.expectedY),
      /*kzgCommitment*/ blobCompressionProof.commitment,
      /*kzgProof*/ blobCompressionProof.kzgProofContract
    )
  }

  /**
   *  function submitBlobs(
   *     BlobSubmissionData[] calldata _blobSubmissionData,
   *     bytes32 _parentShnarf,
   *     bytes32 _finalBlobShnarf
   *   ) external;
   */
  return Function(
    LineaRollup.FUNC_SUBMITBLOBS,
    Arrays.asList<Type<*>>(
      DynamicArray(BlobSubmissionData::class.java, blobsSubmissionData),
      Bytes32(blobs.first().blobCompressionProof!!.prevShnarf),
      Bytes32(blobs.last().blobCompressionProof!!.expectedShnarf)
    ),
    emptyList<TypeReference<*>>()
  )
}

internal fun buildFinalizeBlobsFunction(
  aggregationProof: ProofToFinalize,
  aggregationLastBlob: BlobRecord,
  parentShnarf: ByteArray,
  parentL1RollingHash: ByteArray,
  parentL1RollingHashMessageNumber: Long
): Function {
  val aggregationEndBlobInfo = LineaRollup.ShnarfData(
    /*parentShnarf*/ aggregationLastBlob.blobCompressionProof!!.prevShnarf,
    /*snarkHash*/ aggregationLastBlob.blobCompressionProof!!.snarkHash,
    /*finalStateRootHash*/ aggregationLastBlob.blobCompressionProof!!.finalStateRootHash,
    /*dataEvaluationPoint*/ aggregationLastBlob.blobCompressionProof!!.expectedX,
    /*dataEvaluationClaim*/ aggregationLastBlob.blobCompressionProof!!.expectedY
  )

  val finalizationData = LineaRollup.FinalizationDataV2(
    /*parentStateRootHash*/ aggregationProof.parentStateRootHash,
    /*lastFinalizedShnarf*/ parentShnarf,
    /*finalBlockInData*/ aggregationProof.finalBlockNumber.toBigInteger(),
    /*shnarfData*/ aggregationEndBlobInfo,
    /*lastFinalizedTimestamp*/ aggregationProof.parentAggregationLastBlockTimestamp.epochSeconds.toBigInteger(),
    /*finalTimestamp*/ aggregationProof.finalTimestamp.epochSeconds.toBigInteger(),
    /*lastFinalizedL1RollingHash*/ parentL1RollingHash,
    /*l1RollingHash*/ aggregationProof.l1RollingHash,
    /*lastFinalizedL1RollingHashMessageNumber*/ parentL1RollingHashMessageNumber.toBigInteger(),
    /*l1RollingHashMessageNumber*/ aggregationProof.l1RollingHashMessageNumber.toBigInteger(),
    /*l2MerkleTreesDepth*/ aggregationProof.l2MerkleTreesDepth.toBigInteger(),
    /*l2MerkleRoots*/ aggregationProof.l2MerkleRoots,
    /*l2MessagingBlocksOffsets*/ aggregationProof.l2MessagingBlocksOffsets
  )

  /**
   *  function finalizeBlocksWithProof(
   *     bytes calldata _aggregatedProof,
   *     uint256 _aggregatedVerifierIndex,
   *     FinalizationData calldata _finalizationData
   *   ) external;
   */
  val function = Function(
    LineaRollup.FUNC_FINALIZEBLOCKSWITHPROOF,
    Arrays.asList<Type<*>>(
      DynamicBytes(aggregationProof.aggregatedProof),
      Uint256(aggregationProof.aggregatedVerifierIndex.toLong()),
      finalizationData
    ),
    emptyList<TypeReference<*>>()
  )
  return function
}
