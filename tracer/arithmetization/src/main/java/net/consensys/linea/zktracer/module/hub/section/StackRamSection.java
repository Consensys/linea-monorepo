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

package net.consensys.linea.zktracer.module.hub.section;

import static com.google.common.base.Preconditions.*;
import static net.consensys.linea.zktracer.Trace.MMU_INST_MLOAD;
import static net.consensys.linea.zktracer.Trace.MMU_INST_MSTORE;
import static net.consensys.linea.zktracer.Trace.MMU_INST_MSTORE8;
import static net.consensys.linea.zktracer.Trace.WORD_SIZE;
import static net.consensys.linea.zktracer.runtime.callstack.CallFrame.extractContiguousLimbsFromMemory;

import java.util.Optional;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.MxpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.Range;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.internal.Words;

public class StackRamSection extends TraceSection {

  public StackRamSection(Hub hub) {
    super(hub, (short) 3);

    this.addStack(hub);

    final OpCode instruction = hub.opCode();
    final short exceptions = hub.pch().exceptions();

    final ImcFragment imcFragment = ImcFragment.empty(hub);
    this.addFragment(imcFragment);

    final MxpCall mxpCall = new MxpCall(hub);
    imcFragment.callMxp(mxpCall);

    checkArgument(mxpCall.isMxpx() == Exceptions.memoryExpansionException(exceptions));

    // MXPX or OOGX case
    if (Exceptions.memoryExpansionException(exceptions)
        || Exceptions.outOfGasException(exceptions)) {
      return;
    }

    // the unexceptional case
    checkArgument(Exceptions.none(exceptions));

    final CallFrame currentFrame = hub.currentFrame();
    final EWord offset = EWord.of(currentFrame.frame().getStackItem(0));
    final long longOffset = Words.clampedToLong(offset);
    final Bytes currentRam =
        extractContiguousLimbsFromMemory(currentFrame.frame(), new Range(longOffset, WORD_SIZE));
    final int currentContextNumber = currentFrame.contextNumber();
    final EWord value =
        instruction.equals(OpCode.MLOAD)
            ? EWord.of(currentFrame.frame().shadowReadMemory(longOffset, WORD_SIZE))
            : EWord.of(currentFrame.frame().getStackItem(1));

    MmuCall mmuCall;

    switch (instruction) {
      case MSTORE -> mmuCall =
          new MmuCall(hub, MMU_INST_MSTORE)
              .targetId(currentContextNumber)
              .targetOffset(offset)
              .limb1(value.hi())
              .limb2(value.lo())
              .targetRamBytes(Optional.of(currentRam));

      case MSTORE8 -> mmuCall =
          new MmuCall(hub, MMU_INST_MSTORE8)
              .targetId(currentContextNumber)
              .targetOffset(offset)
              .limb1(value.hi())
              .limb2(value.lo())
              .targetRamBytes(Optional.of(currentRam));

      case MLOAD -> mmuCall =
          new MmuCall(hub, MMU_INST_MLOAD)
              .sourceId(currentContextNumber)
              .sourceOffset(offset)
              .limb1(value.hi())
              .limb2(value.lo())
              .sourceRamBytes(Optional.of(currentRam));

      default -> throw new IllegalStateException("Not a STACK_RAM instruction");
    }

    imcFragment.callMmu(mmuCall);
  }
}
