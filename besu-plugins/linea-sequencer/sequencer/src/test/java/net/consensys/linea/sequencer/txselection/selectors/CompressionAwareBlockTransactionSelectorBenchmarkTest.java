/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.txselection.selectors;

import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.BLOCK_COMPRESSED_SIZE_OVERFLOW;
import static org.hyperledger.besu.plugin.data.TransactionSelectionResult.SELECTED;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.PriorityQueue;
import java.util.Random;
import linea.blob.BlobCompressorVersion;
import linea.blob.GoBackedBlobCompressor;
import net.consensys.linea.utils.CachingTransactionCompressor;
import net.consensys.linea.utils.TransactionCompressor;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SignatureAlgorithm;
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.services.txselection.SelectorsStateManager;
import org.junit.jupiter.api.Test;

/**
 * Benchmarks {@link CompressionAwareBlockTransactionSelector} directly with production-like block
 * building flow:
 *
 * <p>preProcessing -> postProcessing -> onSelected, repeated until first rejection (full block).
 */
class CompressionAwareBlockTransactionSelectorBenchmarkTest {
  private static final int BLOB_SIZE_LIMIT_BYTES = 128 * 1024;
  private static final int HEADER_OVERHEAD_BYTES = 1024;
  private static final int SAMPLES_PER_SCENARIO =
      intProp("linea.bench.samplesPerScenario", 2000);
  private static final int SENDER_POOL_SIZE =
      intProp("linea.bench.senderPoolSize", 200);
  private static final int WARMUP_BLOCKS = intProp("linea.bench.warmupBlocks", 5);
  private static final int MEASURE_BLOCKS = intProp("linea.bench.measureBlocks", 20);
  private static final int PROGRESS_STEP_PERCENT =
      intProp("linea.bench.progressStepPercent", 1);

  private static final long CHAIN_ID = 59144L;
  private static final SignatureAlgorithm SIGNATURE_ALGORITHM =
      SignatureAlgorithmFactory.getInstance();
  private static final List<KeyPair> SENDER_KEYS = buildSenderKeys(SENDER_POOL_SIZE);
  private static final TransactionCompressor TX_COMPRESSOR = new CachingTransactionCompressor();

  @Test
  void benchmarkSelectorDirectlyOnFullBlocks() {
    final ProcessableBlockHeader header = mockHeader();
    final TransactionProcessingResult processingResult = mock(TransactionProcessingResult.class);
    final var sharedBlobCompressor =
        GoBackedBlobCompressor.getInstance(BlobCompressorVersion.V1_2, BLOB_SIZE_LIMIT_BYTES);

    final List<Scenario> scenarios =
        List.of(
            new Scenario("erc20-transfer", buildErc20Transfers(SAMPLES_PER_SCENARIO)),
            new Scenario(
                "calldata-3kb", buildCalldataTransactions(3 * 1024, SAMPLES_PER_SCENARIO)),
            new Scenario(
                "mixed-tx-types",
                buildMixedTransactions(
                    buildPlainTransfers(SAMPLES_PER_SCENARIO),
                    buildErc20Transfers(SAMPLES_PER_SCENARIO),
                    buildCalldataTransactions(3 * 1024, SAMPLES_PER_SCENARIO),
                    buildCalldataTransactions(500, SAMPLES_PER_SCENARIO))));

    System.out.println("CompressionAwareBlockTransactionSelector full-block benchmark");
    System.out.println(
        "blobLimitBytes="
            + BLOB_SIZE_LIMIT_BYTES
            + ", headerOverheadBytes="
            + HEADER_OVERHEAD_BYTES
            + ", warmupBlocks="
            + WARMUP_BLOCKS
            + ", measureBlocks="
            + MEASURE_BLOCKS
            + ", samplesPerScenario="
            + SAMPLES_PER_SCENARIO
            + ", senderPoolSize="
            + SENDER_POOL_SIZE);
    System.out.println("-------------------------------------------------------------------------");

    for (final Scenario scenario : scenarios) {
      final List<TestTransactionEvaluationContext> contexts = wrapTransactions(header, scenario.transactions());
      System.out.println("Warmup scenario: " + scenario.name());
      runBlocks(
          scenario.name() + " warmup",
          contexts,
          processingResult,
          WARMUP_BLOCKS,
          false,
          sharedBlobCompressor);

      System.out.println("Running scenario: " + scenario.name());
      final BlockBenchmarkResult result =
          runBlocks(
              scenario.name(),
              contexts,
              processingResult,
              MEASURE_BLOCKS,
              true,
              sharedBlobCompressor);

      printResult(scenario, result);
      System.out.println("-------------------------------------------------------------------------");
    }
  }

