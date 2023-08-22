package net.consensys.zkevm.coordinator.clients.prover

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class ProverFilesNameProviderImplV1Test {
  private val filesNameProvider = ProverFilesNameProviderImplV1(
    tracesVersion = "1.1.1",
    stateManagerVersion = "2.2.2",
    proverVersion = "3.3.3",
    proofFileExtension = "json"
  )

  @Test
  fun getRequestFileName() {
    assertThat(filesNameProvider.getRequestFileName(1uL, 2uL))
      .isEqualTo("1-2-etv1.1.1-stv2.2.2-getZkProof.json")
  }

  @Test
  fun getResponseFileName() {
    assertThat(filesNameProvider.getResponseFileName(1uL, 2uL))
      .isEqualTo("1-2-etv1.1.1-stv2.2.2-getZkProof.json.3.3.3.json")
  }
}
