package net.consensys.linea

import com.github.michaelbull.result.Result
import io.vertx.core.json.JsonObject

fun interface TracesCounter {
  fun countTraces(traces: JsonObject): Result<VersionedResult<BlockCounters>, TracesError>
}

fun interface TracesConflator {
  fun conflateTraces(traces: List<JsonObject>): Result<VersionedResult<JsonObject>, TracesError>
}
