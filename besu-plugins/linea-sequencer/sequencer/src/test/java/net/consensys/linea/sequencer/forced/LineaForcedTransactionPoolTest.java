/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.forced;

import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.DENIED_LOG_TOPIC;
import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import java.util.ArrayList;
import java.util.List;
import net.consensys.linea.utils.TestTransactionFactory;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.BlockTransactionSelectionService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class LineaForcedTransactionPoolTest {

  private static final long TEST_TIMESTAMP = 1_700_000_000L; // Fixed timestamp for tests

  private LineaForcedTransactionPool pool;
  private TestTransactionFactory txFactory;

  @BeforeEach
  void setUp() {
    pool = new LineaForcedTransactionPool(100, null);
    txFactory = new TestTransactionFactory();
  }

  @Test
  void addForcedTransactions_addsToQueueAndReturnsHashes() {
    final List<ForcedTransaction> ftxs = createForcedTransactions(3);

    final List<Hash> result = pool.addForcedTransactions(ftxs);

    assertThat(result).hasSize(3);
    assertThat(result)
        .containsExactly(ftxs.get(0).txHash(), ftxs.get(1).txHash(), ftxs.get(2).txHash());
    assertThat(pool.pendingCount()).isEqualTo(3);
  }

  @Test
  void processForBlock_includesSelectedTransaction() {
    final ForcedTransaction ftx = createForcedTransaction();
    pool.addForcedTransactions(List.of(ftx));

    pool.processForBlock(100L, TEST_TIMESTAMP, alwaysSelect());

    assertThat(pool.pendingCount()).isZero();
    assertInclusionStatus(ftx.txHash(), 100L, ForcedTransactionInclusionResult.INCLUDED);
  }

  @Test
  void processForBlock_includesMultipleTransactionsInOrder() {
    final List<ForcedTransaction> ftxs = createForcedTransactions(3);
    pool.addForcedTransactions(ftxs);

    pool.processForBlock(50L, TEST_TIMESTAMP, alwaysSelect());

    assertThat(pool.pendingCount()).isZero();
    for (final ForcedTransaction ftx : ftxs) {
      assertInclusionStatus(ftx.txHash(), 50L, ForcedTransactionInclusionResult.INCLUDED);
    }
  }

  @Test
  void processForBlock_rejectsTransactionAtIndex0WithFinalStatus() {
    final ForcedTransaction ftx = createForcedTransaction();
    pool.addForcedTransactions(List.of(ftx));

    pool.processForBlock(100L, TEST_TIMESTAMP, alwaysReject("NONCE_TOO_LOW"));

    assertThat(pool.pendingCount()).isZero();
    assertInclusionStatus(ftx.txHash(), 100L, ForcedTransactionInclusionResult.BAD_NONCE);
  }

  @Test
  void processForBlock_retriesTransactionFailedAtIndexGreaterThan0() {
    final List<ForcedTransaction> ftxs = createForcedTransactions(2);
    pool.addForcedTransactions(ftxs);

    // Block 1: first tx succeeds, second tx fails at index 1
    pool.processForBlock(100L, TEST_TIMESTAMP, selectThenReject("NONCE"));

    // First tx included, second tx still pending (will retry)
    assertThat(pool.pendingCount()).isEqualTo(1);
    assertInclusionStatus(ftxs.get(0).txHash(), 100L, ForcedTransactionInclusionResult.INCLUDED);
    assertThat(pool.getInclusionStatus(ftxs.get(1).txHash())).isEmpty();

    // Block 2: second tx fails at index 0 - final rejection
    pool.processForBlock(101L, TEST_TIMESTAMP, alwaysReject("NONCE"));

    assertThat(pool.pendingCount()).isZero();
    assertInclusionStatus(ftxs.get(1).txHash(), 101L, ForcedTransactionInclusionResult.BAD_NONCE);
  }

  @Test
  void processForBlock_stopsOnFirstFailureToPreventMultipleInvalidityProofs() {
    final List<ForcedTransaction> ftxs = createForcedTransactions(3);
    pool.addForcedTransactions(ftxs);

    // Block 1: first tx succeeds, second tx fails - third tx should NOT be evaluated
    pool.processForBlock(100L, TEST_TIMESTAMP, selectThenReject("BALANCE"));

    assertThat(pool.pendingCount()).isEqualTo(2); // tx2 and tx3 still pending
    assertInclusionStatus(ftxs.get(0).txHash(), 100L, ForcedTransactionInclusionResult.INCLUDED);
    assertThat(pool.getInclusionStatus(ftxs.get(1).txHash())).isEmpty(); // Will retry
    assertThat(pool.getInclusionStatus(ftxs.get(2).txHash())).isEmpty(); // Not tried yet

    pool.processForBlock(101L, TEST_TIMESTAMP, alwaysReject("NONCE"));

    assertThat(pool.pendingCount()).isEqualTo(1); // tx3 still pending
    assertInclusionStatus(ftxs.get(1).txHash(), 101L, ForcedTransactionInclusionResult.BAD_NONCE);
    assertThat(pool.getInclusionStatus(ftxs.get(2).txHash())).isEmpty(); // Not tried yet

    pool.processForBlock(102L, TEST_TIMESTAMP, alwaysReject("PRECOMPILE"));

    assertThat(pool.pendingCount()).isEqualTo(0);
    assertInclusionStatus(
        ftxs.get(2).txHash(), 102L, ForcedTransactionInclusionResult.BAD_PRECOMPILE);
  }

  @Test
  void processForBlock_twoConsecutiveFailingTransactionsGetStatusInDifferentBlocks() {
    final List<ForcedTransaction> ftxs = createForcedTransactions(2);
    pool.addForcedTransactions(ftxs);

    // Block 1: first tx fails at index 0 - gets final status
    pool.processForBlock(100L, TEST_TIMESTAMP, alwaysReject("NONCE"));

    assertThat(pool.pendingCount()).isEqualTo(1);
    assertInclusionStatus(ftxs.get(0).txHash(), 100L, ForcedTransactionInclusionResult.BAD_NONCE);
    assertThat(pool.getInclusionStatus(ftxs.get(1).txHash())).isEmpty();

    // Block 2: second tx fails at index 0 - gets final status in different block
    pool.processForBlock(101L, TEST_TIMESTAMP, alwaysReject("BALANCE"));

    assertThat(pool.pendingCount()).isZero();
    assertInclusionStatus(ftxs.get(1).txHash(), 101L, ForcedTransactionInclusionResult.BAD_BALANCE);
  }

  @Test
  void processForBlock_multipleCallsForSameBlockOnlyProcessOneFailure() {
    // This test verifies the key invariant: only one invalidity proof per block
    // When processForBlock is called multiple times for the same block (as happens during
    // iterative block building), we should only process one failure per block.
    final List<ForcedTransaction> ftxs = createForcedTransactions(3);
    pool.addForcedTransactions(ftxs);

    // First call to processForBlock for block 100: first tx fails
    pool.processForBlock(100L, TEST_TIMESTAMP, alwaysReject("NONCE"));

    // Simulate iterative block building: second call for same block 100
    // This should be skipped because we already had a failure in this block
    pool.processForBlock(100L, TEST_TIMESTAMP, alwaysReject("NONCE"));

    // Third call for same block 100 - still should be skipped
    pool.processForBlock(100L, TEST_TIMESTAMP, alwaysReject("NONCE"));

    // Only first tx should have status, others still pending
    assertThat(pool.pendingCount()).isEqualTo(2);
    assertInclusionStatus(ftxs.get(0).txHash(), 100L, ForcedTransactionInclusionResult.BAD_NONCE);
    assertThat(pool.getInclusionStatus(ftxs.get(1).txHash())).isEmpty();
    assertThat(pool.getInclusionStatus(ftxs.get(2).txHash())).isEmpty();

    // Now move to block 101 - second tx should be processed
    pool.processForBlock(101L, TEST_TIMESTAMP, alwaysReject("BALANCE"));

    assertThat(pool.pendingCount()).isEqualTo(1);
    assertInclusionStatus(ftxs.get(1).txHash(), 101L, ForcedTransactionInclusionResult.BAD_BALANCE);
    assertThat(pool.getInclusionStatus(ftxs.get(2).txHash())).isEmpty();

    // Move to block 102 - third tx processed
    pool.processForBlock(102L, TEST_TIMESTAMP, alwaysReject("PRECOMPILE"));

    assertThat(pool.pendingCount()).isZero();
    assertInclusionStatus(
        ftxs.get(2).txHash(), 102L, ForcedTransactionInclusionResult.BAD_PRECOMPILE);
  }

  @Test
  void processForBlock_mapsNonceErrorsCorrectly() {
    final ForcedTransaction ftx = createForcedTransaction();
    pool.addForcedTransactions(List.of(ftx));

    pool.processForBlock(100L, TEST_TIMESTAMP, alwaysReject("NONCE_TOO_HIGH"));

    assertInclusionStatus(ftx.txHash(), 100L, ForcedTransactionInclusionResult.BAD_NONCE);
  }

  @Test
  void processForBlock_mapsBalanceErrorsCorrectly() {
    final ForcedTransaction ftx = createForcedTransaction();
    pool.addForcedTransactions(List.of(ftx));

    pool.processForBlock(100L, TEST_TIMESTAMP, alwaysReject("UPFRONT_COST_EXCEEDS_BALANCE"));

    assertInclusionStatus(ftx.txHash(), 100L, ForcedTransactionInclusionResult.BAD_BALANCE);
  }

  @Test
  void processForBlock_mapsFilteredAddressErrorsCorrectly() {
    final ForcedTransaction ftx = createForcedTransaction();
    pool.addForcedTransactions(List.of(ftx));

    pool.processForBlock(100L, TEST_TIMESTAMP, alwaysRejectWith(DENIED_LOG_TOPIC));

    assertInclusionStatus(ftx.txHash(), 100L, ForcedTransactionInclusionResult.FILTERED_ADDRESSES);
  }

  @Test
  void processForBlock_retriesOnUnknownRejectionReasonAtIndex0() {
    final ForcedTransaction ftx = createForcedTransaction();
    pool.addForcedTransactions(List.of(ftx));

    // First block: unknown rejection reason at index 0 - should retry
    pool.processForBlock(100L, TEST_TIMESTAMP, alwaysReject("SOME_UNKNOWN_REASON"));

    assertThat(pool.pendingCount()).isEqualTo(1); // Still pending
    assertThat(pool.getInclusionStatus(ftx.txHash())).isEmpty(); // No status recorded

    // Second block: same unknown reason at index 0 - still retry
    pool.processForBlock(101L, TEST_TIMESTAMP, alwaysReject("ANOTHER_UNKNOWN_REASON"));

    assertThat(pool.pendingCount()).isEqualTo(1); // Still pending
    assertThat(pool.getInclusionStatus(ftx.txHash())).isEmpty(); // Still no status

    // Third block: now it succeeds
    pool.processForBlock(102L, TEST_TIMESTAMP, alwaysSelect());

    assertThat(pool.pendingCount()).isZero();
    assertInclusionStatus(ftx.txHash(), 102L, ForcedTransactionInclusionResult.INCLUDED);
  }

  @Test
  void processForBlock_finallyRejectsWhenKnownReasonAtIndex0AfterRetry() {
    final ForcedTransaction ftx = createForcedTransaction();
    pool.addForcedTransactions(List.of(ftx));

    // First block: unknown rejection reason - will retry
    pool.processForBlock(100L, TEST_TIMESTAMP, alwaysReject("SOME_UNKNOWN_REASON"));

    assertThat(pool.pendingCount()).isEqualTo(1);
    assertThat(pool.getInclusionStatus(ftx.txHash())).isEmpty();

    // Second block: known rejection reason (NONCE) - final rejection
    pool.processForBlock(101L, TEST_TIMESTAMP, alwaysReject("NONCE_TOO_LOW"));

    assertThat(pool.pendingCount()).isZero();
    assertInclusionStatus(ftx.txHash(), 101L, ForcedTransactionInclusionResult.BAD_NONCE);
  }

  @Test
  void getInclusionStatus_returnsEmptyForUnknownTransaction() {
    final Hash unknownHash = Hash.fromHexStringLenient("0x999");
    assertThat(pool.getInclusionStatus(unknownHash)).isEmpty();
  }

  @Test
  void processForBlock_doesNothingWhenQueueEmpty() {
    pool.processForBlock(100L, TEST_TIMESTAMP, alwaysSelect());
    assertThat(pool.pendingCount()).isZero();
  }

  private void assertInclusionStatus(
      final Hash txHash,
      final long expectedBlockNumber,
      final ForcedTransactionInclusionResult expectedResult) {
    final var status = pool.getInclusionStatus(txHash);
    assertThat(status).isPresent();
    assertThat(status.get().blockNumber()).isEqualTo(expectedBlockNumber);
    assertThat(status.get().inclusionResult()).isEqualTo(expectedResult);
  }

  private ForcedTransaction createForcedTransaction() {
    final Transaction tx = txFactory.createTransaction();
    return new ForcedTransaction(tx.getHash(), tx, Long.MAX_VALUE);
  }

  private List<ForcedTransaction> createForcedTransactions(final int count) {
    final List<ForcedTransaction> ftxs = new ArrayList<>(count);
    for (int i = 0; i < count; i++) {
      ftxs.add(createForcedTransaction());
    }
    return ftxs;
  }

  private BlockTransactionSelectionService alwaysSelect() {
    final BlockTransactionSelectionService bts = mock(BlockTransactionSelectionService.class);
    when(bts.evaluatePendingTransaction(any())).thenReturn(TransactionSelectionResult.SELECTED);
    return bts;
  }

  private BlockTransactionSelectionService alwaysReject(final String reason) {
    final BlockTransactionSelectionService bts = mock(BlockTransactionSelectionService.class);
    when(bts.evaluatePendingTransaction(any()))
        .thenReturn(TransactionSelectionResult.invalidTransient(reason));
    return bts;
  }

  private BlockTransactionSelectionService alwaysRejectWith(
      final TransactionSelectionResult result) {
    final BlockTransactionSelectionService bts = mock(BlockTransactionSelectionService.class);
    when(bts.evaluatePendingTransaction(any())).thenReturn(result);
    return bts;
  }

  private BlockTransactionSelectionService selectThenReject(final String rejectReason) {
    final BlockTransactionSelectionService bts = mock(BlockTransactionSelectionService.class);
    when(bts.evaluatePendingTransaction(any()))
        .thenReturn(TransactionSelectionResult.SELECTED)
        .thenReturn(TransactionSelectionResult.invalidTransient(rejectReason));
    return bts;
  }
}
