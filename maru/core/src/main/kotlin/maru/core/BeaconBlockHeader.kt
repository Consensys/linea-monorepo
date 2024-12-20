package maru.consensus.core

data class BeaconBlockHeader(
  val number: ULong,
  val round: ULong,
  val proposer: Validator,
  val parentRoot: ByteArray,
  val stateRoot: ByteArray,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as BeaconBlockHeader

    if (number != other.number) return false
    if (round != other.round) return false
    if (proposer != other.proposer) return false
    if (!parentRoot.contentEquals(other.parentRoot)) return false
    if (!stateRoot.contentEquals(other.stateRoot)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = number.hashCode()
    result = 31 * result + round.hashCode()
    result = 31 * result + proposer.hashCode()
    result = 31 * result + parentRoot.contentHashCode()
    result = 31 * result + stateRoot.contentHashCode()
    return result
  }
}
