package net.consensys.linea

import java.math.BigDecimal
import java.math.BigInteger

fun ULong.toHexString(): String = "0x${this.toString(16)}"

/**
 * Parses an hexadecimal string as [ULong] number and returns the result.
 * @throws NumberFormatException if the string is not a valid hexadecimal representation of a number.
 */
fun ULong.Companion.fromHexString(value: String): ULong = value.replace("0x", "").toULong(16)

fun BigInteger.toULong(): ULong = this.toString().toULong()
fun ULong.toBigInteger(): BigInteger = BigInteger(this.toString())

fun <T : Comparable<T>> ClosedRange<T>.toIntervalString(): String {
  val size = if (start <= endInclusive) {
    this.endInclusive.toString().toBigDecimal() - this.start.toString().toBigDecimal() + 1.toBigDecimal()
  } else {
    this.start.toString().toBigDecimal() - this.endInclusive.toString().toBigDecimal() + 1.toBigDecimal()
  }
  return "[${this.start}..${this.endInclusive}]($size)"
}

private val OneGWei = BigDecimal.valueOf(1_000_000_000)
fun BigInteger.toGWei(): BigDecimal = this.toBigDecimal().divide(OneGWei)
