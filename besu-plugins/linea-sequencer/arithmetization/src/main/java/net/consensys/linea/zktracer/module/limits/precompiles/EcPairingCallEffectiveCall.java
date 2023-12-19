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

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@Slf4j
@RequiredArgsConstructor
public final class EcPairingCallEffectiveCall implements Module {
  private final Hub hub;
  @Getter private final Stack<EcPairingLimit> counts = new Stack<>();
  private static final int PRECOMPILE_BASE_GAS_FEE = 45000; // cf EIP-1108
  private static final int PRECOMPILE_MILLER_LOOP_GAS_FEE = 34000; // cf EIP-1108
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

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    final OpCode opCode = hub.opCode();

    switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address target = Words.toAddress(frame.getStackItem(1));
        if (target.equals(Address.ALTBN128_PAIRING)) {
          long length = 0;
          switch (opCode) {
            case CALL, CALLCODE -> length = Words.clampedToLong(frame.getStackItem(4));
            case DELEGATECALL, STATICCALL -> length = Words.clampedToLong(frame.getStackItem(3));
          }

          final long nMillerLoop = (length / ECPAIRING_NB_BYTES_PER_MILLER_LOOP);
          if (nMillerLoop * ECPAIRING_NB_BYTES_PER_MILLER_LOOP != length) {
            log.warn("[ECPairing] Argument is not a right size: " + length);
            return;
          }

          final long gasPaid = Words.clampedToLong(frame.getStackItem(0));
          if (gasPaid >= PRECOMPILE_BASE_GAS_FEE + PRECOMPILE_MILLER_LOOP_GAS_FEE * nMillerLoop) {
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
