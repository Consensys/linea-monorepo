package net.consensys.zkevm.coordinator.clients

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class TracesApiRequestPriorityComparatorTest {
  private val reqBuilder = TracesGeneratorJsonRpcClientV2.RequestBuilder("v2")
  private val comparator = TracesGeneratorJsonRpcClientV2.requestPriorityComparator

  @Test
  fun `should compare mixed methods by block number`() {
    val counters1 = reqBuilder.buildGetTracesCountersV2Request(blockNumber = 1u)
    val conflation1 = reqBuilder.buildGenerateConflatedTracesToFileV2Request(startBlockNumber = 1u, endBlockNumber = 2u)

    // same block => equal
    assertThat(comparator.compare(counters1, conflation1)).isEqualTo(0)
    assertThat(comparator.compare(conflation1, counters1)).isEqualTo(0)

    val counters2 = reqBuilder.buildGetTracesCountersV2Request(blockNumber = 2u)
    val conflation2 = reqBuilder.buildGenerateConflatedTracesToFileV2Request(startBlockNumber = 3u, endBlockNumber = 4u)

    // 1 - 3 < 0
    assertThat(comparator.compare(counters1, conflation2)).isLessThan(0)
    // 3 - 2 > 0
    assertThat(comparator.compare(conflation2, counters2)).isGreaterThan(0)
  }

  @Test
  fun `should compare by block number - trace counters method`() {
    val req1 = reqBuilder.buildGetTracesCountersV2Request(blockNumber = 1u)
    val req2 = reqBuilder.buildGetTracesCountersV2Request(blockNumber = 2u)

    assertThat(comparator.compare(req1, req2)).isLessThan(0)
    assertThat(comparator.compare(req2, req1)).isGreaterThan(0)

    assertThat(comparator.compare(req1, reqBuilder.buildGetTracesCountersV2Request(blockNumber = 1u))).isEqualTo(0)
  }

  @Test
  fun `should compare by block number - trace conflation method`() {
    val req1 = reqBuilder.buildGenerateConflatedTracesToFileV2Request(startBlockNumber = 1u, endBlockNumber = 20u)
    val req2 = reqBuilder.buildGenerateConflatedTracesToFileV2Request(startBlockNumber = 2u, endBlockNumber = 20u)

    assertThat(comparator.compare(req1, req2)).isLessThan(0)
    assertThat(comparator.compare(req2, req1)).isGreaterThan(0)

    assertThat(
      comparator.compare(
        req1,
        reqBuilder.buildGenerateConflatedTracesToFileV2Request(startBlockNumber = 1u, endBlockNumber = 10u),
      ),
    ).isEqualTo(0)
  }

  @Test
  fun `should return 0 for unrelated methods`() {
    val unrelated = net.consensys.linea.jsonrpc.JsonRpcRequestListParams(
      "2.0",
      1,
      "unrelated_method",
      emptyList<Any>(),
    )
    val counters = reqBuilder.buildGetTracesCountersV2Request(blockNumber = 1u)

    assertThat(comparator.compare(unrelated, counters)).isEqualTo(0)
    assertThat(comparator.compare(counters, unrelated)).isEqualTo(0)
  }
}