  private BlockBenchmarkResult runBlocks(
      final String labelPrefix,
      final List<TestTransactionEvaluationContext> contexts,
      final TransactionProcessingResult processingResult,
      final int blocksToRun,
      final boolean measure,
      final GoBackedBlobCompressor sharedBlobCompressor) {
    final TimingAccumulator blockTime = new TimingAccumulator(blocksToRun);
    final TimingAccumulator preProcessingTime = new TimingAccumulator(blocksToRun * 32);
    final NumericAccumulator cumulativePreProcessingTimePerBlock = new NumericAccumulator(blocksToRun);
    final NumericAccumulator selectedTxPerBlock = new NumericAccumulator(blocksToRun);
    long blockCutsByCompressionOverflow = 0L;
    long nonOverflowRejections = 0L;
    final Map<String, Long> rejectionReasons = new HashMap<>();
    int cursor = 0;
    int nextProgressPercent = PROGRESS_STEP_PERCENT;
    final long startedAtNs = System.nanoTime();

    for (int blockIdx = 0; blockIdx < blocksToRun; blockIdx++) {
      final SelectorsStateManager stateManager = new SelectorsStateManager();
      final var selector =
          new CompressionAwareBlockTransactionSelector(
              stateManager,
              BLOB_SIZE_LIMIT_BYTES,
              HEADER_OVERHEAD_BYTES,
              TX_COMPRESSOR,
              sharedBlobCompressor);
      stateManager.blockSelectionStarted();

      int selectedInBlock = 0;
      long cumulativePreProcessingNsInBlock = 0L;
      final long blockStartNs = System.nanoTime();
      while (true) {
        final TestTransactionEvaluationContext context = contexts.get(cursor % contexts.size());
        cursor++;

        final long preStartNs = System.nanoTime();
        final var preResult = selector.evaluateTransactionPreProcessing(context);
        if (measure) {
          final long preElapsedNs = System.nanoTime() - preStartNs;
          preProcessingTime.record(preElapsedNs);
          cumulativePreProcessingNsInBlock += preElapsedNs;
        }

        if (preResult == SELECTED) {
          selector.evaluateTransactionPostProcessing(context, processingResult);
          selector.onTransactionSelected(context, processingResult);
          selectedInBlock++;
          continue;
        }

        selector.onTransactionNotSelected(context, preResult);
        rejectionReasons.merge(preResult.toString(), 1L, Long::sum);
        if (preResult == BLOCK_COMPRESSED_SIZE_OVERFLOW) {
          blockCutsByCompressionOverflow++;
          break;
        }
        nonOverflowRejections++;
        // Keep filling current block; only overflow is a block cut-off condition.
      }

      if (measure) {
        blockTime.record(System.nanoTime() - blockStartNs);
        cumulativePreProcessingTimePerBlock.record(cumulativePreProcessingNsInBlock);
        selectedTxPerBlock.record(selectedInBlock);
      }

      nextProgressPercent =
          maybeReportProgress(
              labelPrefix + (measure ? " measure" : " warmup"),
              blockIdx + 1,
              blocksToRun,
              startedAtNs,
              nextProgressPercent);
    }

    return new BlockBenchmarkResult(
        blockTime.toTimingStats(),
        preProcessingTime.toTimingStats(),
        cumulativePreProcessingTimePerBlock.toTimingStats(),
        selectedTxPerBlock.toTimingStats(),
        blockCutsByCompressionOverflow,
        nonOverflowRejections,
        rejectionReasons);
  }

