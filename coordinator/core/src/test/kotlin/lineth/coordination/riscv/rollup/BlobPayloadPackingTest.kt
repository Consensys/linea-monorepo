package lineth.coordination.riscv.rollup

import linea.domain.Constants
import linea.kotlin.encodeHex
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.Arguments
import org.junit.jupiter.params.provider.MethodSource
import java.security.MessageDigest
import java.util.stream.Stream

class BlobPayloadPackingTest {

  companion object {
    // Reference vectors generated with rollup_spec `pack_blob_payload` (rollup_spec/src/rollup_spec/rollup.py).
    @JvmStatic
    fun specVectors(): Stream<Arguments> = Stream.of(
      Arguments.of(
        ByteArray(0),
        "0x3fc00000000000000000000000000000000000000000000000000000000000000000000000000000",
        "0x700a187d18ec30a349376803b8bd345a7b1cdc38c8559073abe0bb0965e7529e",
      ),
      Arguments.of(
        ByteArray(768) { it.toByte() },
        "0x00004080c1014181c2024282c3034383c4044484c5054585c6064686c70747873202122232425262",
        "0xbc1251723c659b72f0c657f77d79dce1d36eb527137944b4cc93599743f7f325",
      ),
      Arguments.of(
        ByteArray(BlobPayloadPacking.BLOB_PAYLOAD_CAPACITY) { (it * 7 + 3).toByte() },
        "0x00c2844607c98b4d0ed0925415d7995b1cdea06223e5a7692aecae7031f3b5770e3eaf1f8ff060d1",
        "0xbe11638fb5abd3a2b6ac7a904e05bd55314427be9dea3d81b3d7dbe18f3d1f1e",
      ),
    )
  }

  @Test
  fun `payload capacity matches the spec`() {
    assertThat(BlobPayloadPacking.BLOB_PAYLOAD_CAPACITY).isEqualTo(130047)
  }

  @ParameterizedTest
  @MethodSource("specVectors")
  fun `pack matches the spec reference and unpack round-trips`(
    payload: ByteArray,
    expectedPrefixHex: String,
    expectedSha256Hex: String,
  ) {
    val blob = BlobPayloadPacking.pack(payload)

    assertThat(blob).hasSize(Constants.Eip4844BlobSize)
    assertThat(blob.copyOf(40).encodeHex()).isEqualTo(expectedPrefixHex)
    assertThat(MessageDigest.getInstance("SHA-256").digest(blob).encodeHex()).isEqualTo(expectedSha256Hex)
    assertThat(BlobPayloadPacking.unpack(blob)).isEqualTo(payload)
  }

  @Test
  fun `pack rejects a payload above capacity`() {
    assertThatThrownBy { BlobPayloadPacking.pack(ByteArray(BlobPayloadPacking.BLOB_PAYLOAD_CAPACITY + 1)) }
      .isInstanceOf(IllegalArgumentException::class.java)
      .hasMessageContaining("exceeds packing capacity")
  }

  @Test
  fun `unpack rejects a blob without terminal`() {
    assertThatThrownBy { BlobPayloadPacking.unpack(ByteArray(Constants.Eip4844BlobSize)) }
      .isInstanceOf(IllegalArgumentException::class.java)
      .hasMessageContaining("invalid blob terminal symbol")
  }

  @Test
  fun `unpack rejects a field element wider than 254 bits`() {
    val blob = BlobPayloadPacking.pack(byteArrayOf(0x01)).also { it[32] = 0x80.toByte() }

    assertThatThrownBy { BlobPayloadPacking.unpack(blob) }
      .isInstanceOf(IllegalArgumentException::class.java)
      .hasMessageContaining("exceeds 254 bits")
  }
}
