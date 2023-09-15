/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.module.hub.chunks;

import java.math.BigInteger;

import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.module.runtime.callstack.CallStack;

public record ContextFragment(CallStack callStack, CallFrame callFrame, boolean update)
    implements TraceFragment {
  @Override
  public Trace.TraceBuilder trace(Trace.TraceBuilder trace) {
    EWord eAddress = callFrame.addressAsEWord();
    EWord eCodeAddress = callFrame.codeAddressAsEWord();

    CallFrame parent = callStack.getParentOf(callFrame.getId());
    EWord parentAddress = parent.addressAsEWord();

    return trace
        .peekAtContext(true)
        .pContextContextNumber(BigInteger.valueOf(callFrame.getContextNumber()))
        .pContextCallStackDepth(BigInteger.ZERO) // TODO: should be CallFrame::getDepth
        .pContextIsStatic(callFrame.getType().isStatic() ? BigInteger.ONE : BigInteger.ZERO)
        .pContextAccountAddressHi(eAddress.hiBigInt())
        .pContextAccountAddressLo(eAddress.loBigInt())
        .pContextByteCodeAddressHi(eCodeAddress.hiBigInt())
        .pContextByteCodeAddressLo(eCodeAddress.loBigInt())
        .pContextAccountDeploymentNumber(BigInteger.valueOf(callFrame.getAccountDeploymentNumber()))
        .pContextByteCodeDeploymentNumber(BigInteger.valueOf(callFrame.getCodeDeploymentNumber()))
        .pContextByteCodeDeploymentStatus(
            callFrame.isCodeDeploymentStatus() ? BigInteger.ONE : BigInteger.ZERO)
        .pContextCallerContextNumber(BigInteger.valueOf(parent.getContextNumber()))
        .pContextCallerAddressHi(parentAddress.hiBigInt())
        .pContextCallerAddressLo(parentAddress.loBigInt())
        .pContextCallValue(callFrame.getValue().toUnsignedBigInteger())
        .pContextCallDataOffset(BigInteger.valueOf(callFrame.getCallDataPointer().offset()))
        .pContextCallDataSize(BigInteger.valueOf(callFrame.getCallDataPointer().length()))
        .pContextReturnAtOffset(BigInteger.valueOf(callFrame.getReturnDataTarget().offset()))
        .pContextReturnAtSize(BigInteger.valueOf(callFrame.getReturnDataTarget().length()))
        .pContextUpdate(update)
        .pContextReturnerContextNumber(
            BigInteger.valueOf(
                callFrame.lastCallee().map(c -> callStack.get(c).getContextNumber()).orElse(0)))
        .pContextReturnDataOffset(BigInteger.valueOf(callFrame.getReturnDataPointer().offset()))
        .pContextReturnDataSize(BigInteger.valueOf(callFrame.getReturnDataPointer().length()));
  }
}
