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

import static net.consensys.linea.zktracer.CurveOperations.isOnC1;
import static net.consensys.linea.zktracer.CurveOperations.isOnG2;

import java.nio.MappedByteBuffer;
import java.util.List;
import java.util.Stack;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@Slf4j
@RequiredArgsConstructor
public final class EcPairingCallEffectiveCall implements Module {
  private final Hub hub;
  @Getter private final Stack<EcPairingLimit> counts = new Stack<>();
  private static final int PRECOMPILE_BASE_GAS_FEE = 45_000; // cf EIP-1108
  private static final int PRECOMPILE_MILLER_LOOP_GAS_FEE = 34_000; // cf EIP-1108
  private static final int ECPAIRING_NB_BYTES_PER_MILLER_LOOP = 192;

  @Override
  public String moduleKey() {
    return "PRECOMPILE_ECPAIRING_EFFECTIVE_CALL";
  }

  @Override
  public void enterTransaction() {
    counts.push(new EcPairingLimit(0, 0));
  }

  @Override
  public void popTransaction() {
    counts.pop();
  }

  public static boolean isHubFailure(final Hub hub) {
    final OpCode opCode = hub.opCode();
    final MessageFrame frame = hub.messageFrame();

    switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address target = Words.toAddress(frame.getStackItem(1));
        if (target.equals(Address.ALTBN128_PAIRING)) {
          long length = hub.transients().op().callDataSegment().length();
          if (length % 192 != 0) {
            return true;
          }
          final long pairingCount = length / ECPAIRING_NB_BYTES_PER_MILLER_LOOP;

          return hub.transients().op().gasAllowanceForCall()
              < PRECOMPILE_BASE_GAS_FEE + PRECOMPILE_MILLER_LOOP_GAS_FEE * pairingCount;
        }
      }
      default -> {}
    }

    return false;
  }

  public static boolean isRamFailure(final Hub hub) {
    final MessageFrame frame = hub.messageFrame();
    long length = hub.transients().op().callDataSegment().length();

    if (length == 0) {
      return true;
    }

    for (int i = 0; i < length; i += 192) {
      final Bytes coordinates = frame.shadowReadMemory(i, 192);
      if (!isOnC1(coordinates.slice(0, 64)) || !isOnG2(coordinates.slice(64, 128))) {
        return true;
      }
    }

    return false;
  }

  public static long gasCost(final Hub hub) {
    final OpCode opCode = hub.opCode();
    final MessageFrame frame = hub.messageFrame();

    switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address target = Words.toAddress(frame.getStackItem(1));
        if (target.equals(Address.ALTBN128_PAIRING)) {
          final long length = hub.transients().op().callDataSegment().length();
          final long nMillerLoop = (length / ECPAIRING_NB_BYTES_PER_MILLER_LOOP);
          if (nMillerLoop * ECPAIRING_NB_BYTES_PER_MILLER_LOOP != length) {
            return 0;
          }

          return PRECOMPILE_BASE_GAS_FEE + PRECOMPILE_MILLER_LOOP_GAS_FEE * nMillerLoop;
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
        if (target.equals(Address.ALTBN128_PAIRING)) {
          long length = hub.transients().op().callDataSegment().length();

          final long nMillerLoop = (length / ECPAIRING_NB_BYTES_PER_MILLER_LOOP);
          if (nMillerLoop * ECPAIRING_NB_BYTES_PER_MILLER_LOOP != length) {
            log.warn("[ECPairing] Argument is not a right size: " + length);
            return;
          }

          if (hub.transients().op().gasAllowanceForCall()
              >= PRECOMPILE_BASE_GAS_FEE + PRECOMPILE_MILLER_LOOP_GAS_FEE * nMillerLoop) {
            final EcPairingLimit lastEcpairingLimit = this.counts.pop();
            this.counts.push(
                new EcPairingLimit(
                    lastEcpairingLimit.nPrecompileCall() + 1,
                    lastEcpairingLimit.nMillerLoop() + nMillerLoop));
          }
        }
      }
      default -> {}
    }
  }

  @Override
  public int lineCount() {
    return this.counts.stream().mapToInt(EcPairingLimit::nPrecompileCall).sum();
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
