package linea.web3j

import linea.domain.Block
import linea.domain.Transaction
import linea.domain.TransactionType
import net.consensys.decodeHex
import net.consensys.toIntFromHex
import net.consensys.toULong
import net.consensys.toULongFromHex
import org.web3j.protocol.core.methods.response.EthBlock

fun EthBlock.Block.toDomain(): Block {
  return Block(
    number = this.number.toULong(),
    hash = this.hash.decodeHex(),
    parentHash = this.parentHash.decodeHex(),
    ommersHash = this.sha3Uncles.decodeHex(),
    miner = this.miner.decodeHex(),
    nonce = this.nonce.toULong(),
    stateRoot = this.stateRoot.decodeHex(),
    transactionsRoot = this.transactionsRoot.decodeHex(),
    receiptsRoot = this.receiptsRoot.decodeHex(),
    logsBloom = this.logsBloom.decodeHex(),
    difficulty = this.difficulty.toULong(),
    gasLimit = this.gasLimit.toULong(),
    gasUsed = this.gasUsed.toULong(),
    timestamp = this.timestamp.toULong(),
    extraData = this.extraData.decodeHex(),
    mixHash = this.mixHash.decodeHex(),
    baseFeePerGas = this.baseFeePerGas?.toULong(), // Optional field for EIP-1559 blocks
    ommers = this.uncles.map { it.decodeHex() }, // List of uncle block hashes
    transactions = run {
      if (this.transactions.isNotEmpty() && this.transactions[0] !is EthBlock.TransactionObject) {
        throw IllegalArgumentException(
          "Expected to be have full EthBlock.TransactionObject." +
            "Got just transaction hashes."
        )
      }
      this.transactions.map { (it as EthBlock.TransactionObject).toDomain() }
    }
  )
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
