package lineth.coordination.riscv.rollup

import linea.domain.Constants

/**
 * Canonical 254-bit packing of a blob payload into a physical EIP-4844 blob (rollup_spec `pack_blob_payload` /
 * `unpack_blob_payload`).
 *
 * The payload followed by a `0xff` terminal is read as one big-endian bit string, left-aligned and zero-padded to
 * 4096 * 254 bits, and split into 4096 consecutive 254-bit values. Each value is written as a 32-byte big-endian
 * field element, so its top 2 bits are always zero.
 */
object BlobPayloadPacking {
  private const val FIELD_ELEMENTS_PER_BLOB = 4096
  private const val BYTES_PER_FIELD_ELEMENT = 32
  private const val BITS_PER_FIELD_ELEMENT = BYTES_PER_FIELD_ELEMENT * 8
  private const val PACKING_BITS_PER_ELEMENT = 254
  private const val PADDING_BITS_PER_ELEMENT = BITS_PER_FIELD_ELEMENT - PACKING_BITS_PER_ELEMENT
  private const val TERMINAL: Byte = 0xff.toByte()

  /** Maximum payload bytes per blob: 4096 * 254 / 8 = 130048, minus one byte for the terminal. */
  const val BLOB_PAYLOAD_CAPACITY = FIELD_ELEMENTS_PER_BLOB * PACKING_BITS_PER_ELEMENT / 8 - 1

  fun pack(payload: ByteArray): ByteArray {
    require(payload.size <= BLOB_PAYLOAD_CAPACITY) {
      "blob payload size=${payload.size} exceeds packing capacity=$BLOB_PAYLOAD_CAPACITY"
    }
    val blob = ByteArray(Constants.Eip4844BlobSize)
    val bitsToWrite = (payload.size + 1) * 8
    for (srcBit in 0 until bitsToWrite) {
      val srcByte = if (srcBit / 8 < payload.size) payload[srcBit / 8] else TERMINAL
      if ((srcByte.toInt() shr (7 - srcBit % 8)) and 1 == 0) continue
      val dstBit = (srcBit / PACKING_BITS_PER_ELEMENT) * BITS_PER_FIELD_ELEMENT +
        PADDING_BITS_PER_ELEMENT + srcBit % PACKING_BITS_PER_ELEMENT
      blob[dstBit / 8] = (blob[dstBit / 8].toInt() or (1 shl (7 - dstBit % 8))).toByte()
    }
    return blob
  }

  fun unpack(blob: ByteArray): ByteArray {
    require(blob.size == Constants.Eip4844BlobSize) {
      "physical blob must be ${Constants.Eip4844BlobSize} bytes, got ${blob.size}"
    }
    val unpacked = ByteArray(BLOB_PAYLOAD_CAPACITY + 1)
    for (element in 0 until FIELD_ELEMENTS_PER_BLOB) {
      require(blob[element * BYTES_PER_FIELD_ELEMENT].toInt() and 0xc0 == 0) {
        "blob field element $element exceeds $PACKING_BITS_PER_ELEMENT bits"
      }
    }
    for (dstBit in 0 until unpacked.size * 8) {
      val srcBit = (dstBit / PACKING_BITS_PER_ELEMENT) * BITS_PER_FIELD_ELEMENT +
        PADDING_BITS_PER_ELEMENT + dstBit % PACKING_BITS_PER_ELEMENT
      if ((blob[srcBit / 8].toInt() shr (7 - srcBit % 8)) and 1 == 0) continue
      unpacked[dstBit / 8] = (unpacked[dstBit / 8].toInt() or (1 shl (7 - dstBit % 8))).toByte()
    }
    val terminalIndex = unpacked.indexOfLast { it != 0.toByte() }
    require(terminalIndex >= 0 && unpacked[terminalIndex] == TERMINAL) { "invalid blob terminal symbol" }
    val payload = unpacked.copyOf(terminalIndex)
    require(pack(payload).contentEquals(blob)) { "noncanonical blob padding" }
    return payload
  }
}
