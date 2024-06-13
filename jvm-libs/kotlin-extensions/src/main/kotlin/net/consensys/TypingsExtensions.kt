package net.consensys

import kotlinx.datetime.Instant
import java.math.BigDecimal
import java.math.BigInteger

// BigInteger extensions
private val OneGWei = BigDecimal.valueOf(1_000_000_000)
fun BigInteger.toGWei(): BigDecimal = this.toBigDecimal().divide(OneGWei)
fun BigInteger.toULong(): ULong = this.toString().toULong()
// BigInteger extensions

// ULong extensions
fun ULong.toBigInteger(): BigInteger = BigInteger(this.toString())
fun ULong.toHexString(): String = "0x${this.toString(16)}"

/**
 * Parses an hexadecimal string as [ULong] number and returns the result.
 * @throws NumberFormatException if the string is not a valid hexadecimal representation of a number.
 */
fun ULong.Companion.fromHexString(value: String): ULong = value.replace("0x", "").toULong(16)

// ULong extensions
fun ByteArray.encodeHex(prefix: Boolean = true): String =
  "${if (prefix) "0x" else ""}${joinToString(separator = "") { eachByte -> "%02x".format(eachByte) }}"

fun String.decodeHex(): ByteArray {
  check(length % 2 == 0) { "Must have an even length" }
  return removePrefix("0x").chunked(2)
    .map { it.toInt(16).toByte() }
    .toByteArray()
}

fun <T : Comparable<T>> ClosedRange<T>.toIntervalString(): String {
  val size = if (start <= endInclusive) {
    this.endInclusive.toString().toBigDecimal() - this.start.toString().toBigDecimal() + 1.toBigDecimal()
  } else {
    this.start.toString().toBigDecimal() - this.endInclusive.toString().toBigDecimal() + 1.toBigDecimal()
  }
  return "[${this.start}..${this.endInclusive}]${size.toInt()}"
}

fun ByteArray.assertSize(expectedSize: UInt, fieldName: String = ""): ByteArray = apply {
  require(size == expectedSize.toInt()) { "$fieldName expected to have $expectedSize bytes, but got $size" }
}

fun ByteArray.assertIs32Bytes(fieldName: String = ""): ByteArray = assertSize(32u, fieldName)

fun ByteArray.assertIs20Bytes(fieldName: String = ""): ByteArray = assertSize(20u, fieldName)

fun ByteArray.setFirstByteToZero(): ByteArray {
  this[0] = 0
  return this
}

fun Instant.trimToSecondPrecision(): Instant {
  return Instant.fromEpochSeconds(this.epochSeconds)
}

fun Instant.trimToMillisecondPrecision(): Instant {
  return Instant.fromEpochMilliseconds(this.toEpochMilliseconds())
}
