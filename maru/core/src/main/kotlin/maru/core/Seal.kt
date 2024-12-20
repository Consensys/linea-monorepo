package maru.consensus.core

data class Seal(val signature: ByteArray) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as Seal

    return signature.contentEquals(other.signature)
  }

  override fun hashCode(): Int {
    return signature.contentHashCode()
  }
}
