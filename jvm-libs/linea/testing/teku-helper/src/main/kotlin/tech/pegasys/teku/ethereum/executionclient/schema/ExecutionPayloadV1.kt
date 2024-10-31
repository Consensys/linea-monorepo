package tech.pegasys.teku.ethereum.executionclient.schema

import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.ByteArrayExt
import net.consensys.toBigInteger
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.apache.tuweni.units.bigints.UInt256
import tech.pegasys.teku.infrastructure.bytes.Bytes20
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import tech.pegasys.teku.spec.TestSpecFactory
import tech.pegasys.teku.spec.util.DataStructureUtil
import java.math.BigInteger

fun executionPayloadV1(
  blockNumber: Long = 0,
  parentHash: ByteArray = ByteArrayExt.random32(),
  feeRecipient: ByteArray = ByteArrayExt.random(20),
  stateRoot: ByteArray = ByteArrayExt.random32(),
  receiptsRoot: ByteArray = ByteArrayExt.random32(),
  logsBloom: ByteArray = ByteArrayExt.random32(),
  prevRandao: ByteArray = ByteArrayExt.random32(),
  gasLimit: ULong = 0UL,
  gasUsed: ULong = 0UL,
  timestamp: Instant = Clock.System.now(),
  extraData: ByteArray = ByteArrayExt.random32(),
  baseFeePerGas: BigInteger = BigInteger.valueOf(256),
  blockHash: ByteArray = ByteArrayExt.random32(),
  transactions: List<ByteArray> = emptyList()
): ExecutionPayloadV1 {
  return ExecutionPayloadV1(
    Bytes32.wrap(parentHash),
    Bytes20(Bytes.wrap(feeRecipient)),
    Bytes32.wrap(stateRoot),
    Bytes32.wrap(receiptsRoot),
    Bytes.wrap(logsBloom),
    Bytes32.wrap(prevRandao),
    UInt64.valueOf(blockNumber),
    UInt64.valueOf(gasLimit.toBigInteger()),
    UInt64.valueOf(gasUsed.toBigInteger()),
    UInt64.valueOf(timestamp.epochSeconds),
    Bytes.wrap(extraData),
    UInt256.valueOf(baseFeePerGas),
    Bytes32.wrap(blockHash),
    transactions.map { Bytes.wrap(it) }
  )
}

fun executionPayloadV1(
  blockNumber: Long = 0,
  parentHash: Bytes32 = Bytes32.random(),
  feeRecipient: Bytes20 = Bytes20(Bytes.random(20)),
  stateRoot: Bytes32 = Bytes32.random(),
  receiptsRoot: Bytes32 = Bytes32.random(),
  logsBloom: Bytes = Bytes32.random(),
  prevRandao: Bytes32 = Bytes32.random(),
  gasLimit: UInt64 = UInt64.valueOf(0),
  gasUsed: UInt64 = UInt64.valueOf(0),
  timestamp: UInt64 = UInt64.valueOf(0),
  extraData: Bytes = Bytes32.random(),
  baseFeePerGas: UInt256 = UInt256.valueOf(256),
  blockHash: Bytes32 = Bytes32.random(),
  transactions: List<Bytes> = emptyList()
): ExecutionPayloadV1 {
  return ExecutionPayloadV1(
    parentHash,
    feeRecipient,
    stateRoot,
    receiptsRoot,
    logsBloom,
    prevRandao,
    UInt64.valueOf(blockNumber),
    gasLimit,
    gasUsed,
    timestamp,
    extraData,
    baseFeePerGas,
    blockHash,
    transactions
  )
}

fun randomExecutionPayload(
  transactionsRlp: List<Bytes> = emptyList(),
  blockNumber: Long? = null
): ExecutionPayloadV1 {
  val executionPayload = dataStructureUtil.randomExecutionPayload()
  return ExecutionPayloadV1(
    /* parentHash = */ executionPayload.parentHash,
    /* feeRecipient = */
    executionPayload.feeRecipient,
    /* stateRoot = */
    executionPayload.stateRoot,
    /* receiptsRoot = */
    executionPayload.receiptsRoot,
    /* logsBloom = */
    executionPayload.logsBloom,
    /* prevRandao = */
    executionPayload.prevRandao,
    /* blockNumber = */
    blockNumber?.let(UInt64::valueOf) ?: executionPayload.blockNumber.cropToPositiveSignedLong(),
    /* gasLimit = */
    executionPayload.gasLimit.cropToPositiveSignedLong(),
    /* gasUsed = */
    executionPayload.gasUsed.cropToPositiveSignedLong(),
    /* timestamp = */
    executionPayload.timestamp.cropToPositiveSignedLong(),
    /* extraData = */
    executionPayload.extraData,
    /* baseFeePerGas = */
    executionPayload.baseFeePerGas,
    /* blockHash = */
    executionPayload.blockHash,
    /* transactions = */
    transactionsRlp
  )
}

val dataStructureUtil: DataStructureUtil = DataStructureUtil(TestSpecFactory.createMinimalBellatrix())

// Teku UInt64 has a bug allow negative number to be created
// random test payload creates such cases we need to fix it
private fun UInt64.cropToPositiveSignedLong(): UInt64 {
  val longValue = this.longValue()
  return if (longValue < 0) {
    return UInt64.valueOf(-longValue)
  } else {
    this
  }
}
