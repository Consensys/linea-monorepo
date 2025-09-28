package net.consensys.linea.ethereum.gaspricing.staticcap

import linea.OneKWei
import linea.domain.BlockParameter
import linea.domain.BlockParameter.Companion.toBlockParameter
import linea.kotlin.encodeHex
import linea.web3j.ExtendedWeb3J
import net.consensys.linea.ethereum.gaspricing.HistoricVariableCostProvider
import net.consensys.linea.ethereum.gaspricing.MinerExtraDataV1
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.atomic.AtomicReference

class HistoricVariableCostProviderImpl(
  private val web3jClient: ExtendedWeb3J,
) : HistoricVariableCostProvider {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private var lastVariableCost: AtomicReference<Pair<BigInteger, Double>> =
    AtomicReference(BigInteger.ZERO to 0.0)

  private fun getHistoricVariableCostInWei(blockParameter: BlockParameter): SafeFuture<Double> {
    return web3jClient.ethGetBlock(blockParameter)
      .thenApply { block ->
        try {
          MinerExtraDataV1.decodeV1(block!!.extraData.encodeHex())
            .variableCostInKWei.toDouble() * OneKWei
        } catch (th: Throwable) {
          if (block != null) {
            log.debug(
              "Will return historic variable cost as zero due to failure in decoding extra data: {}",
              th.message,
            )
            0.0
          } else {
            throw th
          }
        }
      }
  }

  override fun getLatestVariableCost(): SafeFuture<Double> {
    return web3jClient.ethBlockNumber()
      .thenCompose { currentBlockNumber ->
        if (lastVariableCost.get().first == currentBlockNumber) {
          log.debug(
            "Use cached lastVariableCost={} currentBlockNumber={}",
            lastVariableCost.get().second,
            currentBlockNumber,
          )
          SafeFuture.completedFuture(lastVariableCost.get().second)
        } else {
          getHistoricVariableCostInWei(currentBlockNumber.toBlockParameter())
            .thenPeek { variableCost ->
              log.debug(
                "variableCost={} blockNumber={}",
                variableCost,
                currentBlockNumber,
              )
              lastVariableCost.set(currentBlockNumber to variableCost)
            }
            .whenException { th ->
              log.error(
                "Get the variable cost from latest L2 block extra data failure: {}",
                th.message,
                th,
              )
            }
        }
      }
  }
}
