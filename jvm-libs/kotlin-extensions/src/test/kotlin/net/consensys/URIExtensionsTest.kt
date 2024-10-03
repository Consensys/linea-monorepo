package net.consensys

import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test
import java.net.URI

class URIExtensionsTest {
  @Test
  fun `getPortWithSchemaDefaults`() {
    assertThat(URI.create("http://example.com").getPortWithSchemaDefaults()).isEqualTo(80)
    assertThat(URI.create("https://example.com").getPortWithSchemaDefaults()).isEqualTo(443)
    assertThat(URI.create("http://example.com:8080").getPortWithSchemaDefaults()).isEqualTo(8080)
    assertThat(URI.create("https://example.com:8080").getPortWithSchemaDefaults()).isEqualTo(8080)
    assertThat(URI.create("myschema://example.com:8080").getPortWithSchemaDefaults()).isEqualTo(8080)
    assertThatThrownBy { (URI.create("mySchema://example.com").getPortWithSchemaDefaults()) }
      .isInstanceOf(IllegalArgumentException::class.java)
      .hasMessage("Unsupported scheme: mySchema")
  }
}
