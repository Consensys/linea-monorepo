package build.linea.s11n.jackson

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.module.SimpleModule
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import kotlinx.datetime.Instant
import java.math.BigInteger

val ethNumberAsHexSerialisersModule = SimpleModule().apply {
  this.addSerializer(Instant::class.java, InstantISO8601Serializer)
  this.addDeserializer(Instant::class.java, InstantISO8601Deserializer)
  this.addSerializer(Int::class.java, IntToHexSerializer)
  this.addSerializer(Long::class.java, LongToHexSerializer)
  this.addSerializer(ULong::class.java, ULongToHexSerializer)
  this.addSerializer(BigInteger::class.java, BigIntegerToHexSerializer)
}

val ethByteAsHexSerialisersModule = SimpleModule().apply {
  this.addSerializer(Byte::class.java, ByteToHexSerializer)
  this.addSerializer(UByte::class.java, UByteToHexSerializer)
  this.addSerializer(ByteArray::class.java, ByteArrayToHexSerializer)
}

val defaultObjectMapper: ObjectMapper = jacksonObjectMapper()
  .registerModules(ethByteAsHexSerialisersModule)

val ethApiObjectMapper: ObjectMapper = jacksonObjectMapper()
  .registerModules(
    ethNumberAsHexSerialisersModule,
    ethByteAsHexSerialisersModule
  )
