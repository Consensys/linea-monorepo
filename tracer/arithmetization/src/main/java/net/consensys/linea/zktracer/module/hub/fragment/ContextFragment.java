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

import static net.consensys.linea.zktracer.types.AddressUtils.highPart;
import static net.consensys.linea.zktracer.types.AddressUtils.lowPart;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.types.Either;
import net.consensys.linea.zktracer.types.MemorySpan;
import org.hyperledger.besu.datatypes.Address;

@Getter
@Setter
@Accessors(fluent = true)
@AllArgsConstructor
public class ContextFragment implements TraceFragment {
  private final Hub hub;
  private final CallStack callStack;
  // Left: callFrameId, Right: contextNumber
  private Either<Integer, Integer> callFrameReference;
  private int returnDataContextNumber;
  private MemorySpan returnDataSegment;
  private boolean updateReturnData;

  public static ContextFragment readContextDataByContextNumber(
      final Hub hub, final int contextNumber) {
    CallStack callStack = hub.callStack();
    return new ContextFragment(
        hub,
        callStack,
        Either.right(contextNumber),
        callStack.getByContextNumber(contextNumber).returnDataContextNumber(),
        callStack.current().returnDataSpan().snapshot(),
        false);
  }

  public static ContextFragment readContextDataById(final Hub hub, final int contextId) {
    CallStack callStack = hub.callStack();
    return new ContextFragment(
        hub,
        callStack,
        Either.left(contextId),
        callStack.getById(contextId).returnDataContextNumber(),
        callStack.current().returnDataSpan().snapshot(),
        false);
  }

  public static ContextFragment readCurrentContextData(final Hub hub) {
    return readContextDataById(hub, hub.callStack().current().id());
  }

  public static ContextFragment initializeNewExecutionContext(final Hub hub) {
    return new ContextFragment(
        hub,
        hub.callStack(),
        Either.right(hub.newChildContextNumber()),
        0,
        MemorySpan.empty(),
        false);
  }

  public static ContextFragment executionProvidesEmptyReturnData(final Hub hub) {
    CallStack callStack = hub.callStack();
    return new ContextFragment(
        hub,
        callStack,
        Either.left(callStack.parent().id()),
        hub.callStack().current().contextNumber(),
        MemorySpan.empty(),
        true);
  }

  public static ContextFragment executionProvidesEmptyReturnData(final Hub hub, int contextNumber) {
    CallStack callStack = hub.callStack();
    int parentId = callStack.getByContextNumber(contextNumber).callerId();
    return new ContextFragment(
        hub, callStack, Either.left(parentId), contextNumber, MemorySpan.empty(), true);
  }

  public static ContextFragment nonExecutionProvidesEmptyReturnData(final Hub hub) {
    CallStack callStack = hub.callStack();
    return new ContextFragment(
        hub,
        callStack,
        Either.left(callStack.current().id()),
        hub.newChildContextNumber(),
        MemorySpan.empty(),
        true);
  }

  public static ContextFragment executionProvidesReturnData(final Hub hub) {
    CallStack callStack = hub.callStack();
    return new ContextFragment(
        hub,
        callStack,
        Either.left(callStack.parent().id()),
        hub.currentFrame().contextNumber(),
        callStack.current().outputDataSpan(),
        true);
  }

  public static ContextFragment executionProvidesReturnData(
      final Hub hub, int receiverContextNumber, int providerContextNumber) {
    CallStack callStack = hub.callStack();
    return new ContextFragment(
        hub,
        callStack,
        Either.right(receiverContextNumber),
        providerContextNumber,
        callStack.current().returnDataSpan().snapshot(),
        true);
    // TODO: is this what we want ?
    //  also: will the latestReturnData have been updated ?
  }

  public static ContextFragment updateReturnData(
      final Hub hub, final int returnDataContextNumber, final MemorySpan returnDataMetaInfo) {
    return new ContextFragment(
        hub,
        hub.callStack(),
        Either.right(hub.callStack().current().contextNumber()),
        returnDataContextNumber,
        returnDataMetaInfo,
        true);
  }

  @Override
  public Trace trace(Trace trace) {
    final CallFrame callFrame =
        this.callFrameReference.map(this.callStack::getById, this.callStack::getByContextNumber);
    final CallFrame parent = callStack.getParentCallFrameById(callFrame.id());

    final Address address = callFrame.accountAddress();
    final Address codeAddress = callFrame.byteCodeAddress();
    final Address callerAddress = callFrame.callerAddress();

    return trace
        .peekAtContext(true)
        .pContextContextNumber(callFrame.contextNumber())
        .pContextCallStackDepth((short) callFrame.depth())
        .pContextIsRoot(callFrame.isRoot())
        .pContextIsStatic(callFrame.type().isStatic())
        .pContextAccountAddressHi(highPart(address))
        .pContextAccountAddressLo(lowPart(address))
        .pContextAccountDeploymentNumber(callFrame.accountDeploymentNumber())
        .pContextByteCodeAddressHi(highPart(codeAddress))
        .pContextByteCodeAddressLo(lowPart(codeAddress))
        .pContextByteCodeDeploymentNumber(callFrame.byteCodeDeploymentNumber())
        .pContextByteCodeDeploymentStatus(callFrame.isDeployment() ? 1 : 0)
        .pContextByteCodeCodeFragmentIndex(callFrame.getCodeFragmentIndex(hub))
        .pContextCallerAddressHi(highPart(callerAddress))
        .pContextCallerAddressLo(lowPart(callerAddress))
        .pContextCallValue(callFrame.value())
        .pContextCallDataContextNumber(parent.contextNumber())
        .pContextCallDataOffset(callFrame.callDataInfo().memorySpan().offset())
        .pContextCallDataSize(callFrame.callDataInfo().memorySpan().length())
        .pContextReturnAtOffset(callFrame.returnDataTargetInCaller().offset())
        .pContextReturnAtCapacity(callFrame.returnDataTargetInCaller().length())
        .pContextUpdate(updateReturnData)
        .pContextReturnDataContextNumber(returnDataContextNumber)
        //             callFrame.id() == 0
        //                 ? callFrame.universalParentReturnDataContextNumber
        //                 : callFrame.lastCallee().map(c ->
        // callStack.getById(c).contextNumber()).orElse(0))
        .pContextReturnDataOffset(returnDataSegment.offset())
        .pContextReturnDataSize(returnDataSegment.length());
  }
}
