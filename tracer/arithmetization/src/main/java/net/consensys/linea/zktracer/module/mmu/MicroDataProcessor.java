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

import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.runtime.callstack.CallFrameType;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.runtime.stack.StackOperation;
import org.apache.commons.lang3.function.TriFunction;

class MicroDataProcessor {
  private final MicroData microData;

  private final PreComputations preComputations;

  MicroDataProcessor() {
    this.microData = new MicroData();
    this.preComputations = new PreComputations();
  }

  void initializeProcessing(final CallStack callStack, final MicroData microData) {
    executeProcessingActionPerType(callStack, microData, MmuPreComputation::process);
  }

  void initializePreProcessing(final CallStack callStack) {
    executeProcessingActionPerType(callStack, microData, MmuPreComputation::preProcess);
  }

  MicroData dispatchOpCode(
      final OpCode opCode, final Map<Integer, StackOperation> stackOps, final CallStack callStack) {
    int preComputation = typeOf(opCode, callStack);

    microData.precomputation(preComputation);

    long returnLength = callStack.caller().returnDataTarget().length();
    microData.instructionContext(
        InstructionContext.builder()
            .self(callStack.current().contextNumber())
            .caller(callStack.caller().contextNumber())
            .returnOffset(callStack.caller().returnTarget().absolute())
            .returnCapacity((int) returnLength)
            .returner(callStack.current().returner())
            .build());

    microData.opCode(opCode);
    microData.setInfo(callStack);

    MmuPreComputation currentPreComputation = preComputations.typeMap().get(preComputation);

    return currentPreComputation.dispatch(microData, opCode, stackOps, callStack);
  }

  private void executeProcessingActionPerType(
      final CallStack callStack,
      final MicroData microData,
      final TriFunction<MmuPreComputation, MicroData, CallStack, MicroData> processorFunc) {
    for (MmuPreComputation p : preComputations.types()) {
      if (p.preComputationTypes().contains(microData.precomputation())) {
        processorFunc.apply(p, microData, callStack);
        return;
      }
    }
  }

  private int typeOf(final OpCode opCode, final CallStack callStack) {
    if (opCode == OpCode.RETURN && callStack.current().type() == CallFrameType.INIT_CODE) {
      return MmuTrace.type3;
    }

    return switch (opCode) {
      case MLOAD, MSTORE, MSTORE8 -> MmuTrace.type1;
      case RETURN, REVERT -> MmuTrace.type2;
      case CREATE, CREATE2, SHA3, LOG0, LOG1, LOG2, LOG3, LOG4 -> MmuTrace.type3;
      case CODECOPY, EXTCODECOPY -> MmuTrace.type4CC;
      case CALLDATACOPY -> MmuTrace.type4CD;
      case RETURNDATACOPY -> MmuTrace.type4RD;
      case CALLDATALOAD -> MmuTrace.type5;
      default -> throw new IllegalArgumentException("Unsupported opcode: %s".formatted(opCode));
    };
  }
}
