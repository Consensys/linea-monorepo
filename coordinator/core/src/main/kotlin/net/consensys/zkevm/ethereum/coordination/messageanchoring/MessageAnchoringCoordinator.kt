package net.consensys.zkevm.ethereum.coordination.messageanchoring

import org.apache.tuweni.bytes.Bytes32
import org.web3j.protocol.core.methods.response.TransactionReceipt
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

data class SendMessageEvent(val messageHash: Bytes32, val messageNumber: ULong, val blockNumber: ULong)

data class MessageHashAnchoredEvent(val messageHash: Bytes32)

interface L1EventQuerier {
  fun getSendMessageEventsForAnchoredMessage(
    messageHash: MessageHashAnchoredEvent?
  ): SafeFuture<List<SendMessageEvent>>
}

interface L2MessageAnchorer {
  /**
   * Anchor L1 messages into L2 by sending the message hashes and the final rolling hash
   * Wait until the transaction is safe on L2
   */
  fun anchorMessages(
    sendMessageEvents: List<SendMessageEvent>,
    finalRollingHash: ByteArray
  ): SafeFuture<TransactionReceipt>
}

interface L2Querier {
  // Finds the last Anchored message event
  fun findLastFinalizedAnchoredEvent(): SafeFuture<MessageHashAnchoredEvent?>

  // Retrieves a message hash status
  // This may not actually be needed due to the idempotent ignoring of non-unknown status hashes
  fun getMessageHashStatus(messageHash: Bytes32): SafeFuture<BigInteger>
}
