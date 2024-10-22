package net.consensys.linea.blob

import net.consensys.linea.testing.filesystem.findPathTo
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions
import org.hyperledger.besu.ethereum.rlp.RLP
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Disabled
import kotlin.io.path.readBytes

class BlobDecompressorDataDecodingTest {
  private lateinit var decompressor: BlobDecompressor

  @BeforeEach
  fun beforeEach() {
    decompressor = GoNativeBlobDecompressorFactory.getInstance(BlobDecompressorVersion.V1_1_0)
  }

  @Disabled("Until Besu supports deserializing transactions without signatures validation")
  fun `can deserialize native lib testdata blobs`() {
    val blob = findPathTo("prover")!!
      .resolve("lib/compressor/blob/testdata/v0/sample-blob-0151eda71505187b5.bin")
      .readBytes()
    val decompressedBlob = decompressor.decompress(blob)
    val blocksRlpEncoded = rlpDecodeAsListOfBytes(decompressedBlob)
    blocksRlpEncoded.forEachIndexed { index, blockRlp ->
      val rlpInput = RLP.input(Bytes.wrap(blockRlp))
      val decodedBlock = Block.readFrom(rlpInput, MainnetBlockHeaderFunctions())
      println("$index: $decodedBlock")
    }
  }

  @Disabled("for local dev validation")
  fun `can decode  RLP`() {
    val blockBytes = Bytes.wrap(
      // INSERT HERE THE RLP ENCODED BLOCK
      // 0x01ff.decodeHex()
    )
    RLP.validate(blockBytes)
    val rlpInput = RLP.input(blockBytes)
    val decodedBlock = Block.readFrom(rlpInput, MainnetBlockHeaderFunctions())
    println(decodedBlock)
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
