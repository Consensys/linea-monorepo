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

import static net.consensys.linea.zktracer.module.Util.slice;

import java.math.BigInteger;
import java.util.Stack;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.ModuleTrace;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@Slf4j
public class Modexp implements Module {
  private final Stack<Integer> counts = new Stack<Integer>();
  private final int proverMaxInputBitSize = 4096;
  private final int ewordSize = 32;
  private final int gQuadDivisor = 3;

  @Override
  public String jsonKey() {
    return "modexp";
  }

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
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());

    switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final Address target = Words.toAddress(frame.getStackItem(1));
        if (target == Address.MODEXP) {
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
          final Bytes inputData = frame.shadowReadMemory(offset, length);

          final int baseLength = slice(inputData, 0, ewordSize).toInt();
          if (baseLength * 8 > proverMaxInputBitSize) {
            log.info(
                "Too big argument, base bit length =" + baseLength + " > " + proverMaxInputBitSize);
            this.counts.pop();
            this.counts.push(Integer.MAX_VALUE);
            return;
          }
          final int expLength = slice(inputData, ewordSize, ewordSize).toInt();
          if (expLength * 8 > proverMaxInputBitSize) {
            log.info(
                "Too big argument, exp bit length =" + expLength + " > " + proverMaxInputBitSize);
            this.counts.pop();
            this.counts.push(Integer.MAX_VALUE);
            return;
          }
          final int moduloLength = slice(inputData, 2 * ewordSize, ewordSize).toInt();
          if (expLength * 8 > proverMaxInputBitSize) {
            log.info(
                "Too big argument, modulo bit length ="
                    + moduloLength
                    + " > "
                    + proverMaxInputBitSize);
            this.counts.pop();
            this.counts.push(Integer.MAX_VALUE);
            return;
          }
          final Bytes exp = slice(inputData, 3 * ewordSize + baseLength, expLength);

          final long gasPaid = Words.clampedToLong(frame.getStackItem(0));

          if (gasPaid >= gasPrice(baseLength, expLength, moduloLength, exp)) {
            this.counts.push(this.counts.pop() + 1);
          }
        }
      }
      default -> {}
    }
  }

  private long gasPrice(int lB, int lE, int lM, Bytes e) {
    final long maxLbLmSquarred = (long) Math.sqrt((double) (Math.max(lB, lM) + 7) / 8);
    final long secondArg = (maxLbLmSquarred * expLengthPrime(lE, e)) / 3;
    return Math.max(200, secondArg);
  }

  private int expLengthPrime(int lE, Bytes e) {
    int output = 0;
    if (lE <= 32) {
      if (e.toUnsignedBigInteger().equals(BigInteger.ZERO)) {
        return 0;
      } else {
        return (e.toUnsignedBigInteger().bitLength() - 1);
      }
    } else {
      if (e.slice(0, ewordSize).toUnsignedBigInteger().compareTo(BigInteger.ZERO) != 0) {
        return 8 * (lE - 32) + e.slice(0, ewordSize).toUnsignedBigInteger().bitLength() - 1;
      } else {
        return 8 * (lE - 32);
      }
    }
  }

  @Override
  public int lineCount() {
    return this.counts.stream().mapToInt(x -> x).sum();
  }

  @Override
  public ModuleTrace commit() {
    throw new IllegalStateException("should never be called");
  }
}
