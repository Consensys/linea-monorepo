package linea.domain

import kotlinx.datetime.Instant
import net.consensys.linea.BlockNumberAndHash

data class Block(
  val number: ULong,
  val hash: ByteArray,
  val parentHash: ByteArray,
  val ommersHash: ByteArray,
  val miner: ByteArray,
  val stateRoot: ByteArray,
  val transactionsRoot: ByteArray,
  val receiptsRoot: ByteArray,
  val logsBloom: ByteArray,
  val difficulty: ULong,
  val gasLimit: ULong,
  val gasUsed: ULong,
  val timestamp: ULong,
  val extraData: ByteArray,
  val mixHash: ByteArray,
  val nonce: ULong,
  val baseFeePerGas: ULong? = null, // Optional field for EIP-1559 blocks
  val transactions: List<Transaction> = emptyList(), // List of transaction hashes
  val ommers: List<ByteArray> = emptyList() // List of uncle block hashes
) {
  companion object {
    // companion object  to allow static extension functions
  }

  val numberAndHash = BlockNumberAndHash(this.number, this.hash)
  val headerSummary = BlockHeaderSummary(this.number, this.hash, Instant.fromEpochSeconds(this.timestamp.toLong()))
}

data class BlockHeaderSummary(
  val number: ULong,
  val hash: ByteArray,
  val timestamp: Instant
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as BlockHeaderSummary

    if (number != other.number) return false
    if (!hash.contentEquals(other.hash)) return false
    if (timestamp != other.timestamp) return false

    return true
  }

  override fun hashCode(): Int {
    var result = number.hashCode()
    result = 31 * result + hash.contentHashCode()
    result = 31 * result + timestamp.hashCode()
    return result
  }

  override fun toString(): String {
    return "BlockHeaderSummary(number=$number, hash=${hash.contentToString()}, timestamp=$timestamp)"
  }
}
