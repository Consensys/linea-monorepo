package net.consensys.linea.staterecover.clients.blobscan

import io.vertx.core.json.JsonObject
import net.consensys.decodeHex
import net.consensys.encodeHex
import net.consensys.linea.staterecover.clients.BlobFetcher
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

class BlobScanClient(
  val restClient: RestClient<JsonObject>,
  val log: Logger = LogManager.getLogger(BlobScanClient::class.java)
) : BlobFetcher {
  fun getBlobById(id: String): SafeFuture<ByteArray> {
    return restClient
      .get("/blobs/$id")
      .thenApply { response ->
        if (response.statusCode == 200) {
          response.body!!.getString("data").decodeHex()
        } else {
          throw RuntimeException(
            "error fetching blobId=$id " +
              "errorMessage=${response.body?.getString("message") ?: ""}"
          )
        }
      }
  }

  override fun fetchBlobsByHash(blobVersionedHashes: List<ByteArray>): SafeFuture<List<ByteArray>> {
    return SafeFuture
      .collectAll(blobVersionedHashes.map { hash -> getBlobById(hash.encodeHex()) }.stream())
  }
}
