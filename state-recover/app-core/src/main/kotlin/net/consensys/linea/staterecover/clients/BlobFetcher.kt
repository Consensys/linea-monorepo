package net.consensys.linea.staterecover.clients

import tech.pegasys.teku.infrastructure.async.SafeFuture

interface BlobFetcher {
  fun fetchBlobsByHash(blobVersionedHashes: List<ByteArray>): SafeFuture<List<ByteArray>>
}
