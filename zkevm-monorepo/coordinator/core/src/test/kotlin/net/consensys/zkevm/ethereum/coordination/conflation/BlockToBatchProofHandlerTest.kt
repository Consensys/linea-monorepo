package net.consensys.zkevm.ethereum.coordination.conflation

import io.vertx.core.json.JsonObject
import net.consensys.encodeHex
import net.consensys.linea.traces.TracingModule
import net.consensys.zkevm.coordinator.clients.GetTracesCountersResponse
import net.consensys.zkevm.ethereum.coordination.conflation.BlockToBatchSubmissionCoordinator.Companion.parseTracesCountersResponseToJson
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import kotlin.random.Random
import kotlin.random.nextUInt

class BlockToBatchProofHandlerTest {
  private val tracesCountersValid: Map<String, Long> =
    TracingModule.values()
      .fold(mutableMapOf()) { acc: MutableMap<String, Long>,
        evmModule: TracingModule ->
        acc[evmModule.name] = Random.nextUInt(0u, UInt.MAX_VALUE).toLong()
        acc
      }
      .also {
        // add edge case of max UInt
        it[TracingModule.EXT.name] = UInt.MAX_VALUE.toLong()
      }
  private val tracesCountersResponseUnsorted = GetTracesCountersResponse(
    10000U,
    tracesCountersValid
      .mapKeys { TracingModule.valueOf(it.key) }
      .mapValues { it.value.toUInt() },
    "0.1.0"
  )
  private val tracesCountersResponseSortedReversedlyInEnumOrder = GetTracesCountersResponse(
    10000U,
    tracesCountersValid
      .mapKeys { TracingModule.valueOf(it.key) }
      .mapValues { it.value.toUInt() }
      .toSortedMap(reverseOrder()),
    "0.1.0"
  )

  @Test
  fun parseTracesCountersResponseToJson_with_unsorted_tc_should_return_sorted_tc_json() {
    val blockNumber = 1000L
    val blockHash = Random.Default.nextBytes(32).encodeHex()
    val returnedJson = parseTracesCountersResponseToJson(
      blockNumber,
      blockHash,
      tracesCountersResponseUnsorted
    )

    val expectedJson = JsonObject.of(
      "tracesEngineVersion",
      tracesCountersResponseUnsorted.tracesEngineVersion,
      "blockNumber",
      blockNumber,
      "blockHash",
      blockHash,
      "blockL1Size",
      tracesCountersResponseUnsorted.blockL1Size.toLong(),
      "tracesCounters",
      tracesCountersResponseUnsorted.tracesCounters
        .map { it.key.name to it.value.toLong() }
        .sortedBy { it.first }
        .toMap()
    )
    assertThat(expectedJson.toString()).isEqualTo(returnedJson.toString())
  }

  @Test
  fun parseTracesCountersResponseToJson_with_reversely_sorted_tc_should_return_sorted_tc_json() {
    val blockNumber = 1000L
    val blockHash = Random.Default.nextBytes(32).encodeHex()
    val returnedJson = parseTracesCountersResponseToJson(
      blockNumber,
      blockHash,
      tracesCountersResponseSortedReversedlyInEnumOrder
    )

    val expectedJson = JsonObject.of(
      "tracesEngineVersion",
      tracesCountersResponseSortedReversedlyInEnumOrder.tracesEngineVersion,
      "blockNumber",
      blockNumber,
      "blockHash",
      blockHash,
      "blockL1Size",
      tracesCountersResponseSortedReversedlyInEnumOrder.blockL1Size.toLong(),
      "tracesCounters",
      tracesCountersResponseSortedReversedlyInEnumOrder.tracesCounters
        .map { it.key.name to it.value.toLong() }
        .sortedBy { it.first }
        .toMap()
    )
    assertThat(expectedJson.toString()).isEqualTo(returnedJson.toString())
  }
}