  private static int maybeReportProgress(
      final String label,
      final int completed,
      final int total,
      final long startedAtNs,
      final int nextProgressPercent) {
    final int percent = (completed * 100) / total;
    final boolean shouldReport = percent >= nextProgressPercent || completed == total;
    if (!shouldReport) {
      return nextProgressPercent;
    }

    final long elapsedNs = System.nanoTime() - startedAtNs;
    final double elapsedSec = elapsedNs / 1_000_000_000.0;
    final double estimatedTotalSec = elapsedSec * total / Math.max(1, completed);
    final double etaSec = Math.max(0.0, estimatedTotalSec - elapsedSec);
    System.out.printf(
        "  [%s] %3d%% (%d/%d) elapsed=%.1fs eta=%.1fs%n",
        label, percent, completed, total, elapsedSec, etaSec);
    return nextProgressPercent + PROGRESS_STEP_PERCENT;
  }

  private static int intProp(final String key, final int defaultValue) {
    final String raw = System.getProperty(key);
    if (raw == null || raw.isBlank()) {
      return defaultValue;
    }
    try {
      final int parsed = Integer.parseInt(raw);
      return parsed > 0 ? parsed : defaultValue;
    } catch (NumberFormatException ignored) {
      return defaultValue;
    }
  }

  private static void printResult(final Scenario scenario, final BlockBenchmarkResult result) {
    System.out.printf(
        "%-15s txEncoded=[%6d..%6d] | pre(us): min=%7.2f p95=%7.2f avg=%7.2f max=%7.2f | cumulativePre/block(ms): min=%7.2f p95=%7.2f avg=%7.2f max=%7.2f | block(ms): min=%7.2f p95=%7.2f avg=%7.2f max=%7.2f | selectedTx/block: min=%6.1f p95=%6.1f avg=%6.1f max=%6.1f | overflows=%d%n",
        scenario.name(),
        scenario.minEncodedSize(),
        scenario.maxEncodedSize(),
        nanosToMicros(result.preProcessingTime().min()),
        nanosToMicros(result.preProcessingTime().p95()),
        nanosToMicros(result.preProcessingTime().avg()),
        nanosToMicros(result.preProcessingTime().max()),
        nanosToMillis(result.cumulativePreProcessingTimePerBlock().min()),
        nanosToMillis(result.cumulativePreProcessingTimePerBlock().p95()),
        nanosToMillis(result.cumulativePreProcessingTimePerBlock().avg()),
        nanosToMillis(result.cumulativePreProcessingTimePerBlock().max()),
        nanosToMillis(result.blockTime().min()),
        nanosToMillis(result.blockTime().p95()),
        nanosToMillis(result.blockTime().avg()),
        nanosToMillis(result.blockTime().max()),
        result.selectedTxPerBlock().min(),
        result.selectedTxPerBlock().p95(),
        result.selectedTxPerBlock().avg(),
        result.selectedTxPerBlock().max(),
        result.blockCutsByCompressionOverflow());
    if (result.nonOverflowRejections() > 0) {
      System.out.println(
          "  non-overflow rejections="
              + result.nonOverflowRejections()
              + " by reason="
              + result.rejectionReasons());
    }
  }

  private static double nanosToMicros(final double nanos) {
    return nanos / 1_000.0;
  }

  private static double nanosToMillis(final double nanos) {
    return nanos / 1_000_000.0;
  }

  private static List<TestTransactionEvaluationContext> wrapTransactions(
      final ProcessableBlockHeader header, final List<Transaction> transactions) {
    final List<TestTransactionEvaluationContext> contexts = new ArrayList<>(transactions.size());
    for (final Transaction tx : transactions) {
      final PendingTransaction pendingTx = mock(PendingTransaction.class);
      when(pendingTx.getTransaction()).thenReturn(tx);
      contexts.add(new TestTransactionEvaluationContext(header, pendingTx));
    }
    return contexts;
  }

  private static ProcessableBlockHeader mockHeader() {
    final ProcessableBlockHeader header = mock(ProcessableBlockHeader.class);
    when(header.getNumber()).thenReturn(1L);
    when(header.getTimestamp()).thenReturn(1_700_000_000L);
    when(header.getCoinbase()).thenReturn(Address.ZERO);
    when(header.getGasLimit()).thenReturn(30_000_000L);
    when(header.getParentHash()).thenReturn(Hash.wrap(Bytes32.ZERO));
    return header;
  }

