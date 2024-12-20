package maru.consensus.core

data class Validator(val address: ByteArray) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as Validator

    return address.contentEquals(other.address)
  }

  override fun hashCode(): Int {
    return address.contentHashCode()
  }
}
