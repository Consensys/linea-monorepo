/*
 * Copyright ConsenSys Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package net.consensys.linea.zktracer.module.blake2fmodexpdata;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_BLAKE_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_BLAKE_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_MODEXP_BASE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_MODEXP_EXPONENT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_MODEXP_MODULUS;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_MODEXP_RESULT;

import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.worldstate.WorldView;

@RequiredArgsConstructor
public class BlakeModexpData implements Module {
  private final Wcp wcp;
  private StackedSet<BlakeModexpDataOperation> operations = new StackedSet<>();
  private List<BlakeModexpDataOperation> sortedOperations = new ArrayList<>();
  private int numberOfOperationsAtStartTx = 0;

  @Override
  public String moduleKey() {
    return "BLAKE_MODEXP_DATA";
  }

  @Override
  public void enterTransaction() {
    this.operations.enter();
  }

  @Override
  public void traceEndTx(
      WorldView worldView,
      Transaction tx,
      boolean isSuccessful,
      Bytes output,
      List<Log> logs,
      long gasUsed) {
    final List<BlakeModexpDataOperation> newOperations =
        new ArrayList<>(this.operations.sets.getLast())
            .stream().sorted(Comparator.comparingLong(BlakeModexpDataOperation::id)).toList();

    this.sortedOperations.addAll(newOperations);
    final int numberOfOperationsAtEndTx = sortedOperations.size();
    for (int i = numberOfOperationsAtStartTx; i < numberOfOperationsAtEndTx; i++) {
      final long previousID = i == 0 ? 0 : sortedOperations.get(i - 1).id();
      this.wcp.callLT(previousID, sortedOperations.get(i).id());
    }
  }

  @Override
  public void traceStartTx(WorldView worldView, Transaction tx) {
    this.numberOfOperationsAtStartTx = operations.size();
  }

  @Override
  public void popTransaction() {
    this.sortedOperations.removeAll(this.operations.sets.getLast());
    this.operations.pop();
  }

  @Override
  public int lineCount() {
    return this.operations.lineCount();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  public void call(final BlakeModexpDataOperation operation) {
    this.operations.add(operation);
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    Trace trace = new Trace(buffers);
    int stamp = 0;
    for (BlakeModexpDataOperation o : this.sortedOperations) {
      stamp++;
      o.trace(trace, stamp);
    }
  }

  public Bytes getInputDataByIdAndPhase(final int id, final int phase) {
    final BlakeModexpDataOperation op = getOperationById(id);
    return switch (phase) {
      case PHASE_MODEXP_BASE -> op.modexpComponents.get().base();
      case PHASE_MODEXP_EXPONENT -> op.modexpComponents.get().exp();
      case PHASE_MODEXP_MODULUS -> op.modexpComponents.get().mod();
      case PHASE_MODEXP_RESULT -> Bytes.EMPTY; // TODO
      case PHASE_BLAKE_DATA -> op.blake2fComponents.get().data();
      case PHASE_BLAKE_RESULT -> Bytes.EMPTY; // TODO
      default -> throw new IllegalStateException("Unexpected value: " + phase);
    };
  }

  private BlakeModexpDataOperation getOperationById(final int id) {
    for (BlakeModexpDataOperation operation : this.operations) {
      if (id == operation.id) {
        return operation;
      }
    }
    throw new RuntimeException("BlakeModexpOperation not found");
  }
}
