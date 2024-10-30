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

import static com.google.common.base.Preconditions.checkArgument;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_RIGHT_PADDED_WORD_EXTRACTION;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.WORD_SIZE;
import static net.consensys.linea.zktracer.module.hub.fragment.ContextFragment.readCurrentContextData;
import static net.consensys.linea.zktracer.runtime.callstack.CallFrame.extractContiguousLimbsFromMemory;
import static org.hyperledger.besu.evm.internal.Words.clampedToLong;

import java.util.Arrays;
import java.util.Optional;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.opcodes.CallDataLoadOobCall;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.runtime.callstack.CallFrameType;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.MemorySpan;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.internal.Words;

public class CallDataLoadSection extends TraceSection {

  public CallDataLoadSection(Hub hub) {
    super(hub, (short) (hub.opCode().equals(OpCode.CALLDATALOAD) ? 4 : 3));
    this.addStack(hub);

    final short exception = hub.pch().exceptions();

    final ImcFragment imcFragment = ImcFragment.empty(hub);
    this.addFragment(imcFragment);

    final CallDataLoadOobCall oobCall = new CallDataLoadOobCall();
    imcFragment.callOob(oobCall);

    if (Exceptions.none(exception)) {
      if (!oobCall.isCdlOutOfBounds()) {
        final long callDataSize = hub.currentFrame().callDataInfo().memorySpan().length();
        final long callDataOffset = hub.currentFrame().callDataInfo().memorySpan().offset();
        final EWord sourceOffset = EWord.of(hub.currentFrame().frame().getStackItem(0));
        final long callDataCN = hub.currentFrame().callDataInfo().callDataContextNumber();

        final EWord read =
            EWord.of(
                Bytes.wrap(
                    Arrays.copyOfRange(
                        hub.currentFrame().callDataInfo().data().toArray(),
                        Words.clampedToInt(sourceOffset),
                        Words.clampedToInt(sourceOffset) + WORD_SIZE)));

        final CallFrame callDataCallFrame = hub.callStack().getByContextNumber(callDataCN);

        final MmuCall call =
            new MmuCall(hub, MMU_INST_RIGHT_PADDED_WORD_EXTRACTION)
                .sourceId((int) callDataCN)
                .sourceRamBytes(
                    Optional.of(
                        callDataCallFrame.type() == CallFrameType.TRANSACTION_CALL_DATA_HOLDER
                            ? callDataCallFrame.callDataInfo().data()
                            : extractContiguousLimbsFromMemory(
                                callDataCallFrame.frame(),
                                MemorySpan.fromStartLength(
                                    clampedToLong(sourceOffset) + callDataOffset, WORD_SIZE))))
                .sourceOffset(sourceOffset)
                .referenceOffset(callDataOffset)
                .referenceSize(callDataSize)
                .limb1(read.hi())
                .limb2(read.lo());

        imcFragment.callMmu(call);
      }
    } else {
      // Sanity check
      checkArgument(Exceptions.outOfGasException(exception));
    }

    final ContextFragment context = readCurrentContextData(hub);
    this.addFragment(context);
  }
}
