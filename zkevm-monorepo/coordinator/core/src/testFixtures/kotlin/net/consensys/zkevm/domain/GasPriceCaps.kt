package net.consensys.zkevm.domain

import net.consensys.zkevm.ethereum.gaspricing.GasPriceCaps
import java.math.BigInteger

val defaultGasPriceCaps = GasPriceCaps(
  maxPriorityFeePerGasCap = BigInteger.valueOf(10000000000),
  maxFeePerGasCap = BigInteger.valueOf(10000000000),
  maxFeePerBlobGasCap = BigInteger.valueOf(1000000000)
)
