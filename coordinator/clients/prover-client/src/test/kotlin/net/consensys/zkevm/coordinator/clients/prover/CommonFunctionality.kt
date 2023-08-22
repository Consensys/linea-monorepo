package net.consensys.zkevm.coordinator.clients.prover

import net.consensys.zkevm.toULong
import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import tech.pegasys.teku.spec.TestSpecFactory
import tech.pegasys.teku.spec.util.DataStructureUtil
import kotlin.math.abs

class SimpleFileNameProvider() : ProverFilesNameProvider {
  override fun getRequestFileName(startBlockNumber: ULong, endBlockNumber: ULong): String {
    return "$startBlockNumber-$endBlockNumber-getZkProof.json"
  }

  override fun getResponseFileName(startBlockNumber: ULong, endBlockNumber: ULong): String {
    return "$startBlockNumber-$endBlockNumber-proof.json"
  }
}

fun randomExecutionPayloads(numberOfBlocks: Int): List<ExecutionPayloadV1> {
  return (1..numberOfBlocks)
    .map { index ->
      randomExecutionPayloadWithProperTx(CommonTestData.validTransactionRlp, index.toLong())
    }
    .toMutableList()
    .apply { this.sortBy { it.blockNumber } }
}

fun randomExecutionPayloadWithProperTx(
  transactionRlp: String,
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
    blockNumber?.let(UInt64::valueOf) ?: UInt64.valueOf(abs(executionPayload.blockNumber.longValue())),
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
    listOf(Bytes.fromHexString(transactionRlp))
  )
}

private val dataStructureUtil: DataStructureUtil = DataStructureUtil(TestSpecFactory.createMinimalBellatrix())

private fun UInt64.cropToPositiveSignedLong(): UInt64 {
  val longValue = this.toULong().toLong()
  return if (longValue < 0) {
    return UInt64.valueOf(-this.toULong().toLong())
  } else {
    UInt64.valueOf(this.toULong().toLong())
  }
}