  private static List<Transaction> buildPlainTransfers(final int count) {
    final Random random = new Random(0xA11CE001L);
    final List<Transaction> txs = new ArrayList<>(count);
    for (int i = 0; i < count; i++) {
      final long nonce = random.nextLong(10_000_000L);
      final Address recipient = randomAddress(random);
      final Wei value = Wei.of(random.nextLong(100_000_000_000L));
      final Eip1559Fees fees = randomEip1559FeesInRange(random);
      final long gasLimit = 21_000L + random.nextInt(4_000);
      txs.add(
          buildTransaction(
              nonce,
              recipient,
              Bytes.EMPTY,
              gasLimit,
              value,
              fees.maxPriorityFeePerGas(),
              fees.maxFeePerGas(),
              SENDER_KEYS.get(i % SENDER_KEYS.size())));
    }
    return txs;
  }

  private static List<Transaction> buildErc20Transfers(final int count) {
    final Random random = new Random(0xEFC2001L);
    final List<Transaction> txs = new ArrayList<>(count);
    for (int i = 0; i < count; i++) {
      final long nonce = random.nextLong(10_000_000L);
      final Address tokenContract = randomAddress(random);
      final Eip1559Fees fees = randomEip1559FeesInRange(random);
      final long gasLimit = 90_000L + random.nextInt(40_000);
      txs.add(
          buildTransaction(
              nonce,
              tokenContract,
              erc20TransferPayload(random),
              gasLimit,
              Wei.ZERO,
              fees.maxPriorityFeePerGas(),
              fees.maxFeePerGas(),
              SENDER_KEYS.get(i % SENDER_KEYS.size())));
    }
    return txs;
  }

  private static List<Transaction> buildCalldataTransactions(
      final int calldataSizeBytes, final int count) {
    final Random random = new Random(0xCA11DA7AL + calldataSizeBytes);
    final List<Transaction> txs = new ArrayList<>(count);
    for (int i = 0; i < count; i++) {
      final long nonce = random.nextLong(10_000_000L);
      final Address recipient = randomAddress(random);
      final Wei value = Wei.of(random.nextInt(10_000));
      final Eip1559Fees fees = randomEip1559FeesInRange(random);
      final long gasLimit = 4_000_000L + random.nextInt(3_000_000);
      txs.add(
          buildTransaction(
              nonce,
              recipient,
              randomPayload(calldataSizeBytes, 0xBEEFL + nonce),
              gasLimit,
              value,
              fees.maxPriorityFeePerGas(),
              fees.maxFeePerGas(),
              SENDER_KEYS.get(i % SENDER_KEYS.size())));
    }
    return txs;
  }

  private static List<Transaction> buildMixedTransactions(
      final List<Transaction> plainTransfers,
      final List<Transaction> erc20Transfers,
      final List<Transaction> calldata3kb,
      final List<Transaction> calldata500b) {
    final int mixedCount =
        Math.min(
            Math.min(plainTransfers.size(), erc20Transfers.size()),
            Math.min(calldata3kb.size(), calldata500b.size()));
    final List<Transaction> mixed = new ArrayList<>(mixedCount * 4);
    for (int i = 0; i < mixedCount; i++) {
      mixed.add(plainTransfers.get(i));
      mixed.add(erc20Transfers.get(i));
      mixed.add(calldata500b.get(i));
      mixed.add(calldata3kb.get(i));
    }
    return mixed;
  }

  private static Transaction buildTransaction(
      final long nonce,
      final Address to,
      final Bytes payload,
      final long gasLimit,
      final Wei value,
      final Wei maxPriorityFeePerGas,
      final Wei maxFeePerGas,
      final KeyPair signerKey) {
    return Transaction.builder()
        .type(TransactionType.EIP1559)
        .chainId(BigInteger.valueOf(CHAIN_ID))
        .nonce(nonce)
        .gasLimit(gasLimit)
        .maxPriorityFeePerGas(maxPriorityFeePerGas)
        .maxFeePerGas(maxFeePerGas)
        .to(to)
        .value(value)
        .payload(payload)
        .signAndBuild(signerKey);
  }

  private static Eip1559Fees randomEip1559FeesInRange(final Random random) {
    final long maxPriorityGwei = random.nextInt(1001);
    final long maxFeeGwei = maxPriorityGwei + random.nextInt((int) (1001 - maxPriorityGwei));
    return new Eip1559Fees(gweiToWei(maxPriorityGwei), gweiToWei(maxFeeGwei));
  }

