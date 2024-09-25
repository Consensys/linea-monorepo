package net.consensys

fun ByteArray.assertSize(expectedSize: UInt, fieldName: String = ""): ByteArray = apply {
  require(size == expectedSize.toInt()) { "$fieldName expected to have $expectedSize bytes, but got $size" }
}

fun ByteArray.assertIs32Bytes(fieldName: String = ""): ByteArray = assertSize(32u, fieldName)

fun ByteArray.assertIs20Bytes(fieldName: String = ""): ByteArray = assertSize(20u, fieldName)

fun ByteArray.setFirstByteToZero(): ByteArray {
  this[0] = 0
  return this
}

/**
 * Slices the ByteArray into sliceSize bytes chunks and returns the sliceNumber-th chunk.
 */
fun ByteArray.sliceOf(
  sliceSize: Int,
  sliceNumber: Int,
  allowIncompleteLastSlice: Boolean = false
): ByteArray {
  assert(sliceSize > 0) {
    "sliceSize=$sliceSize should be greater than 0"
  }

  val startIndex = sliceNumber * sliceSize
  val endIndex = (sliceNumber * sliceSize + sliceSize - 1)
    .let {
      if (it >= this.size && allowIncompleteLastSlice) {
        this.size - 1
      } else {
        it
      }
    }

  assert(startIndex <= this.size && endIndex <= this.size) {
    "slice $startIndex..$endIndex is out of array size=${this.size}"
  }

  return this.sliceArray(startIndex..endIndex)
}

/**
 * Slices the ByteArray into 32 bytes chunks and returns the sliceNumber-th chunk.
 */
fun ByteArray.sliceOf32(sliceNumber: Int): ByteArray {
  return this.sliceOf(sliceSize = 32, sliceNumber)
}
