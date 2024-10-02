package net.consensys.linea.staterecover.clients.elclient

import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.BlockParameter
import net.consensys.linea.staterecover.domain.BlockL1RecoveredData
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface ExecutionLayerClient {
  fun getBlockNumberAndHash(blockParameter: BlockParameter): SafeFuture<BlockNumberAndHash>
  fun lineaEngineImportBlocksFromBlob(blocks: List<BlockL1RecoveredData>): SafeFuture<Unit>
  fun lineaEngineForkChoiceUpdated(headBlockHash: ByteArray, finalizedBlockHash: ByteArray): SafeFuture<Unit>
}
