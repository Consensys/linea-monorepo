package net.consensys.zkevm.domain

import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.setFirstByteToZero
import net.consensys.trimToSecondPrecision
import net.consensys.zkevm.coordinator.clients.BlobCompressionProof
import org.web3j.crypto.Hash
import kotlin.random.Random
import kotlin.time.Duration.Companion.seconds

fun createBlobRecord(
  startBlockNumber: ULong,
  endBlockNumber: ULong,
  conflationCalculationVersion: String = "0.1.0",
  blobHash: ByteArray = Random.nextBytes(32).setFirstByteToZero(),
  status: BlobStatus = BlobStatus.COMPRESSION_PROVEN,
  shnarf: ByteArray = Random.nextBytes(32),
  parentStateRootHash: ByteArray = Random.nextBytes(32).setFirstByteToZero(),
  parentDataHash: ByteArray = Random.nextBytes(32),
  compressedData: ByteArray = Random.nextBytes(32).setFirstByteToZero(),
  eip4844Enabled: Boolean = false,
  startBlockTime: Instant = Clock.System.now().trimToSecondPrecision(),
  batchesCount: UInt = 1U,
  parentBlobRecord: BlobRecord? = null,
  blobCompressionProof: BlobCompressionProof? = BlobCompressionProof(
    compressedData = compressedData,
    conflationOrder = BlockIntervals(startBlockNumber, listOf(endBlockNumber)),
    prevShnarf = Random.nextBytes(32),
    parentStateRootHash = parentBlobRecord?.blobCompressionProof?.finalStateRootHash ?: parentStateRootHash,
    finalStateRootHash = Random.nextBytes(32).setFirstByteToZero(),
    parentDataHash = parentBlobRecord?.blobCompressionProof?.dataHash ?: parentDataHash,
    dataHash = Hash.sha3(compressedData),
    snarkHash = Random.nextBytes(32),
    expectedX = Random.nextBytes(32),
    expectedY = Random.nextBytes(32),
    expectedShnarf = shnarf,
    decompressionProof = Random.nextBytes(512),
    proverVersion = "mock-0.0.0",
    verifierID = 6789,
    eip4844Enabled = eip4844Enabled,
    commitment = if (eip4844Enabled) Random.nextBytes(48) else ByteArray(0),
    kzgProofContract = if (eip4844Enabled) Random.nextBytes(48) else ByteArray(0),
    kzgProofSidecar = if (eip4844Enabled) Random.nextBytes(48) else ByteArray(0)
  )
): BlobRecord {
  val startBlockTime = startBlockTime.plus(
    ((startBlockNumber - 1UL).toLong() * 12).seconds
  )
  val endBlockTime = startBlockTime.plus(
    ((endBlockNumber - startBlockNumber).toLong() * 12).seconds
  )
  return BlobRecord(
    startBlockNumber = startBlockNumber,
    endBlockNumber = endBlockNumber,
    conflationCalculatorVersion = conflationCalculationVersion,
    blobHash = blobHash,
    startBlockTime = startBlockTime,
    endBlockTime = endBlockTime,
    batchesCount = batchesCount,
    status = status,
    expectedShnarf = shnarf,
    blobCompressionProof = blobCompressionProof
  )
}
