import build.linea.staterecover.core.BlockL1RecoveredData
import build.linea.staterecover.core.TransactionL1RecoveredData
import kotlinx.datetime.Instant
import net.consensys.tuweni.bytes.toULong
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.Wei
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.core.Transaction
import kotlin.jvm.optionals.getOrDefault

fun Block.toDomain(): BlockL1RecoveredData {
  return BlockL1RecoveredData(
    blockNumber = this.header.number.toULong(),
    blockHash = this.header.hash.toArray(),
    coinbase = this.header.coinbase.toArray(),
    blockTimestamp = Instant.fromEpochMilliseconds(this.header.timestamp),
    gasLimit = this.header.gasLimit.toULong(),
    difficulty = this.header.difficulty.toULong(),
    transactions = this.body.transactions.map { it.toDomain() }
  )
}

fun Transaction.toDomain(): TransactionL1RecoveredData {
  return TransactionL1RecoveredData(
    type = this.type.serializedType.toUByte(),
    nonce = this.nonce.toULong(),
    maxPriorityFeePerGas = this.maxPriorityFeePerGas.getOrDefault(Wei.ZERO).toULong(),
    maxFeePerGas = this.maxFeePerGas.getOrDefault(Wei.ZERO).toULong(),
    gasLimit = this.gasLimit.toULong(),
    from = this.signature.s.toByteArray(),
    to = this.to.getOrDefault(Address.ZERO).toArray(),
    value = this.value.asBigInteger,
    data = this.data.getOrDefault(Bytes.EMPTY).toArray(),
    accessList = this.accessList.getOrDefault(emptyList()).map { accessListEntry ->
      TransactionL1RecoveredData.AccessTuple(
        address = accessListEntry.address.toArray(),
        storageKeys = accessListEntry.storageKeys.map { storageKey -> storageKey.toArray() }
      )
    }
  )
}
