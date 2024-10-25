package net.consensys

class KMath {
  companion object {
    fun addExact(a: UInt, b: UInt): UInt {
      val result = a + b
      if (result < a || result < b) {
        throw ArithmeticException("UInt overflow")
      }
      return result
    }

    fun addExact(a: ULong, b: ULong): ULong {
      val result = a + b
      if (result < a || result < b) {
        throw ArithmeticException("ULong overflow")
      }
      return result
    }
  }
}

fun ULong.plusExact(other: ULong): ULong = KMath.addExact(this, other)
fun UInt.plusExact(other: UInt): UInt = KMath.addExact(this, other)
