package net.consensys.linea.jsonrpc.client

import com.fasterxml.jackson.core.JsonGenerator
import com.fasterxml.jackson.databind.JsonNode
import com.fasterxml.jackson.databind.JsonSerializer
import com.fasterxml.jackson.databind.SerializerProvider
import com.fasterxml.jackson.databind.module.SimpleModule
import com.fasterxml.jackson.databind.node.JsonNodeType
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import io.vertx.core.json.jackson.VertxModule
import java.util.HexFormat

internal val objectMapper = jacksonObjectMapper()
  .registerModules(VertxModule())
  .registerModules(
    SimpleModule().apply {
      this.addSerializer(ByteArray::class.java, ByteArrayToHexStringSerializer())
    }
  )

class ByteArrayToHexStringSerializer : JsonSerializer<ByteArray>() {
  private val hexFormatter = HexFormat.of()
  override fun serialize(value: ByteArray?, gen: JsonGenerator?, serializers: SerializerProvider?) {
    gen?.writeString(value?.let { "0x" + hexFormatter.formatHex(it) })
  }
}

fun JsonNodeType.isPrimitive(): Boolean {
  return when (this) {
    JsonNodeType.STRING, JsonNodeType.NUMBER, JsonNodeType.BOOLEAN, JsonNodeType.NULL -> true
    else -> false
  }
}

fun JsonNode.toPrimitiveOrJsonNode(): Any? {
  return if (this.nodeType.isPrimitive()) {
    objectMapper.convertValue(this, Any::class.java)
  } else {
    this
  }
}
