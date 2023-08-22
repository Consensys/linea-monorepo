package net.consensys.linea

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.unwrap
import com.github.michaelbull.result.unwrapError
import io.vertx.core.json.JsonObject
import tech.pegasys.teku.infrastructure.async.SafeFuture

class BlockTraces(
  val blockNumber: ULong,
  // val blockHash: Bytes32,
  val traces: JsonObject
)

interface TracesRepository {
  fun getTraces(blockNumber: UInt): SafeFuture<Result<BlockTraces, TracesError>>

  fun getTraces(
    startBlockNumber: UInt,
    endBlockNumber: UInt
  ): SafeFuture<Result<List<BlockTraces>, TracesError>> {
    val futures: List<SafeFuture<Result<BlockTraces, TracesError>>> =
      (startBlockNumber..endBlockNumber).map { getTraces(it) }

    return SafeFuture.collectAll(futures.stream()).thenApply { results: List<Result<BlockTraces, TracesError>> ->
      val error = (results.find { it is Err })?.unwrapError()
      if (error != null) {
        Err(error)
      } else {
        Ok(results.map { it.unwrap() })
      }
    }
  }
}

data class TracesConflation(
  val startBlockNumber: ULong,
  val endBlockNumber: ULong,
  val traces: VersionedResult<JsonObject>
)

interface TracesRepositoryV1 {
  fun getTraces(block: BlockNumberAndHash): SafeFuture<Result<BlockTraces, TracesError>>
  fun getTraces(blocks: List<BlockNumberAndHash>): SafeFuture<Result<List<BlockTraces>, TracesError>>
}

interface ConflatedTracesRepository {
  fun findConflatedTraces(startBlockNumber: ULong, endBlockNumber: ULong, tracesVersion: String): SafeFuture<String?>
  fun saveConflatedTraces(conflation: TracesConflation): SafeFuture<String>
}
