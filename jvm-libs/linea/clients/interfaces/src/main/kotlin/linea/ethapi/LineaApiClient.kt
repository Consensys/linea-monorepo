package linea.ethapi

import linea.domain.TransactionForEthCall
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

interface LineaApiClient {
  fun lineaEstimateGas(transaction: TransactionForEthCall): SafeFuture<BigInteger>
}