  private static Wei gweiToWei(final long gwei) {
    return Wei.of(gwei * 1_000_000_000L);
  }

  private static List<KeyPair> buildSenderKeys(final int senderPoolSize) {
    final List<KeyPair> keys = new ArrayList<>(senderPoolSize);
    final Random random = new Random(0x5EEDC0DEL);
    for (int i = 0; i < senderPoolSize; i++) {
      BigInteger privateKey;
      do {
        privateKey = new BigInteger(256, random);
      } while (privateKey.signum() <= 0);
      keys.add(SIGNATURE_ALGORITHM.createKeyPair(SIGNATURE_ALGORITHM.createPrivateKey(privateKey)));
    }
    return keys;
  }

  private static Address randomAddress(final Random random) {
    final byte[] raw = new byte[20];
    random.nextBytes(raw);
    raw[0] = (byte) (raw[0] | 0x10);
    return Address.wrap(Bytes.wrap(raw));
  }

  private static Bytes erc20TransferPayload(final Random random) {
    final String recipientHex = randomAddress(random).toHexString().substring(2);
    final long amount = random.nextLong(1, Long.MAX_VALUE);
    final String amountHex = String.format("%064x", amount);
    return Bytes.fromHexString("0xa9059cbb" + "000000000000000000000000" + recipientHex + amountHex);
  }

  private static Bytes randomPayload(final int size, final long seed) {
    final byte[] payload = new byte[size];
    new Random(seed).nextBytes(payload);
    return Bytes.wrap(payload);
  }

  private record Scenario(String name, List<Transaction> transactions) {
    int minEncodedSize() {
      int min = Integer.MAX_VALUE;
      for (final Transaction tx : transactions) {
        min = Math.min(min, tx.encoded().size());
      }
      return min;
    }

    int maxEncodedSize() {
      int max = Integer.MIN_VALUE;
      for (final Transaction tx : transactions) {
        max = Math.max(max, tx.encoded().size());
      }
      return max;
    }
  }

  private record Eip1559Fees(Wei maxPriorityFeePerGas, Wei maxFeePerGas) {}

  private record BlockBenchmarkResult(
      NumericStats blockTime,
      NumericStats preProcessingTime,
      NumericStats cumulativePreProcessingTimePerBlock,
      NumericStats selectedTxPerBlock,
      long blockCutsByCompressionOverflow,
      long nonOverflowRejections,
      Map<String, Long> rejectionReasons) {}

  private record NumericStats(double min, double max, double avg, double p95) {}

  private static class TimingAccumulator {
    private final int p95TailSize;
    private final PriorityQueue<Double> largestTail = new PriorityQueue<>();
    private int count = 0;
    private double min = Double.POSITIVE_INFINITY;
    private double max = Double.NEGATIVE_INFINITY;
    private double total = 0.0;

    TimingAccumulator(final int expectedSamples) {
      this.p95TailSize = expectedSamples - (int) Math.ceil(expectedSamples * 0.95d) + 1;
    }

    void record(final long value) {
      record((double) value);
    }

    void record(final double value) {
      count++;
      min = Math.min(min, value);
      max = Math.max(max, value);
      total += value;
      addToTail(value);
    }

    NumericStats toTimingStats() {
      return toNumericStats();
    }

    private NumericStats toNumericStats() {
      if (count == 0) {
        return new NumericStats(0.0, 0.0, 0.0, 0.0);
      }
      final double p95 = largestTail.isEmpty() ? 0.0 : largestTail.peek();
      return new NumericStats(min, max, total / count, p95);
    }

    private void addToTail(final double value) {
      if (largestTail.size() < p95TailSize) {
        largestTail.add(value);
        return;
      }
      if (value > largestTail.peek()) {
        largestTail.poll();
        largestTail.add(value);
      }
    }
  }

  private static class NumericAccumulator {
    private final TimingAccumulator delegate;

    NumericAccumulator(final int expectedSamples) {
      this.delegate = new TimingAccumulator(expectedSamples);
    }

    void record(final int value) {
      delegate.record((double) value);
    }

    void record(final long value) {
      delegate.record((double) value);
    }

    NumericStats toTimingStats() {
      return delegate.toTimingStats();
    }
  }
}
