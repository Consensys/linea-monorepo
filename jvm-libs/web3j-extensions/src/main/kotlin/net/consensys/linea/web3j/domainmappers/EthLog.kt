package net.consensys.linea.web3j.domainmappers

import net.consensys.decodeHex
import net.consensys.eth.EthLog
import net.consensys.toULong
import org.web3j.protocol.core.methods.response.Log

fun Log.toDomain(): EthLog {
  return EthLog(
    removed = this.isRemoved,
    logIndex = this.logIndex.toULong(),
    transactionIndex = this.transactionIndex.toULong(),
    transactionHash = this.transactionHash.decodeHex(),
    blockHash = this.blockHash.decodeHex(),
    blockNumber = this.blockNumber.toULong(),
    address = this.address.decodeHex(),
    data = this.data.decodeHex(),
    topics = this.topics.map(String::decodeHex)
  )
}
