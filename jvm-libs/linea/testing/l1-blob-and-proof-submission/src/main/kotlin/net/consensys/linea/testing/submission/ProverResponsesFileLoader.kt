package net.consensys.linea.testing.submission

import kotlinx.datetime.Instant
import net.consensys.linea.testing.filesystem.getPathTo
import net.consensys.zkevm.coordinator.clients.prover.serialization.BlobCompressionProofJsonResponse
import net.consensys.zkevm.coordinator.clients.prover.serialization.ProofToFinalizeJsonResponse
import net.consensys.zkevm.domain.Aggregation
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.Constants.LINEA_BLOCK_INTERVAL
import net.consensys.zkevm.domain.createAggregation
import net.consensys.zkevm.domain.createBlobRecords
import java.io.File

fun proverResponsesFromDir(dir: String): List<File> {
  return getPathTo(dir)
    .toFile()
    .listFiles()
    ?.filter { it.name.endsWith(".json") }
    ?: emptyList()
}

fun <T> loadProverResponses(responsesDir: String, mapper: (String) -> T): List<T> {
  return proverResponsesFromDir(responsesDir)
    .map { mapper.invoke(it.readText()) }
}

fun loadAggregations(aggregationsDir: String): List<Aggregation> {
  return loadProverResponses(aggregationsDir) {
    createAggregation(aggregationProof = ProofToFinalizeJsonResponse.fromJsonString(it).toDomainObject())
  }.sortedBy { it.startBlockNumber }
}

fun loadBlobs(
  blobsDir: String,
  firstBlockStartBlockTime: Instant
): List<BlobRecord> {
  return loadProverResponses(blobsDir) {
    BlobCompressionProofJsonResponse.fromJsonString(it).toDomainObject()
  }
    .let { compressionProofs ->
      createBlobRecords(
        compressionProofs = compressionProofs,
        firstBlockStartBlockTime = firstBlockStartBlockTime
      )
    }
    .sortedBy { it.startBlockNumber }
}

fun loadBlobsAndAggregations(
  blobsResponsesDir: String,
  aggregationsResponsesDir: String
): Pair<List<BlobRecord>, List<Aggregation>> {
  val aggregations = loadAggregations(aggregationsResponsesDir)
  val firstAggregationBlockTime = aggregations.first().let { agg ->
    agg.aggregationProof!!.finalTimestamp
      .minus(LINEA_BLOCK_INTERVAL.times((agg.endBlockNumber - agg.startBlockNumber).toInt()))
  }
  val blobs = loadBlobs(blobsResponsesDir, firstAggregationBlockTime)
  return blobs to aggregations
}
