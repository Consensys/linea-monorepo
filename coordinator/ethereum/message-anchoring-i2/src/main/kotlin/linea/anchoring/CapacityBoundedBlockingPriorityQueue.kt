package linea.anchoring

import java.util.concurrent.PriorityBlockingQueue

/**
 * A bounded priority queue that allows adding elements even when the queue is full.
 */
class CapacityBoundedBlockingPriorityQueue<T : Comparable<T>>(
  private val targetCapacity: UInt,
  private val absoluteMaxCapacity: UInt = targetCapacity * 2u
) : PriorityBlockingQueue<T>(targetCapacity.toInt()) {
  init {
    val validRange = 0..Int.MAX_VALUE
    require(targetCapacity.toInt() in validRange) { "targetCapacity=$targetCapacity must in range $validRange" }
    require(absoluteMaxCapacity.toInt() in validRange) {
      "absoluteMaxCapacity=$absoluteMaxCapacity must in range $validRange"
    }
    require(absoluteMaxCapacity >= targetCapacity) {
      "absoluteMaxCapacity=$absoluteMaxCapacity must be greater than or equal to targetCapacity=$targetCapacity"
    }
  }

  override fun add(element: T): Boolean {
    require(super.size < absoluteMaxCapacity.toInt()) {
      "Queue is full size=$size, cannot add element: $element"
    }
    return super.add(element)
  }

  override fun addAll(elements: Collection<T>): Boolean {
    require((remainingMaxCapacity() - super.size) >= elements.size) {
      "Queue absolute MaxremainingCapacity=${this.remainingMaxCapacity()} is less than elements size=${elements.size}"
    }

    return super.addAll(elements)
  }

  override fun remainingCapacity(): Int {
    return (targetCapacity.toInt() - super.size)
  }

  private fun remainingMaxCapacity(): Int {
    return (absoluteMaxCapacity.toInt() - super.size)
  }
}
