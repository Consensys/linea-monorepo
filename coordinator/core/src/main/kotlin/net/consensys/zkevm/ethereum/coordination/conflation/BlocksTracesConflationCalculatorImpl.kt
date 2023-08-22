package net.consensys.zkevm.ethereum.coordination.conflation

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.linea.traces.TracesCounters
import net.consensys.linea.traces.TracingModule
import net.consensys.zkevm.ethereum.coordination.blockcreation.SafeBlockProvider
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.concurrent.timer
import kotlin.math.roundToInt
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

val NOOP_CONSUMER: (ConflationCalculationResult) -> SafeFuture<*> =
  { _: ConflationCalculationResult ->
    SafeFuture.completedFuture(Unit)
  }

data class DataLimits(
  /**
   * Total number of bytes that can be used for blocks conflation;
   * This should be around:
   *  L1_transaction_size_limit(128KB)
   *    - (signature and rlp overhead)
   *    - proof size
   *    - proof type
   *    - blockDataOffsetEncodingSize (BlockData[] rlp overhead)
   */
  val totalLimitBytes: UInt,

  /**
   * Number of extra bytes each block will require to send to L1, due smart contract interface;
   * This should be around:
   *  blockRootHash (32 bytes) +
   *  l2BlockTimestamp 4 bytes +
   */
  val perBlockOverheadBytes: UInt,

  /**
   * Minimum number of bytes the smallest block with single ETH transfer would require;
   * If totalAccumulatedData + ninBlockL1Size > totalLimitBytes, then conflation will be triggered
   */
  val minBlockL1SizeBytes: UInt
)

data class ConflationCalculatorConfig(
  val tracesConflationLimit: TracesCounters,
  val dataConflationLimits: DataLimits,
  val conflationDeadline: Duration,
  val conflationDeadlineCheckInterval: Duration,
  val conflationDeadlineLastBlockConfirmationDelay: Duration,
  val blocksLimit: UInt? = null
) {
  init {
    require(tracesConflationLimit.keys.toSet() == TracingModule.values().toSet()) {
      "Conflation caps must have all EVM tracing modules"
    }
    require(conflationDeadlineCheckInterval < conflationDeadline) {
      "Clock ticker interval must be smaller than conflation deadline"
    }
  }
}

