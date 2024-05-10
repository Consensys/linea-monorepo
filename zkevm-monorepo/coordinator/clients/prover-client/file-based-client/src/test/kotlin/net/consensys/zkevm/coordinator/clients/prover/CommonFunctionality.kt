package net.consensys.zkevm.coordinator.clients.prover

import net.consensys.zkevm.domain.BridgeLogsData
import okhttp3.internal.toHexString
import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.ethereum.executionclient.schema.randomExecutionPayload
import kotlin.random.Random

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
      randomExecutionPayload(listOf(Bytes.fromHexString(CommonTestData.validTransactionRlp)), index.toLong())
    }
    .toMutableList()
    .apply { this.sortBy { it.blockNumber } }
}

fun randomBridgeLogsDataList(numberOfBlocks: Int): List<List<BridgeLogsData>> {
  return (1..numberOfBlocks)
    .map { index ->
      listOf(
        BridgeLogsData(
          removed = false,
          logIndex = "0x0",
          transactionIndex = "0x0",
          transactionHash = "0x" + Random.nextBytes(32).joinToString("") {
            java.lang.String.format("%02x", it)
          },
          blockHash = "0x" + Random.nextBytes(32).joinToString("") {
            java.lang.String.format("%02x", it)
          },
          blockNumber = "0x" + index.toHexString(),
          address = "0x" + Random.nextBytes(20).joinToString("") {
            java.lang.String.format("%02x", it)
          },
          data = "0x" + Random.nextBytes(128).joinToString("") {
            java.lang.String.format("%02x", it)
          },
          topics = listOf(
            "0x" + Random.nextBytes(32).joinToString("") {
              java.lang.String.format("%02x", it)
            }
          )
        )
      )
    }
    .toMutableList()
}
