/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.txselection.selectors;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
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
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Block;
import org.hyperledger.besu.ethereum.core.BlockBody;
import org.hyperledger.besu.ethereum.core.BlockHeader;
import org.hyperledger.besu.ethereum.core.BlockHeaderBuilder;
import org.hyperledger.besu.ethereum.core.Difficulty;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions;
import org.hyperledger.besu.evm.log.LogsBloomFilter;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;

/**
 * Micro-benchmark for fast and slow path code blocks of {@link
 * CompressionAwareBlockTransactionSelector} executed directly (without selector orchestration).
 *
 * <p>Run manually with:
 *
 * <pre>
 * ./gradlew :besu-plugins:linea-sequencer:sequencer:test \
 *   --tests net.consensys.linea.sequencer.txselection.selectors.CompressionAwareBlockTransactionSelectorBenchmarkTest
 * </pre>
 */
class CompressionAwareBlockTransactionSelectorBenchmarkTest {

  private static final long PLACEHOLDER_SEED = 0xDEADBEEFL;
  private static final int BLOB_SIZE_LIMIT_BYTES = 127 * 1024;
  private static final int HEADER_OVERHEAD_BYTES = 1024;
  private static final int FAST_PATH_EFFECTIVE_LIMIT_BYTES =
      BLOB_SIZE_LIMIT_BYTES - HEADER_OVERHEAD_BYTES;
  private static final int SAMPLES_PER_SCENARIO = 1000;

  private static final int WARMUP_ITERATIONS = 10000;
  private static final int MEASURE_ITERATIONS = 10000;

  private static final long CHAIN_ID = 59144L;
  private static final Wei GAS_PRICE = Wei.of(1_000_000_000L);
  private static final Address RECIPIENT =
      Address.fromHexString("0x000000000000000000000000000000000000dead");
  private static final Address ERC20_CONTRACT =
      Address.fromHexString("0x000000000000000000000000000000000000c0de");

  private static final SignatureAlgorithm SIGNATURE_ALGORITHM =
      SignatureAlgorithmFactory.getInstance();
  private static final BigInteger PRIVATE_KEY =
      new BigInteger("8f2a55949038a9610f50fb23b5883af3b4ecb3c3bb792cbcefbd1542c692be63", 16);
  private static final KeyPair KEY_PAIR =
      SIGNATURE_ALGORITHM.createKeyPair(SIGNATURE_ALGORITHM.createPrivateKey(PRIVATE_KEY));
  private static final TransactionCompressor TX_COMPRESSOR = new CachingTransactionCompressor();

  @Disabled("Used for manual assessment")
  void benchmarkFastAndSlowPathCodeDirectly() {
    final GoBackedBlobCompressor blobCompressor =
        GoBackedBlobCompressor.getInstance(
            BlobCompressorVersion.V1_2,
            BLOB_SIZE_LIMIT_BYTES);

    final MockPendingHeader pendingHeader = new MockPendingHeader(1L, 1_700_000_000L, 30_000_000L);

    final List<Scenario> scenarios =
        List.of(
            new Scenario("plain-transfer", buildPlainTransfers(1_000L, SAMPLES_PER_SCENARIO)),
            new Scenario("erc20-transfer", buildErc20Transfers(100_000L, SAMPLES_PER_SCENARIO)),
            new Scenario(
                "calldata-3kb", buildCalldataTransactions(200_000L, 3 * 1024, SAMPLES_PER_SCENARIO)),
            new Scenario(
                "calldata-500b", buildCalldataTransactions(300_000L, 500, SAMPLES_PER_SCENARIO)));

    System.out.println("CompressionAwareBlockTransactionSelector direct-path benchmark");
    System.out.println("warmupIterations=" + WARMUP_ITERATIONS + ", measureIterations=" + MEASURE_ITERATIONS);
    System.out.println(
        "blobLimitBytes="
            + BLOB_SIZE_LIMIT_BYTES
            + ", headerOverheadBytes="
            + HEADER_OVERHEAD_BYTES
            + ", fastPathEffectiveLimitBytes="
            + FAST_PATH_EFFECTIVE_LIMIT_BYTES
            + ", samplesPerScenario="
            + SAMPLES_PER_SCENARIO);
    System.out.println("-------------------------------------------------------------------------");

    warmupBothPaths(blobCompressor, pendingHeader, scenarios);

    for (final Scenario scenario : scenarios) {
      final FastPathResult fastResult =
          benchmarkFastPathOnly(scenario.transactions(), FAST_PATH_EFFECTIVE_LIMIT_BYTES);
      final SlowPathResult slowResult =
          benchmarkSlowPathOnly(
              blobCompressor, pendingHeader, Collections.emptyList(), scenario.transactions());

      printResult(
          scenario.name(), scenario.minEncodedSize(), scenario.maxEncodedSize(), fastResult, slowResult);
      System.out.println("-------------------------------------------------------------------------");
    }
  }

