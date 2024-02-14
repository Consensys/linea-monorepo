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

import static net.consensys.linea.zktracer.module.Util.slice;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.List;
import java.util.Stack;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.modexpdata.ModexpData;
import net.consensys.linea.zktracer.module.modexpdata.ModexpDataOperation;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@Slf4j
@RequiredArgsConstructor
@Accessors(fluent = true)
public class ModexpEffectiveCall implements Module {
  private final Hub hub;

  @Getter private final ModexpData data = new ModexpData();
  private final Stack<Integer> counts = new Stack<>();
  private static final BigInteger PROVER_MAX_INPUT_BIT_SIZE = BigInteger.valueOf(4096 / 8);
  private static final int EVM_WORD_SIZE = 32;

  private int lastModexpDataCallHubStamp = 0;

  @Override
  public String moduleKey() {
    return "PRECOMPILE_MODEXP_EFFECTIVE_CALL";
  }

  @Override
  public void enterTransaction() {
    counts.push(0);
    this.data.enterTransaction();
  }

  @Override
  public void popTransaction() {
    counts.pop();
    this.data.popTransaction();
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    final OpCode opCode = hub.opCode();

    if (opCode.isAnyOf(OpCode.CALL, OpCode.STATICCALL, OpCode.DELEGATECALL, OpCode.CALLCODE)) {
      final Address target = Words.toAddress(frame.getStackItem(1));
      if (target.equals(Address.MODEXP)) {
        final Bytes inputData = hub.transients().op().callData();

        // Get the Base length
        final BigInteger baseLength = slice(inputData, 0, EVM_WORD_SIZE).toUnsignedBigInteger();
        if (isOutOfProverInputBounds(baseLength)) {
          log.info(
              "Too big argument, base bit length = {} > {}", baseLength, PROVER_MAX_INPUT_BIT_SIZE);
          this.counts.pop();
          this.counts.push(Integer.MAX_VALUE);
          return;
        }

        // Get the Exponent length
        final BigInteger expLength =
            slice(inputData, EVM_WORD_SIZE, EVM_WORD_SIZE).toUnsignedBigInteger();
        if (isOutOfProverInputBounds(expLength)) {
          log.info(
              "Too big argument, expComponent bit length = {} > {}",
              expLength,
              PROVER_MAX_INPUT_BIT_SIZE);
          this.counts.pop();
          this.counts.push(Integer.MAX_VALUE);
          return;
        }

        // Get the Modulo length
        final BigInteger modLength =
            slice(inputData, 2 * EVM_WORD_SIZE, EVM_WORD_SIZE).toUnsignedBigInteger();
        if (isOutOfProverInputBounds(modLength)) {
          log.info(
              "Too big argument, modulo bit length = {} > {}",
              modLength,
              PROVER_MAX_INPUT_BIT_SIZE);
          this.counts.pop();
          this.counts.push(Integer.MAX_VALUE);
          return;
        }

        final int baseLengthInt = baseLength.intValueExact();
        final int expLengthInt = expLength.intValueExact();
        final int modLengthInt = modLength.intValueExact();

        // Get the Base.
        final Bytes baseComponent = slice(inputData, 3 * EVM_WORD_SIZE, baseLengthInt);

        // Get the Exponent.
        final Bytes expComponent =
            slice(inputData, 3 * EVM_WORD_SIZE + baseLengthInt, expLength.intValueExact());

        // Get the Modulus.
        final Bytes modComponent =
            slice(
                inputData,
                3 * EVM_WORD_SIZE + baseLengthInt + expLengthInt,
                modLength.intValueExact());
        final long gasPrice = gasPrice(baseLengthInt, expLengthInt, modLengthInt, expComponent);

        // If enough gas, add 1 to the call of the precompile.
        if (hub.transients().op().gasAllowanceForCall() >= gasPrice) {
          this.lastModexpDataCallHubStamp =
              this.data.call(
                  new ModexpDataOperation(
                      hub.stamp(),
                      lastModexpDataCallHubStamp,
                      baseComponent,
                      expComponent,
                      modComponent));
          this.counts.push(this.counts.pop() + 1);
        }
      }
    }
  }

  public static long gasCost(final Hub hub) {
    final OpCode opCode = hub.opCode();
    final MessageFrame frame = hub.messageFrame();

    if (opCode.isAnyOf(OpCode.CALL, OpCode.STATICCALL, OpCode.DELEGATECALL, OpCode.CALLCODE)) {
      final Address target = Words.toAddress(frame.getStackItem(1));
      if (target.equals(Address.MODEXP)) {
        final Bytes inputData = hub.transients().op().callData();

        // Get the Base length
        final BigInteger baseLength = slice(inputData, 0, EVM_WORD_SIZE).toUnsignedBigInteger();
        if (isOutOfProverInputBounds(baseLength)) {
          return 0;
        }

        // Get the Exponent length
        final BigInteger expLength =
            slice(inputData, EVM_WORD_SIZE, EVM_WORD_SIZE).toUnsignedBigInteger();
        if (isOutOfProverInputBounds(expLength)) {
          return 0;
        }

        // Get the Modulo length
        final BigInteger modLength =
            slice(inputData, 2 * EVM_WORD_SIZE, EVM_WORD_SIZE).toUnsignedBigInteger();
        if (isOutOfProverInputBounds(modLength)) {
          return 0;
        }

        final int baseLengthInt = baseLength.intValueExact();
        final int expLengthInt = expLength.intValueExact();
        final int modLengthInt = modLength.intValueExact();

        // Get the Exponent.
        final Bytes expComponent =
            slice(inputData, 3 * EVM_WORD_SIZE + baseLengthInt, expLength.intValueExact());

        return gasPrice(baseLengthInt, expLengthInt, modLengthInt, expComponent);
      }
    }

    return 0;
  }

  private static long gasPrice(int baseLength, int expLength, int moduloLength, Bytes e) {
    final long maxLbLmSquared =
        (long) Math.pow((double) (Math.max(baseLength, moduloLength) + 7) / 8, 2);
    final long secondArg = (maxLbLmSquared * expLengthPrime(expLength, e)) / 3;

    return Math.max(200, secondArg);
  }

  private static int expLengthPrime(int expLength, Bytes e) {
    if (expLength <= 32) {
      return e.isZero() ? 0 : e.toUnsignedBigInteger().bitLength() - 1;
    }

    final Bytes leadingWord = e.slice(0, EVM_WORD_SIZE);
    return 8 * (expLength - 32) + Math.max(leadingWord.bitLength() - 1, 0);
  }

  private static boolean isOutOfProverInputBounds(BigInteger modexpComponentLength) {
    return modexpComponentLength.compareTo(PROVER_MAX_INPUT_BIT_SIZE) > 0;
  }

  @Override
  public int lineCount() {
    return this.counts.stream().mapToInt(x -> x).sum();
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
