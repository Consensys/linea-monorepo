package net.consensys

import java.util.HexFormat

fun String.decodeHex(): ByteArray {
  check(length % 2 == 0) { "Must have an even length" }
  return HexFormat.of().parseHex(removePrefix("0x"))
}

fun String.containsAny(strings: List<String>, ignoreCase: Boolean): Boolean {
  return strings.any { this.contains(it, ignoreCase) }
}
