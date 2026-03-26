package net.consensys.zkevm.ethereum.coordination.aggregation

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class HardForkAggregationTargetEndBlocksTest {
  @Test
  fun `invoke returns configured target end blocks when nothing registered`() {
    val store = HardForkAggregationTargetEndBlocks(configuredTargetEndBlocks = listOf(10uL, 20uL))
    assertThat(store()).containsExactly(10uL, 20uL)
  }

  @Test
  fun `register deduplicates and invoke returns ascending order`() {
    val store = HardForkAggregationTargetEndBlocks(configuredTargetEndBlocks = listOf(10uL, 20uL))
    store.registerAggregationEndBlockInclusive(15uL)
    store.registerAggregationEndBlockInclusive(15uL)
    store.registerAggregationEndBlockInclusive(5uL)
    assertThat(store()).containsExactly(5uL, 10uL, 15uL, 20uL)
  }
}
