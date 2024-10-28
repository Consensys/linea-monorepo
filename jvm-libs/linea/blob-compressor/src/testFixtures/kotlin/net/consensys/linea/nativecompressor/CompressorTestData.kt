package net.consensys.linea.nativecompressor

import java.nio.ByteBuffer
import java.nio.ByteOrder

object CompressorTestData {
  val blocksRlpEncoded: Array<ByteArray> = loadTestData()

  private fun loadTestData(): Array<ByteArray> {
    val data = Thread.currentThread().getContextClassLoader().getResourceAsStream("rlp_blocks.bin")!!.readAllBytes()

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
