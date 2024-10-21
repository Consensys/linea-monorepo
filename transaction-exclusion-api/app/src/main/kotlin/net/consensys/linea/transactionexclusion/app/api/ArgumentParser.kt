package net.consensys.linea.transactionexclusion.app.api

import kotlinx.datetime.Instant
import net.consensys.assertIs32Bytes
import net.consensys.decodeHex
import net.consensys.linea.transactionexclusion.ModuleOverflow
import net.consensys.linea.transactionexclusion.RejectedTransaction
import net.consensys.linea.transactionexclusion.TransactionInfo
import net.consensys.linea.transactionexclusion.dto.ModuleOverflowJsonDto
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.ethereum.core.encoding.TransactionDecoder
import java.time.format.DateTimeFormatter
import java.time.format.DateTimeParseException

object ArgumentParser {
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
      return ModuleOverflowJsonDto.parseListFromJsonString(
        ModuleOverflowJsonDto.parseToJsonString(target)
      ).map { it.toDomainObject() }
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

  fun getBlockNumber(blockNumberStr: String?): ULong? {
    try {
      return blockNumberStr?.toULong()
    } catch (ex: NumberFormatException) {
      throw IllegalArgumentException("Block number cannot be parsed to an unsigned number: ${ex.message}")
    }
  }

  fun getTimestampFromISO8601(timestamp: String): Instant {
    try {
      DateTimeFormatter.ISO_DATE_TIME.parse(timestamp)
      return Instant.parse(timestamp)
    } catch (ex: DateTimeParseException) {
      throw IllegalArgumentException("Timestamp is not in ISO-8601: ${ex.message}")
    }
  }

  fun getTxRejectionStage(txRejectionStage: String): RejectedTransaction.Stage {
    try {
      return RejectedTransaction.Stage.valueOf(txRejectionStage)
    } catch (ex: IllegalArgumentException) {
      throw IllegalArgumentException("Unsupported transaction rejection stage: $txRejectionStage")
    }
  }
}
