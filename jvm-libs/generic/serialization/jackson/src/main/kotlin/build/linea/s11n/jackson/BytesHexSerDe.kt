package build.linea.s11n.jackson

import com.fasterxml.jackson.core.JsonGenerator
import com.fasterxml.jackson.databind.JsonSerializer
import com.fasterxml.jackson.databind.SerializerProvider
import java.util.HexFormat

private val hexFormatter = HexFormat.of()

object ByteArrayToHexSerializer : JsonSerializer<ByteArray>() {
  override fun serialize(value: ByteArray?, gen: JsonGenerator?, serializers: SerializerProvider?) {
    gen?.writeString(value?.let { "0x" + hexFormatter.formatHex(it) })
  }
}

object ByteToHexSerializer : JsonSerializer<Byte>() {
  @OptIn(ExperimentalStdlibApi::class)
  override fun serialize(value: Byte?, gen: JsonGenerator?, serializers: SerializerProvider?) {
    gen?.writeString(value?.let { "0x" + it.toUByte().toHexString() })
  }
}

object UByteToHexSerializer : JsonSerializer<UByte>() {
  @OptIn(ExperimentalStdlibApi::class)
  override fun serialize(value: UByte?, gen: JsonGenerator?, serializers: SerializerProvider?) {
    gen?.writeString(value?.let { "0x" + it.toHexString() })
  }
}
