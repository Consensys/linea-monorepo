package linea.ethapi

import linea.domain.Block
import linea.domain.BlockParameter
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface EthApiClient : EthLogsClient {
  fun getBlockByNumber(
    blockParameter: BlockParameter,
    includeTransactions: Boolean = false
  ): SafeFuture<Block?>
}
