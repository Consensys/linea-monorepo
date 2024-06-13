package net.consensys.linea.nativecompressor

import com.sun.jna.Library
import com.sun.jna.Structure

// Equivalent to the C response struct used in CGO
class CalculateShnarfResult(
  @JvmField var commitment: String,
  @JvmField var kzgProofContract: String,
  @JvmField var kzgProofSideCar: String,
  @JvmField var dataHash: String,
  @JvmField var snarkHash: String,
  @JvmField var expectedX: String,
  @JvmField var expectedY: String,
  @JvmField var expectedShnarf: String,
  // error message is empty if there is no error
  // it cannot be null because JNA call blow up with segfault.
  @JvmField var errorMessage: String
) : Structure() {

  // JNA requires a default constructor
  constructor() : this("", "", "", "", "", "", "", "", "")

  override fun getFieldOrder(): List<String> {
    return listOf(
      "commitment",
      "kzgProofContract",
      "kzgProofSideCar",
      "dataHash",
      "snarkHash",
      "expectedX",
      "expectedY",
      "expectedShnarf",
      "errorMessage"
    )
  }
}

interface GoNativeBlobShnarfCalculator {
  // Equivalent to the Go CalculateShnarf function acting as bridge
  // between C and Go code.
  // It prepares a Response object by computing all the fields except for the
  // proof.
  fun CalculateShnarf(
    eip4844Enabled: Boolean,
    compressedData: String,
    parentStateRootHash: String,
    finalStateRootHash: String,
    prevShnarf: String,
    conflationOrderStartingBlockNumber: Long,
    conflationOrderUpperBoundariesLen: Int,
    conflationOrderUpperBoundaries: LongArray
  ): CalculateShnarfResult
}

// Just to not expose JNA to the clients
internal interface GoNativeBlobShnarfCalculatorJna : GoNativeBlobShnarfCalculator, Library

class GoNativeShnarfCalculatorFactory {
  companion object {
    const val LIB_NAME = "libshnarf_calculator_native_jna"

    @Volatile
    private var instance: GoNativeBlobShnarfCalculator? = null

    fun getInstance(): GoNativeBlobShnarfCalculator {
      if (instance == null) {
        synchronized(this) {
          if (instance == null) {
            instance = NativeLibUtil.loadJnaLibFromResource(LIB_NAME, GoNativeBlobShnarfCalculatorJna::class.java)
          }
        }
      }
      return instance!!
    }
  }
}
