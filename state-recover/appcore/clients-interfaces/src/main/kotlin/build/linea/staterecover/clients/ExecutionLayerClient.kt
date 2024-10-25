package build.linea.staterecover.clients

import build.linea.staterecover.BlockL1RecoveredData
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.BlockParameter
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface ExecutionLayerClient {
  fun getBlockNumberAndHash(blockParameter: BlockParameter): SafeFuture<BlockNumberAndHash>
  fun lineaEngineImportBlocksFromBlob(blocks: List<BlockL1RecoveredData>): SafeFuture<Unit>
  fun lineaEngineForkChoiceUpdated(headBlockHash: ByteArray, finalizedBlockHash: ByteArray): SafeFuture<Unit>
}
