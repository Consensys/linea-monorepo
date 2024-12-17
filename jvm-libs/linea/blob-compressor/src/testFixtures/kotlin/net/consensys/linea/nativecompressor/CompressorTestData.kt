package net.consensys.linea.nativecompressor

import linea.rlp.RLP
import java.nio.ByteBuffer
import java.nio.ByteOrder

object CompressorTestData {
  val blocksRlpEncoded: Array<ByteArray> = loadTestData()
  val blocksRlpEncodedV2: List<ByteArray> = RLP.decodeList(readResourcesFile("blocks_rlp.bin"))

  private fun readResourcesFile(fileName: String): ByteArray {
    return Thread.currentThread().getContextClassLoader().getResourceAsStream(fileName)
      ?.readAllBytes()
      ?: throw IllegalArgumentException("File not found in jar resources: file=$fileName")
  }

  private fun loadTestData(): Array<ByteArray> {
    val data = readResourcesFile("rlp_blocks.bin")

    // first 4 bytes are the number of blocks
    val numBlocks = ByteBuffer.wrap(data, 0, 4).order(ByteOrder.LITTLE_ENDIAN).int

    // the rest of the file is the blocks
    // (we repeat them to fill more data)
    val blocks = Array(numBlocks * 2) { ByteArray(0) }

    for (j in 0 until 2) {
      var offset = 4
      for (i in 0 until numBlocks) {
        // first 4 bytes are the length of the block
        val blockLen = ByteBuffer.wrap(data, offset, 4).order(ByteOrder.LITTLE_ENDIAN).int

        // the rest of the block is the block
        blocks[i + j * numBlocks] = ByteArray(blockLen)
        System.arraycopy(data, offset + 4, blocks[i + j * numBlocks], 0, blockLen)
        offset += 4 + blockLen
      }
    }
    return blocks
  }
}
