package linea.blob

import org.junit.jupiter.api.Test
import java.util.PriorityQueue
import kotlin.random.Random
import kotlin.time.Duration
import kotlin.time.measureTime

/**
 * Benchmark to compare TxCompressor JNA wrapper performance with the Go benchmark.
 * This helps identify if the JNA overhead or transaction encoding is causing performance issues.
 */
class TxCompressorJnaBenchmarkTest {

  companion object {
    private const val BLOB_LIMIT = 128 * 1024
    private const val WARMUP_BLOCKS = 5
    private const val MEASURE_BLOCKS = 10
  }

  @Test
  fun `benchmark TxCompressor appendTransaction with detailed timing`() {
    val scenarios = listOf(
      "erc20-transfers" to ::generateErc20Transfers,
      "calldata-3kb" to { count: Int -> generateCalldataTxs(count, 3 * 1024) },
      "calldata-500b" to { count: Int -> generateCalldataTxs(count, 500) },
      "mixed" to ::generateMixedTxs,
    )

    println("=== TxCompressor JNA Benchmark (appendTransaction only, post-processing style) ===")
    println("Blob limit: $BLOB_LIMIT bytes, Warmup blocks: $WARMUP_BLOCKS, Measure blocks: $MEASURE_BLOCKS")
    println()

    for ((name, generator) in scenarios) {
      println("--- Scenario: $name ---")

      val txDataList = generator(5000)

      // Warmup
      println("Warmup...")
      benchmarkAppendTransaction(txDataList, WARMUP_BLOCKS, measure = false)

      // Measure
      println("Measuring...")
      val result = benchmarkAppendTransaction(txDataList, MEASURE_BLOCKS, measure = true)

      printDetailedResult(name, result)
      println()
    }
  }

  @Test
  fun `benchmark TxCompressor compressedSize (stateless estimate)`() {
    val scenarios = listOf(
      "erc20-transfers" to ::generateErc20Transfers,
      "calldata-3kb" to { count: Int -> generateCalldataTxs(count, 3 * 1024) },
      "calldata-500b" to { count: Int -> generateCalldataTxs(count, 500) },
      "mixed" to ::generateMixedTxs,
    )

    println("=== TxCompressor JNA Benchmark (compressedSize - stateless per-tx estimate) ===")
    println("Blob limit: $BLOB_LIMIT bytes")
    println("This simulates the fast path: stateless compression of individual transactions")
    println()

    val compressor = GoBackedTxCompressor.getInstance(TxCompressorVersion.V2, BLOB_LIMIT)

    for ((name, generator) in scenarios) {
      println("--- Scenario: $name ---")

      val txDataList = generator(2000)

      // Warmup
      println("Warmup...")
      benchmarkCompressedSize(compressor, txDataList, WARMUP_BLOCKS, measure = false)

      // Measure
      println("Measuring...")
      val result = benchmarkCompressedSize(compressor, txDataList, MEASURE_BLOCKS, measure = true)

      println("$name (stateless compressedSize):")
      println("  Total txs: ${result.totalTxs}")
      println(
        "  Per-tx (us): min=%7.2f  p95=%7.2f  avg=%7.2f  max=%7.2f".format(
          result.perTxTimingNs.min / 1000.0,
          result.perTxTimingNs.p95 / 1000.0,
          result.perTxTimingNs.avg / 1000.0,
          result.perTxTimingNs.max / 1000.0,
        ),
      )
      println()
    }
  }

  private data class StatelessBenchmarkResult(
    val totalTxs: Int,
    val perTxTimingNs: TimingStats,
  )

  private fun benchmarkCompressedSize(
    compressor: TxCompressor,
    txDataList: List<TransactionData>,
    iterations: Int,
    measure: Boolean,
  ): StatelessBenchmarkResult {
    val totalTxs = txDataList.size * iterations
    val perTxTimings = if (measure) TimingAccumulator(totalTxs) else null

    repeat(iterations) {
      for (txData in txDataList) {
        val txStartNs = System.nanoTime()
        compressor.compressedSize(txData.from + txData.rlpForSigning)
        val txElapsedNs = System.nanoTime() - txStartNs
        perTxTimings?.record(txElapsedNs)
      }
    }

    return StatelessBenchmarkResult(
      totalTxs = totalTxs,
      perTxTimingNs = perTxTimings?.toStats() ?: TimingStats(0, 0, 0.0, 0),
    )
  }

  private data class DetailedBenchmarkResult(
    val blocks: Int,
    val avgTxsPerBlock: Double,
    val totalTxs: Int,
    val perTxTimingNs: TimingStats,
    val perBlockTimingNs: TimingStats,
    val totalTime: Duration,
  )

  private data class TimingStats(
    val min: Long,
    val max: Long,
    val avg: Double,
    val p95: Long,
  )

