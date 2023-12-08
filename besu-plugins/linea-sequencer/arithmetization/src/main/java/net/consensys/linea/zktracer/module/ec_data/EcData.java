/*
 * Copyright Consensys Software Inc.
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

package net.consensys.linea.zktracer.module.ec_data;

import java.nio.MappedByteBuffer;
import java.util.List;
import java.util.Set;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.ext.Ext;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@RequiredArgsConstructor
public class EcData implements Module {
  private static final Set<Address> EC_PRECOMPILES =
      Set.of(Address.ECREC, Address.ALTBN128_ADD, Address.ALTBN128_MUL, Address.ALTBN128_PAIRING);

  private final StackedSet<EcDataOperation> operations = new StackedSet<>();
  private final Hub hub;
  private final Wcp wcp;
  private final Ext ext;
  private int previousContextNumber = 0;

  @Override
  public String moduleKey() {
    return "EC_DATA";
  }

  @Override
  public void enterTransaction() {
    this.operations.enter();
  }

  @Override
  public void popTransaction() {
    this.operations.pop();
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    final OpCode opCode = hub.opCode();
    // TODO: in the PCH
    if (frame.stackSize() < 5) {
      return;
    }

    switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address to = Words.toAddress(frame.getStackItem(1));
        if (!EC_PRECOMPILES.contains(to)) {
          return;
        }

        long length = 0;
        long offset = 0;
        switch (opCode) {
          case CALL, CALLCODE -> {
            length = Words.clampedToLong(frame.getStackItem(4));
            offset = Words.clampedToLong(frame.getStackItem(3));
          }
          case DELEGATECALL, STATICCALL -> {
            length = Words.clampedToLong(frame.getStackItem(3));
            offset = Words.clampedToLong(frame.getStackItem(2));
          }
        }

        if (to.equals(Address.ALTBN128_PAIRING) && (length == 0 || length % 192 != 0)) {
          return;
        }

        final Bytes input = frame.shadowReadMemory(offset, length);

        this.operations.add(
            EcDataOperation.of(
                this.wcp,
                this.ext,
                to,
                input,
                this.hub.currentFrame().contextNumber(),
                this.previousContextNumber));
        this.previousContextNumber = this.hub.currentFrame().contextNumber();
      }
      default -> {}
    }
  }

  @Override
  public int lineCount() {
    int sum = 0;
    for (EcDataOperation op : this.operations) {
      sum += op.nRows();
    }
    return sum;
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    for (EcDataOperation op : this.operations) {
      op.trace(trace);
    }
  }
}
