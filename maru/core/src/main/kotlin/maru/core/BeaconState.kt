package maru.consensus.core

/**
 * After every BeaconBlock there is a transition in the BeaconState by applying the operations from
 * the BeaconBlock These operations could be a new execution payload, adding/removing validators
 * etc.
 */
data class BeaconState(
  val latestBeaconBlockHeader: BeaconBlockHeader,
  val latestBeaconBlockRoot: ByteArray,
  val validators: Set<Validator>,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as BeaconState

    if (latestBeaconBlockHeader != other.latestBeaconBlockHeader) return false
    if (!latestBeaconBlockRoot.contentEquals(other.latestBeaconBlockRoot)) return false
    if (validators != other.validators) return false

    return true
  }

  override fun hashCode(): Int {
    var result = latestBeaconBlockHeader.hashCode()
    result = 31 * result + latestBeaconBlockRoot.contentHashCode()
    result = 31 * result + validators.hashCode()
    return result
  }
}
