package linea.anchoring

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows

class CapacityBoundedBlockingPriorityQueueTest {

  @Test
  fun `should add elements within target capacity`() {
    val queue = CapacityBoundedBlockingPriorityQueue<Int>(targetCapacity = 5u)

    queue.add(1)
    queue.add(2)

    assertThat(queue).containsExactly(1, 2)
    assertThat(queue.remainingCapacity()).isEqualTo(3)
  }

  @Test
  fun `should throw exception when adding element beyond absolute max capacity`() {
    val queue = CapacityBoundedBlockingPriorityQueue<Int>(targetCapacity = 2u, absoluteMaxCapacity = 3u)

    queue.add(1)
    queue.add(2)
    queue.add(3)

    val exception = assertThrows<IllegalArgumentException> {
      queue.add(4)
    }

    assertThat(exception).hasMessageContaining("Queue is full")
  }

  @Test
  fun `should add all elements within remaining max capacity`() {
    val queue = CapacityBoundedBlockingPriorityQueue<Int>(targetCapacity = 3u, absoluteMaxCapacity = 5u)

    queue.addAll(listOf(1, 2))

    assertThrows<IllegalArgumentException> {
      queue.addAll(listOf(3, 4, 5, 6, 7))
    }

    queue.addAll(listOf(3, 4, 5))

    assertThat(queue).containsExactly(1, 2, 3, 4, 5)
    assertThat(queue.remainingCapacity()).isEqualTo(0)
  }

  @Test
  fun `should throw exception when adding collection beyond remaining max capacity`() {
    val queue = CapacityBoundedBlockingPriorityQueue<Int>(targetCapacity = 3u, absoluteMaxCapacity = 4u)

    queue.addAll(listOf(1, 2, 3))

    val exception = assertThrows<IllegalArgumentException> {
      queue.addAll(listOf(4, 5, 6))
    }

    assertThat(exception).hasMessageContaining("Queue absolute MaxRemainingCapacity")
  }

  @Test
  fun `should maintain priority order when adding elements`() {
    val queue = CapacityBoundedBlockingPriorityQueue<Int>(targetCapacity = 5u)

    queue.add(5)
    queue.add(1)
    queue.add(3)

    val orderedElements = mutableListOf<Int>()
    while (queue.peek() != null) {
      orderedElements.add(queue.poll())
    }
    assertThat(orderedElements).containsExactly(1, 3, 5)
  }

  @Test
  fun `should calculate remaining capacity correctly`() {
    val queue = CapacityBoundedBlockingPriorityQueue<Int>(targetCapacity = 5u)

    queue.add(1)
    queue.add(2)

    assertThat(queue.remainingCapacity()).isEqualTo(3)
  }

  @Test
  fun `should throw exception when target capacity is invalid`() {
    assertThrows<IllegalArgumentException> {
      CapacityBoundedBlockingPriorityQueue<Int>(targetCapacity = 0u)
    }
  }

  @Test
  fun `should throw exception when absolute max capacity is less than target capacity`() {
    val exception = assertThrows<IllegalArgumentException> {
      CapacityBoundedBlockingPriorityQueue<Int>(targetCapacity = 5u, absoluteMaxCapacity = 4u)
    }

    assertThat(exception).hasMessageContaining("absoluteMaxCapacity")
  }
}
