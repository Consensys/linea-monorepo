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

package net.consensys.linea.zktracer.module.preclimits;

import java.nio.MappedByteBuffer;
import java.util.List;
import java.util.Stack;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@Slf4j
public final class EcpairingCall implements Module {
  public final Stack<EcpairingLimit> counts = new Stack<>();
  private final int precompileBaseGasFee = 45000; // cf EIP-1108
  private final int precompileMillerLoopGasFee = 34000; // cf EIP-1108
  private final int ecPairingNbBytesperMillerLoop = 192;

  @Override
  public String jsonKey() {
    return "ecpairingCall";
  }

  @Override
  public void enterTransaction() {
    counts.push(new EcpairingLimit(0, 0));
  }

  @Override
  public void popTransaction() {
    counts.pop();
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());

    switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address target = Words.toAddress(frame.getStackItem(1));
        if (target == Address.ALTBN128_PAIRING) {
          long length = 0;
          switch (opCode) {
            case CALL, CALLCODE -> length = Words.clampedToLong(frame.getStackItem(4));
            case DELEGATECALL, STATICCALL -> length = Words.clampedToLong(frame.getStackItem(3));
          }

          final int nMillerLoop = (int) (length / ecPairingNbBytesperMillerLoop);
          if (nMillerLoop * ecPairingNbBytesperMillerLoop != length) {
            log.info("Argument is not a right size: " + length);
            return;
          }

          final long gasPaid = Words.clampedToLong(frame.getStackItem(0));
          if (gasPaid >= precompileBaseGasFee + precompileMillerLoopGasFee * nMillerLoop) {
            final EcpairingLimit lastEcpairingLimit = this.counts.pop();
            this.counts.push(
                new EcpairingLimit(
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
    return this.counts.stream().mapToInt(EcpairingLimit::nPrecompileCall).sum();
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
