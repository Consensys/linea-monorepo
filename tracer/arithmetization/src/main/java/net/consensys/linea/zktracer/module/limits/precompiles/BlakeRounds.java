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
import java.util.ArrayDeque;
import java.util.Deque;
import java.util.List;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.blake2fmodexpdata.BlakeComponents;
import net.consensys.linea.zktracer.module.blake2fmodexpdata.BlakeModexpData;
import net.consensys.linea.zktracer.module.blake2fmodexpdata.BlakeModexpDataOperation;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.precompiles.Blake2fMetadata;
import net.consensys.linea.zktracer.module.hub.precompiles.PrecompileMetadata;
import net.consensys.linea.zktracer.module.hub.transients.OperationAncillaries;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.MemorySpan;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@RequiredArgsConstructor
@Accessors(fluent = true)
public final class BlakeRounds implements Module {
  private static final int BLAKE2f_INPUT_SIZE = 213;
  private final Hub hub;

  @Getter private final BlakeModexpData blakeModexpData;

  @Getter private final Deque<BlakeLimit> counts = new ArrayDeque<>();

  @Override
  public String moduleKey() {
    return "PRECOMPILE_BLAKE_ROUNDS";
  }

  @Override
  public void traceStartConflation(final long blockCount) {
    counts.push(new BlakeLimit(0, 0));
  }

  @Override
  public void enterTransaction() {
    counts.push(counts.getFirst());
  }

  @Override
  public void popTransaction() {
    counts.pop();
  }

  public static boolean isHubFailure(final Hub hub) {
    final OpCode opCode = hub.opCode();
    final MessageFrame frame = hub.messageFrame();

    if (opCode.isCall()) {
      final Address target = Words.toAddress(frame.getStackItem(1));
      if (target.equals(Address.BLAKE2B_F_COMPRESSION)) {
        final long length = hub.transients().op().callDataSegment().length();
        return length != BLAKE2f_INPUT_SIZE;
      }
    }

    return false;
  }

  public static boolean isRamFailure(final Hub hub) {
    final OpCode opCode = hub.opCode();
    final MessageFrame frame = hub.messageFrame();

    if (isHubFailure(hub)) {
      return false;
    }

    if (opCode.isCall()) {
      final Address target = Words.toAddress(frame.getStackItem(1));
      if (target.equals(Address.BLAKE2B_F_COMPRESSION)) {
        final long offset = hub.transients().op().callDataSegment().offset();
        final int f =
            frame.shadowReadMemory(offset, BLAKE2f_INPUT_SIZE).get(BLAKE2f_INPUT_SIZE - 1);
        final int r =
            frame
                .shadowReadMemory(offset, BLAKE2f_INPUT_SIZE)
                .slice(0, 4)
                .toInt(); // The number of round is equal to the gas to pay
        return !((f == 0 || f == 1) && hub.transients().op().gasAllowanceForCall() >= r);
      }
    }

    return false;
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

    if (opCode.isCall()) {
      final Address target = Words.toAddress(hub.messageFrame().getStackItem(1));
      if (target.equals(Address.BLAKE2B_F_COMPRESSION)) {
        final long length = hub.transients().op().callDataSegment().length();

        if (length == BLAKE2f_INPUT_SIZE) {
          final int f = hub.transients().op().callData().get(BLAKE2f_INPUT_SIZE - 1);
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

    return new Blake2fMetadata(0, 0);
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    final OpCode opCode = hub.opCode();

    if (opCode.isCall()) {
      final Address target = Words.toAddress(frame.getStackItem(1));

      if (target.equals(Address.BLAKE2B_F_COMPRESSION)) {
        final OperationAncillaries opInfo = hub.transients().op();
        final long length = opInfo.callDataSegment().length();

        if (length == BLAKE2f_INPUT_SIZE) {
          final Bytes inputData = opInfo.callData();
          final int f = inputData.get(BLAKE2f_INPUT_SIZE - 1);
          if (f == 0 || f == 1) {
            final Bytes r = inputData.slice(0, 4); // The number of round is equal to the gas to pay

            final int rInt = r.toInt();

            if (opInfo.gasAllowanceForCall() >= rInt) {
              final Bytes data = inputData.slice(4, BLAKE2f_INPUT_SIZE - 5);
              this.blakeModexpData.call(
                  new BlakeModexpDataOperation(
                      hub.stamp(), null, new BlakeComponents(data, r, Bytes.of(f))));

              final BlakeLimit currentLimit = this.counts.pop();
              this.counts.push(
                  new BlakeLimit(
                      currentLimit.numberOfRounds() + rInt,
                      currentLimit.numberOfEffectiveCalls() + 1));
            }
          }
        }
      }
    }
  }

  @Override
  public int lineCount() {
    final long totalR = counts.stream().mapToLong(BlakeLimit::numberOfRounds).sum();
    if (totalR > Integer.MAX_VALUE) {
      throw new RuntimeException("Ludicrous BlakeLimit calls");
    }

    return (int) totalR;
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    throw new UnsupportedOperationException("should never be called");
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    throw new UnsupportedOperationException("should never be called");
  }
}
