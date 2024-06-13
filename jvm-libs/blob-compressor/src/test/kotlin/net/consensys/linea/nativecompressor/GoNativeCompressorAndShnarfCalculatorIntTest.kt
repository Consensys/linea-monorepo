package net.consensys.linea.nativecompressor

import net.consensys.decodeHex
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import java.util.Base64

class GoNativeCompressorAndShnarfCalculatorIntTest {
  val DATA_LIMIT = 16 * 1024
  private lateinit var compressor: GoNativeBlobCompressor
  private lateinit var shnarfCalculator: GoNativeBlobShnarfCalculator

  @BeforeEach
  fun beforeEach() {
    compressor = GoNativeBlobCompressorFactory.newInstance()
      .apply { this.Init(DATA_LIMIT, GoNativeBlobCompressorFactory.getDictionaryPath()) }
    shnarfCalculator = GoNativeShnarfCalculatorFactory.getInstance()
  }

  fun encodeBase64(bytes: ByteArray): String {
    return Base64.getEncoder().encodeToString(bytes)
  }

  @Test
  fun `should compress and shnarf`() {
    val block = CompressorTestData.blocksRlpEncoded.first()
    assertTrue(compressor.Write(block, block.size))
    assertThat(compressor.Error()).isNullOrEmpty()

    val compressedData = ByteArray(compressor.Len())
    compressor.Bytes(compressedData)

    val result: CalculateShnarfResult = shnarfCalculator.CalculateShnarf(
      eip4844Enabled = false,
      compressedData = encodeBase64(compressedData),
      parentStateRootHash = Bytes32.random().toHexString(),
      finalStateRootHash = Bytes32.random().toHexString(),
      prevShnarf = Bytes32.random().toHexString(),
      conflationOrderStartingBlockNumber = 1,
      conflationOrderUpperBoundariesLen = 2,
      conflationOrderUpperBoundaries = longArrayOf(10, 20)
    )
    assertThat(result).isNotNull
    assertThat(result.errorMessage).isEmpty()
    assertThat(result.commitment).isNotNull()
    assertThat(result.commitment.decodeHex()).hasSize(0)
    assertThat(result.kzgProofContract).isNotNull()
    assertThat(result.kzgProofContract.decodeHex()).hasSize(0)
    assertThat(result.kzgProofSideCar).isNotNull()
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
  fun `compressed size estimation should be consistent with blob maker output`() {
    val block = CompressorTestData.blocksRlpEncoded.first()

    // Write the block to the blob maker and get the effective compressed size
    assertTrue(compressor.Write(block, block.size))
    assertThat(compressor.Error()).isNullOrEmpty()
    val compressedSizeWithHeader = compressor.Len()
    assertTrue(compressedSizeWithHeader > 0)
    compressor.Reset()

    // Perform the same operation using the WorstCompressedBlockSize function
    val compressedSize = compressor.WorstCompressedBlockSize(block, block.size)
    assertTrue(compressedSize < block.size)
    assertTrue(compressedSize > 0)

    assertTrue(compressedSizeWithHeader > compressedSize)

    // sanity checks, these may not be "future proof" tests
    // but they are useful to ensure that the blob maker is behaving as expected
    // and that the code is correctly interfacing with it
    val estimatedHeaderSizePacked = 64 // technically, unpacked it's like ~40 bytes

    // min compressed size should always be strictly bigger than the size returned
    // by the blob maker minus the header size
    assertTrue((compressedSizeWithHeader - estimatedHeaderSizePacked) < compressedSize)
  }

  @Test
  fun `should compress and shnarf with eip4844`() {
    val block = CompressorTestData.blocksRlpEncoded.first()
    assertTrue(compressor.Write(block, block.size))
    assertThat(compressor.Error()).isNullOrEmpty()

    val compressedData = ByteArray(compressor.Len())
    compressor.Bytes(compressedData)

    val result: CalculateShnarfResult = shnarfCalculator.CalculateShnarf(
      eip4844Enabled = true,
      compressedData = encodeBase64(compressedData),
      parentStateRootHash = Bytes32.random().toHexString(),
      finalStateRootHash = Bytes32.random().toHexString(),
      prevShnarf = Bytes32.random().toHexString(),
      conflationOrderStartingBlockNumber = 1,
      conflationOrderUpperBoundariesLen = 2,
      conflationOrderUpperBoundaries = longArrayOf(10, 20)
    )
    assertThat(result).isNotNull
    assertThat(result.errorMessage).isEmpty()
    assertThat(result.commitment).isNotNull()
    assertThat(result.commitment.decodeHex()).hasSize(48)
    assertThat(result.kzgProofContract).isNotNull()
    assertThat(result.kzgProofContract.decodeHex()).hasSize(48)
    assertThat(result.kzgProofSideCar).isNotNull()
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
}
