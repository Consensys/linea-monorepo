package linea.anchoring.clients

import linea.EthLogsSearcher
import linea.anchoring.events.L1RollingHashUpdatedEvent
import linea.anchoring.events.MessageSentEvent
import linea.domain.BlockParameter
import linea.domain.BlockParameter.Companion.toBlockParameter
import linea.domain.EthLogEvent
import linea.kotlin.toHexStringUInt256
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration

/**
 * This class is responsible for fetching L1 MessageSent events from the Ethereum blockchain.
 * It leverages the invariant of increasing message numbers to efficiently search for events on L1 by never going back
 *
 * This is relevant when MessageSent events are very sparse on the chain, it requires 10_000K l1 blocks
 * to collect 100 messages.
 */
internal class L1MessageSentEventsFetcher(
  private val l1SmartContractAddress: String,
  private val l1EventsSearcher: EthLogsSearcher,
  private val highestBlockNumber: BlockParameter
) {
  private val lastSearch: AtomicPair = AtomicPair()

  private class AtomicPair(
    private var highestL1AlreadySearchedBlockNumber: ULong = 0UL,
    private var lastStartingMessageNumber: ULong = 0UL
  ) {
    @Synchronized
    fun lastStartingMessageNumber(): ULong = lastStartingMessageNumber

    @Synchronized
    fun highestL1AlreadySearchedBlockNumber(): ULong = highestL1AlreadySearchedBlockNumber

    @Synchronized
    fun set(
      highestL1AlreadySearchedBlockNumber: ULong,
      lastStartingMessageNumber: ULong
    ) {
      this.highestL1AlreadySearchedBlockNumber = highestL1AlreadySearchedBlockNumber
      this.lastStartingMessageNumber = lastStartingMessageNumber
    }
  }

  fun findL1MessageSentEvents(
    startingMessageNumber: ULong,
    messagesToFetch: UInt,
    fetchTimeout: Duration,
    blockChunkSize: UInt
  ): SafeFuture<List<EthLogEvent<MessageSentEvent>>> {
    require(startingMessageNumber >= lastSearch.lastStartingMessageNumber()) {
      "startingMessageNumber=$startingMessageNumber must greater than " +
        "or equal to lastStartingMessageNumber=${lastSearch.lastStartingMessageNumber()}"
    }

    return findL1RollingHashUpdatedEvent(
      fromBlock = lastSearch.highestL1AlreadySearchedBlockNumber(),
      messageNumber = startingMessageNumber
    ).thenCompose { event ->
      if (event == null) {
        return@thenCompose SafeFuture.completedFuture(emptyList())
      }

      l1EventsSearcher.getLogsRollingForward(
        fromBlock = event.log.blockNumber.toBlockParameter(),
        toBlock = highestBlockNumber,
        address = l1SmartContractAddress,
        topics = listOf(
          MessageSentEvent.topic
        ),
        chunkSize = blockChunkSize,
        searchTimeout = fetchTimeout,
        stopAfterTargetLogsCount = messagesToFetch
      ).thenApply { result ->
        lastSearch.set(
          highestL1AlreadySearchedBlockNumber = result.endBlockNumber,
          lastStartingMessageNumber = startingMessageNumber
        )

        result.logs.map(MessageSentEvent::fromEthLog)
      }
    }
  }

  private fun findL1RollingHashUpdatedEvent(
    fromBlock: ULong,
    messageNumber: ULong
  ): SafeFuture<EthLogEvent<L1RollingHashUpdatedEvent>?> {
    return l1EventsSearcher.getLogs(
      fromBlock = fromBlock.toBlockParameter(),
      toBlock = BlockParameter.Tag.FINALIZED,
      address = l1SmartContractAddress,
      topics = listOf(
        L1RollingHashUpdatedEvent.topic,
        messageNumber.toHexStringUInt256()
      )
    ).thenApply {
      it.firstOrNull()?.let { log -> L1RollingHashUpdatedEvent.fromEthLog(log) }
    }
  }
}
