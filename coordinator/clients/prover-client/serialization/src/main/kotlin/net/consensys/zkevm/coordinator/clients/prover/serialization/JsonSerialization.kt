package net.consensys.zkevm.coordinator.clients.prover.serialization

import com.fasterxml.jackson.core.JsonGenerator
import com.fasterxml.jackson.databind.DeserializationFeature
import com.fasterxml.jackson.databind.JsonSerializer
import com.fasterxml.jackson.databind.MapperFeature
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.SerializerProvider
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule
import com.fasterxml.jackson.module.kotlin.jacksonMapperBuilder
import org.apache.tuweni.bytes.Bytes

object JsonSerialization {
  val proofResponseMapperV1: ObjectMapper =
    jacksonMapperBuilder()
      .enable(MapperFeature.ACCEPT_CASE_INSENSITIVE_ENUMS)
      .configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false)
      .addModule(JavaTimeModule())
      .build()
}

class TuweniBytesSerializer : JsonSerializer<Bytes>() {
  override fun serialize(value: Bytes, gen: JsonGenerator, provider: SerializerProvider) {
    gen.writeString(value.toHexString().lowercase())
  }
}
