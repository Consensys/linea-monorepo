package net.consensys.linea

import com.fasterxml.jackson.annotation.JsonProperty
import com.fasterxml.jackson.databind.ObjectMapper
import kotlinx.datetime.Instant
import net.consensys.encodeHex

data class ModuleOverflow(
  @JsonProperty("module")
  val module: String,
  @JsonProperty("count")
  val count: Long,
  @JsonProperty("limit")
  val limit: Long
) {
  // Jackson ObjectMapper requires a default constructor
  constructor() : this("", 0L, 0L)

  override fun toString(): String {
    return "module=$module count=$count limit=$limit"
  }

  companion object {
    fun parseListFromJsonString(jsonString: String): List<ModuleOverflow> {
      return ObjectMapper().readValue(
        jsonString,
        Array<ModuleOverflow>::class.java
      ).toList()
    }

    fun parseToJsonString(target: Any): String {
      if (target is String) {
        return target
      }
      return ObjectMapper().writeValueAsString(target)
    }
  }
}

data class TransactionInfo(
  val hash: ByteArray,
  val from: ByteArray,
  val to: ByteArray,
  val nonce: ULong
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as TransactionInfo

    if (!hash.contentEquals(other.hash)) return false
    if (!from.contentEquals(other.from)) return false
    if (!to.contentEquals(other.to)) return false
    return nonce == other.nonce
  }

  override fun hashCode(): Int {
    var result = hash.contentHashCode()
    result = 31 * result + from.contentHashCode()
    result = 31 * result + to.contentHashCode()
    result = 31 * result + nonce.hashCode()
    return result
  }

  override fun toString(): String {
    return "hash=${hash.encodeHex()} from=${from.encodeHex()} " +
      "to=${to.encodeHex()} nonce=$nonce"
  }
}

data class RejectedTransaction(
  val txRejectionStage: Stage,
  val timestamp: Instant,
  val blockNumber: ULong?,
  val transactionRLP: ByteArray,
  val reasonMessage: String,
  val overflows: List<ModuleOverflow>,
  var transactionInfo: TransactionInfo? = null
) {
  enum class Stage {
    SEQUENCER,
    RPC,
    P2P
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as RejectedTransaction

    if (txRejectionStage != other.txRejectionStage) return false
    if (timestamp != other.timestamp) return false
    if (blockNumber != other.blockNumber) return false
    if (!transactionRLP.contentEquals(other.transactionRLP)) return false
    if (reasonMessage != other.reasonMessage) return false
    return overflows == other.overflows
  }

  override fun hashCode(): Int {
    var result = txRejectionStage.hashCode()
    result = 31 * result + timestamp.hashCode()
    result = 31 * result + blockNumber.hashCode()
    result = 31 * result + transactionRLP.contentHashCode()
    result = 31 * result + reasonMessage.hashCode()
    result = 31 * result + overflows.hashCode()
    return result
  }

  override fun toString(): String {
    return "txRejectionStage=$txRejectionStage timestamp=${timestamp.toEpochMilliseconds()} blockNumber=$blockNumber" +
      " transactionRLP=${transactionRLP.encodeHex()}" +
      " transactionInfo=$transactionInfo" +
      " reasonMessage=\"$reasonMessage\"" +
      " overflows=[${overflows.joinToString(",") { "{$it}" }}]"
  }
}
