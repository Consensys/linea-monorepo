package net.consensys.zkevm.ethereum.coordination.conflation

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import net.consensys.linea.traces.TracesCounters
import net.consensys.linea.traces.TracingModule
import net.consensys.linea.traces.allTracesEmpty
import net.consensys.linea.traces.allTracesWithinLimits
import net.consensys.linea.traces.emptyTracesCounts
import net.consensys.linea.traces.sumTracesCounters
import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.domain.ConflationTrigger
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

class ConflationCalculatorByExecutionTraces(
  val tracesCountersLimit: TracesCounters,
  private val log: Logger = LogManager.getLogger(ConflationCalculatorByExecutionTraces::class.java)
) : ConflationCalculator {
  init {
    require(tracesCountersLimit.keys.toSet() == TracingModule.values().toSet()) {
      "Conflation caps must have all EVM tracing modules"
    }
  }

  override val id: String = ConflationTrigger.TRACES_LIMIT.name
  private var inprogressTracesCounters: TracesCounters = emptyTracesCounts().toMutableMap()

  override fun checkOverflow(blockCounters: BlockCounters): ConflationCalculator.OverflowTrigger? {
    return if (canAppendTraces(blockCounters.tracesCounters)) {
      return null
    } else {
      val tracesValidationResult = checkTracesAreWithinCaps(blockCounters.blockNumber, blockCounters.tracesCounters)
      if (tracesValidationResult is Err) {
        log.warn(tracesValidationResult.error)
        ConflationCalculator.OverflowTrigger(ConflationTrigger.TRACES_LIMIT, true)
      } else {
        ConflationCalculator.OverflowTrigger(ConflationTrigger.TRACES_LIMIT, false)
      }
    }
  }

  override fun appendBlock(blockCounters: BlockCounters) {
    val appendResult = sumTracesCounters(inprogressTracesCounters, blockCounters.tracesCounters)
    if (isOversizedBlockOnTopOfNonEmptyConflation(appendResult)) {
      // if single block overflows traces conflation limits,
      // we allow in because it will trigger conflation next and will be flushed
      throw IllegalStateException("Block ${blockCounters.blockNumber} overflows traces conflation limits.")
    } else {
      inprogressTracesCounters = appendResult
    }
  }

  private fun isOversizedBlockOnTopOfNonEmptyConflation(
    countersAfterConflation: TracesCounters
  ): Boolean {
    return !allTracesWithinLimits(countersAfterConflation, tracesCountersLimit) && !allTracesEmpty(
      inprogressTracesCounters
    )
  }

  private fun canAppendTraces(tracesCounters: TracesCounters): Boolean {
    return allTracesWithinLimits(
      sumTracesCounters(inprogressTracesCounters, tracesCounters),
      tracesCountersLimit
    )
  }

  override fun reset() {
    this.inprogressTracesCounters = emptyTracesCounts().toMutableMap()
  }

  override fun copyCountersTo(
    counters: ConflationCounters
  ) {
    counters.tracesCounters = inprogressTracesCounters.toMap()
  }

  private fun checkTracesAreWithinCaps(blockNumber: ULong, tracesCounters: TracesCounters): Result<Unit, String> {
    val overSizeTraces = mutableListOf<Triple<TracingModule, UInt, UInt>>()
    for (moduleEntry in tracesCounters.entries) {
      val moduleCap = tracesCountersLimit[moduleEntry.key]!!
      if (moduleEntry.value > moduleCap) {
        overSizeTraces.add(Triple(moduleEntry.key, moduleEntry.value, moduleCap))
      }
    }
    return if (overSizeTraces.isNotEmpty()) {
      val errorMessage = overSizeTraces.joinToString(
        separator = ", ",
        prefix = "oversized block: block=$blockNumber, oversize traces TRACE(count, limit, overflow): [",
        postfix = "]"
      ) { (moduleName, count, limit) ->
        "$moduleName($count, $limit, ${count - limit})"
      }
      Err(errorMessage)
    } else {
      Ok(Unit)
    }
  }
}
