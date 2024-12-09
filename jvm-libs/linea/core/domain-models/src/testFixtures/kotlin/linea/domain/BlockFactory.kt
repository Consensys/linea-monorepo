package linea.domain

import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.ByteArrayExt

val zeroHash = ByteArray(32) { 0 }

fun createBlock(
  number: ULong = 0UL,
  hash: ByteArray = ByteArrayExt.random32(),
  gasLimit: ULong = 60_000_000UL,
  gasUsed: ULong = 30_000_000UL,
  difficulty: ULong = 2UL,
  parentHash: ByteArray = ByteArrayExt.random32(),
  stateRoot: ByteArray = ByteArrayExt.random32(),
  receiptsRoot: ByteArray = ByteArrayExt.random32(),
  logsBloom: ByteArray = ByteArrayExt.random32(),
  ommersHash: ByteArray = ByteArrayExt.random32(),
  timestamp: Instant = Clock.System.now(),
  extraData: ByteArray = ByteArrayExt.random32(),
  baseFeePerGas: ULong = 7UL,
  transactionsRoot: ByteArray = ByteArrayExt.random32(),
  transactions: List<Transaction> = emptyList()
): Block {
  return Block(
    number = number,
    hash = hash,
    parentHash = parentHash,
    ommersHash = ommersHash,
    miner = zeroHash,
    stateRoot = stateRoot,
    transactionsRoot = transactionsRoot,
    receiptsRoot = receiptsRoot,
    logsBloom = logsBloom,
    difficulty = difficulty,
    gasLimit = gasLimit,
    gasUsed = gasUsed,
    timestamp = timestamp.epochSeconds.toULong(),
    extraData = extraData,
    mixHash = zeroHash,
    nonce = 0UL,
    baseFeePerGas = baseFeePerGas,
    transactions = transactions,
    ommers = emptyList()
  )
}