  /**
   * Direct fast path logic equivalent:
   *
   * <pre>
   * newCumulative = cumulative + txCompressedSize
   * return newCumulative <= fastPathLimit
   * </pre>
   */
  private static FastPathResult benchmarkFastPathOnly(
      final List<Transaction> candidates, final long fastPathLimit) {
    final long cumulativeRawCompressedSize = 0L;
    long selectedCount = 0;
    final long startNs = System.nanoTime();
    for (int i = 0; i < MEASURE_ITERATIONS; i++) {
      final Transaction candidate = candidates.get(i % candidates.size());
      final int txCompressedSize = TX_COMPRESSOR.getCompressedSize(candidate);
      if (fastPathDecision(cumulativeRawCompressedSize, txCompressedSize, fastPathLimit)) {
        selectedCount++;
      }
    }
    final long elapsedNs = System.nanoTime() - startNs;

    return new FastPathResult(elapsedNs / (double) MEASURE_ITERATIONS, selectedCount);
  }

  /**
   * Direct slow path logic equivalent:
   *
   * <pre>
   * tentative = new ArrayList<>(selectedTransactions)
   * tentative.add(candidate)
   * blockRlp = buildBlockRlp(header, tentative)
   * blobCompressor.reset()
   * return blobCompressor.canAppendBlock(blockRlp)
   * </pre>
   */
  private static SlowPathResult benchmarkSlowPathOnly(
      final GoBackedBlobCompressor blobCompressor,
      final MockPendingHeader pendingHeader,
      final List<Transaction> selectedTransactions,
      final List<Transaction> candidates) {
    long fitsCount = 0;
    double avgRlpBytes = 0.0;
    final long startNs = System.nanoTime();
    for (int i = 0; i < MEASURE_ITERATIONS; i++) {
      final Transaction candidate = candidates.get(i % candidates.size());
      final SlowPathDecisionResult result =
          slowPathDecision(blobCompressor, pendingHeader, selectedTransactions, candidate);
      if (result.fits()) {
        fitsCount++;
      }
      avgRlpBytes += result.rlpSize();
    }
    final long elapsedNs = System.nanoTime() - startNs;

    return new SlowPathResult(
        elapsedNs / (double) MEASURE_ITERATIONS, fitsCount, avgRlpBytes / MEASURE_ITERATIONS);
  }

  /**
   * Global warmup that exercises both paths before any timed measurements start. This reduces JVM
   * and native compressor startup bias that can skew the first measured scenario.
   */
  private static void warmupBothPaths(
      final GoBackedBlobCompressor blobCompressor,
      final MockPendingHeader pendingHeader,
      final List<Scenario> scenarios) {
    final long cumulativeRawCompressedSize = 0L;
    final long fastPathLimit = FAST_PATH_EFFECTIVE_LIMIT_BYTES;

    for (int i = 0; i < WARMUP_ITERATIONS; i++) {
      for (final Scenario scenario : scenarios) {
        final Transaction candidate =
            scenario.transactions().get(i % scenario.transactions().size());
        final int txCompressedSize = TX_COMPRESSOR.getCompressedSize(candidate);
        fastPathDecision(cumulativeRawCompressedSize, txCompressedSize, fastPathLimit);
        slowPathDecision(blobCompressor, pendingHeader, Collections.emptyList(), candidate);
      }
    }
  }

