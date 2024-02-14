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

package net.consensys.linea.zktracer.module.mmu;

import java.util.Map;
import java.util.Set;

import lombok.AccessLevel;
import lombok.AllArgsConstructor;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.runtime.stack.StackOperation;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.commons.lang3.BooleanUtils;

@AllArgsConstructor(access = AccessLevel.PACKAGE)
class Type5PreComputation implements MmuPreComputation {
  private static final Set<Integer> TYPES = Set.of(MmuTrace.type5);

  @Override
  public MicroData dispatch(
      final MicroData microData,
      final OpCode opCode,
      final Map<Integer, StackOperation> stackOps,
      final CallStack callStack) {
    microData.callStackDepth(callStack.depth());
    microData.value(stackOps.get(3).value().copy());

    if (microData.callStackDepth() == 1) {
      microData.sourceContext(0);
    } else {
      microData.sourceContext(callStack.caller().contextNumber());
    }

    microData.pointers(Pointers.builder().stack1(stackOps.get(0).value().copy()).build());

    final EWord offset = EWord.of(microData.pointers().stack1().copy());
    microData.callDataSize((int) callStack.caller().returnDataSource().length());

    int callDataSize = microData.callDataSize();
    final EWord callDataSizeEWord = EWord.of(callDataSize);

    if (callDataSize == 0 || callDataSizeEWord.lessOrEqualThan(offset)) {
      microData.skip(true);

      return microData;
    }

    boolean[] bits = microData.bits();
    bits[0] = EWord.THIRTY_TWO.add(offset).lessOrEqualThan(callDataSizeEWord);

    int bits0Int = BooleanUtils.toInteger(bits[0]);
    int exp =
        bits[0]
            ? callDataSizeEWord.subtract(offset).subtract(32).getAsBigInteger().intValue()
                + bits0Int
                - 1
            : EWord.THIRTY_TWO.add(offset).subtract(callDataSize).getAsBigInteger().intValue()
                + bits0Int
                - 1;

    microData.setAccsAtIndex(0, EWord.of(exp));

    EWord acc2 = bits[0] ? EWord.THIRTY_ONE : EWord.of(callDataSize - 1).subtract(offset);

    microData.setAccsAtIndex(1, acc2);

    bits[1] = !microData.getAccsAtIndex(1).divide(16).isZero();

    UnsignedByte[] nibbles = microData.nibbles();
    nibbles[1] = UnsignedByte.of(microData.getAccsAtIndex(1).mod(16).toLong());

    bits[2] = nibbles[1].toInteger() == 15;

    microData.readPad(
        callDataSize == 1
            ? ReadPad.builder().totalNumberLimbs(4).totalNumberPaddingMicroInstructions(0).build()
            : ReadPad.builder().totalNumberLimbs(1).totalNumberPaddingMicroInstructions(0).build());

    microData.callDataOffset((int) callStack.caller().callDataSource().absolute());

    int fullOffset = offset.add(microData.callDataOffset()).toInt();
    microData.setAccsAtIndex(2, EWord.of(fullOffset / 16));

    nibbles[2] = UnsignedByte.of(fullOffset % 16);
    microData.aligned(nibbles[2].toInteger() == 0);

    if (!microData.aligned()) {
      bits[3] = 15 - nibbles[2].toInteger() < nibbles[1].toInteger();
    }

    int bits3Int = BooleanUtils.toInteger(bits[3]);
    nibbles[3] =
        UnsignedByte.of(
            (2 * bits3Int - 1) * nibbles[1].toInteger() - 15 + nibbles[2].toInteger() - bits3Int);

    microData.sourceLimbOffset(microData.getAccsAtIndex(2).copy());
    microData.sourceByteOffset(nibbles[2]);

    return microData;
  }

  @Override
  public MicroData preProcess(MicroData microData, CallStack callStack) {
    return microData;
  }

  @Override
  public MicroData process(MicroData microData, CallStack callStack) {
    updateMicroOpType5(microData, callStack);
    updateOffsetType5(microData, callStack);

    return microData;
  }

  @Override
  public Set<Integer> preComputationTypes() {
    return TYPES;
  }

  private void updateOffsetType5(MicroData microData, final CallStack callStack) {
    EWord acc3 = microData.getAccsAtIndex(2);
    UnsignedByte[] nibbles = microData.nibbles();

    if (microData.isLastRead()) {
      microData.size(1 + nibbles[1].toInteger());
    }

    if (callStack.depth() == 1) {
      switch (microData.remainingReads()) {
        case 3 -> microData.sourceLimbOffset(acc3);
        case 2 -> microData.sourceLimbOffset(acc3.add(1));
        case 1 -> microData.sourceLimbOffset(acc3.add(2));
        case 0 -> microData.sourceLimbOffset(EWord.ZERO);
        default -> throw new IllegalStateException(
            "Unexpected value: %s".formatted(microData.remainingReads()));
      }
    } else {
      if (!microData.isLastRead()) {
        throw new IllegalStateException("Should never happen");
      }
      microData.sourceLimbOffset(acc3);
    }
    microData.sourceByteOffset(nibbles[2]);
  }

  private void updateMicroOpType5(MicroData microData, final CallStack callStack) {
    microData.exoIsTxcd(!microData.isLastRead());

    switch (microData.remainingReads()) {
      case 3 -> {
        checkCallStackDepth(callStack);
        microData.microOp(MmuTrace.StoreXInAThreeRequired);
      }
      case 2 -> {
        checkCallStackDepth(callStack);
        microData.microOp(MmuTrace.StoreXInB);
      }
      case 1 -> {
        checkCallStackDepth(callStack);
        microData.microOp(MmuTrace.StoreXInC);
      }
      case 0 -> {
        boolean[] bits = microData.bits();
        if (callStack.depth() != 1) {
          if (microData.aligned()) {
            if (!bits[1] && bits[2]) {
              microData.microOp(MmuTrace.FirstPaddedSecondZero);
            } else if (!bits[1]) {
              microData.microOp(MmuTrace.PushOneRamToStack);
            } else if (!bits[2]) {
              microData.microOp(MmuTrace.FirstFastSecondPadded);
            } else {
              microData.microOp(MmuTrace.PushTwoRamToStack);
            }
          } else {
            if (!bits[1] && !bits[2]) {
              microData.microOp(
                  bits[3]
                      ? MmuTrace.NA_RamToStack_2To1PaddedAndZero
                      : MmuTrace.NA_RamToStack_1To1PaddedAndZero);
            } else if (!bits[1]) {
              microData.microOp(MmuTrace.NA_RamToStack_2To1FullAndZero);
            } else if (!bits[2]) {
              microData.microOp(
                  bits[3] ? MmuTrace.NA_RamToStack_3To2Padded : MmuTrace.NA_RamToStack_2To2Padded);
            } else {
              microData.microOp(MmuTrace.NA_RamToStack_3To2Full);
            }
          }
        } else {
          if (microData.aligned() && bits[2]) {
            microData.microOp(MmuTrace.ExceptionalRamToStack3To2FullFast);
          } else {
            microData.microOp(MmuTrace.Exceptional_RamToStack_3To2Full);
          }
        }
      }
      default -> throw new IllegalStateException(
          "Unexpected value: %s".formatted(microData.remainingReads()));
    }
  }

  private static void checkCallStackDepth(CallStack callStack) {
    if (callStack.depth() != 1) {
      throw new IllegalStateException("Should never happen");
    }
  }
}
