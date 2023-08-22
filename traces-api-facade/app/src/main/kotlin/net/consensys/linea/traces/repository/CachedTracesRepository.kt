package net.consensys.linea.traces.repository

import com.github.benmanes.caffeine.cache.AsyncCacheLoader
import com.github.benmanes.caffeine.cache.AsyncLoadingCache
import com.github.benmanes.caffeine.cache.Caffeine
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.onFailure
import com.github.michaelbull.result.onSuccess
import net.consensys.linea.BlockTraces
import net.consensys.linea.TracesError
import net.consensys.linea.TracesRepository
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.time.Duration
import java.util.concurrent.Executor

data class ReadTracesCacheConfig(val size: UInt, val expirationDuration: Duration)

class CachedTracesRepository(
  private val delegate: TracesRepository,
  config: ReadTracesCacheConfig
) : TracesRepository {
  private val log: Logger = LogManager.getLogger(CachedTracesRepository::class.java)
  private val loader: AsyncCacheLoader<UInt, BlockTraces> = AsyncCacheLoader { key, ex: Executor ->
    val resultFuture = SafeFuture<BlockTraces>()
    ex.execute {
      delegate
        .getTraces(key)
        .thenAccept { result ->
          result
            .onSuccess {
              log.trace("cache loaded {} -> Ok", key)
              resultFuture.complete(it)
            }
            .onFailure { tracesError ->
              log.debug("cache load failure {} -> {}", key, tracesError)
              // Error shall not be cached. e.g traces file not yet generated
              resultFuture.completeExceptionally(
                Exception("${tracesError.errorType}, ${tracesError.errorDetail}")
              )
            }
        }
        .handleException(resultFuture::completeExceptionally)
    }

    resultFuture
  }
  private val cache: AsyncLoadingCache<UInt, BlockTraces> =
    Caffeine.newBuilder()
      .maximumSize(config.size.toLong())
      .expireAfterWrite(config.expirationDuration)
      .buildAsync(loader)

  override fun getTraces(blockNumber: UInt): SafeFuture<Result<BlockTraces, TracesError>> {
    return SafeFuture.of(cache.get(blockNumber)).thenApply { Ok(it) }
  }
}
