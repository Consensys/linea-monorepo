package net.consensys.zkevm.ethereum.gaspricing

import net.consensys.linea.FeeHistory
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

interface FeesFetcher {
  fun getL1EthGasPriceData(): SafeFuture<FeeHistory>
}

interface FeesCalculator {
  fun calculateFees(feeHistory: FeeHistory): BigInteger
}

interface GasPriceUpdater {
  fun updateMinerGasPrice(gasPrice: BigInteger): SafeFuture<Unit>
}
