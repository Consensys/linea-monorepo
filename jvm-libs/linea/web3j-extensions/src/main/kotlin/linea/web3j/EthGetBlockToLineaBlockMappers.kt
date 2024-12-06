package linea.web3j

import linea.domain.Block
import linea.domain.Transaction
import linea.domain.TransactionType
import net.consensys.decodeHex
import net.consensys.toIntFromHex
import net.consensys.toULong
import net.consensys.toULongFromHex
import org.web3j.protocol.core.methods.response.EthBlock

fun EthBlock.Block.toDomain(): Block = mapToDomain(this)

fun mapToDomain(web3jBlock: EthBlock.Block): Block {
  val block = Block(
    number = web3jBlock.number.toULong(),
    hash = web3jBlock.hash.decodeHex(),
    parentHash = web3jBlock.parentHash.decodeHex(),
    ommersHash = web3jBlock.sha3Uncles.decodeHex(),
    miner = web3jBlock.miner.decodeHex(),
    nonce = web3jBlock.nonce.toULong(),
    stateRoot = web3jBlock.stateRoot.decodeHex(),
    transactionsRoot = web3jBlock.transactionsRoot.decodeHex(),
    receiptsRoot = web3jBlock.receiptsRoot.decodeHex(),
    logsBloom = web3jBlock.logsBloom.decodeHex(),
    difficulty = web3jBlock.difficulty.toULong(),
    gasLimit = web3jBlock.gasLimit.toULong(),
    gasUsed = web3jBlock.gasUsed.toULong(),
    timestamp = web3jBlock.timestamp.toULong(),
    extraData = web3jBlock.extraData.decodeHex(),
    mixHash = web3jBlock.mixHash.decodeHex(),
    baseFeePerGas = web3jBlock.baseFeePerGas?.toULong(), // Optional field for EIP-1559 blocks
    ommers = web3jBlock.uncles.map { it.decodeHex() }, // List of uncle block hashes
    transactions = run {
      if (web3jBlock.transactions.isNotEmpty() && web3jBlock.transactions[0] !is EthBlock.TransactionObject) {
        throw IllegalArgumentException(
          "Expected to be have full EthBlock.TransactionObject." +
            "Got just transaction hashes."
        )
      }
      web3jBlock.transactions.map { (it as EthBlock.TransactionObject).toDomain() }
    }
  )
  return block
}

fun EthBlock.TransactionObject.toDomain(): Transaction {
  val maxFeePerGas = this.maxFeePerGas?.toULong()
  // Web3J throws an exception if maxPriorityFeePerGas null, instead of a check like in maxFeePerGas
  // we need to check if maxFeePerGas is null to avoid the exception
  val maxPriorityFeePerGas = if (maxFeePerGas != null) this.maxPriorityFeePerGas?.toULong() else null
  return Transaction(
    nonce = this.nonce.toULong(),
    gasPrice = this.gasPrice.toULong(),
    gasLimit = this.gas.toULong(),
    to = this.to?.decodeHex(),
    value = this.value,
    input = this.input.decodeHex(),
    r = this.r.removePrefix("0x").toBigInteger(16),
    s = this.s.removePrefix("0x").toBigInteger(16),
    v = this.v.toULong(),
    yParity = this.getyParity()?.toULongFromHex(),
    type = mapType(this.type), // Optional field for EIP-2718 typed transactions
    chainId = this.chainId?.toULong(), // Optional field for EIP-155 transactions
    maxFeePerGas = maxFeePerGas, // Optional field for EIP-1559 transactions
    maxPriorityFeePerGas = maxPriorityFeePerGas // Optional field for EIP-1559 transactions
  )
}

fun mapType(type: String?): TransactionType {
  return type
    ?.let { TransactionType.fromEthApiSerializedValue(it.toIntFromHex()) }
    ?: TransactionType.FRONTIER
}