  private static boolean fastPathDecision(
      final long cumulativeRawCompressedSize, final int txCompressedSize, final long fastPathLimit) {
    final long newCumulative = cumulativeRawCompressedSize + txCompressedSize;
    return newCumulative <= fastPathLimit;
  }

  private static SlowPathDecisionResult slowPathDecision(
      final GoBackedBlobCompressor blobCompressor,
      final MockPendingHeader pendingHeader,
      final List<Transaction> selectedTransactions,
      final Transaction candidate) {
    final List<Transaction> tentativeTxs = new ArrayList<>(selectedTransactions);
    tentativeTxs.add(candidate);
    final byte[] blockRlp = buildBlockRlp(pendingHeader, tentativeTxs);

    blobCompressor.reset();
    final boolean fits = blobCompressor.canAppendBlock(blockRlp);
    return new SlowPathDecisionResult(fits, blockRlp.length);
  }

  private static void printResult(
      final String scenarioName,
      final int minEncodedSize,
      final int maxEncodedSize,
      final FastPathResult fastResult,
      final SlowPathResult slowResult) {
    final double fastMicros = fastResult.avgNsPerOperation() / 1_000.0;
    final double slowMicros = slowResult.avgNsPerOperation() / 1_000.0;

    System.out.printf(
        "%-15s txEncoded=[%6d..%6d] | fast=%9.2f us/op selected=%5d/%d | slow=%10.2f us/op fits=%5d/%d avgRlpBytes=%10.0f%n",
        scenarioName,
        minEncodedSize,
        maxEncodedSize,
        fastMicros,
        fastResult.selectedCount(),
        MEASURE_ITERATIONS,
        slowMicros,
        slowResult.fitsCount(),
        MEASURE_ITERATIONS,
        slowResult.avgRlpBytes());
  }

  private static byte[] buildBlockRlp(
      final MockPendingHeader pendingHeader, final List<Transaction> transactions) {
    final Random random = new Random(PLACEHOLDER_SEED);

    final BlockHeader header =
        BlockHeaderBuilder.create()
            .parentHash(randomHash(random))
            .ommersHash(randomHash(random))
            .coinbase(Address.wrap(Bytes.wrap(randomBytes(random, 20))))
            .stateRoot(randomHash(random))
            .transactionsRoot(randomHash(random))
            .receiptsRoot(randomHash(random))
            .logsBloom(
                LogsBloomFilter.fromHexString(Bytes.wrap(randomBytes(random, 256)).toHexString()))
            .difficulty(Difficulty.of(random.nextLong(Long.MAX_VALUE)))
            .number(pendingHeader.number())
            .gasLimit(pendingHeader.gasLimit())
            .gasUsed(random.nextLong(Long.MAX_VALUE))
            .timestamp(pendingHeader.timestamp())
            .extraData(Bytes.wrap(randomBytes(random, 32)))
            .mixHash(randomHash(random))
            .nonce(random.nextLong())
            .baseFee(Wei.of(random.nextLong(Long.MAX_VALUE)))
            .blockHeaderFunctions(new MainnetBlockHeaderFunctions())
            .buildBlockHeader();

    final BlockBody body = new BlockBody(transactions, Collections.emptyList());
    final Block block = new Block(header, body);
    return block.toRlp().toArray();
  }

  private static Hash randomHash(final Random random) {
    return Hash.wrap(Bytes32.wrap(randomBytes(random, 32)));
  }

  private static byte[] randomBytes(final Random random, final int length) {
    final byte[] bytes = new byte[length];
    random.nextBytes(bytes);
    return bytes;
  }

