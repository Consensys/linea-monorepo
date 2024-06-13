package net.consensys.linea.nativecompressor

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.Test

class GoNativeBlobCompressorFactoryTest {

  @Test
  fun `getInstance should create a GoNativeBlobCompressor singleton`() {
    val compressor = GoNativeBlobCompressorFactory.getInstance()
    assertThat(compressor).isNotNull

    // is singleton
    assertThat(GoNativeBlobCompressorFactory.getInstance()).isSameAs(compressor)
    compressor.Init(100, GoNativeBlobCompressorFactory.getDictionaryPath())
  }

  @Test
  @Disabled("We cannot have more than single instance of GoNativeBlobCompressor in the same JVM")
  fun `newInstance should create a GoNativeBlobCompressor with a dictionary`() {
    val compressor = GoNativeBlobCompressorFactory.newInstance()
    compressor.Init(100, GoNativeBlobCompressorFactory.getDictionaryPath())

    assertThat(compressor).isNotNull
    assertThat(compressor).isNotSameAs(GoNativeBlobCompressorFactory.newInstance())
  }
}
