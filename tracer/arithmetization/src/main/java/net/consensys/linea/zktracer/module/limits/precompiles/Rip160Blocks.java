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

package net.consensys.linea.zktracer.module.limits.precompiles;

import java.nio.MappedByteBuffer;
import java.util.List;
import java.util.Stack;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@RequiredArgsConstructor
public final class Rip160Blocks implements Module {
  private final Hub hub;
  private final Stack<Integer> counts = new Stack<>();

  @Override
  public String moduleKey() {
    return "PRECOMPILE_RIPEMD_BLOCKS";
  }

  private static final int PRECOMPILE_BASE_GAS_FEE = 600;
  private static final int PRECOMPILE_GAS_FEE_PER_EWORD = 120;
  private static final int RIPEMD160_BLOCKSIZE = 64 * 8;
  // If the length is > 2⁶4, we just use the lower 64 bits.
  private static final int RIPEMD160_LENGTH_APPEND = 64;
  private static final int RIPEMD160_ND_PADDED_ONE = 1;

  @Override
  public void enterTransaction() {
    counts.push(0);
  }

  @Override
  public void popTransaction() {
    counts.pop();
  }

  public static boolean hasEnoughGas(final Hub hub) {
    return hub.transients().op().gasAllowanceForCall() <= gasCost(hub);
  }

  public static long gasCost(final Hub hub) {
    final OpCode opCode = hub.opCode();

    switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address target = Words.toAddress(hub.messageFrame().getStackItem(1));
        if (target.equals(Address.RIPEMD160)) {
          final long dataByteLength = hub.transients().op().callDataSegment().length();
          final long wordCount = (dataByteLength + 31) / 32;
          return PRECOMPILE_BASE_GAS_FEE + PRECOMPILE_GAS_FEE_PER_EWORD * wordCount;
        }
      }
    }

    return 0;
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    final OpCode opCode = hub.opCode();

    switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address target = Words.toAddress(frame.getStackItem(1));
        if (target.equals(Address.RIPEMD160)) {
          final long dataByteLength = hub.transients().op().callDataSegment().length();
          if (dataByteLength == 0) {
            return;
          } // skip trivial hash TODO: check the prover does skip it
          final int blockCount =
              (int)
                      (dataByteLength * 8
                          + RIPEMD160_ND_PADDED_ONE
                          + RIPEMD160_LENGTH_APPEND
                          + (RIPEMD160_BLOCKSIZE - 1))
                  / RIPEMD160_BLOCKSIZE;

          final long wordCount = (dataByteLength + 31) / 32;
          final long gasNeeded = PRECOMPILE_BASE_GAS_FEE + PRECOMPILE_GAS_FEE_PER_EWORD * wordCount;

          if (hub.transients().op().gasAllowanceForCall() >= gasNeeded) {
            this.counts.push(this.counts.pop() + blockCount);
          }
        }
      }
      default -> {}
    }
  }

  @Override
  public int lineCount() {
    int r = 0;
    for (int i = 0; i < this.counts.size(); i++) {
      r += this.counts.get(i);
    }
    return r;
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    throw new IllegalStateException("should never be called");
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    throw new IllegalStateException("should never be called");
  }
}