  private static Transaction createPlainTransfer(final long nonce) {
    return buildTransaction(nonce, RECIPIENT, Bytes.EMPTY, 21_000L);
  }

  private static List<Transaction> buildPlainTransfers(final long nonceStart, final int count) {
    final Random random = new Random(0xA11CE001L);
    final List<Transaction> txs = new ArrayList<>(count);
    for (int i = 0; i < count; i++) {
      final long nonce = nonceStart + i;
      final Address recipient = randomAddress(random);
      final Wei value = Wei.of(1L + random.nextInt(100_000));
      final Wei gasPrice = Wei.of(GAS_PRICE.toLong() + random.nextInt(1_000));
      final long gasLimit = 21_000L + random.nextInt(2_000);
      txs.add(buildTransaction(nonce, recipient, Bytes.EMPTY, gasLimit, value, gasPrice));
    }
    return txs;
  }

  private static List<Transaction> buildErc20Transfers(final long nonceStart, final int count) {
    final Random random = new Random(0xEFC2001L);
    final List<Transaction> txs = new ArrayList<>(count);
    for (int i = 0; i < count; i++) {
      final long nonce = nonceStart + i;
      final Wei gasPrice = Wei.of(GAS_PRICE.toLong() + random.nextInt(2_000));
      final long gasLimit = 90_000L + random.nextInt(30_000);
      txs.add(
          buildTransaction(
              nonce, ERC20_CONTRACT, erc20TransferPayload(i), gasLimit, Wei.ZERO, gasPrice));
    }
    return txs;
  }

  private static List<Transaction> buildCalldataTransactions(
      final long nonceStart, final int calldataSizeBytes, final int count) {
    final Random random = new Random(0xCA11DA7AL + calldataSizeBytes);
    final List<Transaction> txs = new ArrayList<>(count);
    for (int i = 0; i < count; i++) {
      final long nonce = nonceStart + i;
      final Address recipient = randomAddress(random);
      final Wei value = Wei.of(random.nextInt(10_000));
      final Wei gasPrice = Wei.of(GAS_PRICE.toLong() + random.nextInt(5_000));
      final long gasLimit = 28_000_000L + random.nextInt(2_000_000);
      txs.add(
          buildTransaction(
              nonce,
              recipient,
              randomPayload(calldataSizeBytes, 0xBEEFL + nonce),
              gasLimit,
              value,
              gasPrice));
    }
    return txs;
  }

  private static Transaction buildTransaction(
      final long nonce, final Address to, final Bytes payload, final long gasLimit) {
    return buildTransaction(nonce, to, payload, gasLimit, Wei.ZERO, GAS_PRICE);
  }

  private static Transaction buildTransaction(
      final long nonce,
      final Address to,
      final Bytes payload,
      final long gasLimit,
      final Wei value,
      final Wei gasPrice) {
    return Transaction.builder()
        .type(TransactionType.FRONTIER)
        .chainId(BigInteger.valueOf(CHAIN_ID))
        .nonce(nonce)
        .gasLimit(gasLimit)
        .gasPrice(gasPrice)
        .to(to)
        .value(value)
        .payload(payload)
        .signAndBuild(KEY_PAIR);
  }

  private static Address randomAddress(final Random random) {
    final byte[] raw = new byte[20];
    random.nextBytes(raw);
    // Keep it out of precompile range by setting a non-zero high byte.
    raw[0] = (byte) (raw[0] | 0x10);
    return Address.wrap(Bytes.wrap(raw));
  }

  private static Bytes erc20TransferPayload(final int i) {
    final String recipientHex = String.format("%040x", 0xBEEFL + i);
    final String amountHex = String.format("%064x", 1_000_000L + i);
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

  private record MockPendingHeader(long number, long timestamp, long gasLimit) {}

  private record FastPathResult(double avgNsPerOperation, long selectedCount) {}

  private record SlowPathDecisionResult(boolean fits, int rlpSize) {}

  private record SlowPathResult(double avgNsPerOperation, long fitsCount, double avgRlpBytes) {}
}
