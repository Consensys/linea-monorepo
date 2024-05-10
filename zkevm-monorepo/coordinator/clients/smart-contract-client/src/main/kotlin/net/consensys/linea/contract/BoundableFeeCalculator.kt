package net.consensys.linea.contract

import net.consensys.linea.FeeHistory
import net.consensys.toGWei
import net.consensys.toIntervalString
import net.consensys.zkevm.ethereum.gaspricing.FeesCalculator
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.math.BigInteger

class BoundableFeeCalculator(
  val config: Config,
  private val feesCalculator: FeesCalculator
) : FeesCalculator {
  data class Config(
    val feeUpperBound: BigInteger,
    val feeLowerBound: BigInteger,
    val feeMargin: BigInteger
  ) {
    init {
      require(
        feeUpperBound >= BigInteger.ZERO && feeLowerBound >= BigInteger.ZERO &&
          feeMargin >= BigInteger.ZERO
      ) {
        "feeUpperBound, feeLowerBound, and feeMargin must be no less than 0."
      }

      require(feeUpperBound >= feeLowerBound) {
        "feeUpperBound must be no less than feeLowerBound."
      }
    }
  }

  private val log: Logger = LogManager.getLogger(this::class.java)

  override fun calculateFees(feeHistory: FeeHistory): BigInteger {
    val blockRange = feeHistory.blocksRange()
    val fee = feesCalculator.calculateFees(feeHistory) + config.feeMargin
    val gasPriceToUpdate = if (fee > config.feeUpperBound) {
      log.debug(
        "Gas fee update: fee={}GWei, l1Blocks={}. fee is higher than feeUpperBound={}GWei." +
          " Will default to the upper bound value",
        fee.toGWei(),
        blockRange.toIntervalString(),
        config.feeUpperBound.toGWei()
      )
      config.feeUpperBound
    } else if (fee < config.feeLowerBound) {
      log.debug(
        "Gas fee update: fee={}GWei, l1Blocks={}. fee is lower than feeLowerBound={}GWei." +
          " Will default to the lower bound value",
        fee.toGWei(),
        blockRange.toIntervalString(),
        config.feeLowerBound.toGWei()
      )
      config.feeLowerBound
    } else {
      log.debug(
        "Gas fee update: gasPrice={}GWei, l1Blocks={}.",
        fee.toGWei(),
        blockRange.toIntervalString()
      )
      fee
    }
    return gasPriceToUpdate
  }
}
