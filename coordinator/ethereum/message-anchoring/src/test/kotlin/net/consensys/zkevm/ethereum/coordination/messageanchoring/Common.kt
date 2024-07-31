package net.consensys.zkevm.ethereum.coordination.messageanchoring

import org.apache.tuweni.bytes.Bytes32

fun createRandomSendMessageEvents(numberOfRandomHashes: ULong): List<SendMessageEvent> {
  return (0UL..numberOfRandomHashes)
    .map { n ->
      SendMessageEvent(
        Bytes32.random(),
        messageNumber = n + 1UL,
        blockNumber = n + 1UL
      )
    }
}
