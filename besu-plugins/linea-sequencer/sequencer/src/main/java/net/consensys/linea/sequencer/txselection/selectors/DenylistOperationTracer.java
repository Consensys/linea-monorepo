/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.txselection.selectors;

import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.tracing.OperationTracer;
import org.hyperledger.besu.evm.worldstate.WorldView;

/**
 * Operation tracer that collects all recipient addresses entered during EVM execution. Used by
 * {@link DenylistExecutionSelector} to check called addresses against the denylist after
 * transaction processing.
 */
public class DenylistOperationTracer implements OperationTracer {

  private final ThreadLocal<Set<Address>> calledAddresses =
      ThreadLocal.withInitial(HashSet::new);

  @Override
  public void traceStartTransaction(final WorldView worldView, final Transaction transaction) {
    calledAddresses.get().clear();
  }

  @Override
  public void traceContextEnter(final MessageFrame frame) {
    calledAddresses.get().add(frame.getRecipientAddress());
  }

  @Override
  public void traceEndTransaction(
      final WorldView worldView,
      final Transaction tx,
      final boolean status,
      final Bytes output,
      final List<Log> logs,
      final long gasUsed,
      final Set<Address> selfDestructs,
      final long timeNs) {
    calledAddresses.remove();
  }

  public Set<Address> getCalledAddresses() {
    return Collections.unmodifiableSet(calledAddresses.get());
  }
}
