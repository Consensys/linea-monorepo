package net.consensys.zkevm.coordinator.clients.prover

import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test

class ProverFileNameProvidersTest {

  @Test
  fun test_getExecutionProofFileName() {
    Assertions.assertEquals(
      "11-17-getZkProof.json",
      ExecutionProofResponseFileNameProvider.getResponseFileName(11u, 17u)
    )
  }

  @Test
  fun test_getCompressionProofFileName() {
    val fileNameProvider = CompressionProofFileNameProvider
    Assertions.assertEquals(
      "11-17-getZkBlobCompressionProof.json",
      fileNameProvider.getResponseFileName(11u, 17u)
    )
  }

  @Test
  fun test_getResponseFileName() {
    Assertions.assertEquals(
      "11-27-getZkAggregatedProof.json",
      AggregationProofResponseFileNameProviderV2.getResponseFileName(11u, 27u)
    )
  }

  @Test
  fun test_getAggregationResponseV3FileName() {
    Assertions.assertEquals(
      "11-27-abcd-getZkAggregatedProof.json",
      AggregationProofResponseFileNameProviderV3.getResponseFileName(11u, 27u, "abcd")
    )
  }
}
