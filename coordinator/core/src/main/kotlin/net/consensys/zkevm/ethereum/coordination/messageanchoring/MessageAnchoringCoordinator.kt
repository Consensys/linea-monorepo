package net.consensys.zkevm.ethereum.coordination.messageanchoring

import org.apache.tuweni.bytes.Bytes32
import org.web3j.protocol.core.methods.response.TransactionReceipt
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

data class SendMessageEvent(val messageHash: Bytes32)

data class MessageHashAnchoredEvent(val messageHash: Bytes32)

interface L1EventQuerier {
  fun getSendMessageEventsForAnchoredMessage(
    messageHash: MessageHashAnchoredEvent?
  ): SafeFuture<List<SendMessageEvent>>
}

interface L2MessageAnchorer {
  // Anchors message hashes with the coordinator client
  // "L2 Version 2 Coordinator for message anchoring" in 1 password is the signer
  fun anchorMessages(messageHashes: List<Bytes32>): SafeFuture<TransactionReceipt>
}

// todo discuss removing this as it is a V2 implementation and it would be good to clean up
interface AnchoringRepository {
  // Finds the last Anchored message event
  fun findLastContinuouslyAnchoredMessageHash(): SafeFuture<MessageHashAnchoredEvent?>

  // Insert the list of hashes and prune the continuously sent hashes atomically
  fun trackMessageHashes(numberedMessageHashes: List<SendMessageEvent>): SafeFuture<Unit>
}

interface L2Querier {
  // Finds the last Anchored message event
  fun findLastFinalizedAnchoredEvent(): SafeFuture<MessageHashAnchoredEvent?>

  // Retrieves a message hash status
  // This may not actually be needed due to the idempotent ignoring of non-unknown status hashes
  fun getMessageHashStatus(messageHash: Bytes32): SafeFuture<BigInteger>
}
