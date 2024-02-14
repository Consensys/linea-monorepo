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

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.Either;
import net.consensys.linea.zktracer.types.MemorySpan;
import org.apache.tuweni.bytes.Bytes;

public record ContextFragment(
    CallStack callStack,
    // Left: callFrameId, Right: contextNumber
    Either<Integer, Integer> callFrameReference,
    MemorySpan returnDataSegment,
    boolean updateCallerReturndata)
    implements TraceFragment {

  public static ContextFragment readContextData(final CallStack callStack) {

    return new ContextFragment(
        callStack,
        Either.left(callStack.current().id()),
        callStack.current().latestReturnDataSource().snapshot(),
        false);
  }

  public static ContextFragment intializeExecutionContext(final Hub hub) {
    return new ContextFragment(
        hub.callStack(),
        Either.right(hub.stamp() + 1),
        MemorySpan.fromStartEnd(
            0, hub.transients().tx().transaction().getData().map(Bytes::size).orElse(0)),
        false);
  }

  public static ContextFragment executionEmptyReturnData(final CallStack callStack) {
    return new ContextFragment(
        callStack, Either.left(callStack.parent().id()), MemorySpan.empty(), true);
  }

  public static ContextFragment nonExecutionEmptyReturnData(final CallStack callStack) {
    return new ContextFragment(
        callStack, Either.left(callStack.parent().id()), MemorySpan.empty(), true);
  }

  public static ContextFragment executionReturnData(final CallStack callStack) {
    return new ContextFragment(
        callStack,
        Either.left(callStack.parent().id()),
        callStack.current().returnDataSource(),
        true);
  }

  public static ContextFragment enterContext(
      final CallStack callStack, final CallFrame calledCallFrame) {
    return new ContextFragment(
        callStack, Either.left(calledCallFrame.id()), MemorySpan.empty(), false);
  }

  public static ContextFragment providesReturnData(final CallStack callStack) {
    return new ContextFragment(
        callStack,
        Either.left(callStack.current().id()),
        callStack.current().latestReturnDataSource().snapshot(),
        true);
  }

  @Override
  public Trace trace(Trace trace) {
    final CallFrame callFrame =
        this.callFrameReference.map(this.callStack::getById, this.callStack::getByContextNumber);
    final CallFrame parent = callStack.getParentOf(callFrame.id());

    final EWord eAddress = callFrame.addressAsEWord();
    final EWord eCodeAddress = callFrame.codeAddressAsEWord();
    final EWord parentAddress = parent.addressAsEWord();

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
        .pContextCallDataOffset(Bytes.ofUnsignedLong(callFrame.callDataSource().offset()))
        .pContextCallDataSize(Bytes.ofUnsignedLong(callFrame.callDataSource().length()))
        .pContextReturnAtOffset(
            Bytes.ofUnsignedLong(callFrame.requestedReturnDataTarget().offset()))
        .pContextReturnAtSize(Bytes.ofUnsignedLong(callFrame.requestedReturnDataTarget().length()))
        .pContextUpdate(updateCallerReturndata)
        .pContextReturnerContextNumber(
            Bytes.ofUnsignedInt(
                callFrame.lastCallee().map(c -> callStack.getById(c).contextNumber()).orElse(0)))
        .pContextReturnDataOffset(Bytes.ofUnsignedLong(returnDataSegment.offset()))
        .pContextReturnDataSize(Bytes.ofUnsignedLong(returnDataSegment.length()));
  }
}
