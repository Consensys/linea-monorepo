package net.consensys.linea.staterecover.clients

import tech.pegasys.teku.infrastructure.async.SafeFuture

interface TransactionDetailsClient {
  fun getBlobVersionedHashesByTransactionHash(transactionHash: ByteArray): SafeFuture<List<ByteArray>>
}
