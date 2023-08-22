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

interface ProverFilesNameProvider : GetProofRequestFileNameProvider, ProofResponseFileNameProvider

class ProverFilesNameProviderImplV1(
  private val tracesVersion: String,
  private val stateManagerVersion: String,
  private val proverVersion: String,
  private val proofFileExtension: String
) : ProverFilesNameProvider {

  override fun getRequestFileName(startBlockNumber: ULong, endBlockNumber: ULong): String {
    return "$startBlockNumber-$endBlockNumber-etv$tracesVersion-" +
      "stv$stateManagerVersion-getZkProof.json"
  }

  override fun getResponseFileName(startBlockNumber: ULong, endBlockNumber: ULong): String {
    return getRequestFileName(startBlockNumber, endBlockNumber) +
      ".$proverVersion.$proofFileExtension"
  }
}
