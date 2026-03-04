package linea.blob

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertNotNull
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import kotlin.time.Duration.Companion.days
import kotlin.time.Instant

class BlobCompressorSelectorTest {

  @Test
  fun `returns correct BlobCompressor for matching timestamp`() {
    val v1 = BlobCompressorVersion.V1_2
    val v2 = BlobCompressorVersion.V2
    val t1 = Instant.parse("2025-01-01T00:00:00Z")
    val t2 = Instant.parse("2026-01-01T00:00:00Z")
    val dataLimit = 100
    val selector = BlobCompressorSelectorByTimestamp(
      mapOf(v1 to t1, v2 to t2),
      dataLimit,
    )
    // Should select v2 for t2, v1 for t1 and mid-timestamps
    val compressorV2 = selector.getBlobCompressor(t2)
    val compressorV1 = selector.getBlobCompressor(t1)
    val compressorMid = selector.getBlobCompressor(Instant.parse("2025-06-01T00:00:00Z"))
    assertNotNull(compressorV2)
    assertNotNull(compressorV1)
    assertNotNull(compressorMid)
    assertEquals(compressorV1, compressorMid)
    assert(compressorV2 != compressorV1)
  }

  @Test
  fun `returns the same BlobCompressor object`() {
    val v1 = BlobCompressorVersion.V1_2
    val t1 = Instant.parse("2025-01-01T00:00:00Z")
    val dataLimit = 100
    val selector = BlobCompressorSelectorByTimestamp(
      mapOf(v1 to t1),
      dataLimit,
    )
    // Should select v1 for t2, v1 for t1 and mid-timestamps
    val compressorV1 = selector.getBlobCompressor(t1)
    assertNotNull(compressorV1)
    assertEquals(compressorV1, selector.getBlobCompressor(t1))
    assertEquals(compressorV1, selector.getBlobCompressor(t1.plus(1.days)))
  }

  @Test
  fun `throws if no version matches timestamp`() {
    val v1 = BlobCompressorVersion.V1_2
    val t1 = Instant.parse("2025-01-01T00:00:00Z")
    val dataLimit = 100
    val selector = BlobCompressorSelectorByTimestamp(mapOf(v1 to t1), dataLimit)
    val before = Instant.parse("2024-01-01T00:00:00Z")
    assertThrows<IllegalStateException> {
      selector.getBlobCompressor(before)
    }
  }
}
