package linea.staterecover

import build.linea.staterecover.BlockL1RecoveredData
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.BlockParameter
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class StateRecoveryStatus(
  val headBlockNumber: ULong,
  val stateRecoverStartBlockNumber: ULong?
)
interface ExecutionLayerClient {
  fun getBlockNumberAndHash(blockParameter: BlockParameter): SafeFuture<BlockNumberAndHash>
  fun lineaEngineImportBlocksFromBlob(blocks: List<BlockL1RecoveredData>): SafeFuture<Unit>
  fun lineaGetStateRecoveryStatus(): SafeFuture<StateRecoveryStatus>
  fun lineaEnableStateRecovery(stateRecoverStartBlockNumber: ULong): SafeFuture<StateRecoveryStatus>
}
