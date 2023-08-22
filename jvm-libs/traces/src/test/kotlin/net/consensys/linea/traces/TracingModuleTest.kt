package net.consensys.linea.traces

import com.sksamuel.hoplite.ConfigLoaderBuilder
import com.sksamuel.hoplite.addFileSource
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import java.nio.file.Path

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
}
