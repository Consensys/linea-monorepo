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
public final class Rip160 implements Module {
  private final Hub hub;
  private final Stack<Integer> counts = new Stack<>();

  @Override
  public String moduleKey() {
    return "PRECOMPILE_RIPEMD";
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

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    final OpCode opCode = hub.opCode();

    switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address target = Words.toAddress(frame.getStackItem(1));
        if (target.equals(Address.RIPEMD160)) {
          long dataByteLength = 0;
          switch (opCode) {
            case CALL, CALLCODE -> dataByteLength = Words.clampedToLong(frame.getStackItem(4));
            case DELEGATECALL, STATICCALL -> dataByteLength =
                Words.clampedToLong(frame.getStackItem(3));
          }

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
          final long gasPaid = Words.clampedToLong(frame.getStackItem(0));
          final long gasNeeded = PRECOMPILE_BASE_GAS_FEE + PRECOMPILE_GAS_FEE_PER_EWORD * wordCount;

          if (gasPaid >= gasNeeded) {
            this.counts.push(this.counts.pop() + blockCount);
          }
        }
      }
      default -> {}
    }
  }

  @Override
  public int lineCount() {
    return this.counts.stream().mapToInt(x -> x).sum();
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
