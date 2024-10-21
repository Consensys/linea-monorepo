package tech.pegasys.teku.ethereum.executionclient.schema

import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.apache.tuweni.units.bigints.UInt256
import tech.pegasys.teku.infrastructure.bytes.Bytes20
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import tech.pegasys.teku.spec.TestSpecFactory
import tech.pegasys.teku.spec.util.DataStructureUtil

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

private val dataStructureUtil: DataStructureUtil = DataStructureUtil(TestSpecFactory.createMinimalBellatrix())

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
