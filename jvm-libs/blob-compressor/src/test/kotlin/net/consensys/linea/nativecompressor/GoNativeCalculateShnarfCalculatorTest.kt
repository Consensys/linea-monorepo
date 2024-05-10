package net.consensys.linea.nativecompressor

import net.consensys.decodeHex
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Assertions.assertNotNull
import org.junit.jupiter.api.Assertions.assertSame
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.Test

class GoNativeCalculateShnarfCalculatorTest {

  @Test
  fun `should create a GoNativeShnarfCalculator`() {
    val calculator = GoNativeShnarfCalculatorFactory.getInstance()
    assertNotNull(calculator)

    // is singleton
    assertSame(GoNativeShnarfCalculatorFactory.getInstance(), calculator)
  }

  @Test
  fun `shouldCalculateShnarf`() {
    val calculator = GoNativeShnarfCalculatorFactory.getInstance()
    val result = calculator.CalculateShnarf(
      eip4844Enabled = false,
      compressedData = "ADAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA=",
      parentStateRootHash = Bytes32.random().toHexString(),
      finalStateRootHash = Bytes32.random().toHexString(),
      prevShnarf = Bytes32.random().toHexString(),
      conflationOrderStartingBlockNumber = 1L,
      conflationOrderUpperBoundariesLen = 3,
      conflationOrderUpperBoundaries = longArrayOf(10L, 20L, 30L)
    )

    assertNotNull(result)
    assertThat(result.errorMessage).isEmpty()

    // real response fields
    assertNotNull(result.commitment)
    assertThat(result.commitment.decodeHex()).hasSize(0)
    assertNotNull(result.kzgProofContract)
    assertThat(result.kzgProofContract.decodeHex()).hasSize(0)
    assertNotNull(result.kzgProofSideCar)
    assertThat(result.kzgProofSideCar.decodeHex()).hasSize(0)
    assertThat(result.dataHash).isNotNull
    assertThat(result.dataHash.decodeHex()).hasSize(32)
    assertThat(result.snarkHash).isNotNull
    assertThat(result.snarkHash.decodeHex()).hasSize(32)
    assertThat(result.expectedX).isNotNull()
    assertThat(result.expectedX.decodeHex()).hasSize(32)
    assertThat(result.expectedY).isNotNull()
    assertThat(result.expectedY.decodeHex()).hasSize(32)
    assertThat(result.expectedShnarf).isNotNull
    assertThat(result.expectedShnarf.decodeHex()).hasSize(32)
  }

  @Test
  fun `shouldCalculateShnarfWithEip4844`() {
    val calculator = GoNativeShnarfCalculatorFactory.getInstance()
    val result = calculator.CalculateShnarf(
      eip4844Enabled = true,
      compressedData = "ADAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA=",
      parentStateRootHash = Bytes32.random().toHexString(),
      finalStateRootHash = Bytes32.random().toHexString(),
      prevShnarf = Bytes32.random().toHexString(),
      conflationOrderStartingBlockNumber = 1L,
      conflationOrderUpperBoundariesLen = 3,
      conflationOrderUpperBoundaries = longArrayOf(10L, 20L, 30L)
    )

    assertNotNull(result)
    assertThat(result.errorMessage).isEmpty()

    // real response fields
    assertNotNull(result.commitment)
    assertThat(result.commitment.decodeHex()).hasSize(48)
    assertNotNull(result.kzgProofContract)
    assertThat(result.kzgProofContract.decodeHex()).hasSize(48)
    assertNotNull(result.kzgProofSideCar)
    assertThat(result.kzgProofSideCar.decodeHex()).hasSize(48)
    assertThat(result.dataHash).isNotNull
    assertThat(result.dataHash.decodeHex()).hasSize(32)
    assertThat(result.snarkHash).isNotNull
    assertThat(result.snarkHash.decodeHex()).hasSize(32)
    assertThat(result.expectedX).isNotNull()
    assertThat(result.expectedX.decodeHex()).hasSize(32)
    assertThat(result.expectedY).isNotNull()
    assertThat(result.expectedY.decodeHex()).hasSize(32)
    assertThat(result.expectedShnarf).isNotNull
    assertThat(result.expectedShnarf.decodeHex()).hasSize(32)
  }

  @Disabled("This test is meant to run locally to check for memory leaks")
  fun `shouldCalculateShnarf_checkMemory`() {
    val calculator = GoNativeShnarfCalculatorFactory.getInstance()
    // if we enable forcedLeakBuffer, we will ger OOM after 4-6 iterations
    // Exception in thread "JNA Cleaner" java.lang.OutOfMemoryError: Java heap space
    // val forcedLeakBuffer = mutableListOf<CalculateShnarfResult>()
    for (j in 0..100) {
      for (i in 0..100_000) {
        val result = calculator.CalculateShnarf(
          eip4844Enabled = false,
          compressedData = "ADAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA=",
          parentStateRootHash = Bytes32.random().toHexString(),
          finalStateRootHash = Bytes32.random().toHexString(),
          prevShnarf = Bytes32.random().toHexString(),
          conflationOrderStartingBlockNumber = 0L,
          conflationOrderUpperBoundariesLen = 2,
          conflationOrderUpperBoundaries = longArrayOf(10L, 20L)
        )
        // forcedLeakBuffer.add(result)
      }
      System.gc()
      Thread.sleep(1000)
      println(
        "total=${Runtime.getRuntime().totalMemory()}, " +
          "free=${Runtime.getRuntime().freeMemory()}, " +
          "max=${Runtime.getRuntime().maxMemory()}"
      )
    }
  }
}
