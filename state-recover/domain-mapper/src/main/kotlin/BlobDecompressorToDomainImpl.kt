import build.linea.staterecover.core.BlobDecompressorToDomain
import build.linea.staterecover.core.BlockL1RecoveredData
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions
import org.hyperledger.besu.ethereum.rlp.RLP
import org.hyperledger.besu.ethereum.rlp.RLPException

class BlobDecompressorToDomainImpl : BlobDecompressorToDomain {
  override fun decompress(blobs: List<ByteArray>): List<BlockL1RecoveredData> {
    TODO("Not yet implemented")
  }

  override fun decompress(blob: ByteArray): List<BlockL1RecoveredData> {
    val blobBytes = Bytes.wrap(blob)
    try {
      RLP.validate(blobBytes)
    } catch (e: RLPException) {
      throw IllegalArgumentException("Invalid RLP encoded blob", e)
    }

    val blocks = mutableListOf<Block>()
    val rlpInput = RLP.input(blobBytes)
    while (rlpInput.nextIsList()) {
      val block = Block.readFrom(rlpInput, MainnetBlockHeaderFunctions())
      blocks.add(block)
    }

    return blocks.map { it.toDomain() }
  }

  fun decodeRLPBlock(encodedBlock: ByteArray): Block {
    val blockBytes = Bytes.wrap(encodedBlock)
    try {
      RLP.validate(blockBytes)
    } catch (e: RLPException) {
      throw IllegalArgumentException("Invalid RLP encoded block", e)
    }
    val rlpInput = RLP.input(blockBytes)
    val decodedBlock = Block.readFrom(rlpInput, MainnetBlockHeaderFunctions())
    return decodedBlock
  }
}
