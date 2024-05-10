package net.consensys.zkevm.domain

import kotlinx.datetime.Instant

typealias BlobsToAggregate = BlockIntervalData

data class ExecutionProofVersions(
  val conflationCalculatorVersion: String,
  val executionProverVersion: String
)

data class VersionedExecutionProofs(
  val executionProofs: BlockIntervals,
  val executionVersion: List<ExecutionProofVersions>
)

/**
 * Represents an Aggregation request to the Prover
 * @property parentAggregationLastBlockTimestamp The timestamp of the last block of the previous aggregation.
 */
data class ProofsToAggregate(
  val compressionProofs: BlockIntervals,
  val executionProofs: BlockIntervals,
  val executionVersion: List<ExecutionProofVersions>,
  val parentAggregationLastBlockTimestamp: Instant,
  val parentAggregationLastL1RollingHashMessageNumber: ULong,
  val parentAggregationLastL1RollingHash: ByteArray
) {
  fun getStartEndBlockInterval(): BlockInterval {
    val startBlockNumber = compressionProofs.startingBlockNumber
    val endBlockNumber = compressionProofs.upperBoundaries.last()
    return BlockInterval.between(startBlockNumber, endBlockNumber)
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as ProofsToAggregate

    if (compressionProofs != other.compressionProofs) return false
    if (executionProofs != other.executionProofs) return false
    if (executionVersion != other.executionVersion) return false
    if (parentAggregationLastBlockTimestamp != other.parentAggregationLastBlockTimestamp) return false
    if (parentAggregationLastL1RollingHashMessageNumber != other.parentAggregationLastL1RollingHashMessageNumber) {
      return false
    }
    if (!parentAggregationLastL1RollingHash.contentEquals(other.parentAggregationLastL1RollingHash)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = compressionProofs.hashCode()
    result = 31 * result + executionProofs.hashCode()
    result = 31 * result + executionVersion.hashCode()
    result = 31 * result + parentAggregationLastBlockTimestamp.hashCode()
    result = 31 * result + parentAggregationLastL1RollingHashMessageNumber.hashCode()
    result = 31 * result + parentAggregationLastL1RollingHash.contentHashCode()
    return result
  }
}

data class ProofToFinalize(
  val aggregatedProof: ByteArray,
  val parentStateRootHash: ByteArray,
  val aggregatedVerifierIndex: Int,
  val aggregatedProofPublicInput: ByteArray,
  val dataHashes: List<ByteArray>,
  val dataParentHash: ByteArray,
  val firstBlockNumber: Long,
  val finalBlockNumber: Long,
  val parentAggregationLastBlockTimestamp: Instant,
  val finalTimestamp: Instant,
  val l1RollingHash: ByteArray,
  val l1RollingHashMessageNumber: Long,
  val l2MerkleRoots: List<ByteArray>,
  val l2MerkleTreesDepth: Int,
  val l2MessagingBlocksOffsets: ByteArray
) : BlockInterval {
  override val startBlockNumber: ULong = firstBlockNumber.toULong()
  override val endBlockNumber: ULong = finalBlockNumber.toULong()

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as ProofToFinalize

    if (!aggregatedProof.contentEquals(other.aggregatedProof)) return false
    if (!parentStateRootHash.contentEquals(other.parentStateRootHash)) return false
    if (aggregatedVerifierIndex != other.aggregatedVerifierIndex) return false
    if (!aggregatedProofPublicInput.contentEquals(other.aggregatedProofPublicInput)) return false
    dataHashes.forEachIndexed { index, bytes ->
      if (!bytes.contentEquals(other.dataHashes[index])) return false
    }
    if (!dataParentHash.contentEquals(other.dataParentHash)) return false
    if (firstBlockNumber != other.firstBlockNumber) return false
    if (finalBlockNumber != other.finalBlockNumber) return false
    if (parentAggregationLastBlockTimestamp != other.parentAggregationLastBlockTimestamp) return false
    if (finalTimestamp != other.finalTimestamp) return false
    if (!l1RollingHash.contentEquals(other.l1RollingHash)) return false
    if (l1RollingHashMessageNumber != other.l1RollingHashMessageNumber) {
      return false
    }
    l2MerkleRoots.forEachIndexed { index, bytes ->
      if (!bytes.contentEquals(other.l2MerkleRoots[index])) return false
    }
    if (l2MerkleTreesDepth != other.l2MerkleTreesDepth) return false
    return l2MessagingBlocksOffsets.contentEquals(other.l2MessagingBlocksOffsets)
  }

  override fun hashCode(): Int {
    var result = aggregatedProof.contentHashCode()
    result = 31 * result + parentStateRootHash.contentHashCode()
    result = 31 * result + aggregatedVerifierIndex
    result = 31 * result + aggregatedProofPublicInput.contentHashCode()
    result = 31 * result + dataHashes.hashCode()
    result = 31 * result + dataParentHash.contentHashCode()
    result = 31 * result + firstBlockNumber.hashCode()
    result = 31 * result + finalBlockNumber.hashCode()
    result = 31 * result + parentAggregationLastBlockTimestamp.hashCode()
    result = 31 * result + finalTimestamp.hashCode()
    result = 31 * result + l1RollingHash.contentHashCode()
    result = 31 * result + l1RollingHashMessageNumber.hashCode()
    result = 31 * result + l2MerkleRoots.hashCode()
    result = 31 * result + l2MerkleTreesDepth
    result = 31 * result + l2MessagingBlocksOffsets.contentHashCode()
    return result
  }
}

data class Aggregation(
  override val startBlockNumber: ULong,
  override val endBlockNumber: ULong,
  val status: Status,
  val aggregationCalculatorVersion: String,
  val batchCount: ULong,
  val aggregationProof: ProofToFinalize?
) : BlockInterval {
  enum class Status {
    Proven,
    Proving
  }
}
