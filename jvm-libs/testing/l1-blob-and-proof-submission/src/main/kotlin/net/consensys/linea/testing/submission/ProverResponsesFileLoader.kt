package net.consensys.linea.testing.submission

import net.consensys.linea.testing.filesystem.findPathFileOrDir
import net.consensys.zkevm.coordinator.clients.prover.serialization.BlobCompressionProofJsonResponse
import net.consensys.zkevm.coordinator.clients.prover.serialization.ProofToFinalizeJsonResponse
import net.consensys.zkevm.domain.Aggregation
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.Constants.LINEA_BLOCK_INTERVAL
import net.consensys.zkevm.domain.createAggregation
import net.consensys.zkevm.domain.createBlobRecords
import java.io.File

fun proverResponsesFromDir(dir: String): List<File> {
  return (
    findPathFileOrDir(dir)
      ?: throw IllegalArgumentException("Directory not found: $dir")
    )
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

fun loadBlobs(blobsDir: String, aggregations: List<Aggregation>): List<BlobRecord> {
  return loadProverResponses(blobsDir) {
    BlobCompressionProofJsonResponse.fromJsonString(it).toDomainObject()
  }
    .let { compressionProofs ->
      val firstAggregationBlockTime = aggregations.first().let { agg ->
        agg.aggregationProof!!.finalTimestamp
          .minus(LINEA_BLOCK_INTERVAL.times((agg.endBlockNumber - agg.startBlockNumber).toInt()))
      }
      createBlobRecords(
        compressionProofs = compressionProofs,
        firstBlockStartBlockTime = firstAggregationBlockTime
      )
    }
    .sortedBy { it.startBlockNumber }
}
