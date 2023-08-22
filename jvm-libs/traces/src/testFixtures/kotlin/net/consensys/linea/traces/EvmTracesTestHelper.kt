package net.consensys.linea.traces

import kotlin.random.Random
import kotlin.random.nextUInt

fun fakeTracesCounters(defaultValue: UInt?): TracesCounters {
  return TracingModule.values().fold(mutableMapOf()) { counters,
    evmTracingModule ->
    counters[evmTracingModule] = defaultValue ?: Random.nextUInt(0u, UInt.MAX_VALUE)
    counters
  }
}
