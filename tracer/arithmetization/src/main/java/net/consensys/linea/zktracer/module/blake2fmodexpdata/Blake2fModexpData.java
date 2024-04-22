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

import static net.consensys.linea.zktracer.module.blake2fmodexpdata.Trace.PHASE_BLAKE_DATA;
import static net.consensys.linea.zktracer.module.blake2fmodexpdata.Trace.PHASE_BLAKE_RESULT;
import static net.consensys.linea.zktracer.module.blake2fmodexpdata.Trace.PHASE_MODEXP_BASE;
import static net.consensys.linea.zktracer.module.blake2fmodexpdata.Trace.PHASE_MODEXP_EXPONENT;
import static net.consensys.linea.zktracer.module.blake2fmodexpdata.Trace.PHASE_MODEXP_MODULUS;
import static net.consensys.linea.zktracer.module.blake2fmodexpdata.Trace.PHASE_MODEXP_RESULT;

import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import org.apache.tuweni.bytes.Bytes;

public class Blake2fModexpData implements Module {
  private StackedSet<Blake2fModexpDataOperation> state = new StackedSet<>();

  @Override
  public String moduleKey() {
    return "BLAKE2f_MODEXP_DATA";
  }

  @Override
  public void enterTransaction() {
    this.state.enter();
  }

  @Override
  public void popTransaction() {
    this.state.pop();
  }

  @Override
  public int lineCount() {
    return this.state.lineCount();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  public int call(final Blake2fModexpDataOperation operation) {
    this.state.add(operation);

    return operation.prevHubStamp();
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    Trace trace = new Trace(buffers);
    int stamp = 0;
    for (Blake2fModexpDataOperation o : this.state) {
      stamp++;
      o.trace(trace, stamp);
    }
  }

  public Bytes getInputDataByIdAndPhase(final int id, final int phase) {
    final Blake2fModexpDataOperation op = getOperationById(id);
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

  private Blake2fModexpDataOperation getOperationById(final int id) {
    for (Blake2fModexpDataOperation operation : this.state) {
      if (id == operation.hubStampPlusOne) {
        return operation;
      }
    }
    throw new RuntimeException("BlakeModexpOperation not found");
  }
}
