package net.consensys.zkevm.coordinator.clients.prover

fun interface GetProofRequestFileNameProvider {
  fun getRequestFileName(
    startBlockNumber: ULong,
    endBlockNumber: ULong
  ): String
}

fun interface ProofResponseFileNameProvider {
  fun getResponseFileName(
    startBlockNumber: ULong,
    endBlockNumber: ULong
  ): String
}

fun interface ProofResponseFileNameProviderV2 {
  fun getResponseFileName(
    startBlockNumber: ULong,
    endBlockNumber: ULong,
    hash: String
  ): String
}

interface ProverFilesNameProvider : GetProofRequestFileNameProvider, ProofResponseFileNameProvider

fun fileName(startBlockNumber: ULong, endBlockNumber: ULong, fileNameSuffix: String, hash: String = "") =
  if (hash.isEmpty()) {
    "$startBlockNumber-$endBlockNumber-$fileNameSuffix"
  } else {
    "$startBlockNumber-$endBlockNumber-$hash-$fileNameSuffix"
  }

open class GenericProofFileNameProvider(
  private val fileNameSuffix: String
) : ProofResponseFileNameProvider, ProofResponseFileNameProviderV2 {
  override fun getResponseFileName(startBlockNumber: ULong, endBlockNumber: ULong): String =
    fileName(startBlockNumber, endBlockNumber, fileNameSuffix)

  override fun getResponseFileName(startBlockNumber: ULong, endBlockNumber: ULong, hash: String): String =
    fileName(startBlockNumber, endBlockNumber, fileNameSuffix, hash)
}

object ExecutionProofResponseFileNameProvider : GenericProofFileNameProvider("getZkProof.json")

object CompressionProofFileNameProvider :
  ProofResponseFileNameProvider by GenericProofFileNameProvider("getZkBlobCompressionProof.json")

object AggregationProofResponseFileNameProviderV2 : GenericProofFileNameProvider("getZkAggregatedProof.json")

object AggregationProofResponseFileNameProviderV3 :
  ProofResponseFileNameProviderV2 by GenericProofFileNameProvider("getZkAggregatedProof.json")

class ExecutionProofFileNameProvider(
  private val tracesVersion: String,
  private val stateManagerVersion: String
) : ProverFilesNameProvider, ProofResponseFileNameProvider by ExecutionProofResponseFileNameProvider {
  override fun getRequestFileName(startBlockNumber: ULong, endBlockNumber: ULong): String {
    return "$startBlockNumber-$endBlockNumber-etv$tracesVersion-stv$stateManagerVersion-getZkProof.json"
  }
}
