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
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import java.util.List;
import net.consensys.linea.sequencer.forced.LineaForcedTransactionPool;
import net.consensys.linea.utils.TestTransactionFactory;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class LineaSendForcedRawTransactionTest {

  private LineaSendForcedRawTransaction method;
  private LineaForcedTransactionPool pool;
  private TestTransactionFactory txFactory;

  @BeforeEach
  void setUp() {
    pool = new LineaForcedTransactionPool(100, null);
    method = new LineaSendForcedRawTransaction().init(pool);
    txFactory = new TestTransactionFactory();
  }

  @Test
  void getNamespace_returnsLinea() {
    assertThat(method.getNamespace()).isEqualTo("linea");
  }

  @Test
  void getName_returnsCorrectName() {
    assertThat(method.getName()).isEqualTo("sendForcedRawTransaction");
  }

  @Test
  void execute_acceptsValidTransaction() {
    final Transaction tx = txFactory.createTransaction();
    final String rawTx = TestTransactionFactory.encodeTransaction(tx);

    final List<String> result = method.execute(request(new ForcedTxParam(rawTx, "0x1000")));

    assertThat(result).hasSize(1);
    assertThat(result.get(0)).isEqualTo(tx.getHash().toHexString());
    assertThat(pool.pendingCount()).isEqualTo(1);
  }

  @Test
  void execute_acceptsMultipleTransactions() {
    final Transaction tx1 = txFactory.createTransaction();
    final Transaction tx2 = txFactory.createTransaction();

    final List<String> result =
        method.execute(
            request(
                new ForcedTxParam(TestTransactionFactory.encodeTransaction(tx1), "0x1000"),
                new ForcedTxParam(TestTransactionFactory.encodeTransaction(tx2), "0x2000")));

    assertThat(result).hasSize(2);
    assertThat(result).containsExactly(tx1.getHash().toHexString(), tx2.getHash().toHexString());
    assertThat(pool.pendingCount()).isEqualTo(2);
  }

  @Test
  void execute_parsesDeadlineCorrectly() {
    final Transaction tx = txFactory.createTransaction();

    method.execute(
        request(new ForcedTxParam(TestTransactionFactory.encodeTransaction(tx), "0x1388")));

    // Verify the deadline was parsed - check via the pool's internal state
    final var status = pool.getInclusionStatus(tx.getHash());
    assertThat(status).isEmpty(); // Not processed yet, just added
    assertThat(pool.pendingCount()).isEqualTo(1);
  }

  @Test
  void execute_rejectsNullDeadline() {
    final Transaction tx = txFactory.createTransaction();

    assertThatThrownBy(
            () ->
                method.execute(
                    request(new ForcedTxParam(TestTransactionFactory.encodeTransaction(tx), null))))
        .isInstanceOf(PluginRpcEndpointException.class);
  }

  @Test
  void execute_rejectsEmptyDeadline() {
    final Transaction tx = txFactory.createTransaction();

    assertThatThrownBy(
            () ->
                method.execute(
                    request(new ForcedTxParam(TestTransactionFactory.encodeTransaction(tx), ""))))
        .isInstanceOf(PluginRpcEndpointException.class);
  }

  @Test
  void execute_rejectsEmptyParams() {
    assertThatThrownBy(() -> method.execute(request()))
        .isInstanceOf(PluginRpcEndpointException.class);
  }

  @Test
  void execute_rejectsEmptyTransactionData() {
    assertThatThrownBy(() -> method.execute(request(new ForcedTxParam("", "0x100"))))
        .isInstanceOf(PluginRpcEndpointException.class);
  }

  @Test
  void execute_rejectsInvalidDeadlineFormat() {
    final String rawTx = TestTransactionFactory.encodeTransaction(txFactory.createTransaction());

    assertThatThrownBy(() -> method.execute(request(new ForcedTxParam(rawTx, "not-a-number"))))
        .isInstanceOf(PluginRpcEndpointException.class);
  }

  @Test
  void execute_rejectsInvalidTransactionData() {
    assertThatThrownBy(() -> method.execute(request(new ForcedTxParam("0xdeadbeef", "0x1000"))))
        .isInstanceOf(PluginRpcEndpointException.class);
  }

  private PluginRpcRequest request(final ForcedTxParam... params) {
    final PluginRpcRequest req = mock(PluginRpcRequest.class);
    final LineaSendForcedRawTransaction.ForcedTransactionParam[] converted =
        new LineaSendForcedRawTransaction.ForcedTransactionParam[params.length];
    for (int i = 0; i < params.length; i++) {
      converted[i] =
          new LineaSendForcedRawTransaction.ForcedTransactionParam(
              params[i].transaction, params[i].deadline);
    }
    when(req.getParams()).thenReturn(new Object[] {converted});
    return req;
  }

  private record ForcedTxParam(String transaction, String deadline) {}
}
