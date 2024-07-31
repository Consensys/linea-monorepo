package net.consensys.linea.traces

import com.fasterxml.jackson.databind.ObjectMapper
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

class JsonParserHelper {

  companion object {
    private val log: Logger = LogManager.getLogger(this::class.java)
    val objectMapper: ObjectMapper = ObjectMapper()
    var mapOfValues: Map<String, TracingModuleV1> = mapOf(
      "add" to TracingModuleV1.ADD,
      "bin" to TracingModuleV1.BIN,
      "binRT" to TracingModuleV1.BIN_RT,
      "ext" to TracingModuleV1.EXT,
      "ec_data" to TracingModuleV1.EC_DATA,
      "hash_data" to TracingModuleV1.PUB_HASH,
      "hash_info" to TracingModuleV1.PUB_HASH_INFO,
      "hub" to TracingModuleV1.HUB,
      "log_data" to TracingModuleV1.PUB_LOG,
      "log_info" to TracingModuleV1.PUB_LOG_INFO,
      "mmio" to TracingModuleV1.MMIO,
      "mmu" to TracingModuleV1.MMU,
      "mmuID" to TracingModuleV1.MMU_ID,
      "mod" to TracingModuleV1.MOD,
      "mul" to TracingModuleV1.MUL,
      "mxp" to TracingModuleV1.MXP,
      "phoneyRLP" to TracingModuleV1.PHONEY_RLP,
      "rlp" to TracingModuleV1.RLP,
      "rom" to TracingModuleV1.ROM,
      "shf" to TracingModuleV1.SHF,
      "shfRT" to TracingModuleV1.SHF_RT,
      "txRlp" to TracingModuleV1.TX_RLP,
      "wcp" to TracingModuleV1.WCP
    )

    fun from(key: String): TracingModuleV1? {
      return mapOfValues.get(key)
    }

    fun findKey(json: String, pos: Int): String? {
      // "Trace" is within a {} block, we need to find the name to which this block is associated: "name":{ ..., "Trace":....}
      var curlyPosition = openingCurlyBracketPosition(pos, json)
      if (curlyPosition == -1) {
        // error
        log.warn("Opening curly bracket not found at $pos")
        return null
      }
      // now that we found the {, we look for what is between the previous pair of ", it's the name we look for.
      return extractName(json, curlyPosition)
    }

    private fun extractName(json: String, curlyPosition: Int): String? {
      val nameEnd = json.lastIndexOf('"', curlyPosition)
      if (nameEnd == -1) {
        log.warn("Name end not found at pos $curlyPosition")
        return null
      }
      val nameStart = json.lastIndexOf('"', nameEnd - 1)
      if (nameStart == -1) {
        // error
        log.warn("Name start not found at pos $nameEnd")
        return null
      }
      return json.substring(nameStart + 1, nameEnd).trim()
    }

    private fun openingCurlyBracketPosition(pos: Int, json: String): Int {
      var countOfCurlyBrackets = 0
      var curlyPosition = -1
      for (i in pos downTo 0) {
        if (json.get(i).equals('{')) {
          countOfCurlyBrackets += 1
          if (countOfCurlyBrackets == 1) {
            curlyPosition = i
            break
          }
        } else if (json.get(i).equals('}')) {
          countOfCurlyBrackets -= 1
        }
      }
      return curlyPosition
    }

    fun simpleMatrix(json: String, start: Int, end: Int): Boolean {
      for (i in start..end) {
        if (json.get(i).equals('[') || json.get(i).equals('{')) {
          return false
        }
      }
      return true
    }

    fun countRow(json: String, pos: Int): Int {
      // Trace represent a matrix by columns lie this: {"Trace":{"ACC_1":["0","0","0"... We need to find the first pair of [] and count how many items it contains.
      val startArray = json.indexOf('[', pos + 1)

      if (startArray < 0) {
        log.warn("Start array not found at $pos")
        return -1
      }
      val endArray = json.indexOf(']', startArray + 1)

      if (endArray < 0) {
        log.warn("End array not found at $pos")
        return -1
      }

      if (!simpleMatrix(json, startArray + 1, endArray)) {
        log.warn(
          "We assume that traces are made of colum without nested objects," +
            " but it was not the case for the trace ar $pos."
        )
        return -1
      }

      val size = json.substring(startArray + 1, endArray).filter { ch -> ch == ',' }.count()
      if (size == 0 && json.substring(startArray + 1, endArray).trim().length == 0) {
        // if it's an empty matrix: {"Trace":{"ACC_1":[]
        return 0
      } else {
        return size + 1
      }
    }

    fun getPrecomputedLimit(trace: String, name: String): Long {
      // add quotes to make it more robust. Key are surrounded by quotes.
      val position = trace.indexOf("\"$name\"")

      if (position == -1) {
        log.warn("Name $name not found.")
        return -1L
      }
      val start = trace.indexOf(':', position + name.length)

      var end = -1
      for (i in (start + 1)..trace.length) {
        if (trace.get(i).equals(',') || trace.get(i).equals('}')) {
          end = i
          break
        }
      }
      if (end == -1) {
        return -1
      }

      return trace.substring(start + 1, end).trim().toLongOrNull() ?: -1
    }

    fun getTracesPosition(trace: String): ArrayList<Int> {
      var pos = 0
      val positions: ArrayList<Int> = ArrayList()
      while (pos >= 0 && pos <= trace.length) {
        pos = trace.indexOf("\"Trace\"", pos + 1)
        if (pos >= 0) {
          positions.add(pos)
        }
      }
      return positions
    }

    fun getPrecompiles(trace: String): Map<String, Any>? {
      val preCompilePos = trace.indexOf("PrecompileCalls")
      if (preCompilePos < 0) {
        log.warn("precompile block not found.")
        return null
      } else {
        val startPosition = preCompilePos - 1 // add -1 to get the "

        val endPosition = getPositionOfMatchingClosingBracket(startPosition, trace)
        if (endPosition == -1) {
          log.warn("precompile closing block in json not found at $startPosition")
          return null
        }

        val subString = "{" + trace.substring(startPosition, endPosition + 1) + "}"

        @Suppress("UNCHECKED_CAST")
        val jsonObject = objectMapper.readValue(subString, Map::class.java) as Map<String, Map<String, Any>>
        return jsonObject.get("PrecompileCalls")
      }
    }

    private fun getPositionOfMatchingClosingBracket(startPosition: Int, trace: String): Int {
      var countOfClosingCurlyBrackets = 0
      var curlyPosition = -1
      for (i in startPosition..trace.length - 1) {
        if (trace.get(i).equals('}')) {
          countOfClosingCurlyBrackets += 1
          if (countOfClosingCurlyBrackets == 0) {
            curlyPosition = i
            break
          }
        } else if (trace.get(i).equals('{')) {
          countOfClosingCurlyBrackets -= 1
        }
      }
      return curlyPosition
    }
  }
}
