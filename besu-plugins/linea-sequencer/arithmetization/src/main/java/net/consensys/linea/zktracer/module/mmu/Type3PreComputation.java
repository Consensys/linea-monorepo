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
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.commons.lang3.BooleanUtils;

@AllArgsConstructor(access = AccessLevel.PACKAGE)
class Type3PreComputation implements MmuPreComputation {
  private static final Set<Integer> TYPES = Set.of(MmuTrace.type3);

  @Override
  public MicroData dispatch(
      MicroData microData,
      final OpCode opCode,
      final Map<Integer, StackOperation> stackOps,
      final CallStack callStack) {
    Contexts contexts = microData.contexts();
    contexts.source(callStack.current().contextNumber());
    contexts.target(0);

    microData.value(stackOps.get(3).value().copy());

    int sizeImported = stackOps.get(2).value().toInt();
    microData.sizeImported(sizeImported);

    if (sizeImported == 0) {
      microData.skip(true);
      return microData;
    }

    microData.pointers(Pointers.builder().stack1(stackOps.get(0).value().copy()).build());

    int off1 = microData.pointers().stack1().toInt();

    // Accs
    microData.setAccsAtIndex(0, BigInteger.valueOf(off1 / 16));
    microData.setAccsAtIndex(1, BigInteger.valueOf(sizeImported / 16));

    // Nibbles
    UnsignedByte[] nibbles = microData.nibbles();
    nibbles[0] = UnsignedByte.of(off1 % 16);
    nibbles[1] = UnsignedByte.of(sizeImported % 16);

    microData.aligned(nibbles[0].equals(UnsignedByte.ZERO));

    boolean[] bits = microData.bits();
    bits[0] = !nibbles[1].equals(UnsignedByte.ZERO);

    if (!microData.aligned()) {
      int x = nibbles[0].toInteger() + nibbles[1].toInteger() - 1;
      bits[1] = x >= 16;
      nibbles[2] = UnsignedByte.of(x % 16);
    }

    microData.readPad(
        ReadPad.builder()
            .totalNumberLimbs(microData.getAccsAtIndex(1).toInt() + BooleanUtils.toInteger(bits[0]))
            .totalNumberPaddingMicroInstructions(0)
            .build());

    microData.sourceLimbOffset(microData.getAccsAtIndex(0));
    microData.sourceByteOffset(nibbles[0]);

    return microData;
  }

  @Override
  public MicroData preProcess(final MicroData microData, final CallStack callStack) {
    switch (microData.opCode()) {
      case SHA3 -> microData.exoIsHash(true);
      case LOG0, LOG1, LOG2, LOG3, LOG4 -> microData.exoIsLog(true);
      case CREATE, RETURN -> microData.exoIsRom(true);
      case CREATE2 -> {
        microData.exoIsRom(true);
        microData.exoIsHash(true);
      }
      default ->
          throw new UnsupportedOperationException(
              "OpCode.%s is not supported for MMU type 3 pre-processing."
                  .formatted(microData.opCode()));
    }

    return microData;
  }

  @Override
  public MicroData process(MicroData microData, final CallStack callStack) {
    updateMicroOpType3(microData);
    updateOffsetType3(microData);

    return microData;
  }

  @Override
  public Set<Integer> preComputationTypes() {
    return TYPES;
  }

  private void updateOffsetType3(MicroData microData) {
    UnsignedByte[] nibbles = microData.nibbles();
    if (microData.isLastRead()) {
      microData.size(nibbles[1].toInteger());
    }

    if (!microData.isFirstMicroInstruction()) {
      microData.sourceLimbOffset().add(1);
      microData.targetLimbOffset().add(1);
    }
  }

  private void updateMicroOpType3(MicroData microData) {
    boolean[] bits = microData.bits();
    if (!microData.isLastRead()) {
      if (microData.aligned()) {
        microData.microOp(MmuTrace.RamIsExo);
      } else {
        microData.microOp(MmuTrace.FullExoFromTwo);
      }
    } else if (microData.aligned() && !bits[0]) {
      microData.microOp(MmuTrace.RamIsExo);
    } else if (bits[1]) {
      microData.microOp(MmuTrace.PaddedExoFromTwo);
    } else {
      microData.microOp(MmuTrace.PaddedExoFromOne);
    }
  }
}