  private fun benchmarkAppendTransaction(
    txDataList: List<TransactionData>,
    blocksCount: Int,
    measure: Boolean,
  ): DetailedBenchmarkResult {
    val compressor = GoBackedTxCompressor.getInstance(
      TxCompressorVersion.V2,
      BLOB_LIMIT,
      enableRecompress = false,
    )

    var totalTxs = 0
    var cursor = 0

    val perTxTimings = if (measure) TimingAccumulator(blocksCount * 2000) else null
    val perBlockTimings = if (measure) TimingAccumulator(blocksCount) else null

    val totalTime = measureTime {
      repeat(blocksCount) {
        compressor.reset()
        val blockStartNs = System.nanoTime()

        while (true) {
          val txData = txDataList[cursor % txDataList.size]
          cursor++

          val txStartNs = System.nanoTime()
          val result = compressor.appendTransaction(txData.from, txData.rlpForSigning)
          val txElapsedNs = System.nanoTime() - txStartNs

          if (!result.txAppended) {
            break
          }

          perTxTimings?.record(txElapsedNs)
          totalTxs++
        }

        perBlockTimings?.record(System.nanoTime() - blockStartNs)
      }
    }

    return DetailedBenchmarkResult(
      blocks = blocksCount,
      avgTxsPerBlock = totalTxs.toDouble() / blocksCount,
      totalTxs = totalTxs,
      perTxTimingNs = perTxTimings?.toStats() ?: TimingStats(0, 0, 0.0, 0),
      perBlockTimingNs = perBlockTimings?.toStats() ?: TimingStats(0, 0, 0.0, 0),
      totalTime = totalTime,
    )
  }

  private fun printDetailedResult(name: String, result: DetailedBenchmarkResult) {
    val txStats = result.perTxTimingNs
    val blockStats = result.perBlockTimingNs

    println("$name:")
    println("  Blocks: ${result.blocks}, Total txs: ${result.totalTxs}, Avg txs/block: ${"%.1f".format(result.avgTxsPerBlock)}")
    println(
      "  Per-tx (us):   min=%7.2f  p95=%7.2f  avg=%7.2f  max=%7.2f".format(
        txStats.min / 1000.0,
        txStats.p95 / 1000.0,
        txStats.avg / 1000.0,
        txStats.max / 1000.0,
      ),
    )
    println(
      "  Per-block (ms): min=%7.2f  p95=%7.2f  avg=%7.2f  max=%7.2f".format(
        blockStats.min / 1_000_000.0,
        blockStats.p95 / 1_000_000.0,
        blockStats.avg / 1_000_000.0,
        blockStats.max / 1_000_000.0,
      ),
    )
    println("  Total time: ${result.totalTime}")
  }

  private class TimingAccumulator(expectedSamples: Int) {
    private val p95TailSize = expectedSamples - (expectedSamples * 0.95).toInt() + 1
    private val largestTail = PriorityQueue<Long>()
    private var count = 0
    private var min = Long.MAX_VALUE
    private var max = Long.MIN_VALUE
    private var total = 0L

    fun record(value: Long) {
      count++
      if (value < min) min = value
      if (value > max) max = value
      total += value

      if (largestTail.size < p95TailSize) {
        largestTail.add(value)
      } else if (value > largestTail.peek()) {
        largestTail.poll()
        largestTail.add(value)
      }
    }

    fun toStats(): TimingStats {
      if (count == 0) return TimingStats(0, 0, 0.0, 0)
      val p95 = if (largestTail.isEmpty()) 0L else largestTail.peek()
      return TimingStats(min, max, total.toDouble() / count, p95)
    }
  }

  // ── Scenario generators ────────────────────────────────────────────────────

  private fun generateErc20Transfers(count: Int): List<TransactionData> {
    val random = Random(42L)
    return (0 until count).map { i ->
      TxCompressorTestFixtures.generateRandomizedEip1559Erc20Tx(i.toLong(), random).second
    }
  }

  private fun generatePlainTransfers(count: Int): List<TransactionData> {
    val random = Random(43L)
    return (0 until count).map { i ->
      TxCompressorTestFixtures.generateRandomizedEip1559PlainTransferTx(i.toLong(), random).second
    }
  }

  private fun generateCalldataTxs(count: Int, calldataSize: Int): List<TransactionData> {
    val random = Random(44L)
    return (0 until count).map { i ->
      TxCompressorTestFixtures.generateRandomizedEip1559CalldataTx(i.toLong(), calldataSize, random).second
    }
  }

  private fun generateMixedTxs(count: Int): List<TransactionData> {
    val quarter = count / 4
    val erc20 = generateErc20Transfers(quarter)
    val plain = generatePlainTransfers(quarter)
    val calldata500 = generateCalldataTxs(quarter, 500)
    val calldata3k = generateCalldataTxs(quarter, 3 * 1024)

    val mixed = mutableListOf<TransactionData>()
    for (i in 0 until quarter) {
      mixed.add(erc20[i])
      mixed.add(plain[i])
      mixed.add(calldata500[i])
      mixed.add(calldata3k[i])
    }
    return mixed
  }
}
