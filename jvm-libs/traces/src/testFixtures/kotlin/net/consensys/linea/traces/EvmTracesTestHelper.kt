package net.consensys.linea.traces

import kotlin.random.Random
import kotlin.random.nextUInt

fun fakeTracesCountersV1(
  defaultValue: UInt?,
  moduleValue: Map<TracingModuleV1, UInt> = emptyMap()
): TracesCountersV1 {
  return TracesCountersV1(
    TracingModuleV1.entries.associateWith {
      moduleValue[it] ?: defaultValue ?: Random.nextUInt(0u, UInt.MAX_VALUE)
    }
  )
}

fun fakeTracesCountersV2(
  defaultValue: UInt?,
  moduleValue: Map<TracingModuleV2, UInt> = emptyMap()
): TracesCountersV2 {
  return TracesCountersV2(
    TracingModuleV2.entries.associateWith {
      moduleValue[it] ?: defaultValue ?: Random.nextUInt(0u, UInt.MAX_VALUE)
    }
  )
}
