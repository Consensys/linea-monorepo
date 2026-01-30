package net.consensys.zkevm.coordinator.clients

import io.vertx.core.json.JsonArray
import io.vertx.core.json.JsonObject
import linea.forcedtx.ForcedTransactionInclusionResult
import linea.forcedtx.ForcedTransactionInclusionStatus
import linea.forcedtx.ForcedTransactionResponse
import linea.kotlin.decodeHex
import linea.kotlin.toULongFromHex
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

object ForcedTransactionsResponseParser {

  /**
   * Parses the response from linea_sendForcedRawTransaction RPC call.
   *
   * Expected response format:
   * ```json
   * [
   *   {"forcedTransactionNumber": 6, "hash": "0xabc123..."},
   *   {"forcedTransactionNumber": 7, "error": "Invalid nonce"}
   * ]
   * ```
   */
  fun parseSendForcedRawTransactionResponse(result: Any?): List<ForcedTransactionResponse> {
    if (result == null) {
      return emptyList()
    }

    val resultArray = result as JsonArray
    return resultArray.map { item ->
      val jsonObj = item as JsonObject
      val ftxNumber = jsonObj.getLong("forcedTransactionNumber").toULong()
      val hash = jsonObj.getString("hash")?.decodeHex()
      val error = jsonObj.getString("error")

      ForcedTransactionResponse(
        ftxNumber = ftxNumber,
        ftxHash = hash,
        ftxError = error,
      )
    }
  }

  /**
   * Parses the response from linea_getForcedTransactionInclusionStatus RPC call.
   *
   * Expected response format (found):
   * ```json
   * {
   *   "forcedTransactionNumber": 6,
   *   "blockNumber": "0xeff35f",
   *   "blockTimestamp": 1234567890,
   *   "from": "0x6221a9c005f6e47eb398fd867784cacfdcfff4e7",
   *   "inclusionResult": "Included",
   *   "transactionHash": "0xabc123..."
   * }
   * ```
   *
   * Returns null if result is null (transaction not found).
   */
  fun parseForcedTransactionInclusionStatus(result: Any?): ForcedTransactionInclusionStatus? {
    if (result == null) {
      return null
    }

    val jsonObj = result as JsonObject
    val ftxNumber = jsonObj.getLong("forcedTransactionNumber").toULong()
    val blockNumber = jsonObj.getString("blockNumber").toULongFromHex()
    val blockTimestamp = jsonObj.getLong("blockTimestamp")
    val from = jsonObj.getString("from").decodeHex()
    val inclusionResultStr = jsonObj.getString("inclusionResult")
    val transactionHash = jsonObj.getString("transactionHash").decodeHex()

    val inclusionResult = parseInclusionResult(inclusionResultStr)

    return ForcedTransactionInclusionStatus(
      ftxNumber = ftxNumber,
      blockNumber = blockNumber,
      blockTimestamp = java.time.Instant.ofEpochSecond(blockTimestamp).toKotlinInstant(),
      inclusionResult = inclusionResult,
      ftxHash = transactionHash,
      from = from,
    )
  }

  private fun parseInclusionResult(inclusionResultStr: String): ForcedTransactionInclusionResult {
    return try {
      ForcedTransactionInclusionResult.valueOf(inclusionResultStr)
    } catch (e: IllegalArgumentException) {
      throw IllegalArgumentException(
        "Unknown inclusion result: '$inclusionResultStr'. " +
          "Valid values: ${ForcedTransactionInclusionResult.entries.joinToString(", ")}",
        e,
      )
    }
  }

  private fun java.time.Instant.toKotlinInstant(): kotlin.time.Instant {
    return kotlin.time.Instant.fromEpochSeconds(this.epochSecond, this.nano.toLong())
  }
}
