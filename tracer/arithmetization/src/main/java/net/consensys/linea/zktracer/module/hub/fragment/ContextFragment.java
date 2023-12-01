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

package net.consensys.linea.zktracer.module.hub.fragment;

import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;

public record ContextFragment(
    CallStack callStack, CallFrame callFrame, boolean updateCallerReturndata)
    implements TraceFragment {
  @Override
  public Trace trace(Trace trace) {
    EWord eAddress = callFrame.addressAsEWord();
    EWord eCodeAddress = callFrame.codeAddressAsEWord();

    CallFrame parent = callStack.getParentOf(callFrame.id());
    EWord parentAddress = parent.addressAsEWord();

    return trace
        .peekAtContext(true)
        .pContextContextNumber(Bytes.ofUnsignedInt(callFrame.contextNumber()))
        .pContextCallStackDepth(Bytes.ofUnsignedInt(callFrame.depth()))
        .pContextIsStatic(callFrame.type().isStatic() ? Bytes.of(1) : Bytes.EMPTY)
        .pContextAccountAddressHi(eAddress.hi())
        .pContextAccountAddressLo(eAddress.lo())
        .pContextByteCodeAddressHi(eCodeAddress.hi())
        .pContextByteCodeAddressLo(eCodeAddress.lo())
        .pContextAccountDeploymentNumber(Bytes.ofUnsignedInt(callFrame.accountDeploymentNumber()))
        .pContextByteCodeDeploymentNumber(Bytes.ofUnsignedInt(callFrame.codeDeploymentNumber()))
        .pContextByteCodeDeploymentStatus(callFrame.underDeployment() ? Bytes.of(1) : Bytes.EMPTY)
        .pContextCallerContextNumber(Bytes.ofUnsignedInt(parent.contextNumber()))
        .pContextCallerAddressHi(parentAddress.hi())
        .pContextCallerAddressLo(parentAddress.lo())
        .pContextCallValue(callFrame.value())
        .pContextCallDataOffset(Bytes.ofUnsignedLong(callFrame.callDataPointer().offset()))
        .pContextCallDataSize(Bytes.ofUnsignedLong(callFrame.callDataPointer().length()))
        .pContextReturnAtOffset(Bytes.ofUnsignedLong(callFrame.returnDataTarget().offset()))
        .pContextReturnAtSize(Bytes.ofUnsignedLong(callFrame.returnDataTarget().length()))
        .pContextUpdate(updateCallerReturndata)
        .pContextReturnerContextNumber(
            Bytes.ofUnsignedInt(
                callFrame.lastCallee().map(c -> callStack.get(c).contextNumber()).orElse(0)))
        .pContextReturnDataOffset(Bytes.ofUnsignedLong(callFrame.returnDataPointer().offset()))
        .pContextReturnDataSize(Bytes.ofUnsignedLong(callFrame.returnDataPointer().length()));
  }
}
