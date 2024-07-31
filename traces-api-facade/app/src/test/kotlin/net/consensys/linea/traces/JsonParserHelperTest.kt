package net.consensys.linea.traces

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class JsonParserHelperTest {
  @Test
  fun find_simple_key() {
    // {"add":{"Trace":{"ACC_1"...
    assertEquals("add", JsonParserHelper.findKey("""{"add":{"Trace":{"ACC_1""", 9))
    assertThat(JsonParserHelper.findKey("""{"mmu":{"mmio":{"Trace":{"ACC_1":", 17}}""", 18)).isEqualTo("mmio")
  }

  @Test
  fun find_simple_key_ill_formed() {
    // {"add":{"Trace":{"ACC_1"...
    assertThat(JsonParserHelper.findKey("{\"Trace\":{\"ACC_1\"", 2)).isNull()
    assertThat(JsonParserHelper.findKey("[{\"field\":{\"a\":0}, \"Trace\":{\"ACC_1\":1}}]", 17)).isNull()
  }

  @Test
  fun long_name_is_in_RLP() {
    val json = """
      {"BlockRlp":"+ZIc+QJeoBDQoow7Otewm1p4nUNM2EznGaI6Jxg/2WxGgvNusziKoB3LongNameMTejex116q4W1Z7bM1BrTEkUb",
      "hub": {"mmu":{"mmio":{"Trace":{"ACC_1":["0","0","0","0","0"]}}},
      "LongName": 0} }
    """.trimIndent()

    val value = JsonParserHelper.getPrecomputedLimit(json, "LongName")
    assertEquals(0, value)
  }

  @Test
  fun long_name_is_in_RLP2() {
    val json = """
      {"LongName": 0, "BlockRlp":"+ZIc+QJeoBDQoow7Otewm1p4nUNM2EznGaI6Jxg/2WxGgvNusziKoB3LongNameMTejex116Z7bM1BrTEkUb",
      "hub": {"mmu":{"mmio":{"Trace":{"ACC_1":["0","0","0","0","0"]}}}}}
    """.trimIndent()

    val value = JsonParserHelper.getPrecomputedLimit(json, "LongName")
    assertEquals(0, value)
  }

  @Test
  fun trace_is_in_RLP() {
    val json = """
      {"hub":
      {"mmu":{"mmio":{"Trace":{"ACC_1":["0","0","0","0","0"]}}}},
      "BlockRlp":"+ZIc+QJeoBDQoow7Otewm1p4nUNM2EznGaI6Jxg/2WxGgvNusziKoB3MTejex116q4W1Z7bM1BrTEkUblIp0E/a1x5Ql
      Traceuan0BAAAAAAAAAAA"}
    """.trimIndent()

    val positions = JsonParserHelper.getTracesPosition(json)

    assertEquals(listOf(24), positions)
    assertEquals("mmio", JsonParserHelper.findKey(json, 24))
  }

  @Test
  fun find_nested_key() {
    val json = """
      {"hub":
      {"mmu":{"mmio":{"Trace":{"ACC_1":["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0",
      "0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0",
      "0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0",
      "0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0",
      "0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0",
      "0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0",
      "0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0",
      "0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]}},
      "RamStamp":38,"MicroStamp":74,"Trace":{"ACC_1":["0","0","4","0","0","0","196","0","0","0","0","0","0","0","0","0"
      ,"0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","4","0","0","0","196","0","0","0","192","0","0","0"
      ,"160","0","0","0","128","0","0","0","96","0","0","0","64","0","0","0","32","0","0","0","0","0","0","0","4","0",
      "0","0","4","0","0","0","8","0","0","0","10","0","0","0","0","0","0","0","2","0","0","0","0","0","0","0","0","0",
      "0","0","0","0","0","2","0","0","0","0","0","0","0","0","0","0","4","0","0","0","8","0","0","0","10","0","0","0",
      "4","0","0","0","4","0","0","0","10","0","0","0","12","0","0","0","14","0","0","0","16","0","0","0","18","0","0",
      "0","20","0","0","0","8","0","0","0","22","0","0","0","24","0","0","0","4","0","0","0","10","0","0","0","0","0",
      "0","0","0","0","0","0","0","0","0"]}},"Trace":
    """.trimIndent()
    val positions = JsonParserHelper.getTracesPosition(json)

    assertEquals(listOf(24, 919, 1723), positions)
    assertEquals("mmio", JsonParserHelper.findKey(json, 24))
    assertEquals("mmu", JsonParserHelper.findKey(json, 919))
    assertEquals("hub", JsonParserHelper.findKey(json, 1723))
  }

  @Test
  fun count_ill_formed() {
    assertEquals(-1, JsonParserHelper.countRow("{\"Trace\": {\"ACC_DELTA\": [ ", 7))
  }

  @Test
  fun count() {
    assertEquals(0, JsonParserHelper.countRow("{\"Trace\" :{ \"ACC_DELTA\": []", 7))
    assertEquals(1, JsonParserHelper.countRow("{\"Trace\":{\"ACC_DELTA\":[\"1\"]", 7))
    assertEquals(
      8,
      JsonParserHelper.countRow(
        """
       "mod":{"Trace":{"ACC_1_2":["0" ," 0","0","0","0","0","0","0"],
       "ACC_1_3":["0","0","0","0","0","0","0","0"],
       "ACC_2_2":["0","0","0","0","0", "0","0","0"],
       "ACC_2_3":["0","0","0","0","0","0","0","0"] ,
       " ACC_B_0":["0","0","0","0","0","0","0","1"],
       "ACC_B_1":["0","0","0","0","0","0","0","0"],"AC
        """.trimIndent(),
        13
      )
    )
    assertEquals(-1, JsonParserHelper.countRow("{\"Trace\":{\"ACC_DELTA\":[\"1,[0]\"]", 7))
  }

  @Test
  fun get_long() {
    val result = JsonParserHelper.getPrecomputedLimit(
      """
        "KeccakCount":11,"L2L1logsCount":0,"TxCount":1 ,
        "PrecompileCalls":{"EcRecover":0,"Sha2":0,"RipeMD":0,"Identity":0,"ModExp": 0,
        "EcAdd":0,"EcMul":0,"EcPairing":0,"Blake2f":0}
      """.trimIndent(),
      "TxCount"
    )
    assertEquals(1L, result)
  }

  @Test
  fun get_long_missing() {
    val result = JsonParserHelper.getPrecomputedLimit(
      """
        "KeccakCount":11,"L2L1logsCount":0,"TxCount":1,
        "PrecompileCalls":{"EcRecover":0,"Sha2":0,"RipeMD":0,"Identity":0,"ModExp":0,
        "EcAdd":0,"EcMul":0,"EcPairing":0,"Blake2f":0}
      """.trimIndent(),
      "missing"
    )
    assertEquals(-1L, result)
  }

  @Test
  fun precompiles() {
    val result = JsonParserHelper.getPrecompiles(
      """
      "PrecompileCalls": { "b":1}
      """.trimIndent()
    )
    assertThat(result).isEqualTo(mapOf("b" to 1))
  }

  @Test
  fun precompiles_invalid() {
    val result = JsonParserHelper.getPrecompiles(
      """
      "PrecompileCalls": { "b":1
      """.trimIndent()
    )
    assertThat(result).isNull()
  }

  @Test
  fun precompiles_nested() {
    val result = JsonParserHelper.getPrecompiles(
      """
      "PrecompileCalls": { "b":1, "nested":{"c":"text"}}
      """.trimIndent()
    )
    assertThat(result).isEqualTo(mapOf("b" to 1, "nested" to mapOf("c" to "text")))
  }

  @Test
  fun validateSimpleMatrix() {
    assertThat(JsonParserHelper.simpleMatrix("1,1,1", 0, 4)).isTrue
    assertThat(JsonParserHelper.simpleMatrix(""""ACC_1_3":["0","0","0","0","0","0","0","0"]""", 11, 41)).isTrue
    assertThat(JsonParserHelper.simpleMatrix("1,[0],1", 0, 6)).isFalse()
    assertThat(JsonParserHelper.simpleMatrix("""[1,"field":{"subField":1},1,0,5]""", 1, 32)).isFalse()
  }
}
