package net.consensys.zkevm.ethereum.coordination.blob

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import net.consensys.encodeHex
import net.consensys.zkevm.coordinator.clients.GetZkEVMStateMerkleProofResponse
import net.consensys.zkevm.coordinator.clients.Type2StateManagerClient
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.persistence.BlobsRepository
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64

class ParentBlobDataProviderImpl(
  private val genesisShnarf: ByteArray,
  private val blobsRepository: BlobsRepository,
  private val zkStateClient: Type2StateManagerClient
) : ParentBlobDataProvider {
  private val log: Logger = LogManager.getLogger(this::class.java)

  companion object {
    private data class BlobDataHashAndShnarf(val blobHash: ByteArray, val shnarf: ByteArray) {
      override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (javaClass != other?.javaClass) return false

        other as BlobDataHashAndShnarf

        if (!blobHash.contentEquals(other.blobHash)) return false
        if (!shnarf.contentEquals(other.shnarf)) return false

        return true
      }

      override fun hashCode(): Int {
        var result = blobHash.contentHashCode()
        result = 31 * result + shnarf.contentHashCode()
        return result
      }
    }
  }

  private fun rollupGetZkEVMStateMerkleProof(startBlockNumber: ULong, endBlockNumber: ULong):
    SafeFuture<GetZkEVMStateMerkleProofResponse> {
    return zkStateClient.rollupGetZkEVMStateMerkleProof(
      UInt64.valueOf(startBlockNumber.toLong()),
      UInt64.valueOf(endBlockNumber.toLong())
    ).thenCompose {
      when (it) {
        is Ok -> SafeFuture.completedFuture(it.value)
        is Err -> {
          SafeFuture.failedFuture(it.error.asException())
        }
      }
    }
  }

  private fun getParentBlobData(endBlockNumber: ULong): SafeFuture<BlobDataHashAndShnarf> {
    return if (endBlockNumber == 0UL) {
      log.info(
        "Requested parent shnarf for the genesis block, returning genesisShnarf={}",
        genesisShnarf.encodeHex()
      )
      SafeFuture.completedFuture(BlobDataHashAndShnarf(ByteArray(32), genesisShnarf))
    } else {
      blobsRepository
        .findBlobByEndBlockNumber(endBlockNumber.toLong())
        .thenCompose { blobRecord: BlobRecord? ->
          if (blobRecord != null) {
            SafeFuture.completedFuture(
              BlobDataHashAndShnarf(
                blobHash = blobRecord.blobHash,
                shnarf = blobRecord.expectedShnarf
              )
            )
          } else {
            SafeFuture.failedFuture(
              IllegalStateException("Failed to find the parent blob in db with end block=$endBlockNumber")
            )
          }
        }
    }
  }

  override fun findParentAndZkStateData(
    blockRange: ULongRange
  ): SafeFuture<ParentBlobAndZkStateData> {
    return getParentBlobData(
      blockRange.first.dec()
    ).thenComposeCombined(
      rollupGetZkEVMStateMerkleProof(blockRange.first, blockRange.last)
    ) { blobData: BlobDataHashAndShnarf, zkStateData: GetZkEVMStateMerkleProofResponse ->
      SafeFuture.completedFuture(
        ParentBlobAndZkStateData(
          blobData.blobHash,
          blobData.shnarf,
          zkStateData.zkParentStateRootHash.toArray(),
          zkStateData.zkEndStateRootHash.toArray()
        )
      )
    }
  }
}
