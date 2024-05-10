package net.consensys.zkevm.ethereum.coordination.blockcreation

import net.consensys.linea.BlockNumberAndHash
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface ForkChoiceUpdater {
  fun updateFinalizedBlock(finalizedBlockNumberAndHash: BlockNumberAndHash): SafeFuture<Void>
}
