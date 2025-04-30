/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
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
