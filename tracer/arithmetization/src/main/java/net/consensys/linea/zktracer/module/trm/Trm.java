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

package net.consensys.linea.zktracer.module.trm;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import org.hyperledger.besu.datatypes.AccessListEntry;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class Trm implements Module {
  static final int MAX_CT = 16;
  static final int LLARGE = 16;
  static final int PIVOT_BIT_FLIPS_TO_TRUE = 12;

  private final StackedSet<TrmOperation> trimmings = new StackedSet<>();

  @Override
  public String moduleKey() {
    return "TRM";
  }

  @Override
  public void enterTransaction() {
    this.trimmings.enter();
  }

  @Override
  public void popTransaction() {
    this.trimmings.pop();
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    switch (opCode) {
      case BALANCE, EXTCODESIZE, EXTCODECOPY, EXTCODEHASH, SELFDESTRUCT -> {
        this.trimmings.add(new TrmOperation(EWord.of(frame.getStackItem(0))));
      }
      case CALL, CALLCODE, DELEGATECALL, STATICCALL -> {
        this.trimmings.add(new TrmOperation(EWord.of(frame.getStackItem(1))));
      }
    }
  }

  @Override
  public void traceStartTx(WorldView worldView, Transaction tx) {
    final TransactionType txType = tx.getType();
    switch (txType) {
      case ACCESS_LIST, EIP1559 -> {
        if (tx.getAccessList().isPresent()) {
          for (AccessListEntry entry : tx.getAccessList().get()) {
            this.trimmings.add(new TrmOperation(EWord.of(entry.address())));
          }
        }
      }
      case FRONTIER -> {
        return;
      }
      default -> {
        throw new IllegalStateException("TransactionType not supported: " + txType);
      }
    }
  }

  public static boolean isPrec(EWord data) {
    BigInteger trmAddrParamAsBigInt = data.slice(12, 20).toUnsignedBigInteger();
    return (!trmAddrParamAsBigInt.equals(BigInteger.ZERO)
        && (trmAddrParamAsBigInt.compareTo(BigInteger.TEN) < 0));
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int stamp = 0;
    for (TrmOperation operation : this.trimmings) {
      stamp++;
      operation.trace(trace, stamp);
    }
  }

  @Override
  public int lineCount() {
    return this.trimmings.lineCount();
  }
}
