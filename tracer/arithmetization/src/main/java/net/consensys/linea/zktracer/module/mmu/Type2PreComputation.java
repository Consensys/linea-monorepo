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

import java.math.BigInteger;
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
class Type2PreComputation implements MmuPreComputation {
  private static final Set<Integer> TYPES = Set.of(MmuTrace.type2);

  @Override
  public MicroData dispatch(
      final MicroData microData,
      final OpCode opCode,
      final Map<Integer, StackOperation> stackOps,
      final CallStack callStack) {
    microData.contexts(
        Contexts.builder()
            .source(callStack.current().contextNumber())
            .target(callStack.caller().contextNumber())
            .build());

    microData.sizeImported(stackOps.get(2).value().intValue());

    EWord retOffset = microData.instructionContext().returnOffset();
    EWord ax = retOffset.mod(16);

    int retCapacity = microData.instructionContext().returnCapacity();
    int sizeImported = microData.sizeImported();

    microData.min(Math.min(sizeImported, retCapacity));

    if (microData.min() == 0) {
      microData.skip(true);
      return microData;
    }

    microData.pointers(Pointers.builder().stack1(stackOps.get(0).value().copy()).build());

    int off1 = microData.pointers().stack1().intValue();
    int alignment = off1 % 16;

    microData.readPad(
        ReadPad.builder()
            .totalNumberLimbs((off1 + microData.min() - 1) / 16 - (off1 / 16) + 1)
            .totalNumberPaddingMicroInstructions(0)
            .build());

    ax = ax.add(retOffset.add(microData.min())).subtract(1).mod(16);

    UnsignedByte[] nibbles = microData.nibbles();

    int retAlignment = ax.intValue();

    nibbles[0] = UnsignedByte.of(alignment);
    nibbles[1] = UnsignedByte.of((off1 + microData.min() - 1) % 16);
    nibbles[2] = UnsignedByte.of(retAlignment);
    nibbles[3] = UnsignedByte.of(ax.toLong());

    boolean[] bits = microData.bits();

    bits[0] = retCapacity > sizeImported;
    bits[1] = nibbles[2].toInteger() > nibbles[0].toInteger();
    bits[2] = nibbles[1].toInteger() > nibbles[3].toInteger();
    bits[3] = microData.readPad().totalNumberLimbs() == 1;

    if (bits[3]) {
      int x = nibbles[2].toInteger() + microData.min() - 1;
      bits[4] = x >= 16;
      nibbles[6] = UnsignedByte.of(x % 16);
    } else {
      bits[4] = false;
      nibbles[6] = UnsignedByte.ZERO;
    }

    microData.setAccsAtIndex(0, BigInteger.valueOf(off1 / 16));
    microData.setAccsAtIndex(1, BigInteger.valueOf((off1 + microData.min() - 1) / 16));

    ax = retOffset.divide(16);
    microData.setAccsAtIndex(2, ax.toBigInteger());
    ax = retOffset.add(microData.min()).subtract(1).divide(16);
    microData.setAccsAtIndex(3, ax.toBigInteger());

    if (bits[0]) {
      microData.setAccsAtIndex(4, BigInteger.valueOf(retCapacity - sizeImported - 1));
    } else {
      microData.setAccsAtIndex(4, BigInteger.valueOf(sizeImported - retCapacity));
    }

    if (bits[1]) {
      nibbles[4] = UnsignedByte.of(nibbles[2].toInteger() - nibbles[0].toInteger() - 1);
    } else {
      nibbles[4] = UnsignedByte.of(nibbles[0].toInteger() - nibbles[2].toInteger());
    }

    if (bits[2]) {
      nibbles[5] = UnsignedByte.of(nibbles[1].toInteger() - nibbles[3].toInteger() - 1);
    } else {
      nibbles[5] = UnsignedByte.of(nibbles[3].toInteger() - nibbles[1].toInteger());
    }

    microData.aligned(nibbles[2].equals(nibbles[0]));
    microData.offsets(
        Offsets.builder()
            .source(LimbByte.builder().limb(microData.getAccsAtIndex(0)).uByte(nibbles[0]).build())
            .target(LimbByte.builder().limb(microData.getAccsAtIndex(2)).uByte(nibbles[2]).build())
            .build());

    return microData;
  }

  @Override
  public MicroData preProcess(MicroData microData, final CallStack callStack) {
    return microData;
  }

  @Override
  public MicroData process(MicroData microData, final CallStack callStack) {
    updateMicroOpType2(microData);
    updateOffsetType2(microData);

    return microData;
  }

  @Override
  public Set<Integer> preComputationTypes() {
    return TYPES;
  }

  private void updateOffsetType2(final MicroData microData) {
    if (microData.isFirstMicroInstruction()) {
      microData.sourceLimbOffset().add(1);
    }

    boolean[] bits = microData.bits();
    UnsignedByte[] nibbles = microData.nibbles();

    if (bits[3]) {
      microData.size(microData.min());
    } else {
      if (microData.isFirstMicroInstruction()) {
        microData.size(15 - nibbles[0].toInteger() + 1);
      }

      int processingRow = microData.processingRow();

      if (processingRow == 1) {
        microData
            .targetLimbOffset()
            .add(BooleanUtils.toInteger(microData.aligned()) + BooleanUtils.toInteger(bits[1]));
      }

      if (processingRow > 1) {
        microData.targetLimbOffset().add(1);
      }

      if (!microData.isFirstMicroInstruction()) {
        microData.sourceByteOffset(UnsignedByte.ZERO);
        microData.targetByteOffset(
            UnsignedByte.of(
                nibbles[2].toInteger()
                    + 16
                    - nibbles[0].toInteger()
                    - (16 * BooleanUtils.toInteger(microData.aligned())
                        + BooleanUtils.toInteger(bits[1]))));

        if (microData.isLastRead()) {
          microData.size(nibbles[1].toInteger() + 1);
        } else {
          microData.size(16);
        }
      }
    }
  }

  private void updateMicroOpType2(MicroData microData) {
    boolean[] bits = microData.bits();

    if (bits[3]) {
      if (!bits[1]) {
        microData.microOp(MmuTrace.RamToRamSlideChunk);
      } else if (bits[4]) {
        microData.microOp(MmuTrace.RamToRamSlideOverlappingChunk);
      } else {
        microData.microOp(MmuTrace.RamToRamSlideChunk);
      }
    } else if (microData.isFirstMicroInstruction()) {
      if (bits[1]) {
        microData.microOp(MmuTrace.RamToRamSlideOverlappingChunk);
      } else {
        microData.microOp(MmuTrace.RamToRamSlideChunk);
      }
    } else if (microData.isLastRead()) {
      if (bits[2]) {
        microData.microOp(MmuTrace.RamToRamSlideOverlappingChunk);
      } else {
        microData.microOp(MmuTrace.RamToRamSlideChunk);
      }
    } else if (microData.aligned()) {
      microData.microOp(MmuTrace.RamToRam);
    } else {
      microData.microOp(MmuTrace.RamToRamSlideOverlappingChunk);
    }
  }
}
