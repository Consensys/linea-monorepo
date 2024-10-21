package net.consensys.linea.blob

import net.consensys.decodeHex
import net.consensys.encodeHex
import net.consensys.linea.nativecompressor.CompressorTestData
import net.consensys.linea.testing.filesystem.findPathTo
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.ethereum.rlp.RLP
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import java.security.MessageDigest
import kotlin.io.path.readBytes

class BlobDecompressorDataDecodingTest {
  private lateinit var decompressor: BlobDecompressor

  @BeforeEach
  fun beforeEach() {
    decompressor = GoNativeBlobDecompressorFactory.getInstance(BlobDecompressorVersion.V1_1_0)
  }

  @Test
  fun `can deserialize native lib testdata blobs`() {
    val blob = findPathTo("prover")!!
      .resolve("lib/compressor/blob/testdata/v0/sample-blob-0151eda71505187b5.bin")
      .readBytes()

    val decompressedBlob = decompressor.decompress(blob)
    // size and Hash extracted from GoTests
    assertThat(decompressedBlob).hasSize(603023)
    val sha256 = MessageDigest.getInstance("SHA-256").digest(decompressedBlob).encodeHex()
    assertThat(sha256).isEqualTo("0xa70c413ccbf15aedad1ef4bbfc6e94bd238aad1798b3836e8a52653b34df1158")
    val blockRlpEncoded = rlpDecodeAsListOfBytes(decompressedBlob)
    assertThat(blockRlpEncoded)
    println(blockRlpEncoded.first().encodeHex())
  }

  @Test
  fun rlpEncodeAndDecode() {
    val bytesArray = listOf(
      "0xff010203040506070809aa".decodeHex(),
      "0xff010203040506070809ab".decodeHex(),
      "0xff010203040506070809ac".decodeHex(),
      "0xff010203040506070809acff010203040506070809ac".decodeHex()
    )

    assertThat(rlpDecodeAsListOfBytes(rlpEncode(bytesArray))).hasSameElementsAs(bytesArray)

    val blocksRlpEnconded = CompressorTestData.blocksRlpEncoded.toList()
    assertThat(rlpDecodeAsListOfBytes(rlpEncode(blocksRlpEnconded))).hasSameElementsAs(blocksRlpEnconded)
  }

  private fun rlpEncode(list: List<ByteArray>): ByteArray {
    return RLP.encode { rlpWriter ->
      rlpWriter.startList()
      list.forEach { bytes ->
        rlpWriter.writeBytes(Bytes.wrap(bytes))
      }
      rlpWriter.endList()
    }.toArray()
  }

  private fun rlpDecodeAsListOfBytes(rlpEncoded: ByteArray): List<ByteArray> {
    val decodedBytes = mutableListOf<ByteArray>()
    RLP.input(Bytes.wrap(rlpEncoded), true).also { rlpInput ->
      rlpInput.enterList()
      while (!rlpInput.isEndOfCurrentList) {
        decodedBytes.add(rlpInput.readBytes().toArray())
      }
      rlpInput.leaveList()
    }
    return decodedBytes
  }
}