class BlocksTracesConflationCalculatorImpl(
  override var lastBlockNumber: ULong,
  private val latestBlockProvider: SafeBlockProvider,
  private val config: ConflationCalculatorConfig,
  private val clock: Clock = Clock.System,
  private val log: Logger = LogManager.getLogger(BlocksTracesConflationCalculatorImpl::class.java)
) : TracesConflationCalculator {
  private var currentTracesCount =
    TracingModule.values().associateWith { 0u } as MutableMap<TracingModule, UInt>
  private var currentBlocksCount: UInt = 0u
  private var inprogressBlocksAccumulatedData: UInt = 0u
  private var startBlockNumber: ULong? = null
  private var startBlockTimestamp: Instant? = null
  private var endBlockNumber: ULong? = null
  private var conflationConsumer: (ConflationCalculationResult) -> SafeFuture<*> = NOOP_CONSUMER
  private val configuredEvmModules: Set<TracingModule> = config.tracesConflationLimit.keys

  init {
    timer(
      name = "conflation-deadline-checker",
      initialDelay = config.conflationDeadlineCheckInterval.inWholeMilliseconds,
      period = config.conflationDeadlineCheckInterval.inWholeMilliseconds
    ) {
      checkConflationDeadline()
    }
  }

  @Synchronized
  internal fun checkConflationDeadline() {
    val now = clock.now()
    log.trace(
      "Checking conflation deadline: startBlockTime={}, timeElapsed={}, deadline={} currentBlocksCount={}",
      startBlockTimestamp,
      startBlockTimestamp?.let { now.minus(it) } ?: 0.seconds,
      config.conflationDeadline,
      currentBlocksCount
    )

    val deadlineReachedForFirstBlockInProgress =
      startBlockTimestamp != null && now > startBlockTimestamp!!.plus(config.conflationDeadline)

    latestBlockProvider.getLatestSafeBlockHeader().thenPeek {
      // wait for 2+ block intervals, otherwise if the ticker happens during block creation we get a false positive.

      val noMoreBlocksInL2ChainToConflate =
        it.number == lastBlockNumber && now.minus(config.conflationDeadlineLastBlockConfirmationDelay) > it.timestamp

      if (deadlineReachedForFirstBlockInProgress && noMoreBlocksInL2ChainToConflate) {
        log.debug("Conflation trigger: Deadline reached")
        conflationWithoutOverflow(lastBlockNumber, HashMap(currentTracesCount), ConflationTrigger.TIME_LIMIT)
      }
    }.whenException { th ->
      log.error(
        "SafeBlock request failed. Will Retry conflation deadline on next tick errorMessage={}",
        th.message,
        th
      )
    }
  }

  @Synchronized
  override fun newBlock(blockCounters: BlockCounters) {
    ensureBlockIsInOrder(blockCounters.blockNumber)
    ensureEvmModules(blockCounters.blockNumber, blockCounters.tracesCounters)
    val tracesValidationResult = checkTracesAreWithinCaps(blockCounters.blockNumber, blockCounters.tracesCounters)
    val dataValidationResult = checkDataSizeIsWithinLimit(blockCounters.blockNumber, blockCounters.l1DataSize)

    log.trace(
      "Checking conflation quotas: blockNumber={} blockData={}bytes blockTraces={}, " +
        "accumulatedData={}bytes, accumulatedTraces={}",
      blockCounters.blockNumber,
      blockCounters.l1DataSize,
      blockCounters.tracesCounters,
      currentTracesCount,
      config.tracesConflationLimit
    )

    var oversizeTrigger: ConflationTrigger? = null
    if (tracesValidationResult is Err) {
      oversizeTrigger = ConflationTrigger.TRACES_LIMIT
      log.warn(tracesValidationResult.component2())
    }
    if (dataValidationResult is Err) {
      oversizeTrigger = ConflationTrigger.DATA_LIMIT
      log.warn(dataValidationResult.component2())
    }

    if (oversizeTrigger != null) {
      if (startBlockNumber != null) {
        // flush whatever we have in progress
        conflationWithoutOverflow(lastBlockNumber, currentTracesCount, oversizeTrigger)
      }
      startBlockNumber = blockCounters.blockNumber
      startBlockTimestamp = blockCounters.blockTimestamp
      inprogressBlocksAccumulatedData = blockCounters.l1DataSize + config.dataConflationLimits.perBlockOverheadBytes
      // flush oversize block
      conflationWithoutOverflow(blockCounters.blockNumber, blockCounters.tracesCounters, oversizeTrigger)
    } else {
      if (startBlockNumber == null) {
        startBlockNumber = blockCounters.blockNumber
        startBlockTimestamp = blockCounters.blockTimestamp
      }
      val conflationResult = conflateBlock(blockCounters)
      handleConflationResult(conflationResult, blockCounters)
    }

    lastBlockNumber = blockCounters.blockNumber
  }

  private data class BlockConflationResult(
    val moduleFull: Boolean,
    val moduleOverflow: Boolean,
    val dataFull: Boolean,
    val dataOverflow: Boolean,
    val blocksLimitReached: Boolean,
    val blockAccumulatedTracesCounters: HashMap<TracingModule, UInt>
  )

  private fun conflateBlock(blockCounters: BlockCounters): BlockConflationResult {
    val currentTracesCountSnapshot = HashMap(currentTracesCount)
    var moduleFull = false
    var moduleOverflow = false
    var dataFull = false
    var dataOverflow = false
    val triggerModuleDataList = mutableListOf<ConflationTriggerModuleData>()

    if (isWithinDataLimit(blockCounters.l1DataSize)) {
      inprogressBlocksAccumulatedData += blockCounters.l1DataSize + config.dataConflationLimits.perBlockOverheadBytes

      for (module in blockCounters.tracesCounters.entries) {
        val currentModuleTraceCount = currentTracesCountSnapshot[module.key]!!
        val moduleTraceCap = config.tracesConflationLimit[module.key]!!
        triggerModuleDataList.add(
          ConflationTriggerModuleData(
            module.key,
            currentModuleTraceCount,
            module.value,
            moduleTraceCap
          )
        )

        if (currentModuleTraceCount + module.value <= moduleTraceCap) {
          currentTracesCountSnapshot[module.key] = currentModuleTraceCount + module.value
          if (currentModuleTraceCount + module.value == moduleTraceCap) {
            if (!moduleOverflow) {
              moduleFull = true
              endBlockNumber = blockCounters.blockNumber
            }
          }
        } else {
          moduleOverflow = true
          moduleFull = false
          endBlockNumber = lastBlockNumber
        }
      }

      if (moduleFull || moduleOverflow) {
        logConflationTrigger(blockCounters.blockNumber, triggerModuleDataList)
      }
      dataFull = isDataFull()
    } else {
      endBlockNumber = lastBlockNumber
      dataOverflow = true
    }
    currentBlocksCount += 1u
    val blocksLimitReached = config.blocksLimit != null && currentBlocksCount >= config.blocksLimit

    return BlockConflationResult(
      moduleFull,
      moduleOverflow,
      dataFull,
      dataOverflow,
      blocksLimitReached,
      currentTracesCountSnapshot
    )
  }

  private fun handleConflationResult(conflationResult: BlockConflationResult, blockCounters: BlockCounters) {
    when {
      conflationResult.dataOverflow -> conflationWithOverflow(blockCounters, ConflationTrigger.DATA_LIMIT)
      conflationResult.moduleOverflow -> conflationWithOverflow(blockCounters, ConflationTrigger.TRACES_LIMIT)
      conflationResult.dataFull ->
        conflationWithoutOverflow(
          blockCounters.blockNumber,
          conflationResult.blockAccumulatedTracesCounters,
          ConflationTrigger.DATA_LIMIT
        )

      conflationResult.moduleFull ->
        conflationWithoutOverflow(
          blockCounters.blockNumber,
          conflationResult.blockAccumulatedTracesCounters,
          ConflationTrigger.TRACES_LIMIT
        )

      conflationResult.blocksLimitReached ->
        conflationWithoutOverflow(
          blockCounters.blockNumber,
          conflationResult.blockAccumulatedTracesCounters,
          ConflationTrigger.BLOCKS_LIMIT
        )

      else -> {
        log.trace("No conflation")
        currentTracesCount = conflationResult.blockAccumulatedTracesCounters
      }
    }
  }

  private fun ensureEvmModules(blockNumber: ULong, tracesCounters: Map<TracingModule, UInt>) {
    val blockTracesModules = tracesCounters.keys
    if (configuredEvmModules != blockTracesModules) {
      val missingModules = configuredEvmModules - blockTracesModules
      val extraModules = blockTracesModules - configuredEvmModules
      val errorMessage = "Invalid traces counters for block $blockNumber. " +
        "countersMissing: ${missingModules.joinToString(",", "[", "]")}, " +
        "countersExtra=${extraModules.joinToString(",", "[", "]")}"
      log.error(errorMessage)
      throw IllegalStateException(errorMessage)
    }
  }

  private fun conflationWithOverflow(
    blockCounters: BlockCounters,
    conflationTrigger: ConflationTrigger
  ) {
    val conflationCalculationResult =
      ConflationCalculationResult(
        startBlockNumber!!,
        endBlockNumber!!,
        HashMap(currentTracesCount),
        inprogressBlocksAccumulatedData,
        conflationTrigger
      )
    lastBlockNumber = blockCounters.blockNumber
    startBlockNumber = blockCounters.blockNumber
    startBlockTimestamp = blockCounters.blockTimestamp
    endBlockNumber = null
    currentTracesCount = blockCounters.tracesCounters.toMutableMap()
    currentBlocksCount = 1u
    inprogressBlocksAccumulatedData = blockCounters.l1DataSize + config.dataConflationLimits.perBlockOverheadBytes
    this.conflationConsumer.invoke(conflationCalculationResult)
  }

  private fun conflationWithoutOverflow(
    blockNumber: ULong,
    currentTracesCountSnapshot: TracesCounters,
    conflationTrigger: ConflationTrigger
  ) {
    val conflationCalculationResult =
      ConflationCalculationResult(
        startBlockNumber!!,
        blockNumber,
        HashMap(currentTracesCountSnapshot),
        inprogressBlocksAccumulatedData,
        conflationTrigger
      )

    startBlockNumber = null
    startBlockTimestamp = null
    endBlockNumber = null
    TracingModule.values().forEach { currentTracesCount[it] = 0u }
    currentBlocksCount = 0u
    inprogressBlocksAccumulatedData = 0u
    this.conflationConsumer.invoke(conflationCalculationResult)
  }

  @Synchronized
  override fun getConflationInProgress(): ConflationCalculationResult? {
    if (startBlockNumber == null) {
      return null
    }
    return ConflationCalculationResult(
      startBlockNumber!!,
      endBlockNumber ?: lastBlockNumber,
      currentTracesCount.toMap(),
      inprogressBlocksAccumulatedData,
      ConflationTrigger.TIME_LIMIT
    )
  }

  override fun onConflatedBatch(
    conflationConsumer: (ConflationCalculationResult) -> SafeFuture<*>
  ) {
    if (this.conflationConsumer != NOOP_CONSUMER) {
      throw IllegalStateException("Consumer is already set")
    }
    this.conflationConsumer = conflationConsumer
  }

  private fun isWithinDataLimit(blockL1Data: UInt): Boolean {
    return inprogressBlocksAccumulatedData +
      blockL1Data +
      config.dataConflationLimits.perBlockOverheadBytes <=
      config.dataConflationLimits.totalLimitBytes
  }

  private fun isDataFull(): Boolean {
    return inprogressBlocksAccumulatedData + config.dataConflationLimits.minBlockL1SizeBytes >=
      config.dataConflationLimits.totalLimitBytes
  }

  private fun checkTracesAreWithinCaps(blockNumber: ULong, tracesCounters: TracesCounters): Result<Unit, String> {
    val overSizeTraces = mutableListOf<Triple<TracingModule, UInt, UInt>>()
    for (moduleEntry in tracesCounters.entries) {
      val moduleCap = config.tracesConflationLimit[moduleEntry.key]!!
      if (moduleEntry.value > moduleCap) {
        overSizeTraces.add(Triple(moduleEntry.key, moduleEntry.value, moduleCap))
      }
    }
    return if (overSizeTraces.isNotEmpty()) {
      val errorMessage = overSizeTraces.joinToString(
        ", ",
        "Block $blockNumber has oversize traces TRACE(count, limit, overflow): [",
        "]"
      ) {
        "${it.first}(${it.second}, ${it.third}, ${it.second - it.third})"
      }
      Err(errorMessage)
    } else {
      Ok(Unit)
    }
  }

  private fun checkDataSizeIsWithinLimit(blockNumber: ULong, blockDataSize: UInt): Result<Unit, String> {
    return if (blockDataSize + config.dataConflationLimits.perBlockOverheadBytes >
      config.dataConflationLimits.totalLimitBytes
    ) {
      val overflow = (blockDataSize + config.dataConflationLimits.perBlockOverheadBytes) -
        config.dataConflationLimits.totalLimitBytes
      val errorMessage = "Block $blockNumber has oversize data (bytes): blockL1Size=$blockDataSize " +
        "perBlockOverhead=${config.dataConflationLimits.perBlockOverheadBytes} " +
        "limit=${config.dataConflationLimits.totalLimitBytes} " +
        "overflow=$overflow"
      Err(errorMessage)
    } else {
      Ok(Unit)
    }
  }

  private fun ensureBlockIsInOrder(blockNumber: ULong) {
    if (blockNumber != (lastBlockNumber + 1u)) {
      val error = IllegalArgumentException(
        "Blocks to conflate must be sequential: lastBlockNumber=$lastBlockNumber, new blockNumber=$blockNumber"
      )
      log.error(error.message)
      throw error
    }
  }

  data class ConflationTriggerModuleData(
    val module: TracingModule,
    val moduleCurrentTraceCount: UInt,
    val moduleTraceValueInNewBlock: UInt,
    val moduleTraceCap: UInt
  )

  private fun logConflationTrigger(
    blockNumber: ULong,
    conflationTriggerModuleDataList: MutableList<ConflationTriggerModuleData>
  ) {
    fun calculateTracesCountRatio(counter: UInt, limit: UInt): Double {
      return (counter.toDouble().div(limit.toDouble()) * 100.0).roundToInt() / 100.0
    }

    val moduleDataList = conflationTriggerModuleDataList.sortedByDescending {
      calculateTracesCountRatio(it.moduleCurrentTraceCount + it.moduleTraceValueInNewBlock, it.moduleTraceCap)
    }.map { data ->
      val limit = data.moduleTraceCap
      val counter = data.moduleCurrentTraceCount
      val toAdd = data.moduleTraceValueInNewBlock
      val overflow = if (counter + toAdd >= limit) counter + toAdd - limit else 0
      val ratio = calculateTracesCountRatio(counter + toAdd, limit)
      "${data.module.name}={limit=$limit, counter=$counter, toAdd=$toAdd, ratio=$ratio, overflow=$overflow}"
    }

    log.debug(
      "Conflation trigger: triggeredByBlock={}, batch=[{}..{}]({}), modules={}",
      blockNumber,
      startBlockNumber,
      endBlockNumber,
      endBlockNumber!! - startBlockNumber!! + 1u,
      moduleDataList
    )
  }
}
