package build.linea.s11n.jackson

import com.fasterxml.jackson.databind.ObjectMapper
import net.consensys.decodeHex
import net.javacrumbs.jsonunit.assertj.JsonAssertions.assertThatJson
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class BytesSerDeTest {
  private lateinit var objectMapper: ObjectMapper

  @BeforeEach
  fun setUp() {
    objectMapper = defaultObjectMapper
  }

  @Test
  fun bytesSerialization() {
    data class SomeObject(
      // ByteArray
      val nullBytes: ByteArray? = null,
      val emptyBytes: ByteArray = byteArrayOf(),
      val someBytes: ByteArray = "0x01aaff04".decodeHex(),

      // UByte
      val nullUByte: UByte? = null,
      val someUByte: UByte = 0xaf.toUByte(),
      val minUByte: UByte = UByte.MIN_VALUE,
      val maxUByte: UByte = UByte.MAX_VALUE,

      // Byte
      val nullByte: Byte? = null,
      val someByte: Byte = 0xaf.toByte(),
      val minByte: Byte = Byte.MIN_VALUE,
      val maxByte: Byte = Byte.MAX_VALUE
    )

    val json = objectMapper.writeValueAsString(SomeObject())
    assertThatJson(json).isEqualTo(
      """
      {
          "nullBytes": null,
          "emptyBytes": "0x",
          "someBytes": "0x01aaff04",
          "nullUByte": null,
          "someUByte": "0xaf",
          "minUByte": "0x00",
          "maxUByte": "0xff",
          "nullByte": null,
          "someByte": "0xaf",
          "minByte": "0x80",
          "maxByte": "0x7f"
      }
      """.trimIndent()
    )
  }
}
