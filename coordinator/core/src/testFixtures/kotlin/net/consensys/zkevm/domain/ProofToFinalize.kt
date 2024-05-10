package net.consensys.zkevm.domain

import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.setFirstByteToZero
import net.consensys.trimToSecondPrecision
import kotlin.random.Random
import kotlin.time.Duration.Companion.seconds

fun createProofToFinalize(
  firstBlockNumber: Long,
  finalBlockNumber: Long,
  startBlockTime: Instant = Clock.System.now().trimToSecondPrecision(),
  dataParentHash: ByteArray = Random.nextBytes(32).setFirstByteToZero(),
  parentStateRootHash: ByteArray = Random.nextBytes(32).setFirstByteToZero(),
  dataHashes: List<ByteArray> = (firstBlockNumber..finalBlockNumber).map {
    Random.nextBytes(32)
  },
  parentAggregationLastBlockTimestamp: Instant = startBlockTime.plus(
    ((firstBlockNumber - 1L) * 12).seconds
  ),
  finalTimestamp: Instant = startBlockTime.plus(
    ((finalBlockNumber - 1L) * 12).seconds
  ),
  l2MessagingBlockOffsets: ByteArray = ByteArray(32)
): ProofToFinalize {
  return ProofToFinalize(
    aggregatedProof = Random.nextBytes(32).setFirstByteToZero(),
    parentStateRootHash = parentStateRootHash,
    aggregatedVerifierIndex = 0,
    aggregatedProofPublicInput = Random.nextBytes(32).setFirstByteToZero(),
    dataHashes = dataHashes,
    dataParentHash = dataParentHash,
    firstBlockNumber = firstBlockNumber,
    finalBlockNumber = finalBlockNumber,
    parentAggregationLastBlockTimestamp = parentAggregationLastBlockTimestamp,
    finalTimestamp = finalTimestamp,
    l1RollingHash = ByteArray(32),
    l1RollingHashMessageNumber = 0,
    l2MerkleRoots = listOf(Random.nextBytes(32).setFirstByteToZero()),
    l2MerkleTreesDepth = 0,
    l2MessagingBlocksOffsets = l2MessagingBlockOffsets
  )
}

fun createProofToFinalizeFromBlobs(
  blobRecords: List<BlobRecord>,
  lastFinalizedBlockTime: Instant,
  parentRecord: BlobRecord? = null,
  parentFinalization: ProofToFinalize? = null,
  parentStateRootHash: ByteArray? = null,
  dataParentHash: ByteArray? = null
): ProofToFinalize {
  val endBlockTime = lastFinalizedBlockTime.plus(
    (blobRecords.last().endBlockNumber.toLong() * 12).seconds
  )
  return ProofToFinalize(
    aggregatedProof = Random.nextBytes(32).setFirstByteToZero(),
    parentStateRootHash =
    parentRecord?.blobCompressionProof?.finalStateRootHash
      ?: parentStateRootHash
      ?: Random.nextBytes(32).setFirstByteToZero(),
    aggregatedVerifierIndex = 0,
    aggregatedProofPublicInput = Random.nextBytes(32).setFirstByteToZero(),
    dataHashes = blobRecords.map { it.blobCompressionProof!!.dataHash },
    dataParentHash =
    parentRecord?.blobCompressionProof?.dataHash
      ?: dataParentHash
      ?: Random.nextBytes(32).setFirstByteToZero(),
    firstBlockNumber = blobRecords.first().startBlockNumber.toLong(),
    finalBlockNumber = blobRecords.last().endBlockNumber.toLong(),
    parentAggregationLastBlockTimestamp =
    parentFinalization?.finalTimestamp
      ?: lastFinalizedBlockTime,
    finalTimestamp = endBlockTime,
    l1RollingHash = ByteArray(32),
    l1RollingHashMessageNumber = 0,
    l2MerkleRoots = listOf(Random.nextBytes(32).setFirstByteToZero()),
    l2MerkleTreesDepth = 0,
    l2MessagingBlocksOffsets = ByteArray(32)
  )
}
