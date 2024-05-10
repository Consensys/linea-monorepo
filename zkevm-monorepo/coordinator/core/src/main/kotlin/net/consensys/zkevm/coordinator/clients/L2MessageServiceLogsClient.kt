package net.consensys.zkevm.coordinator.clients

import net.consensys.zkevm.domain.BridgeLogsData
import net.consensys.zkevm.domain.L2RollingHashUpdatedEvent
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface L2MessageServiceLogsClient {
  fun getBridgeLogs(
    blockNumber: Long
  ): SafeFuture<List<BridgeLogsData>>

  @Deprecated(
    "Use findLastRollingHashUpdatedEvent instead. " +
      "Just keeping this untill all test pass we are confident that findLastAnchoredMessageUpToBlock can replace this"
  )
  fun findLastRollingHashUpdatedEvent(
    upToBlockNumberInclusive: Long,
    lookBackBlockNumberLimitInclusive: Long = upToBlockNumberInclusive
  ): SafeFuture<L2RollingHashUpdatedEvent?>
}

interface L2MessageServiceClient : L2MessageServiceLogsClient {

  /**
   * Returns the last anchored message up to the given block number (inclusive).
   * If not message is found, returns a message with message number 0 and an empty rolling hash.
   * Which corresponds to the smart contract value.
   *
   * @param blockNumberInclusive the block number up to which to search for the last anchored message
   * @return the last anchored message up to the given block number (inclusive)
   */
  fun getLastAnchoredMessageUpToBlock(blockNumberInclusive: Long): SafeFuture<L2RollingHashUpdatedEvent>
}
