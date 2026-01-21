/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.rpc.methods;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import java.util.List;
import net.consensys.linea.sequencer.forced.ForcedTransaction;
import net.consensys.linea.sequencer.forced.LineaForcedTransactionPool;
import net.consensys.linea.utils.TestTransactionFactory;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;
import org.hyperledger.besu.plugin.services.txselection.BlockTransactionSelectionService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class LineaGetForcedTransactionInclusionStatusTest {

  private static final long TEST_TIMESTAMP = 1_700_000_000L;

  private LineaGetForcedTransactionInclusionStatus method;
  private LineaForcedTransactionPool pool;
  private TestTransactionFactory txFactory;

  @BeforeEach
  void setUp() {
    pool = new LineaForcedTransactionPool(100, null);
    method = new LineaGetForcedTransactionInclusionStatus().init(pool);
    txFactory = new TestTransactionFactory();
  }

  @Test
  void getNamespace_returnsLinea() {
    assertThat(method.getNamespace()).isEqualTo("linea");
  }

  @Test
  void getName_returnsCorrectName() {
    assertThat(method.getName()).isEqualTo("getForcedTransactionInclusionStatus");
  }

  @Test
  void execute_returnsNullForUnknownTransaction() {
    final String txHash = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef";

    final var result = method.execute(request(txHash));

    assertThat(result).isNull();
  }

  @Test
  void execute_returnsIncludedStatus() {
    final ForcedTransaction ftx =
        addAndProcessTransaction(TransactionSelectionResult.SELECTED, 100L);

    final var result = method.execute(request(ftx.txHash().toHexString()));

    assertThat(result).isNotNull();
    assertThat(result.transactionHash).isEqualTo(ftx.txHash().toHexString());
    assertThat(result.inclusionResult).isEqualTo("INCLUDED");
    assertThat(result.blockNumber).isEqualTo("0x64"); // 100 in hex
    assertThat(result.blockTimestamp).isEqualTo("0x6553f100"); // TEST_TIMESTAMP in hex
    assertThat(result.from).isEqualTo(ftx.transaction().getSender().toHexString());
  }

  @Test
  void execute_returnsBadNonceStatus() {
    final ForcedTransaction ftx =
        addAndProcessTransaction(
            TransactionSelectionResult.invalidTransient("NONCE_TOO_LOW"), 100L);

    final var result = method.execute(request(ftx.txHash().toHexString()));

    assertThat(result).isNotNull();
    assertThat(result.inclusionResult).isEqualTo("BAD_NONCE");
    assertThat(result.blockNumber).isEqualTo("0x64");
  }

  @Test
  void execute_returnsBadBalanceStatus() {
    final ForcedTransaction ftx =
        addAndProcessTransaction(
            TransactionSelectionResult.invalidTransient("UPFRONT_COST_EXCEEDS_BALANCE"), 100L);

    final var result = method.execute(request(ftx.txHash().toHexString()));

    assertThat(result).isNotNull();
    assertThat(result.inclusionResult).isEqualTo("BAD_BALANCE");
  }

  @Test
  void execute_returnsFilteredAddressesStatus() {
    final ForcedTransaction ftx =
        addAndProcessTransaction(
            TransactionSelectionResult.invalidTransient("DENIED_LOG_TOPIC"), 100L);

    final var result = method.execute(request(ftx.txHash().toHexString()));

    assertThat(result).isNotNull();
    assertThat(result.inclusionResult).isEqualTo("FILTERED_ADDRESSES");
  }

  @Test
  void execute_rejectsEmptyTransactionHash() {
    assertThatThrownBy(() -> method.execute(request("")))
        .isInstanceOf(PluginRpcEndpointException.class);
  }

  @Test
  void execute_rejectsNullTransactionHash() {
    assertThatThrownBy(() -> method.execute(request(null)))
        .isInstanceOf(PluginRpcEndpointException.class);
  }

  @Test
  void execute_rejectsInvalidTransactionHash() {
    assertThatThrownBy(() -> method.execute(request("not-a-valid-hash")))
        .isInstanceOf(PluginRpcEndpointException.class);
  }

  private ForcedTransaction addAndProcessTransaction(
      final TransactionSelectionResult result, final long blockNumber) {
    final Transaction tx = txFactory.createTransaction();
    final ForcedTransaction ftx = new ForcedTransaction(tx.getHash(), tx, Long.MAX_VALUE);
    pool.addForcedTransactions(List.of(ftx));

    final BlockTransactionSelectionService bts = mock(BlockTransactionSelectionService.class);
    when(bts.evaluatePendingTransaction(any())).thenReturn(result);
    pool.processForBlock(blockNumber, TEST_TIMESTAMP, bts);
    return ftx;
  }

  private PluginRpcRequest request(final String txHash) {
    final PluginRpcRequest req = mock(PluginRpcRequest.class);
    when(req.getParams()).thenReturn(new Object[] {txHash});
    return req;
  }
}
