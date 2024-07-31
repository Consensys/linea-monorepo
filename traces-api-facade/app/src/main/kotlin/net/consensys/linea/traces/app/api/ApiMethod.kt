package net.consensys.linea.traces.app.api

enum class ApiMethod(val method: String) {
  ROLLUP_GET_TRACES_COUNTERS_BY_BLOCK_NUMBER_V0("rollup_getTracesCountersByBlockNumberV0"),
  ROLLUP_GENERATE_CONFLATED_TRACES_TO_FILE_V0("rollup_generateConflatedTracesToFileV0"),
  ROLLUP_GET_CONFLATED_TRACES_V0("rollup_getConflatedTracesV0"),

  ROLLUP_GET_BLOCK_TRACES_COUNTERS_V1("rollup_getBlockTracesCountersV1"),
  ROLLUP_GENERATE_CONFLATED_TRACES_TO_FILE_V1("rollup_generateConflatedTracesToFileV1"),
  ROLLUP_GET_CONFLATED_TRACES_V1("rollup_getConflatedTracesV1")
}
