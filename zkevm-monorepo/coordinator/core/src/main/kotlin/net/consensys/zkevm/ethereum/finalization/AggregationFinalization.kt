package net.consensys.zkevm.ethereum.finalization

import net.consensys.zkevm.domain.ProofToFinalize
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface AggregationFinalization {

  fun finalizeAggregation(aggregationProof: ProofToFinalize): SafeFuture<*>

  fun finalizeAggregationEthCall(aggregationProof: ProofToFinalize): SafeFuture<*>
}
