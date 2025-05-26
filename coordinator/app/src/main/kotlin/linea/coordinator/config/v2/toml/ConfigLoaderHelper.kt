package linea.coordinator.config.v2.toml

import com.sksamuel.hoplite.ConfigLoaderBuilder
import com.sksamuel.hoplite.ExperimentalHoplite
import com.sksamuel.hoplite.toml.TomlPropertySource
import net.consensys.zkevm.coordinator.app.config.BlockParameterDecoder

fun ConfigLoaderBuilder.addCoordinatorTomlDecoders(): ConfigLoaderBuilder {
  return this
    .addDecoder(BlockParameterDecoder())
    .addDecoder(TomlByteArrayHexDecoder())
    .addDecoder(TomlKotlinDurationDecoder())
    .addDecoder(TomlSignerTypeDecoder())
}

@OptIn(ExperimentalHoplite::class)
inline fun <reified T : Any> parseConfig(toml: String): T {
  return ConfigLoaderBuilder
    .default()
    .withExplicitSealedTypes()
    .addCoordinatorTomlDecoders()
    .addSource(TomlPropertySource(toml))
    .build()
    .loadConfigOrThrow<T>()
}
