package build.linea.staterecover

import java.math.BigInteger

data class TransactionL1RecoveredData(
  val type: UByte,
  val nonce: ULong,
  val maxPriorityFeePerGas: BigInteger,
  val maxFeePerGas: BigInteger,
  val gasLimit: ULong,
  val from: ByteArray,
  val to: ByteArray,
  val value: BigInteger,
  val data: ByteArray,
  val accessList: List<AccessTuple>
) {

  data class AccessTuple(
    val address: ByteArray,
    val storageKeys: List<ByteArray>
  ) {
    override fun equals(other: Any?): Boolean {
      if (this === other) return true
      if (javaClass != other?.javaClass) return false

      other as AccessTuple

      if (!address.contentEquals(other.address)) return false
      if (storageKeys != other.storageKeys) return false

      return true
    }

    override fun hashCode(): Int {
      var result = address.contentHashCode()
      result = 31 * result + storageKeys.hashCode()
      return result
    }
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as TransactionL1RecoveredData

    if (type != other.type) return false
    if (nonce != other.nonce) return false
    if (maxPriorityFeePerGas != other.maxPriorityFeePerGas) return false
    if (maxFeePerGas != other.maxFeePerGas) return false
    if (gasLimit != other.gasLimit) return false
    if (!from.contentEquals(other.from)) return false
    if (!to.contentEquals(other.to)) return false
    if (value != other.value) return false
    if (!data.contentEquals(other.data)) return false
    if (accessList != other.accessList) return false

    return true
  }

  override fun hashCode(): Int {
    var result = type.hashCode()
    result = 31 * result + nonce.hashCode()
    result = 31 * result + maxPriorityFeePerGas.hashCode()
    result = 31 * result + maxFeePerGas.hashCode()
    result = 31 * result + gasLimit.hashCode()
    result = 31 * result + from.contentHashCode()
    result = 31 * result + to.contentHashCode()
    result = 31 * result + value.hashCode()
    result = 31 * result + data.contentHashCode()
    result = 31 * result + accessList.hashCode()
    return result
  }
}
