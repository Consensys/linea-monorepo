package net.consensys.linea.traces

import com.sksamuel.hoplite.ConfigLoaderBuilder
import com.sksamuel.hoplite.addFileSource
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test
import java.nio.file.Path
import javax.swing.UIManager.put

class TracingModuleTest {
  data class TracesConfig(val tracesLimits: TracesCounters)

  @Test
  fun `configs match specifiedModules`() {
    val path = Path.of(System.getProperty("user.dir"))
      .parent
      .parent
      .resolve("config/common/traces-limits-v1.toml")

    val tracesConfig = ConfigLoaderBuilder.default()
      .addFileSource(path.toString())
      .build()
      .loadConfigOrThrow<TracesConfig>()

    assertThat(tracesConfig.tracesLimits.keys).isEqualTo(TracingModule.values().toSet())
    tracesConfig.tracesLimits.forEach { moduleLimit ->
      // PoW 2 requirement only apply to traces passed to the prover
      if (TracingModule.evmModules.contains(moduleLimit.key)) {
        val isPowerOf2 = (moduleLimit.value and (moduleLimit.value - 1u)) == 0u
        assertThat(isPowerOf2)
          .withFailMessage("Trace limit ${moduleLimit.key}=${moduleLimit.value} is not a power of 2!")
          .isTrue()
      }
    }
  }

  @Test
  fun sumTracesCounters_notOverflow() {
    val counters1 = fakeTracesCounters(10u)
    val counters2 = fakeTracesCounters(20u)
    val counters3 = fakeTracesCounters(20u)
    assertThat(sumTracesCounters(counters1, counters2, counters3))
      .isEqualTo(fakeTracesCounters(50u))
  }

  @Test
  fun sumTracesCounters_Overflow_throwsError() {
    val counters1 = fakeTracesCounters(10u)
    val counters2 = fakeTracesCounters(UInt.MAX_VALUE)
    assertThatThrownBy {
      sumTracesCounters(counters1, counters2)
    }.isInstanceOf(ArithmeticException::class.java)
      .withFailMessage("integer overflow")
  }

  @Test
  fun allTracesWithinLimits() {
    val limits = fakeTracesCounters(20u).toMutableMap().apply {
      put(TracingModule.ADD, 10u)
    }
    val countersWithinLimits = fakeTracesCounters(3u)
    val countersOvertLimits = fakeTracesCounters(5u).toMutableMap().apply { put(TracingModule.ADD, 11u) }

    assertThat(allTracesWithinLimits(countersWithinLimits, limits)).isTrue()
    assertThat(allTracesWithinLimits(countersOvertLimits, limits)).isFalse()
  }
}
