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
import net.consensys.linea.zktracer.module.hub.precompiles.Blake2fMetadata;
import net.consensys.linea.zktracer.module.hub.precompiles.PrecompileMetadata;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.MemorySpan;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@RequiredArgsConstructor
public final class Blake2fRounds implements Module {
  private static final int BLAKE2F_VALID_DATASIZE = 213;

  private final Hub hub;
  private final Stack<Integer> counts = new Stack<>();

  @Override
  public String moduleKey() {
    return "PRECOMPILE_BLAKE2F_ROUNDS";
  }

  @Override
  public void enterTransaction() {
    counts.push(0);
  }

  @Override
  public void popTransaction() {
    counts.pop();
  }

  public static boolean isHubFailure(final Hub hub) {
    final OpCode opCode = hub.opCode();
    final MessageFrame frame = hub.messageFrame();

    return switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address target = Words.toAddress(frame.getStackItem(1));
        if (target.equals(Address.BLAKE2B_F_COMPRESSION)) {
          final long length = hub.transients().op().callDataSegment().length();
          yield length != BLAKE2F_VALID_DATASIZE;
        } else {
          yield false;
        }
      }
      default -> false;
    };
  }

  public static boolean isRamFailure(final Hub hub) {
    final OpCode opCode = hub.opCode();
    final MessageFrame frame = hub.messageFrame();

    if (isHubFailure(hub)) {
      return false;
    }

    return switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address target = Words.toAddress(frame.getStackItem(1));
        if (target.equals(Address.BLAKE2B_F_COMPRESSION)) {
          final long offset = hub.transients().op().callDataSegment().offset();
          final int f =
              frame
                  .shadowReadMemory(offset, BLAKE2F_VALID_DATASIZE)
                  .get(BLAKE2F_VALID_DATASIZE - 1);
          final int r =
              frame
                  .shadowReadMemory(offset, BLAKE2F_VALID_DATASIZE)
                  .slice(0, 4)
                  .toInt(); // The number of round is equal to the gas to pay
          yield !((f == 0 || f == 1) && hub.transients().op().gasAllowanceForCall() >= r);
        } else {
          yield false;
        }
      }
      default -> false;
    };
  }

  public static long gasCost(final Hub hub) {
    final MessageFrame frame = hub.messageFrame();
    final Address target = Words.toAddress(frame.getStackItem(1));

    if (target.equals(Address.BLAKE2B_F_COMPRESSION)) {
      final MemorySpan callData = hub.transients().op().callDataSegment();
      final int blake2fDataSize = 213;
      if (callData.length() == blake2fDataSize) {
        final int f =
            frame.shadowReadMemory(callData.offset(), callData.length()).get(blake2fDataSize - 1);
        if (f == 0 || f == 1) {
          return frame
              .shadowReadMemory(callData.offset(), callData.length())
              .slice(0, 4)
              .toInt(); // The number of round is equal to the gas to pay
        }
      }
    }

    return 0;
  }

  public static PrecompileMetadata metadata(final Hub hub) {
    final OpCode opCode = hub.opCode();

    switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address target = Words.toAddress(hub.messageFrame().getStackItem(1));
        if (target.equals(Address.BLAKE2B_F_COMPRESSION)) {
          final long length = hub.transients().op().callDataSegment().length();

          if (length == BLAKE2F_VALID_DATASIZE) {
            final int f = hub.transients().op().callData().get(BLAKE2F_VALID_DATASIZE - 1);
            if (f == 0 || f == 1) {
              final int r =
                  hub.transients()
                      .op()
                      .callData()
                      .slice(0, 4)
                      .toInt(); // The number of round is equal to the gas to pay
              return new Blake2fMetadata(r, f);
            }
          }
        }
      }
    }

    return new Blake2fMetadata(0, 0);
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    final OpCode opCode = hub.opCode();

    switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address target = Words.toAddress(frame.getStackItem(1));
        if (target.equals(Address.BLAKE2B_F_COMPRESSION)) {
          final long length = hub.transients().op().callDataSegment().length();

          if (length == BLAKE2F_VALID_DATASIZE) {
            final int f = hub.transients().op().callData().get(BLAKE2F_VALID_DATASIZE - 1);
            if (f == 0 || f == 1) {
              final int r =
                  hub.transients()
                      .op()
                      .callData()
                      .slice(0, 4)
                      .toInt(); // The number of round is equal to the gas to pay
              if (hub.transients().op().gasAllowanceForCall() >= r) {
                this.counts.push(this.counts.pop() + r);
              }
            }
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
