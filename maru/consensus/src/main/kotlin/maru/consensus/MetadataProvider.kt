/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus

import java.util.concurrent.atomic.AtomicReference
import maru.core.BeaconBlock
import org.apache.tuweni.bytes.Bytes32
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter

fun interface MetadataProvider {
  fun getLatestBlockMetadata(): BlockMetadata
}

data class BlockMetadata(
  val blockNumber: ULong,
  val blockHash: ByteArray,
  val unixTimestampSeconds: Long, // Since the use of Java standard lib, Long is more practical than ULong
) {
  companion object {
    fun fromBeaconBlock(beaconBlock: BeaconBlock): BlockMetadata =
      BlockMetadata(
        beaconBlock.beaconBlockBody.executionPayload.blockNumber,
        beaconBlock.beaconBlockBody
          .executionPayload.blockHash,
        beaconBlock.beaconBlockBody.executionPayload.timestamp
          .toLong(),
      )
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as BlockMetadata

    if (blockNumber != other.blockNumber) return false
    if (!blockHash.contentEquals(other.blockHash)) return false
    if (unixTimestampSeconds != other.unixTimestampSeconds) return false

    return true
  }

  override fun hashCode(): Int {
    var result = blockNumber.hashCode()
    result = 31 * result + blockHash.contentHashCode()
    result = 31 * result + unixTimestampSeconds.hashCode()
    return result
  }
}

class LatestBlockMetadataCache(
  currentBlockMetadata: BlockMetadata,
) : MetadataProvider {
  private var latestBlockMetadataCache = AtomicReference(currentBlockMetadata)

  override fun getLatestBlockMetadata(): BlockMetadata = latestBlockMetadataCache.get()

  fun updateLatestBlockMetadata(blockMetadata: BlockMetadata) {
    latestBlockMetadataCache.set(blockMetadata)
  }
}

class Web3jMetadataProvider(
  private val web3jEthereumApiClient: Web3j,
) {
  fun getLatestBlockMetadata(): BlockMetadata {
    val block =
      web3jEthereumApiClient
        .ethGetBlockByNumber(DefaultBlockParameter.valueOf("latest"), false)
        .send()
        .block
    return BlockMetadata(
      block.number
        .toLong()
        .toULong(),
      Bytes32.fromHexString(block.hash).toArray(),
      block.timestamp.toLong(),
    )
  }
}
