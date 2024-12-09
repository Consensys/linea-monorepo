package linea.rlp

import linea.domain.Block
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface RLPBlockEncoder {
  fun encode(block: Block): ByteArray
  fun encode(blocks: List<Block>): List<ByteArray> = blocks.map { encode(it) }
  fun decode(block: ByteArray): Block
  fun decode(blocks: List<ByteArray>): List<Block> = blocks.map { decode(it) }
}

interface RLPBlockEncoderAsync {
  fun encodeAsync(block: Block): SafeFuture<ByteArray>
  fun encodeAsync(blocks: List<Block>): SafeFuture<List<ByteArray>>
  fun decodeAsync(block: ByteArray): SafeFuture<Block>
  fun decodeAsync(blocks: List<ByteArray>): SafeFuture<List<Block>>
}
