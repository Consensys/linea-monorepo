package net.consensys.linea.transactionexclusion.app.api

import kotlinx.datetime.Instant
import net.consensys.assertIs32Bytes
import net.consensys.decodeHex
import net.consensys.fromHexString
import net.consensys.linea.ModuleOverflow
import net.consensys.linea.RejectedTransaction
import net.consensys.linea.TransactionInfo
import net.consensys.linea.jsonrpc.argument.getArgument
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.ethereum.core.encoding.TransactionDecoder
import java.time.format.DateTimeFormatter
import java.time.format.DateTimeParseException

object ArgumentParser {
  fun getHexString(arguments: List<*>, index: Int, argumentName: String): String {
    return getArgument(String::class, arguments, index, argumentName)
      .also { require(it.startsWith("0x")) { "$argumentName must have '0x' hexadecimal prefix." } }
  }

  fun <T> getHexStringParsed(arguments: List<*>, index: Int, argumentName: String, parser: (String) -> T): T {
    return getArgument(String::class, arguments, index, argumentName)
      .also { require(it.startsWith("0x")) { "$argumentName must have '0x' hexadecimal prefix." } }
      .let(parser)
  }

  fun getBlockNumber(arguments: List<*>, index: Int, argumentName: String): ULong {
    return getHexStringParsed(arguments, index, argumentName) {
      try {
        ULong.fromHexString(it)
      } catch (ex: NumberFormatException) {
        throw NumberFormatException("${ex.message} on argument $argumentName")
      }
    }
  }

  fun getBytes32(arguments: List<*>, index: Int, argumentName: String): Bytes32 {
    return getHexStringParsed(arguments, index, argumentName, Bytes32::fromHexString)
  }

  fun getTransactionRLPInRawBytes(transactionRLP: String): ByteArray {
    try {
      return transactionRLP.decodeHex()
    } catch (ex: Exception) {
      throw IllegalArgumentException("Hex string of RLP-encoded transaction cannot be parsed: ${ex.message}")
    }
  }

  fun getTxHashInRawBytes(txHash: String): ByteArray {
    try {
      return txHash.decodeHex().assertIs32Bytes()
    } catch (ex: Exception) {
      throw IllegalArgumentException("Hex string of transaction hash cannot be parsed: ${ex.message}")
    }
  }

  fun getTransactionInfoFromRLP(rlp: ByteArray): TransactionInfo {
    try {
      return TransactionDecoder.decodeOpaqueBytes(
        Bytes.wrap(rlp)
      ).run {
        TransactionInfo(
          hash = this.hash.toArray(),
          to = this.to.get().toArray(),
          from = this.sender.toArray(),
          nonce = this.nonce.toULong()
        )
      }
    } catch (ex: Exception) {
      throw IllegalArgumentException("RLP-encoded transaction cannot be parsed: ${ex.message}")
    }
  }

  fun getOverflows(target: Any): List<ModuleOverflow> {
    try {
      return ModuleOverflow.parseListFromJsonString(
        ModuleOverflow.parseToJsonString(target)
      )
    } catch (ex: Exception) {
      throw IllegalArgumentException("Overflows cannot be parsed: ${ex.message}")
    }
  }

  fun getReasonMessage(reasonMessage: String): String {
    if (reasonMessage.length > 256) {
      throw IllegalArgumentException("Reason message should not be more than 256 characters: $reasonMessage")
    }
    return reasonMessage
  }

  fun getBlockNumber(blockNumberStr: String): ULong {
    try {
      return blockNumberStr.toULong()
    } catch (ex: NumberFormatException) {
      throw IllegalArgumentException("Block number cannot be parsed to an unsigned number: ${ex.message}")
    }
  }

  fun getTimestampFromISO8601(timestamp: String): Instant {
    return try {
      DateTimeFormatter.ISO_DATE_TIME.parse(timestamp)
      Instant.parse(timestamp)
    } catch (ex: DateTimeParseException) {
      throw IllegalArgumentException("Timestamp is not in ISO-8601: ${ex.message}")
    }
  }

  fun geRejectedTransactionStage(stage: String): RejectedTransaction.Stage {
    return when (stage) {
      "SEQUENCER" -> RejectedTransaction.Stage.Sequencer
      "RPC" -> RejectedTransaction.Stage.Rpc
      "P2P" -> RejectedTransaction.Stage.P2p
      else -> throw IllegalArgumentException("Unsupported rejected transaction stage: $stage")
    }
  }
}
