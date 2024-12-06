package linea.encoding

import linea.encoding.BlockMapper.toBesuBlock
import net.consensys.encodeHex
import net.consensys.toBigInteger
import net.consensys.zkevm.encoding.BlockEncoder
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.crypto.SECP256K1
import org.hyperledger.besu.datatypes.AccessListEntry
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.Hash
import org.hyperledger.besu.datatypes.TransactionType
import org.hyperledger.besu.datatypes.Wei
import org.hyperledger.besu.ethereum.core.BlockBody
import org.hyperledger.besu.ethereum.core.BlockHeaderBuilder
import org.hyperledger.besu.ethereum.core.Difficulty
import org.hyperledger.besu.ethereum.core.Transaction
import org.hyperledger.besu.evm.log.LogsBloomFilter

object BlockRLPEncoder : BlockEncoder {
  override fun encode(block: linea.domain.Block): ByteArray {
    return block.toBesuBlock().toRlp().toArray()
  }
}

object BlockMapper {
  val secp256k1 = SECP256K1()

  fun linea.domain.Block.toBesuBlock(): org.hyperledger.besu.ethereum.core.Block {
    val header = BlockHeaderBuilder.create()
      .parentHash(Hash.wrap(Bytes32.wrap(this.parentHash)))
      .ommersHash(Hash.wrap(Bytes32.wrap(this.ommersHash)))
      .coinbase(Address.wrap(Bytes.wrap(this.miner)))
      .stateRoot(Hash.wrap(Bytes32.wrap(this.stateRoot)))
      .transactionsRoot(Hash.wrap(Bytes32.wrap(this.transactionsRoot)))
      .receiptsRoot(Hash.wrap(Bytes32.wrap(this.receiptsRoot)))
      .logsBloom(LogsBloomFilter.fromHexString(this.logsBloom.encodeHex()))
      .difficulty(Difficulty.fromHexOrDecimalString(this.difficulty.toString()))
      .number(this.number.toLong())
      .gasLimit(this.gasLimit.toLong())
      .gasUsed(this.gasUsed.toLong())
      .timestamp(this.timestamp.toLong())
      .extraData(Bytes.wrap(this.extraData))
      .mixHash(Hash.wrap(Bytes32.wrap(this.mixHash)))
      .nonce(this.nonce.toLong())
      .baseFee(this.baseFeePerGas?.toWei())
      .buildBlockHeader()

    val transactions = this.transactions.map { tx ->
      Transaction.builder()
        .type(tx.type.toBesuType())
        .nonce(tx.nonce.toLong())
        .gasPrice(tx.gasPrice.toWei())
        .gasLimit(tx.gasLimit.toLong())
        .to(tx.to?.let { Address.wrap(Bytes.wrap(it)) })
        .value(tx.value.toWei())
        .payload(Bytes.wrap(tx.input))
        .chainId(tx.chainId?.toBigInteger())
        .maxPriorityFeePerGas(tx.maxPriorityFeePerGas?.toWei())
        .maxFeePerGas(tx.maxFeePerGas?.toWei())
        .accessList(
          tx.accessList.map { entry ->
            AccessListEntry(
              Address.wrap(Bytes.wrap(entry.address)),
              entry.storageKeys.map { Bytes32.wrap(it) }
            )
          }
        )
        .signature(
          secp256k1.createSignature(
            tx.r.encodeHex().toBigInteger(16),
            tx.s.encodeHex().toBigInteger(16),
            tx.v.toByte()
          )
        )
        .build()
    }
    // linea does not support uncles, so we are not converting them
    // throwing an exception just in case we get one and we can fix it
    if (this.ommers.isNotEmpty()) {
      throw IllegalStateException("Uncles are not supported: block=${this.number}")
    }

    val body = BlockBody(transactions, emptyList())

    return org.hyperledger.besu.ethereum.core.Block(header, body)
  }
}

fun ULong.toWei(): Wei = Wei.of(this.toBigInteger())

fun linea.domain.TransactionType.toBesuType(): TransactionType {
  return when (this) {
    linea.domain.TransactionType.FRONTIER -> TransactionType.FRONTIER
    linea.domain.TransactionType.EIP1559 -> TransactionType.EIP1559
    linea.domain.TransactionType.ACCESS_LIST -> TransactionType.ACCESS_LIST
    linea.domain.TransactionType.BLOB -> TransactionType.BLOB
    linea.domain.TransactionType.DELEGATE_CODE -> TransactionType.DELEGATE_CODE
  }
}
