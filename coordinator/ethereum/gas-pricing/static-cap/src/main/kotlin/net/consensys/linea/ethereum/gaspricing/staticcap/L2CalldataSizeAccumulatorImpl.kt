package net.consensys.linea.ethereum.gaspricing.staticcap

import linea.kotlin.toBigInteger
import linea.kotlin.toUInt
import linea.web3j.ExtendedWeb3J
import net.consensys.linea.ethereum.gaspricing.L2CalldataSizeAccumulator
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

class L2CalldataSizeAccumulatorImpl(
  private val web3jClient: ExtendedWeb3J,
  private val config: Config,
) : L2CalldataSizeAccumulator {
  private val log: Logger = LogManager.getLogger(this::class.java)

  data class Config(
    val blockSizeNonCalldataOverhead: UInt,
    val calldataSizeBlockCount: UInt,
  ) {
    init {
      require(calldataSizeBlockCount <= 30u) {
        "calldataSizeBlockCount must be less than 30." +
          " Value=$calldataSizeBlockCount"
      }
    }
  }
  private fun getRecentL2CalldataSize(): SafeFuture<BigInteger> {
    return web3jClient.ethBlockNumber()
      .thenCompose { currentBlockNumber ->
        val futures =
          ((currentBlockNumber.toUInt() - config.calldataSizeBlockCount + 1U)..currentBlockNumber.toUInt())
            .map { blockNumber ->
              web3jClient.ethGetBlockSizeByNumber(blockNumber.toLong())
            }
        SafeFuture.collectAll(futures.stream())
          .thenApply { blockSizes ->
            blockSizes.sumOf {
              it.minus(config.blockSizeNonCalldataOverhead.toULong().toBigInteger())
                .coerceAtLeast(BigInteger.ZERO)
            }.also {
              log.trace(
                "sumOfblockSizes = {} blockSizes = {} blockSizeNonCalldataOverhead = {}",
                it,
                blockSizes,
                config.blockSizeNonCalldataOverhead,
              )
            }
          }
      }
  }

  override fun getSumOfL2CalldataSize(): SafeFuture<BigInteger> {
    return getRecentL2CalldataSize()
      .whenException { th ->
        log.error(
          "Get the sum of L2 calldata size from the last {} blocks failure: {}",
          config.calldataSizeBlockCount,
          th.message,
          th,
        )
      }
  }
}
