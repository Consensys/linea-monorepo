package net.consensys.zkevm.ethereum.coordination.conflation.upgrade

import tech.pegasys.teku.infrastructure.async.SafeFuture

@Deprecated("We may use it for future switches, but maybe it won't ever be useful")
@Suppress("DEPRECATION")
fun interface SwitchProvider {
  enum class ProtocolSwitches(val int: UInt) {
    INITIAL_VERSION(1U),
    DATA_COMPRESSION_PROOF_AGGREGATION(2U)
  }
  fun getSwitch(version: ProtocolSwitches): SafeFuture<ULong?>
}
