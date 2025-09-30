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
import java.util.concurrent.atomic.AtomicReference

class HistoricVariableCostProviderImpl(
  private val web3jClient: ExtendedWeb3J,
) : HistoricVariableCostProvider {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private var lastVariableCost: AtomicReference<Pair<ULong, Double>> =
    AtomicReference(0UL to 0.0)

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

  override fun getVariableCost(blockNumber: ULong): SafeFuture<Double> {
    val (cachedBlockNumber, cachedVariableCost) = lastVariableCost.get()
    return if (cachedBlockNumber == blockNumber) {
      log.debug(
        "Use cached lastVariableCost={} blockNumber={}",
        cachedVariableCost,
        blockNumber,
      )
      SafeFuture.completedFuture(cachedVariableCost)
    } else {
      getHistoricVariableCostInWei(blockNumber.toBlockParameter())
        .thenPeek { variableCost ->
          log.debug(
            "variableCost={} blockNumber={}",
            variableCost,
            blockNumber,
          )
          lastVariableCost.set(blockNumber to variableCost)
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
