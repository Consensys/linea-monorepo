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

@AllArgsConstructor(access = AccessLevel.PACKAGE)
class Type1PreComputation implements MmuPreComputation {

  private static final Set<Integer> TYPES = Set.of(MmuTrace.type1);

  @Override
  public MicroData dispatch(
      MicroData microData,
      final OpCode opCode,
      final Map<Integer, StackOperation> stackOps,
      final CallStack callStack) {
    BigInteger off1 = stackOps.get(0).value().toUnsignedBigInteger();

    microData.aligned(off1.mod(BigInteger.valueOf(16)).equals(BigInteger.ZERO));

    microData.readPad(
        ReadPad.builder().totalNumberLimbs(1).totalNumberPaddingMicroInstructions(0).build());

    microData.value(stackOps.get(3).value().copy());

    microData.pointers(Pointers.builder().stack1(stackOps.get(0).value().copy()).build());

    UnsignedByte nibble0 = UnsignedByte.of(off1.mod(BigInteger.valueOf(16)).intValue());
    microData.nibbles()[0] = nibble0;

    BigInteger accs = off1.divide(BigInteger.valueOf(16));
    microData.setAccsAtIndex(0, accs);

    LimbByte limbByte =
        LimbByte.builder().uByte(microData.nibbles()[0]).limb(EWord.of(accs)).build();

    microData.offsets(Offsets.builder().target(limbByte).build());
    microData.contexts(Contexts.builder().source(callStack.current().contextNumber()).build());

    switch (opCode) {
      case MLOAD -> {
        microData.toRam(false);

        if (microData.aligned()) {
          microData.microOp(MmuTrace.PushTwoRamToStack);
        } else {
          microData.microOp(MmuTrace.NA_RamToStack_3To2Full);
        }
      }
      case MSTORE8 -> {
        microData.toRam(true);
        microData.microOp(MmuTrace.LsbFromStackToRAM);
      }
      case MSTORE -> {
        microData.toRam(true);

        if (microData.aligned()) {
          microData.microOp(MmuTrace.PushTwoStackToRam);
        } else {
          microData.microOp(MmuTrace.FullStackToRam);
        }
      }
      default -> throw new IllegalArgumentException(
          "Opcode %s is not supported for type 1 pre-computation".formatted(opCode));
    }

    return microData;
  }

  @Override
  public MicroData preProcess(MicroData microData, final CallStack callStack) {
    return microData;
  }

  @Override
  public MicroData process(MicroData microData, final CallStack callStack) {
    return microData;
  }

  @Override
  public Set<Integer> preComputationTypes() {
    return TYPES;
  }
}
