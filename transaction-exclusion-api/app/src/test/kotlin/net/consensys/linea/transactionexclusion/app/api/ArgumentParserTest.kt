package net.consensys.linea.transactionexclusion.app.api

import io.vertx.core.json.JsonObject
import kotlinx.datetime.Instant
import net.consensys.decodeHex
import net.consensys.encodeHex
import net.consensys.linea.ModuleOverflow
import net.consensys.linea.RejectedTransaction
import net.consensys.linea.TransactionInfo
import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import kotlin.random.Random

class ArgumentParserTest {
  @Test
  fun getTransactionRLPInRawBytes_should_return_correct_byte_array() {
    val transactionRLPInHexStr =
      "0x02f8388204d2648203e88203e88203e8941195cf65f83b3a5768f3c4" +
        "96d3a05ad6412c64b38203e88c666d93e9cc5f73748162cea9c0017b8201c8"
    Assertions.assertTrue(
      ArgumentParser.getTransactionRLPInRawBytes(transactionRLPInHexStr)
        .contentEquals(transactionRLPInHexStr.decodeHex())
    )
  }

  @Test
  fun getTransactionRLPInRawBytes_should_throw_error_for_invalid_hex_string() {
    // odd number of hex character
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getTransactionRLPInRawBytes(
        "0x02f8388204d2648203e88203e88203e8941195cf65f83b3a5768f3c4" +
          "96d3a05ad6412c64b38203e88c666d93e9cc5f73748162cea9c0017b820"
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("RLP-encoded transaction cannot be parsed")
      )
    }

