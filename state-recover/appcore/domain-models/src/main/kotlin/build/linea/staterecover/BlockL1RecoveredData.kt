package build.linea.staterecover

import kotlinx.datetime.Instant
import net.consensys.encodeHex

data class BlockExtraData(
  val beneficiary: ByteArray
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as BlockExtraData

    return beneficiary.contentEquals(other.beneficiary)
  }

  override fun hashCode(): Int {
    return beneficiary.contentHashCode()
  }

  override fun toString(): String {
    return "BlockExtraData(beneficiary=${beneficiary.encodeHex()})"
  }
}

data class BlockL1RecoveredData(
  val blockNumber: ULong,
  val blockHash: ByteArray,
  val coinbase: ByteArray,
  val blockTimestamp: Instant,
  val gasLimit: ULong,
  val difficulty: ULong,
  val extraData: BlockExtraData,
  val transactions: List<TransactionL1RecoveredData>
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as BlockL1RecoveredData

    if (blockNumber != other.blockNumber) return false
    if (!blockHash.contentEquals(other.blockHash)) return false
    if (!coinbase.contentEquals(other.coinbase)) return false
    if (blockTimestamp != other.blockTimestamp) return false
    if (gasLimit != other.gasLimit) return false
    if (difficulty != other.difficulty) return false
    if (extraData != other.extraData) return false
    if (transactions != other.transactions) return false

    return true
  }

  override fun hashCode(): Int {
    var result = blockNumber.hashCode()
    result = 31 * result + blockHash.contentHashCode()
    result = 31 * result + coinbase.contentHashCode()
    result = 31 * result + blockTimestamp.hashCode()
    result = 31 * result + gasLimit.hashCode()
    result = 31 * result + difficulty.hashCode()
    result = 31 * result + extraData.hashCode()
    result = 31 * result + transactions.hashCode()
    return result
  }

  override fun toString(): String {
    return "BlockL1RecoveredData(" +
      "blockNumber=$blockNumber, " +
      "blockHash=${blockHash.encodeHex()}, " +
      "coinbase=${coinbase.encodeHex()}, " +
      "blockTimestamp=$blockTimestamp, " +
      "gasLimit=$gasLimit, " +
      "difficulty=$difficulty, " +
      "extraData=$extraData, " +
      "transactions=$transactions)"
  }
}
