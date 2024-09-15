package net.consensys.linea.transactionexclusion.dto

import com.fasterxml.jackson.annotation.JsonProperty
import com.fasterxml.jackson.databind.ObjectMapper
import net.consensys.linea.transactionexclusion.ModuleOverflow

data class ModuleOverflowJsonDto(
  // The new property needs to be placed in
  // alphabetical order; otherwise the integration
  // test would fail due to object to json string function
  @JsonProperty("count")
  val count: Long,
  @JsonProperty("limit")
  val limit: Long,
  @JsonProperty("module")
  val module: String
) {
  // Jackson ObjectMapper requires a default constructor
  constructor() : this(0L, 0L, "")

  companion object {
    fun parseListFromJsonString(jsonString: String): List<ModuleOverflowJsonDto> {
      return ObjectMapper().readValue(
        jsonString,
        Array<ModuleOverflowJsonDto>::class.java
      ).toList()
    }

    fun parseToJsonString(target: Any): String {
      if (target is String) {
        return target
      }
      return ObjectMapper().writeValueAsString(target)
    }
  }

  fun toDomainObject(): ModuleOverflow {
    return ModuleOverflow(
      count = count,
      limit = limit,
      module = module
    )
  }
}