    // invalid hex character
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getTransactionRLPInRawBytes(
        "yyf8388204d2648203e88203e88203e8941195cf65f83b3a5768f3c4" +
          "96d3a05ad6412c64b38203e88c666d93e9cc5f73748162cea9c0017b8201xx"
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("RLP-encoded transaction cannot be parsed")
      )
    }
  }

  @Test
  fun getTxHashInRawBytes_should_return_correct_byte_array() {
    val txHashInHexStr = "0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350ad7"
    Assertions.assertTrue(
      ArgumentParser.getTxHashInRawBytes(txHashInHexStr)
        .contentEquals(txHashInHexStr.decodeHex())
    )
  }

  @Test
  fun getTxHashInRawBytes_should_throw_error_for_invalid_hex_string() {
    // hex string of less than 64 hex characters
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getTxHashInRawBytes(
        "0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350a"
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Hex string of transaction hash cannot be parsed")
      )
    }
  }

  @Test
  fun getTransactionInfoFromRLP_should_return_correct_transactionInfo() {
    val transactionRLP =
      (
        "0x02f8388204d2648203e88203e88203e8941195cf65f83b3a5768f3c4" +
          "96d3a05ad6412c64b38203e88c666d93e9cc5f73748162cea9c0017b8201c8"
        ).decodeHex()
    Assertions.assertEquals(
      ArgumentParser.getTransactionInfoFromRLP(transactionRLP),
      TransactionInfo(
        hash = "0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350ad7".decodeHex(),
        from = "0x4d144d7b9c96b26361d6ac74dd1d8267edca4fc2".decodeHex(),
        to = "0x1195cf65f83b3a5768f3c496d3a05ad6412c64b3".decodeHex(),
        nonce = 100UL
      )
    )
  }

  @Test
  fun getTransactionInfoFromRLP_should_throw_error_for_invalid_transactionRLP() {
    // hex string of less than 64 hex characters
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getTransactionInfoFromRLP(
        (
          "0xaaf8388204d2648203e88203e88203e8941195cf65f83b3a5768f3c4" +
            "96d3a05ad6412c64b38203e88c666d93e9cc5f73748162cea9c0017b8201c8"
          ).decodeHex()
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("RLP-encoded transaction cannot be parsed")
      )
    }
  }

  @Test
  fun getOverflows_should_return_correct_list_of_ModuleOverflow() {
    val expectedModuleOverflowList = listOf(
      ModuleOverflow(
        module = "ADD",
        count = 402,
        limit = 70
      ),
      ModuleOverflow(
        module = "MUL",
        count = 587,
        limit = 400
      )
    )

    // valid json string
    val moduleOverflowJsonStr =
      "[{\"module\":\"ADD\",\"count\":402,\"limit\":70},{\"module\":\"MUL\",\"count\":587,\"limit\":400}]"
    Assertions.assertEquals(
      ArgumentParser.getOverflows(moduleOverflowJsonStr),
      expectedModuleOverflowList
    )

    // valid list of ModuleOverflow
    Assertions.assertEquals(
      ArgumentParser.getOverflows(expectedModuleOverflowList),
      expectedModuleOverflowList
    )
  }

  @Test
  fun getOverflows_should_throw_error_for_invalid_moduleOverflowJsonStr() {
    // invalid json string (invalid field name xxx)
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getOverflows(
        "[{\"module\":\"ADD\",\"count\":402,\"xxx\":70},{\"module\":\"MUL\",\"count\":587,\"limit\":400}]"
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Overflows cannot be parsed")
      )
    }

    // invalid json string (invalid module value)
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getOverflows(
        "[{\"module\":123,\"count\":402,\"xxx\":70},{\"module\":\"MUL\",\"count\":587,\"limit\":400}]"
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Overflows cannot be parsed")
      )
    }

    // invalid json string (invalid input)
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getOverflows(JsonObject())
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Overflows cannot be parsed")
      )
    }
  }

  @Test
  fun getReasonMessage_should_return_correct_reason_message() {
    val reasonMessage = "Transaction line count for module ADD=402 is above the limit 70"
    Assertions.assertEquals(
      ArgumentParser.getReasonMessage(reasonMessage),
      reasonMessage
    )

    val reasonMessageWithMaxLen = Random.Default.nextBytes(128).encodeHex(prefix = false)
    Assertions.assertEquals(
      ArgumentParser.getReasonMessage(reasonMessageWithMaxLen),
      reasonMessageWithMaxLen
    )
  }

  @Test
  fun getReasonMessage_should_throw_error_for_string_length_longer_than_256() {
    // reason message string with more than 256 characters
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getReasonMessage(
        Random.Default.nextBytes(128).encodeHex(prefix = false) + "0"
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Reason message should not be more than 256 characters")
      )
    }
  }

  @Test
  fun getBlockNumber_should_return_correct_unsigned_long_or_null() {
    // 10-base number
    val blockNumberStr = "12345"
    Assertions.assertEquals(
      ArgumentParser.getBlockNumber(blockNumberStr)!!,
      blockNumberStr.toULong()
    )

    Assertions.assertNull(
      ArgumentParser.getBlockNumber(null)
    )
  }

  @Test
  fun getBlockNumber_should_throw_error_for_invalid_blockNumberStr() {
    // block number string with hex string
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getBlockNumber(
        "0x12345"
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Block number cannot be parsed to an unsigned number")
      )
    }

    // block number string with random characters
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getBlockNumber(
        "xxyyzz"
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Block number cannot be parsed to an unsigned number")
      )
    }

    // empty block number string
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getBlockNumber("")
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Block number cannot be parsed to an unsigned number")
      )
    }
  }

  @Test
  fun getTimestampFromISO8601_should_return_correct_instant() {
    // timestamp in ISO-8601
    val timestampStr = "2024-09-05T09:22:52Z"
    Assertions.assertEquals(
      ArgumentParser.getTimestampFromISO8601(timestampStr),
      Instant.parse(timestampStr)
    )
  }

  @Test
  fun getTimestampFromISO8601_should_throw_error_for_invalid_timestampStr() {
    // timestamp string not in ISO-8601
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getTimestampFromISO8601(
        "2024-09-05_09:22:52"
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Timestamp is not in ISO-8601")
      )
    }

    // timestamp string in epoch time millisecond
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getTimestampFromISO8601(
        "1725543970103"
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Timestamp is not in ISO-8601")
      )
    }

    // empty timestamp string
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getTimestampFromISO8601("")
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Timestamp is not in ISO-8601")
      )
    }
  }

  @Test
  fun getTxRejectionStage_should_return_correct_rejection_stage() {
    val txRejectionStageStr = "SEQUENCER"
    Assertions.assertEquals(
      ArgumentParser.getTxRejectionStage(txRejectionStageStr),
      RejectedTransaction.Stage.SEQUENCER
    )
  }

  @Test
  fun getTxRejectionStage_should_throw_error_for_invalid_txRejectionStageStr() {
    // rejection stage string in lower case
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getTxRejectionStage(
        "sequencer"
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Unsupported transaction rejection stage")
      )
    }

    // rejection stage string in random characters
    assertThrows<IllegalArgumentException> {
      ArgumentParser.getTxRejectionStage(
        "helloworld"
      )
    }.also { error ->
      Assertions.assertTrue(
        error.message!!.contains("Unsupported transaction rejection stage")
      )
    }
  }
}
