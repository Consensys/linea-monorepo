package net.consensys.linea.transactionexclusion

import kotlinx.datetime.Instant
import net.consensys.decodeHex

internal val defaultRejectedTransaction = RejectedTransaction(
  txRejectionStage = RejectedTransaction.Stage.SEQUENCER,
  timestamp = Instant.parse("2024-08-31T09:18:51Z"),
  blockNumber = 10000UL,
  transactionRLP =
  (
    "0x02f8388204d2648203e88203e88203e8941195cf65f83b3a5768f3c4" +
      "96d3a05ad6412c64b38203e88c666d93e9cc5f73748162cea9c0017b8201c8"
    )
    .decodeHex(),
  reasonMessage = "Transaction line count for module ADD=402 is above the limit 70",
  overflows = listOf(
    ModuleOverflow(
      module = "ADD",
      count = 402,
      limit = 70
    ),
    ModuleOverflow(
      module = "MUL",
      count = 587,
      limit = 401
    ),
    ModuleOverflow(
      module = "EXP",
      count = 9000,
      limit = 8192
    )
  ),
  transactionInfo = TransactionInfo(
    hash = "0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350ad7".decodeHex(),
    from = "0x4d144d7b9c96b26361d6ac74dd1d8267edca4fc2".decodeHex(),
    to = "0x1195cf65f83b3a5768f3c496d3a05ad6412c64b3".decodeHex(),
    nonce = 100UL
  )
)
