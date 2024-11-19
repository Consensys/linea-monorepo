package net.consensys.linea.traces

import com.sksamuel.hoplite.ConfigLoaderBuilder
import com.sksamuel.hoplite.addFileSource
import net.consensys.linea.testing.filesystem.findPathTo
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import java.lang.IllegalArgumentException

class TracesCountersTest {
  data class TracesConfigV1(val tracesLimits: Map<TracingModuleV1, UInt>)
  data class TracesConfigV2(val tracesLimits: Map<TracingModuleV2, UInt>)

  @Test
  fun `configs v1 match specifiedModules`() {
    val path = findPathTo("config/common/traces-limits-v1.toml")

    val tracesConfigV1 = ConfigLoaderBuilder.default()
      .addFileSource(path.toString())
      .build()
      .loadConfigOrThrow<TracesConfigV1>()

    val tracesCountersLimit = TracesCountersV1(tracesConfigV1.tracesLimits)

    tracesCountersLimit.entries().forEach { moduleLimit ->
      // PoW 2 requirement only apply to traces passed to the prover
      if (TracingModuleV1.evmModules.contains(moduleLimit.first)) {
        val isPowerOf2 = (moduleLimit.second and (moduleLimit.second - 1u)) == 0u
        assertThat(isPowerOf2)
          .withFailMessage("Trace limit ${moduleLimit.first}=${moduleLimit.second} is not a power of 2!")
          .isTrue()
      }
    }
  }

  @Test
  fun `configs v2 match specifiedModules`() {
    val path = findPathTo("config/common/traces-limits-v2.toml")

    val tracesConfig = ConfigLoaderBuilder.default()
      .addFileSource(path.toString())
      .build()
      .loadConfigOrThrow<TracesConfigV2>()

    val tracesCountersLimit = TracesCountersV2(tracesConfig.tracesLimits)

    tracesCountersLimit.entries().forEach { moduleLimit ->
      // PoW 2 requirement only apply to traces passed to the prover
      if (TracingModuleV2.evmModules.contains(moduleLimit.first)) {
        val isPowerOf2 = (moduleLimit.second and (moduleLimit.second - 1u)) == 0u
        assertThat(isPowerOf2)
          .withFailMessage("Trace limit ${moduleLimit.first}=${moduleLimit.second} is not a power of 2!")
          .isTrue()
      }
    }
  }

  @Test
  fun add_notOverflow() {
    val counters1 = fakeTracesCountersV1(10u)
    val counters2 = fakeTracesCountersV1(20u)
    val counters3 = fakeTracesCountersV1(20u)
    assertThat(counters1.add(counters2).add(counters3))
      .isEqualTo(fakeTracesCountersV1(50u))
  }

  @Test
  fun add_Overflow_throwsError() {
    val counters1 = fakeTracesCountersV1(10u)
    val counters2 = fakeTracesCountersV1(UInt.MAX_VALUE)
    assertThatThrownBy { counters1.add(counters2) }.isInstanceOf(ArithmeticException::class.java)
      .withFailMessage("integer overflow")
  }

  @Test
  fun add_multipleVersion_throwsError() {
    val counters1 = fakeTracesCountersV1(10u)
    val counters2 = fakeTracesCountersV2(10u)
    assertThatThrownBy { counters1.add(counters2) }.isInstanceOf(IllegalArgumentException::class.java)
      .hasMessageContaining("Cannot add different traces counters")
  }

  @Test
  fun allTracesWithinLimits() {
    val limits = fakeTracesCountersV1(20u, mapOf(Pair(TracingModuleV1.ADD, 10u)))
    val countersWithinLimits = fakeTracesCountersV1(3u)
    val countersOvertLimits = fakeTracesCountersV1(5u, mapOf(Pair(TracingModuleV1.ADD, 11u)))

    assertThat(countersWithinLimits.allTracesWithinLimits(limits)).isTrue()
    assertThat(countersOvertLimits.allTracesWithinLimits(limits)).isFalse()
  }

  @Test
  fun empty_counters() {
    val tracesCountersV1 = TracesCountersV1(
      TracingModuleV1.entries.associateWith { 0u }
    )
    assertThat(tracesCountersV1).isEqualTo(TracesCountersV1.EMPTY_TRACES_COUNT)

    val tracesCountersV2 = TracesCountersV2(
      TracingModuleV2.entries.associateWith { 0u }
    )
    assertThat(tracesCountersV2).isEqualTo(TracesCountersV2.EMPTY_TRACES_COUNT)
  }

  @Test
  fun incomplete_counters_throwsError() {
    assertThrows<IllegalArgumentException> {
      TracesCountersV1(emptyMap())
    }
    assertThrows<IllegalArgumentException> {
      TracesCountersV1(mapOf(Pair(TracingModuleV1.ADD, 10u)))
    }
    assertThrows<IllegalArgumentException> {
      TracesCountersV2(emptyMap())
    }
    assertThrows<IllegalArgumentException> {
      TracesCountersV2(mapOf(Pair(TracingModuleV2.ADD, 10u)))
    }
  }

  @Test
  fun oversizedTraces() {
    val limits = fakeTracesCountersV1(20u, mapOf(Pair(TracingModuleV1.ADD, 10u)))
    val countersWithinLimits = fakeTracesCountersV1(3u)
    val countersOvertLimits = fakeTracesCountersV1(5u, mapOf(Pair(TracingModuleV1.ADD, 11u)))

    assertThat(countersWithinLimits.oversizedTraces(limits)).isEmpty()
    val oversizedTraces = countersOvertLimits.oversizedTraces(limits)
    assertThat(oversizedTraces).hasSize(1)
    assertThat(oversizedTraces.first()).isEqualTo(Triple(TracingModuleV1.ADD, 11u, 10u))
  }
}
