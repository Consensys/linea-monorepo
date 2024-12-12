package linea.rlp

import tech.pegasys.teku.infrastructure.async.SafeFuture

interface BlockRlpEncoder<T> {
  fun encode(block: T): ByteArray
  fun encode(blocks: List<T>): List<ByteArray> = blocks.map { encode(it) }
}

interface BlockRlpDecoder<T> {
  fun decode(block: ByteArray): T
  fun decode(blocks: List<ByteArray>): List<T> = blocks.map { decode(it) }
}

interface BlockRlpEncoderAsync<T> {
  fun encodeAsync(block: T): SafeFuture<ByteArray>
  fun encodeAsync(blocks: List<T>): SafeFuture<List<ByteArray>> =
    SafeFuture.collectAll(blocks.map { encodeAsync(it) }.stream())
}

interface BlockRlpDecoderAsync<T> {
  fun decodeAsync(block: ByteArray): SafeFuture<T>
  fun decodeAsync(blocks: List<ByteArray>): SafeFuture<List<T>> =
    SafeFuture.collectAll(blocks.map { decodeAsync(it) }.stream())
}

interface BesuBlockRlpEncoder : BlockRlpEncoder<org.hyperledger.besu.ethereum.core.Block>
interface BesuBlockRlpEncoderAsync : BlockRlpEncoderAsync<org.hyperledger.besu.ethereum.core.Block>
interface BesuBlockRlpDecoder : BlockRlpDecoder<org.hyperledger.besu.ethereum.core.Block>
interface BesuBlockRlpDecoderAsync : BlockRlpDecoderAsync<org.hyperledger.besu.ethereum.core.Block>
