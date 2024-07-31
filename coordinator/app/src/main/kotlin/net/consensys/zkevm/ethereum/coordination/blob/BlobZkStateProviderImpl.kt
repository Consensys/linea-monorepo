package net.consensys.zkevm.ethereum.coordination.blob

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import net.consensys.zkevm.coordinator.clients.GetZkEVMStateMerkleProofResponse
import net.consensys.zkevm.coordinator.clients.Type2StateManagerClient
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64

class BlobZkStateProviderImpl(private val zkStateClient: Type2StateManagerClient) : BlobZkStateProvider {
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

  override fun getBlobZKState(blockRange: ULongRange): SafeFuture<BlobZkState> {
    return rollupGetZkEVMStateMerkleProof(blockRange.first, blockRange.last).thenApply {
      BlobZkState(
        parentStateRootHash = it.zkParentStateRootHash.toArray(),
        finalStateRootHash = it.zkEndStateRootHash.toArray()
      )
    }
  }
}
